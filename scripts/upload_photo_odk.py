#!/usr/bin/env python3
"""
Upload Photos to ODK Central via Update Submission

This script uploads photos to existing ODK submissions by:
1. Creating an update submission with photo field reference
2. Uploading the photo binary as attachment

Procedure (from ODK Central API docs):
1. POST XML submission with mode=update and photo filename in image field
2. POST binary to /attachments/{filename} endpoint

Usage: python upload_photo_odk.py --submission-id <uuid> --photo-path <path> --photo-field <field_name>
"""

import os
import sys
import uuid
import argparse
import requests
from pathlib import Path
from datetime import datetime

# ODK Central configuration
ODK_BASE_URL = os.getenv('ODK_CENTRAL_URL', os.getenv('ODK_BASE_URL', 'https://data.dayawarga.com'))
ODK_PROJECT_ID = os.getenv('ODK_PROJECT_ID', '3')
ODK_FORM_ID = os.getenv('ODK_FORM_ID', 'form_posko_v1')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')

# Photo field mapping from ArcGIS keywords to ODK form fields
PHOTO_TYPE_MAP = {
    '_61_foto_pos_pengungsian': 'foto_depan',
    '_62_foto_pengungsi': 'foto_area1',
    '_63_foto_dapur': 'foto_dapur',
    '_64_foto_toilet': 'foto_toilet',
    '_65_foto_tempat_tidur': 'foto_area2',
    '_66_foto_lainnya': 'foto_area3',
    'foto_depan': 'foto_depan',
    'foto_area': 'foto_area1',
    'foto_faskes': 'foto_faskes',
    'foto_sampah': 'foto_sampah',
}


