#!/usr/bin/env python3
"""
Crawl Photos from ArcGIS Survey123 Data Dump

This script fetches photos from the ArcGIS Feature Service and saves them locally,
then links them to existing locations in PostgreSQL based on matching criteria.

ArcGIS URL patterns:
- Features: https://services5.arcgis.com/.../FeatureServer/0/query?where=...&f=geojson
- Attachments list: https://services5.arcgis.com/.../FeatureServer/0/{objectid}/attachments?f=json
- Attachment file: https://services5.arcgis.com/.../FeatureServer/0/{objectid}/attachments/{attachment_id}
"""

import os
import sys
import json
import time
import hashlib
import argparse
import subprocess
from pathlib import Path
from urllib.parse import quote
import requests

# ArcGIS Feature Service configuration
ARCGIS_BASE_URL = "https://services5.arcgis.com/aQMqya7Haac8J82d/arcgis/rest/services/survey123_6892e471b0ea4921aa4682468efdab3a_results/FeatureServer/0"

# Output directory for downloaded photos
PHOTO_OUTPUT_DIR = os.getenv("PHOTO_OUTPUT_DIR", "./storage/photos/arcgis_dump")

# PostgreSQL connection via docker
PG_DOCKER_CONTAINER = "senyar-postgres"
PG_USER = "senyar"
PG_DB = "senyar"


def fetch_features(province="Aceh", batch_size=1000):
    """Fetch all features from ArcGIS for a given province"""
    all_features = []
    offset = 0

    while True:
        url = f"{ARCGIS_BASE_URL}/query"
        params = {
            "where": f"_24_provinsi = '{province}'",
            "outFields": "*",
            "f": "geojson",
            "resultRecordCount": batch_size,
            "resultOffset": offset
        }

        print(f"   Fetching features offset={offset}...")
        response = requests.get(url, params=params)
        response.raise_for_status()

        data = response.json()
        features = data.get("features", [])

        if not features:
            break

        all_features.extend(features)
        print(f"   Got {len(features)} features (total: {len(all_features)})")

        # Check if we got fewer than requested (last batch)
        if len(features) < batch_size:
            break

        offset += batch_size
        time.sleep(0.5)  # Rate limiting

    return all_features


def fetch_attachments(object_id):
    """Fetch attachment list for a feature"""
    url = f"{ARCGIS_BASE_URL}/{object_id}/attachments"
    params = {"f": "json"}

    try:
        response = requests.get(url, params=params)
        response.raise_for_status()
        data = response.json()
        return data.get("attachmentInfos", [])
    except Exception as e:
        print(f"   Warning: Failed to fetch attachments for {object_id}: {e}")
        return []


def download_attachment(object_id, attachment_id, filename, output_dir):
    """Download a single attachment"""
    url = f"{ARCGIS_BASE_URL}/{object_id}/attachments/{attachment_id}"
    output_path = Path(output_dir) / filename

    # Skip if already downloaded
    if output_path.exists():
        return str(output_path), True

    try:
        response = requests.get(url, stream=True)
        response.raise_for_status()

        # Ensure directory exists
        output_path.parent.mkdir(parents=True, exist_ok=True)

        with open(output_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)

        return str(output_path), False
    except Exception as e:
        print(f"   Error downloading {filename}: {e}")
        return None, False


def extract_photo_type(keywords, filename):
    """Extract photo type from keywords or filename"""
    keywords = keywords or ""
    filename = filename or ""

    # Map keywords to photo types
    keyword_map = {
        "_61_foto_pos_pengungsian": "tampak_depan",
        "_62_foto_pengungsi": "pengungsi",
        "_63_foto_dapur": "dapur",
        "_64_foto_toilet": "toilet",
        "_65_foto_tempat_tidur": "area_tidur",
        "_66_foto_lainnya": "lainnya",
        "foto_depan": "tampak_depan",
        "foto_area": "area",
        "foto_faskes": "faskes",
        "foto_sampah": "sampah",
    }

    for key, photo_type in keyword_map.items():
        if key in keywords.lower() or key in filename.lower():
            return photo_type

    return "lainnya"


def normalize_name(name):
    """Normalize location name for matching"""
    if not name:
        return ""
    # Lowercase, remove extra spaces, common prefixes
    name = name.lower().strip()
    prefixes = ["pos ", "posko ", "pengungsian ", "pos pengungsian "]
    for prefix in prefixes:
        if name.startswith(prefix):
            name = name[len(prefix):]
    return name


def find_matching_location(feature_props):
    """Find matching location in PostgreSQL based on feature properties"""
    # Try to match by name and desa
    nama = feature_props.get("_21_nama_pos_pengungsian", "")
    desa = feature_props.get("_27_desa", "")
    kecamatan = feature_props.get("_26_kecamatan", "")

    if not nama:
        return None

    # Build query to find matching location
    # Try exact name match first, then fuzzy
    query = f"""
    SELECT id, nama, alamat->>'nama_desa' as desa
    FROM locations
    WHERE deleted_at IS NULL
      AND (
        LOWER(nama) = LOWER('{nama.replace("'", "''")}')
        OR LOWER(nama) LIKE '%{normalize_name(nama).replace("'", "''")}%'
      )
    LIMIT 1
    """

    result = subprocess.run([
        'docker', 'exec', PG_DOCKER_CONTAINER, 'psql', '-U', PG_USER, '-d', PG_DB,
        '-t', '-A', '-c', query
    ], capture_output=True, text=True)

    if result.returncode == 0 and result.stdout.strip():
        parts = result.stdout.strip().split('|')
        if len(parts) >= 2:
            return {"id": parts[0], "nama": parts[1]}

    return None


