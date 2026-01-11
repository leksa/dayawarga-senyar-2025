#!/usr/bin/env python3
"""
Upload Faskes Data from Kemenkes CSV to ODK Central via Form Submission

This creates submissions to form_faskes_v1 which will auto-create entities.

Usage:
  python upload_faskes_odk.py --dry-run              # Preview without uploading
  python upload_faskes_odk.py --sample 1             # Upload 1 sample submission
  python upload_faskes_odk.py --limit 10             # Upload first 10 submissions
  python upload_faskes_odk.py                        # Upload all (365 submissions)
"""

import os
import sys
import uuid
import argparse
import re
import json
import requests
import pandas as pd
from pathlib import Path

# ODK Central configuration
ODK_BASE_URL = os.getenv('ODK_CENTRAL_URL', os.getenv('ODK_BASE_URL', 'https://data.dayawarga.com'))
ODK_PROJECT_ID = os.getenv('ODK_PROJECT_ID', '3')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')

# Faskes form
ODK_FORM_ID = 'form_faskes_v1'
ODK_FORM_VERSION = '2025011021'

# CSV path
CSV_PATH = Path(__file__).parent.parent / 'docs' / 'Faskes_Kemenkes_Aceh.csv'


def escape_xml(value):
    """Escape XML special characters"""
    if not value:
        return ''
    value = str(value)
    value = value.replace('&', '&amp;')
    value = value.replace('<', '&lt;')
    value = value.replace('>', '&gt;')
    value = value.replace('"', '&quot;')
    return value


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
            print(f"Auth failed: {response.status_code} - {response.text}")
            return False

    def submit_form(self, project_id: str, form_id: str, xml: str) -> tuple:
        """Submit XML form data"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions"
        headers = {'Content-Type': 'application/xml'}
        response = self.session.post(url, data=xml.encode('utf-8'), headers=headers)
        if response.status_code in [200, 201]:
            return True, response.json()
        else:
            return False, f"Status {response.status_code}: {response.text[:300]}"


def fix_coordinate(val, coord_type):
    if pd.isna(val):
        return None
    val_str = str(val)
    if re.match(r'^-?\d+\.\d+$', val_str):
        return float(val_str)
    clean = val_str.replace(',', '').replace('"', '').strip()
    if not clean:
        return None
    try:
        if coord_type == 'lon':
            if len(clean) > 2:
                result = float(clean[:2] + '.' + clean[2:])
            else:
                result = float(clean)
        else:
            if len(clean) > 1:
                result = float(clean[:1] + '.' + clean[1:])
            else:
                result = float(clean)
        return result
    except:
        return None


def load_and_transform_csv(csv_path: str) -> pd.DataFrame:
    df = pd.read_csv(csv_path)

    def fix_nama(row):
        jenis = row['JENIS FASYANKES']
        nama = row['NAMA']
        if jenis == 'Puskesmas' and not nama.startswith('Puskesmas'):
            return f'Puskesmas {nama}'
        return nama

    df['nama_faskes'] = df.apply(fix_nama, axis=1)
    df['lon'] = df['LONGITUDE'].apply(lambda x: fix_coordinate(x, 'lon'))
    df['lat'] = df['LATITUDE'].apply(lambda x: fix_coordinate(x, 'lat'))
    df['geometry'] = df.apply(
        lambda row: f"{row['lat']} {row['lon']} 0 0" if pd.notna(row['lat']) and pd.notna(row['lon']) else '',
        axis=1
    )

    jenis_map = {'Puskesmas': 'puskesmas', 'Rumah Sakit': 'rumah_sakit'}
    df['jenis_faskes'] = df['JENIS FASYANKES'].map(jenis_map)

    status_map = {'Beroperasi': 'operasional', 'Belum Beroperasi': 'non_aktif'}
    df['status_faskes'] = df['STATUS OPERASIONAL'].map(status_map)

    df['nama_kota_kab'] = df['KABUPATEN']
    df['nama_provinsi'] = df['PROVINSI']

    valid_coords = (df['lon'] >= 95) & (df['lon'] <= 99) & (df['lat'] >= 1) & (df['lat'] <= 6)
    df['valid_coords'] = valid_coords

    return df


def build_submission_xml(row: pd.Series) -> tuple:
    """Build XML submission for form_faskes_v1 with mode=baru"""
    instance_id = f"uuid:{uuid.uuid4()}"
    entity_id = str(uuid.uuid4())

    # Geometry format: "lat lon altitude accuracy"
    geometry = row['geometry'] if pd.notna(row.get('geometry')) else ''

    xml = f'''<?xml version="1.0" encoding="UTF-8"?>
<data id="{ODK_FORM_ID}" version="{ODK_FORM_VERSION}"
      xmlns:entities="http://www.opendatakit.org/xforms/entities"
      entities:entities-version="2024.1.0">
  <meta>
    <instanceID>{instance_id}</instanceID>
    <entity dataset="faskes_entities" id="{entity_id}" create="1">
      <label>{escape_xml(row['nama_faskes'])}</label>
    </entity>
  </meta>
  <mode>baru</mode>
  <nama_faskes>{escape_xml(row['nama_faskes'])}</nama_faskes>
  <grp_identitas>
    <jenis_faskes>{escape_xml(row['jenis_faskes'])}</jenis_faskes>
    <status_faskes>{escape_xml(row['status_faskes'])}</status_faskes>
    <koordinat>{escape_xml(geometry)}</koordinat>
  </grp_identitas>
  <grp_baseline>
    <baseline_sumber>BNPB/Kemenkes</baseline_sumber>
  </grp_baseline>
  <calc_nama_faskes>{escape_xml(row['nama_faskes'])}</calc_nama_faskes>
  <calc_nama_provinsi>{escape_xml(row['nama_provinsi'])}</calc_nama_provinsi>
  <calc_nama_kota_kab>{escape_xml(row['nama_kota_kab'])}</calc_nama_kota_kab>
  <calc_geometry>{escape_xml(geometry)}</calc_geometry>
</data>'''

    return xml, instance_id


def main():
    parser = argparse.ArgumentParser(description='Upload Faskes data to ODK Central via submission')
    parser.add_argument('--dry-run', action='store_true', help='Preview without uploading')
    parser.add_argument('--sample', type=int, default=0, help='Upload N sample submissions')
    parser.add_argument('--limit', type=int, default=0, help='Limit total submissions')
    args = parser.parse_args()

    print("=" * 60)
    print("ODK Central Faskes Upload (via Submission)")
    print("=" * 60)

    if not args.dry_run and (not ODK_EMAIL or not ODK_PASSWORD):
        print("\nERROR: ODK credentials not set!")
        sys.exit(1)

    print(f"\n1. Loading CSV: {CSV_PATH}")
    df = load_and_transform_csv(str(CSV_PATH))
    print(f"   Loaded {len(df)} rows")

    if args.sample > 0:
        df = df.head(args.sample)
        print(f"   Sample mode: {len(df)} submissions")
    elif args.limit > 0:
        df = df.head(args.limit)
        print(f"   Limited to {len(df)} submissions")

    print(f"\n2. Data Preview:")
    print(f"   Puskesmas: {(df['jenis_faskes'] == 'puskesmas').sum()}")
    print(f"   Rumah Sakit: {(df['jenis_faskes'] == 'rumah_sakit').sum()}")

    if args.dry_run:
        print(f"\n3. [DRY RUN] Sample XML submission:")
        for _, row in df.head(1).iterrows():
            xml, instance_id = build_submission_xml(row)
            print(f"   Instance ID: {instance_id}")
            print(f"   XML:\n{xml}")
        return

    print(f"\n3. Connecting to ODK Central: {ODK_BASE_URL}")
    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        sys.exit(1)
    print("   Authenticated!")

    print(f"\n4. Submitting to {ODK_FORM_ID}...")
    stats = {'success': 0, 'failed': 0}

    for i, (_, row) in enumerate(df.iterrows()):
        nama = row['nama_faskes']
        xml, instance_id = build_submission_xml(row)

        success, result = client.submit_form(ODK_PROJECT_ID, ODK_FORM_ID, xml)

        if success:
            print(f"   [{i+1}/{len(df)}] OK: {nama}")
            stats['success'] += 1
        else:
            print(f"   [{i+1}/{len(df)}] FAILED: {nama}")
            print(f"      Error: {result}")
            stats['failed'] += 1

    print("\n" + "=" * 60)
    print(f"SUCCESS: {stats['success']} | FAILED: {stats['failed']}")
    print("=" * 60)


if __name__ == '__main__':
    main()
