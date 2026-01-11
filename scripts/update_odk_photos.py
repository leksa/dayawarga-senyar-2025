#!/usr/bin/env python3
"""
Update ODK Central submissions with photos from ArcGIS dump.

This script:
1. Reads entity_mapping.json to get mapping between objectid (photo folder) and entity_id (ODK UUID)
2. Downloads current XML from ODK Central
3. Updates XML with foto_depan field and new instanceID
4. PUTs updated XML to create new version
5. POSTs photo file as attachment
"""

import os
import sys
import json
import uuid
import time
import re
import requests
from pathlib import Path
from xml.etree import ElementTree as ET

# Configuration - load from .env
from dotenv import load_dotenv
load_dotenv(Path(__file__).parent.parent / ".env")

ODK_BASE_URL = os.getenv("ODK_BASE_URL", "https://data.dayawarga.com") + "/v1"
ODK_PROJECT_ID = int(os.getenv("ODK_PROJECT_ID", "3"))
ODK_FORM_ID = os.getenv("ODK_FORM_ID", "form_posko_v1")
ODK_USER = os.getenv("ODK_EMAIL", "")
ODK_PASS = os.getenv("ODK_PASSWORD", "")

PHOTO_DIR = Path(__file__).parent.parent / "storage/photos/arcgis_dump"
ENTITY_MAPPING_FILE = PHOTO_DIR / "entity_mapping.json"

# Track processed items
PROGRESS_FILE = PHOTO_DIR / "upload_progress.json"


def load_progress():
    """Load progress from file"""
    if PROGRESS_FILE.exists():
        with open(PROGRESS_FILE) as f:
            return json.load(f)
    return {"processed": [], "failed": [], "skipped": []}


def save_progress(progress):
    """Save progress to file"""
    with open(PROGRESS_FILE, 'w') as f:
        json.dump(progress, f, indent=2)


def get_session():
    """Create authenticated session"""
    session = requests.Session()
    session.auth = (ODK_USER, ODK_PASS)
    session.headers.update({
        "Content-Type": "application/xml"
    })
    return session


def download_submission_xml(session, instance_id):
    """Download current submission XML"""
    url = f"{ODK_BASE_URL}/projects/{ODK_PROJECT_ID}/forms/{ODK_FORM_ID}/submissions/uuid:{instance_id}.xml"
    response = session.get(url)
    if response.status_code == 200:
        return response.text
    else:
        print(f"   Error downloading XML: {response.status_code} - {response.text[:200]}")
        return None


def get_submission_info(session, instance_id):
    """Get submission metadata including current version"""
    url = f"{ODK_BASE_URL}/projects/{ODK_PROJECT_ID}/forms/{ODK_FORM_ID}/submissions/uuid:{instance_id}"
    response = session.get(url, headers={"Content-Type": "application/json"})
    if response.status_code == 200:
        return response.json()
    return None


def update_xml_with_photo(xml_content, photo_filename, new_instance_id, old_instance_id, entity_base_version):
    """Update XML with foto_depan field and versioning info"""

    # Parse XML
    # Remove namespace for easier manipulation
    xml_content = re.sub(r'\sxmlns="[^"]+"', '', xml_content)

    root = ET.fromstring(xml_content)

    # Find meta element
    meta = root.find('.//meta')
    if meta is None:
        print("   Error: No meta element found")
        return None

    # Update instanceID
    instance_id_elem = meta.find('instanceID')
    if instance_id_elem is not None:
        old_uuid = instance_id_elem.text
        instance_id_elem.text = f"uuid:{new_instance_id}"

    # Add deprecatedID
    deprecated_id = ET.SubElement(meta, 'deprecatedID')
    deprecated_id.text = old_uuid

    # Update entity element (change create to update)
    entity = meta.find('entity')
    if entity is not None:
        # Remove create attribute if exists
        if 'create' in entity.attrib:
            del entity.attrib['create']
        # Add update attributes
        entity.set('update', '1')
        entity.set('baseVersion', str(entity_base_version))

    # Check if grp_foto already exists
    grp_foto = root.find('.//grp_foto')
    if grp_foto is None:
        # Create grp_foto section
        grp_foto = ET.SubElement(root, 'grp_foto')

    # Check if foto_depan already has content
    foto_depan = grp_foto.find('foto_depan')
    if foto_depan is None:
        foto_depan = ET.SubElement(grp_foto, 'foto_depan')

    # Only update if empty or we're forcing update
    if not foto_depan.text or foto_depan.text.strip() == '':
        foto_depan.text = photo_filename
    else:
        print(f"   Warning: foto_depan already has value: {foto_depan.text}")
        return None  # Skip if already has photo

    # Convert back to string with proper declaration
    xml_str = ET.tostring(root, encoding='unicode')

    # Add XML declaration and namespace back
    xml_str = '<?xml version="1.0" encoding="UTF-8"?>\n' + xml_str.replace('<data', '<data xmlns="http://opendatakit.org/submissions"', 1)

    return xml_str