def insert_photo_record(location_id, photo_type, filename, file_path, file_size):
    """Insert photo record into PostgreSQL"""
    import uuid
    photo_id = str(uuid.uuid4())

    query = f"""
    INSERT INTO location_photos (id, location_id, photo_type, filename, file_path, file_size, is_cached, created_at, updated_at)
    VALUES ('{photo_id}', '{location_id}', '{photo_type}', '{filename}', '{file_path}', {file_size}, true, NOW(), NOW())
    ON CONFLICT (location_id, filename) DO NOTHING
    RETURNING id
    """

    result = subprocess.run([
        'docker', 'exec', PG_DOCKER_CONTAINER, 'psql', '-U', PG_USER, '-d', PG_DB,
        '-t', '-A', '-c', query
    ], capture_output=True, text=True)

    return result.returncode == 0


def main():
    parser = argparse.ArgumentParser(description="Crawl photos from ArcGIS Survey123")
    parser.add_argument("--province", default="Aceh", help="Province filter")
    parser.add_argument("--output-dir", default=PHOTO_OUTPUT_DIR, help="Output directory for photos")
    parser.add_argument("--dry-run", action="store_true", help="Don't download, just show what would be done")
    parser.add_argument("--limit", type=int, default=0, help="Limit number of features to process (0=all)")
    parser.add_argument("--link-db", action="store_true", help="Link photos to PostgreSQL locations")
    args = parser.parse_args()

    print("=" * 60)
    print("ArcGIS Photo Crawler")
    print("=" * 60)

    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Fetch features
    print(f"\n1. Fetching features from ArcGIS (province={args.province})...")
    features = fetch_features(args.province)
    print(f"   Found {len(features)} features")

    if args.limit > 0:
        features = features[:args.limit]
        print(f"   Limited to {len(features)} features")

    # Process each feature
    print(f"\n2. Processing features and downloading photos...")

    stats = {
        "features_processed": 0,
        "photos_found": 0,
        "photos_downloaded": 0,
        "photos_skipped": 0,
        "photos_linked": 0,
        "locations_matched": 0,
    }

    # Store mapping for later DB linking
    photo_mapping = []

    for i, feature in enumerate(features):
        props = feature.get("properties", {})
        object_id = props.get("objectid")
        nama = props.get("_21_nama_pos_pengungsian", "Unknown")

        if not object_id:
            continue

        print(f"\n   [{i+1}/{len(features)}] {nama} (ObjectID: {object_id})")

        # Fetch attachments
        attachments = fetch_attachments(object_id)

        if not attachments:
            print(f"      No attachments")
            stats["features_processed"] += 1
            continue

        print(f"      Found {len(attachments)} attachments")
        stats["photos_found"] += len(attachments)

        # Find matching location in DB
        matching_location = None
        if args.link_db:
            matching_location = find_matching_location(props)
            if matching_location:
                stats["locations_matched"] += 1
                print(f"      Matched to: {matching_location['nama']}")

        # Download each attachment
        for att in attachments:
            att_id = att.get("id")
            att_name = att.get("name", f"photo_{att_id}.jpg")
            att_keywords = att.get("keywords", "")
            att_size = att.get("size", 0)

            # Create subdirectory by objectid
            feature_dir = output_dir / str(object_id)

            if args.dry_run:
                print(f"      [DRY-RUN] Would download: {att_name}")
                stats["photos_downloaded"] += 1
                continue

            # Download
            file_path, was_cached = download_attachment(object_id, att_id, att_name, feature_dir)

            if file_path:
                if was_cached:
                    print(f"      [CACHED] {att_name}")
                    stats["photos_skipped"] += 1
                else:
                    print(f"      [DOWNLOADED] {att_name}")
                    stats["photos_downloaded"] += 1

                # Store for DB linking
                photo_type = extract_photo_type(att_keywords, att_name)
                photo_mapping.append({
                    "object_id": object_id,
                    "feature_name": nama,
                    "feature_props": props,
                    "attachment_id": att_id,
                    "filename": att_name,
                    "file_path": file_path,
                    "file_size": att_size,
                    "photo_type": photo_type,
                    "location": matching_location
                })

        stats["features_processed"] += 1
        time.sleep(0.3)  # Rate limiting

    # Link photos to PostgreSQL
    if args.link_db and not args.dry_run:
        print(f"\n3. Linking photos to PostgreSQL locations...")
        for photo in photo_mapping:
            if photo["location"]:
                success = insert_photo_record(
                    photo["location"]["id"],
                    photo["photo_type"],
                    photo["filename"],
                    photo["file_path"],
                    photo["file_size"]
                )
                if success:
                    stats["photos_linked"] += 1

    # Save mapping to JSON for reference
    mapping_file = output_dir / "photo_mapping.json"
    with open(mapping_file, 'w') as f:
        json.dump(photo_mapping, f, indent=2, default=str)
    print(f"\n   Saved mapping to: {mapping_file}")

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"  Features processed:  {stats['features_processed']}")
    print(f"  Photos found:        {stats['photos_found']}")
    print(f"  Photos downloaded:   {stats['photos_downloaded']}")
    print(f"  Photos skipped:      {stats['photos_skipped']} (already cached)")
    if args.link_db:
        print(f"  Locations matched:   {stats['locations_matched']}")
        print(f"  Photos linked to DB: {stats['photos_linked']}")
    print("=" * 60)


if __name__ == "__main__":
    main()
