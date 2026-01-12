#!/usr/bin/env python3
"""
Script to import jembatan/jalan data as submissions to ODK Central
Usage: python import_jembatan_submissions.py [--dry-run] [--limit N]
"""

import os
import sys
import csv
import json
import argparse
import requests
from datetime import datetime
from uuid import uuid4

# Configuration
ODK_BASE_URL = os.getenv('ODK_CENTRAL_URL', 'https://data.dayawarga.com')
ODK_PROJECT_ID = os.getenv('ODK_PROJECT_ID', '3')
ODK_FORM_ID = os.getenv('ODK_FORM_ID', 'form_jembatan_v1')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')

# Path to CSV file
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
CSV_FILE = os.path.join(SCRIPT_DIR, '..', 'docs', 'jembatan_entities.csv')


class ODKCentralClient:
    def __init__(self, base_url: str, email: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.email = email
        self.password = password
        self.session = requests.Session()
        self.token = None

    def authenticate(self) -> bool:
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

    def get_existing_submissions(self, project_id: str, form_id: str) -> set:
        """Get set of existing entity IDs from submissions"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}.svc/Submissions"
        response = self.session.get(url)

        existing = set()
        if response.status_code == 200:
            data = response.json()
            for sub in data.get('value', []):
                # Extract entity ID from grp_identifikasi.sel_jembatan
                grp = sub.get('grp_identifikasi', {})
                entity_id = grp.get('sel_jembatan')
                if entity_id:
                    existing.add(entity_id)
        return existing

    def create_submission(self, project_id: str, form_id: str, submission_xml: str) -> dict:
        """Create a new submission via XML"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions"

        headers = {
            'Content-Type': 'application/xml'
        }

        response = self.session.post(url, data=submission_xml.encode('utf-8'), headers=headers)

        if response.status_code in [200, 201]:
            return {'success': True, 'data': response.json()}
        else:
            return {'success': False, 'error': response.text, 'status': response.status_code}

    def approve_submission(self, project_id: str, form_id: str, instance_id: str) -> dict:
        """Approve a submission"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions/{instance_id}"

        response = self.session.patch(url, json={'reviewState': 'approved'})

        if response.status_code == 200:
            return {'success': True}
        else:
            return {'success': False, 'error': response.text}


def map_status_akses(status1: str) -> str:
    """Map status1 to status_akses value"""
    status_lower = (status1 or '').lower()
    if 'tidak' in status_lower or 'putus' in status_lower:
        return 'tidak_dapat_diakses'
    elif 'terbatas' in status_lower:
        return 'dapat_diakses_terbatas'
    else:
        return 'dapat_diakses'


def map_keterangan_bencana(keterangan: str) -> str:
    """Map keterangan to keterangan_bencana value"""
    keterangan_lower = (keterangan or '').lower()
    if 'longsor' in keterangan_lower:
        return 'longsor'
    elif 'banjir' in keterangan_lower:
        return 'banjir_longsor'
    else:
        return 'banjir_longsor'


def map_status_penanganan(penanganan: str) -> str:
    """Map penanganan to status_penanganan value"""
    penanganan_lower = (penanganan or '').lower()
    if 'sudah' in penanganan_lower or 'selesai' in penanganan_lower:
        return 'sudah_ditangani'
    elif 'sedang' in penanganan_lower or 'proses' in penanganan_lower:
        return 'sedang_ditangani'
    else:
        return 'belum_ditangani'


def map_bailey(bailey: str) -> str:
    """Map bailey value"""
    bailey_lower = (bailey or '').lower()
    if 'terpasang' in bailey_lower or 'ada' in bailey_lower or bailey_lower == 'ya':
        return 'terpasang'
    elif 'proses' in bailey_lower:
        return 'dalam_proses'
    else:
        return 'tidak_ada'


def create_submission_xml(row: dict, instance_id: str) -> str:
    """Create ODK submission XML from CSV row"""

    now = datetime.utcnow().strftime('%Y-%m-%dT%H:%M:%S.000Z')

    # Map values
    status_akses = map_status_akses(row.get('status1', ''))
    keterangan_bencana = map_keterangan_bencana(row.get('keterangan', ''))
    status_penanganan = map_status_penanganan(row.get('penanganan', ''))
    bailey = map_bailey(row.get('bailey', ''))

    # Entity ID is in 'nama' column (UUID)
    entity_id = row.get('nama', '')

    # Escape XML special characters
    def escape_xml(s):
        if s is None:
            return ''
        return (str(s)
            .replace('&', '&amp;')
            .replace('<', '&lt;')
            .replace('>', '&gt;')
            .replace('"', '&quot;')
            .replace("'", '&apos;'))

    xml = f'''<?xml version="1.0" encoding="UTF-8"?>
<data id="{ODK_FORM_ID}" version="2026011202">
  <meta>
    <instanceID>uuid:{instance_id}</instanceID>
    <instanceName>{escape_xml(row.get('label', ''))} - {datetime.now().strftime('%Y-%m-%d')}</instanceName>
  </meta>
  <start>{now}</start>
  <end>{now}</end>
  <username>data-import</username>
  <deviceid>import-script</deviceid>
  <baseline_sumber>{escape_xml(row.get('baseline_sumber', 'BNPB/PU'))}</baseline_sumber>
  <update_by>data-import</update_by>
  <update_date>{now}</update_date>
  <grp_identifikasi>
    <sel_jembatan>{escape_xml(entity_id)}</sel_jembatan>
    <c_objectid>{escape_xml(row.get('objectid', ''))}</c_objectid>
    <c_nama>{escape_xml(row.get('label', ''))}</c_nama>
    <c_jenis>{escape_xml(row.get('jenis', ''))}</c_jenis>
    <c_statusjln>{escape_xml(row.get('statusjln', ''))}</c_statusjln>
    <c_kabupaten>{escape_xml(row.get('nama_kabupaten', ''))}</c_kabupaten>
    <c_provinsi>{escape_xml(row.get('nama_provinsi', ''))}</c_provinsi>
    <c_latitude>{escape_xml(row.get('latitude', ''))}</c_latitude>
    <c_longitude>{escape_xml(row.get('longitude', ''))}</c_longitude>
    <c_target_selesai>{escape_xml(row.get('target_selesai', ''))}</c_target_selesai>
  </grp_identifikasi>
  <grp_status>
    <status_akses>{status_akses}</status_akses>
    <keterangan_bencana>{keterangan_bencana}</keterangan_bencana>
    <dampak>{escape_xml(row.get('dampak', ''))}</dampak>
  </grp_status>
  <grp_penanganan>
    <status_penanganan>{status_penanganan}</status_penanganan>
    <penanganan_detail>{escape_xml(row.get('penanganan', ''))}</penanganan_detail>
    <bailey>{bailey}</bailey>
    <progress>0</progress>
  </grp_penanganan>
  <grp_foto>
    <foto_1/>
    <foto_2/>
    <foto_3/>
    <foto_4/>
  </grp_foto>
</data>'''

    return xml


def main():
    parser = argparse.ArgumentParser(description='Import jembatan data as ODK submissions')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be imported without doing it')
    parser.add_argument('--limit', type=int, default=0, help='Limit number of imports (0 = all)')
    parser.add_argument('--auto-approve', action='store_true', default=True, help='Auto-approve submissions after import')
    parser.add_argument('--csv', type=str, default=CSV_FILE, help='Path to CSV file')
    args = parser.parse_args()

    print("=" * 60)
    print("ODK CENTRAL - IMPORT JEMBATAN SUBMISSIONS")
    print("=" * 60)

    # Check credentials
    if not ODK_EMAIL or not ODK_PASSWORD:
        print("\nERROR: ODK credentials not set!")
        print("Set environment variables:")
        print("  export ODK_EMAIL='your-email'")
        print("  export ODK_PASSWORD='your-password'")
        sys.exit(1)

    # Check CSV file
    if not os.path.exists(args.csv):
        print(f"\nERROR: CSV file not found: {args.csv}")
        sys.exit(1)

    # Connect
    print(f"\nConnecting to: {ODK_BASE_URL}")
    print(f"Project: {ODK_PROJECT_ID}, Form: {ODK_FORM_ID}")

    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Get existing submissions to avoid duplicates
    print("\nChecking existing submissions...")
    existing_entities = client.get_existing_submissions(ODK_PROJECT_ID, ODK_FORM_ID)
    print(f"Found {len(existing_entities)} existing submissions")

    # Read CSV
    print(f"\nReading CSV: {args.csv}")
    rows = []
    with open(args.csv, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            rows.append(row)

    print(f"Total rows in CSV: {len(rows)}")

    # Filter out existing
    to_import = []
    for row in rows:
        entity_id = row.get('nama', '')  # UUID is in 'nama' column
        if entity_id and entity_id not in existing_entities:
            to_import.append(row)

    print(f"New rows to import: {len(to_import)}")
    print(f"Skipped (already exists): {len(rows) - len(to_import)}")

    if len(to_import) == 0:
        print("\nNo new data to import!")
        return

    # Apply limit
    if args.limit > 0:
        to_import = to_import[:args.limit]
        print(f"Limited to: {len(to_import)} rows")

    if args.dry_run:
        print("\n[DRY RUN] Would import:")
        for i, row in enumerate(to_import[:10]):
            print(f"  {i+1}. {row.get('label', 'Unknown')} ({row.get('jenis', '')})")
        if len(to_import) > 10:
            print(f"  ... and {len(to_import) - 10} more")
        print("\nTo import for real, remove --dry-run flag")
        return

    # Import submissions
    print("\nImporting submissions...")
    success_count = 0
    error_count = 0
    imported_ids = []

    for i, row in enumerate(to_import):
        instance_id = str(uuid4())
        label = row.get('label', 'Unknown')

        try:
            xml = create_submission_xml(row, instance_id)
            result = client.create_submission(ODK_PROJECT_ID, ODK_FORM_ID, xml)

            if result.get('success'):
                success_count += 1
                imported_ids.append(f"uuid:{instance_id}")
                print(f"✓ [{i+1}/{len(to_import)}] {label[:40]}...")
            else:
                error_count += 1
                error_msg = result.get('error', 'Unknown error')[:80]
                print(f"✗ [{i+1}/{len(to_import)}] {label[:40]}... - {error_msg}")
        except Exception as e:
            error_count += 1
            print(f"✗ [{i+1}/{len(to_import)}] {label[:40]}... - {str(e)[:80]}")

    # Auto-approve imported submissions
    if args.auto_approve and imported_ids:
        print(f"\nAuto-approving {len(imported_ids)} submissions...")
        approved = 0
        for instance_id in imported_ids:
            result = client.approve_submission(ODK_PROJECT_ID, ODK_FORM_ID, instance_id)
            if result.get('success'):
                approved += 1
        print(f"Approved: {approved}/{len(imported_ids)}")

    # Summary
    print("\n" + "=" * 60)
    print(f"SUMMARY: {success_count} imported, {error_count} errors")
    print("=" * 60)


if __name__ == '__main__':
    main()