def upload_submission_version(session, instance_id, xml_content):
    """Upload new submission version"""
    url = f"{ODK_BASE_URL}/projects/{ODK_PROJECT_ID}/forms/{ODK_FORM_ID}/submissions/uuid:{instance_id}"

    response = session.put(
        url,
        data=xml_content.encode('utf-8'),
        headers={"Content-Type": "application/xml"}
    )

    if response.status_code in [200, 201]:
        return True, response.json() if response.text else {}
    else:
        return False, f"{response.status_code} - {response.text[:300]}"


def upload_attachment(session, instance_id, photo_path, photo_filename):
    """Upload photo attachment"""
    url = f"{ODK_BASE_URL}/projects/{ODK_PROJECT_ID}/forms/{ODK_FORM_ID}/submissions/uuid:{instance_id}/attachments/{photo_filename}"

    with open(photo_path, 'rb') as f:
        response = session.post(
            url,
            data=f.read(),
            headers={"Content-Type": "image/jpeg"}
        )

    if response.status_code in [200, 201]:
        return True, response.json() if response.text else {}
    else:
        return False, f"{response.status_code} - {response.text[:300]}"


def process_entry(session, entry, progress):
    """Process a single entry from mapping"""
    entity_id = entry["entity_id"]
    objectid = entry["objectid"]
    nama = entry["nama"]
    photos = entry.get("photos", [])

    # Skip if no photos
    if not photos:
        print(f"   Skipping {nama}: no photos")
        return "skipped", "no photos"

    # Skip if already processed
    if entity_id in progress["processed"]:
        print(f"   Skipping {nama}: already processed")
        return "skipped", "already processed"

    # Use first photo
    original_photo = photos[0]
    photo_path = PHOTO_DIR / str(objectid) / original_photo

    if not photo_path.exists():
        print(f"   Error: Photo not found: {photo_path}")
        return "failed", "photo not found"

    # Create new filename for ODK
    new_photo_filename = f"foto_depan_{entity_id[:8]}.jpg"

    print(f"   Downloading current XML...")
    xml_content = download_submission_xml(session, entity_id)
    if not xml_content:
        return "failed", "could not download XML"

    # Get current submission info for baseVersion
    sub_info = get_submission_info(session, entity_id)
    if not sub_info:
        print(f"   Error: Could not get submission info")
        return "failed", "could not get submission info"

    # Check current entity version
    current_version = sub_info.get("currentVersion", {})

    # For baseVersion, we need the entity's current version
    # This comes from the entity list, but we can use 1 for initial update
    entity_base_version = 1

    # Generate new UUID for this version
    new_instance_id = str(uuid.uuid4())

    print(f"   Updating XML with photo: {new_photo_filename}")
    updated_xml = update_xml_with_photo(
        xml_content,
        new_photo_filename,
        new_instance_id,
        entity_id,
        entity_base_version
    )

    if not updated_xml:
        return "skipped", "foto_depan already has value"

    print(f"   Uploading new version...")
    success, result = upload_submission_version(session, entity_id, updated_xml)
    if not success:
        print(f"   Error uploading XML: {result}")
        return "failed", f"upload failed: {result}"

    print(f"   Uploading attachment...")
    success, result = upload_attachment(session, entity_id, photo_path, new_photo_filename)
    if not success:
        print(f"   Error uploading attachment: {result}")
        return "failed", f"attachment failed: {result}"

    return "success", new_photo_filename


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Update ODK submissions with photos")
    parser.add_argument("--limit", type=int, default=0, help="Limit number of entries to process (0=all)")
    parser.add_argument("--dry-run", action="store_true", help="Don't actually upload, just show what would be done")
    parser.add_argument("--reset-progress", action="store_true", help="Reset progress tracking")
    parser.add_argument("--entity-id", type=str, help="Process only specific entity ID")
    args = parser.parse_args()

    # Check password
    if not ODK_PASS:
        print("Error: ODK_PASS environment variable not set")
        print("Run: export ODK_PASS='your_password'")
        sys.exit(1)

    # Load mapping
    print("Loading entity mapping...")
    with open(ENTITY_MAPPING_FILE) as f:
        mapping = json.load(f)

    # Filter entries with photos
    entries_with_photos = [e for e in mapping if e.get("photos")]
    print(f"Found {len(entries_with_photos)} entries with photos out of {len(mapping)} total")

    # Load progress
    if args.reset_progress:
        progress = {"processed": [], "failed": [], "skipped": []}
        save_progress(progress)
    else:
        progress = load_progress()

    print(f"Progress: {len(progress['processed'])} processed, {len(progress['failed'])} failed, {len(progress['skipped'])} skipped")

    # Filter by entity_id if specified
    if args.entity_id:
        entries_with_photos = [e for e in entries_with_photos if e["entity_id"] == args.entity_id]
        if not entries_with_photos:
            print(f"No entry found with entity_id: {args.entity_id}")
            sys.exit(1)

    # Apply limit
    if args.limit > 0:
        entries_with_photos = entries_with_photos[:args.limit]

    # Create session
    session = get_session()

    # Process entries
    stats = {"success": 0, "failed": 0, "skipped": 0}

    print(f"\nProcessing {len(entries_with_photos)} entries...")
    print("=" * 60)

    for i, entry in enumerate(entries_with_photos):
        entity_id = entry["entity_id"]
        nama = entry["nama"]
        objectid = entry["objectid"]

        print(f"\n[{i+1}/{len(entries_with_photos)}] {nama}")
        print(f"   Entity: {entity_id}")
        print(f"   ObjectID: {objectid}")

        if args.dry_run:
            print(f"   [DRY-RUN] Would process photos: {entry.get('photos', [])}")
            stats["skipped"] += 1
            continue

        try:
            status, message = process_entry(session, entry, progress)

            if status == "success":
                progress["processed"].append(entity_id)
                stats["success"] += 1
                print(f"   SUCCESS: {message}")
            elif status == "failed":
                progress["failed"].append({"entity_id": entity_id, "error": message})
                stats["failed"] += 1
                print(f"   FAILED: {message}")
            else:
                progress["skipped"].append({"entity_id": entity_id, "reason": message})
                stats["skipped"] += 1
                print(f"   SKIPPED: {message}")

            # Save progress after each entry
            save_progress(progress)

            # Rate limiting
            time.sleep(0.5)

        except Exception as e:
            print(f"   ERROR: {str(e)}")
            progress["failed"].append({"entity_id": entity_id, "error": str(e)})
            stats["failed"] += 1
            save_progress(progress)

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"  Success: {stats['success']}")
    print(f"  Failed:  {stats['failed']}")
    print(f"  Skipped: {stats['skipped']}")
    print(f"\nProgress saved to: {PROGRESS_FILE}")
    print("=" * 60)


if __name__ == "__main__":
    main()