class ODKCentralClient:
    """Client for ODK Central API"""

    def __init__(self, base_url: str, email: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.email = email
        self.password = password
        self.session = requests.Session()
        self.token = None

    def authenticate(self) -> bool:
        """Authenticate and get session token"""
        url = f"{self.base_url}/v1/sessions"
        response = self.session.post(url, json={
            'email': self.email,
            'password': self.password
        })

        if response.status_code == 200:
            data = response.json()
            self.token = data.get('token')
            self.session.headers['Authorization'] = f'Bearer {self.token}'
            return True
        else:
            print(f"Authentication failed: {response.status_code} - {response.text}")
            return False

    def get_submission(self, project_id: str, form_id: str, instance_id: str):
        """Get existing submission data"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions/{instance_id}"
        response = self.session.get(url)
        if response.status_code == 200:
            return response.json()
        return None

    def get_submission_xml(self, project_id: str, form_id: str, instance_id: str):
        """Get existing submission XML"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions/{instance_id}.xml"
        response = self.session.get(url)
        if response.status_code == 200:
            return response.text
        return None

    def create_update_submission(self, project_id: str, form_id: str, xml: str):
        """Create an update submission"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions"
        headers = {'Content-Type': 'application/xml'}
        response = self.session.post(url, data=xml.encode('utf-8'), headers=headers)
        return response

    def upload_attachment(self, project_id: str, form_id: str, instance_id: str, filename: str, file_path: str):
        """Upload attachment to submission"""
        from urllib.parse import quote
        # URL-encode instance_id (uuid: becomes uuid%3A)
        encoded_id = quote(instance_id, safe='')
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions/{encoded_id}/attachments/{filename}"

        # Determine content type
        ext = Path(file_path).suffix.lower()
        content_types = {
            '.jpg': 'image/jpeg',
            '.jpeg': 'image/jpeg',
            '.png': 'image/png',
            '.gif': 'image/gif',
        }
        content_type = content_types.get(ext, 'application/octet-stream')

        headers = {'Content-Type': content_type}

        with open(file_path, 'rb') as f:
            response = self.session.post(url, data=f, headers=headers)

        return response


def escape_xml(value: str) -> str:
    """Escape XML special characters"""
    if not value:
        return ''
    value = str(value)
    value = value.replace('&', '&amp;')
    value = value.replace('<', '&lt;')
    value = value.replace('>', '&gt;')
    value = value.replace('"', '&quot;')
    return value


def build_update_xml(entity_id: str, photo_field: str, photo_filename: str, form_id: str = ODK_FORM_ID, submitter_name: str = "Dayawarga") -> str:
    """
    Build minimal update XML for adding photo to existing entity.

    For mode=update, we need:
    - mode = 'update'
    - sel_posko = entity_id (to select which entity to update)
    - The photo field with filename inside grp_foto group
    """
    instance_id = f"uuid:{uuid.uuid4()}"

    xml = f'''<?xml version="1.0" encoding="UTF-8"?>
<data id="{form_id}" version="2025011002"
      xmlns:entities="http://www.opendatakit.org/xforms/entities"
      entities:entities-version="2024.1.0">
  <meta>
    <instanceID>{instance_id}</instanceID>
  </meta>
  <mode>update</mode>
  <sel_posko>{entity_id}</sel_posko>
  <grp_foto>
    <{photo_field}>{escape_xml(photo_filename)}</{photo_field}>
  </grp_foto>
</data>'''

    return xml, instance_id


def detect_photo_field(filename: str) -> str:
    """Detect ODK photo field from filename"""
    filename_lower = filename.lower()
    for pattern, field in PHOTO_TYPE_MAP.items():
        if pattern.lower() in filename_lower:
            return field
    return 'foto_depan'  # Default


def main():
    parser = argparse.ArgumentParser(description='Upload photo to ODK Central submission')
    parser.add_argument('--entity-id', required=True, help='Entity ID (UUID without prefix)')
    parser.add_argument('--photo-path', required=True, help='Path to photo file')
    parser.add_argument('--photo-field', help='ODK form field name (auto-detect if not specified)')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be done without submitting')
    args = parser.parse_args()

    print("=" * 60)
    print("ODK Central Photo Upload")
    print("=" * 60)

    # Check credentials
    if not args.dry_run and (not ODK_EMAIL or not ODK_PASSWORD):
        print("\nERROR: ODK credentials not set!")
        print("Set environment variables:")
        print("  export ODK_EMAIL='your-email'")
        print("  export ODK_PASSWORD='your-password'")
        sys.exit(1)

    # Validate photo path
    photo_path = Path(args.photo_path)
    if not photo_path.exists():
        print(f"ERROR: Photo file not found: {photo_path}")
        sys.exit(1)

    # Detect or use provided photo field
    photo_field = args.photo_field or detect_photo_field(photo_path.name)
    photo_filename = photo_path.name

    print(f"\nEntity ID: {args.entity_id}")
    print(f"Photo: {photo_path}")
    print(f"Photo Field: {photo_field}")
    print(f"Filename: {photo_filename}")

    # Build update XML
    xml, instance_id = build_update_xml(args.entity_id, photo_field, photo_filename)

    print(f"\nGenerated Instance ID: {instance_id}")

    if args.dry_run:
        print("\n=== DRY RUN - XML to be submitted ===")
        print(xml)
        print("\n[DRY RUN] Would upload photo after submission")
        return

    # Connect to ODK Central
    print(f"\nConnecting to ODK Central: {ODK_BASE_URL}")
    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Step 1: Create update submission with photo reference
    print(f"\n1. Creating update submission...")
    response = client.create_update_submission(ODK_PROJECT_ID, ODK_FORM_ID, xml)

    if response.status_code not in [200, 201]:
        print(f"ERROR: Failed to create submission: {response.status_code}")
        print(response.text)
        sys.exit(1)

    print(f"   Submission created: {instance_id}")

    # Step 2: Upload photo attachment
    # Note: instance_id needs to be URL-encoded if it contains special chars
    clean_instance_id = instance_id.replace('uuid:', '')

    print(f"\n2. Uploading photo attachment...")
    response = client.upload_attachment(
        ODK_PROJECT_ID,
        ODK_FORM_ID,
        instance_id,  # Use full instance_id with uuid: prefix
        photo_filename,
        str(photo_path)
    )

    if response.status_code not in [200, 201]:
        print(f"ERROR: Failed to upload attachment: {response.status_code}")
        print(response.text)
        sys.exit(1)

    print(f"   Photo uploaded: {photo_filename}")

    print("\n" + "=" * 60)
    print("SUCCESS!")
    print("=" * 60)


if __name__ == '__main__':
    main()
