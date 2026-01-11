#!/usr/bin/env python3
"""
Update Faskes Entities in ODK Central with complete location data.

Updates the following fields (based on ODK dataset schema):
- nama_faskes: Facility name
- jenis_faskes: Type (puskesmas or rumah_sakit)
- nama_kecamatan: Kecamatan name
- nama_desa: Desa name
- nama_provinsi: Province name
- nama_kota_kab: City/Regency name
- id_desa: Desa ID code

Usage:
  python update_faskes_kecamatan_desa.py --dry-run    # Preview without updating
  python update_faskes_kecamatan_desa.py              # Update all entities
"""

import os
import sys
import csv
import argparse
import requests
from pathlib import Path
from urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# Load .env file if exists
ENV_FILE = Path(__file__).parent.parent / ".env"
if ENV_FILE.exists():
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                key, value = line.split('=', 1)
                value = value.strip().strip('"').strip("'")
                os.environ.setdefault(key.strip(), value)

# ODK Central configuration
ODK_BASE_URL = os.getenv("ODK_BASE_URL", os.getenv("ODK_CENTRAL_URL", "https://data.dayawarga.com"))
ODK_EMAIL = os.getenv("ODK_EMAIL", "")
ODK_PASSWORD = os.getenv("ODK_PASSWORD", "")
ODK_PROJECT_ID = os.getenv("ODK_PROJECT_ID", "3")
DATASET_NAME = "faskes_entities"

# CSV paths
DOCS_DIR = Path(__file__).parent.parent / "docs"
FASKES_CSV = DOCS_DIR / "Faskes_Kemenkes_Aceh.csv"
PROVINSI_CSV = DOCS_DIR / "provinsi.csv"
KOTA_KAB_CSV = DOCS_DIR / "kota_kab.csv"
KECAMATAN_CSV = DOCS_DIR / "kecamatan.csv"
DESA_CSV = DOCS_DIR / "desa.csv"


class ODKCentralClient:
    def __init__(self, base_url: str, email: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.verify = False
        self.email = email
        self.password = password

    def authenticate(self) -> bool:
        response = self.session.post(
            f"{self.base_url}/v1/sessions",
            json={"email": self.email, "password": self.password}
        )
        if response.status_code == 200:
            token = response.json()["token"]
            self.session.headers.update({"Authorization": f"Bearer {token}"})
            return True
        print(f"Auth failed: {response.status_code} - {response.text}")
        return False

    def get_entities(self) -> list:
        """Get all entities from faskes_entities dataset"""
        response = self.session.get(
            f"{self.base_url}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities"
        )
        response.raise_for_status()
        return response.json()

    def get_entity(self, entity_uuid: str) -> dict:
        """Get single entity details"""
        response = self.session.get(
            f"{self.base_url}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities/{entity_uuid}"
        )
        if response.status_code == 200:
            return response.json()
        return {}

    def update_entity(self, entity_uuid: str, label: str, data: dict, base_version: int) -> tuple:
        """Update entity properties via PATCH"""
        url = f"{self.base_url}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities/{entity_uuid}"

        payload = {
            "label": label,
            "data": data
        }

        response = self.session.patch(
            url,
            json=payload,
            params={"baseVersion": base_version}
        )

        if response.status_code in [200, 201]:
            return True, None
        else:
            return False, f"Status {response.status_code}: {response.text[:200]}"


def load_location_lookups():
    """Load all location CSV files and create lookups"""
    lookups = {}

    # Provinsi: nama -> kode
    lookups['provinsi'] = {}
    with open(PROVINSI_CSV, 'r', encoding='utf-8') as f:
        for row in csv.DictReader(f):
            nama = row.get('nama', '').upper()
            kode = row.get('kode', '')
            if nama and kode:
                lookups['provinsi'][nama] = kode

    # Kota/Kab: nama -> {kode, id_provinsi}
    lookups['kota_kab'] = {}
    lookups['kota_kab_by_kode'] = {}
    with open(KOTA_KAB_CSV, 'r', encoding='utf-8') as f:
        for row in csv.DictReader(f):
            nama = row.get('nama', '').upper()
            kode = row.get('kode', '')
            id_prov = row.get('id_provinsi', '')
            if nama and kode:
                lookups['kota_kab'][nama] = {'kode': kode, 'id_provinsi': id_prov}
                lookups['kota_kab_by_kode'][kode] = {'nama': row.get('nama', ''), 'id_provinsi': id_prov}

    # Kecamatan: nama -> {kode, id_kota_kab}
    lookups['kecamatan'] = {}
    lookups['kecamatan_by_kode'] = {}
    with open(KECAMATAN_CSV, 'r', encoding='utf-8') as f:
        for row in csv.DictReader(f):
            nama = row.get('nama', '').upper()
            kode = row.get('kode', '')
            id_kota_kab = row.get('id_kota_kab', '')
            if nama and kode:
                # Key by (id_kota_kab, nama) to handle same name in different kab
                key = (id_kota_kab, nama)
                lookups['kecamatan'][key] = {'kode': kode, 'id_kota_kab': id_kota_kab, 'nama': row.get('nama', '')}
                lookups['kecamatan_by_kode'][kode] = {'nama': row.get('nama', ''), 'id_kota_kab': id_kota_kab}

    # Desa: nama -> {kode, id_kec}
    lookups['desa'] = {}
    lookups['desa_by_kode'] = {}
    with open(DESA_CSV, 'r', encoding='utf-8') as f:
        for row in csv.DictReader(f):
            nama = row.get('nama', '').upper()
            kode = row.get('kode', '')
            id_kec = row.get('id_kec', '')
            if nama and kode:
                # Key by (id_kec, nama) to handle same name in different kec
                key = (id_kec, nama)
                lookups['desa'][key] = {'kode': kode, 'id_kec': id_kec, 'nama': row.get('nama', '')}
                lookups['desa_by_kode'][kode] = {'nama': row.get('nama', ''), 'id_kec': id_kec}

    return lookups


def normalize_kab_name(nama):
    """Normalize kabupaten name for lookup"""
    nama = nama.upper().strip()
    # Add KAB. prefix if not present and not a Kota
    if not nama.startswith('KAB.') and not nama.startswith('KOTA'):
        nama = f'KAB. {nama}'
    return nama


def load_faskes_data(lookups) -> dict:
    """Load faskes CSV and create lookup by nama_faskes with all location IDs"""
    data = {}

    with open(FASKES_CSV, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            # Build faskes name same as upload script
            jenis = row.get('JENIS FASYANKES', '')
            nama = row.get('NAMA', '')

            if jenis == 'Puskesmas' and not nama.startswith('Puskesmas'):
                nama_faskes = f'Puskesmas {nama}'
            else:
                nama_faskes = nama

            # Determine jenis_faskes value
            if jenis == 'Puskesmas':
                jenis_faskes = 'puskesmas'
            elif jenis == 'Rumah Sakit':
                jenis_faskes = 'rumah_sakit'
            else:
                jenis_faskes = ''

            provinsi = row.get('PROVINSI', '').strip()
            kabupaten = row.get('KABUPATEN', '').strip()
            kecamatan = row.get('KECAMATAN', '').strip()
            desa = row.get('DESA', '').strip()

            # Look up IDs
            sel_provinsi = ''
            sel_kota_kab = ''
            sel_kecamatan = ''
            sel_desa = ''

            # Provinsi ID
            if provinsi:
                sel_provinsi = lookups['provinsi'].get(provinsi.upper(), '')

            # Kota/Kab ID
            if kabupaten:
                kab_norm = normalize_kab_name(kabupaten)
                kab_info = lookups['kota_kab'].get(kab_norm)
                if kab_info:
                    sel_kota_kab = kab_info['kode']

            # Kecamatan ID - need to find by (id_kota_kab, nama)
            if kecamatan and sel_kota_kab:
                kec_key = (sel_kota_kab, kecamatan.upper())
                kec_info = lookups['kecamatan'].get(kec_key)
                if kec_info:
                    sel_kecamatan = kec_info['kode']

            # Desa ID - need to find by (id_kec, nama)
            if desa and sel_kecamatan:
                desa_key = (sel_kecamatan, desa.upper())
                desa_info = lookups['desa'].get(desa_key)
                if desa_info:
                    sel_desa = desa_info['kode']

            data[nama_faskes] = {
                'nama_faskes': nama_faskes,
                'jenis_faskes': jenis_faskes,
                'nama_kecamatan': kecamatan,
                'nama_desa': desa,
                'nama_provinsi': provinsi,
                'nama_kota_kab': kabupaten,
                'id_desa': sel_desa,
            }

    return data


def main():
    parser = argparse.ArgumentParser(description='Update faskes entities with complete location data')
    parser.add_argument('--dry-run', action='store_true', help='Preview without updating')
    args = parser.parse_args()

    print("=" * 70)
    print("Update Faskes Entities with Complete Location Data")
    print("=" * 70)

    # Load location lookups
    print("\n1. Loading location lookup tables...")
    lookups = load_location_lookups()
    print(f"   Provinsi: {len(lookups['provinsi'])} entries")
    print(f"   Kota/Kab: {len(lookups['kota_kab'])} entries")
    print(f"   Kecamatan: {len(lookups['kecamatan'])} entries")
    print(f"   Desa: {len(lookups['desa'])} entries")

    # Load faskes data
    print(f"\n2. Loading faskes CSV: {FASKES_CSV}")
    faskes_data = load_faskes_data(lookups)
    print(f"   Found {len(faskes_data)} faskes")

    # Stats for data found
    with_provinsi = sum(1 for d in faskes_data.values() if d['nama_provinsi'])
    with_kota_kab = sum(1 for d in faskes_data.values() if d['nama_kota_kab'])
    with_kecamatan = sum(1 for d in faskes_data.values() if d['nama_kecamatan'])
    with_desa = sum(1 for d in faskes_data.values() if d['nama_desa'])
    with_id_desa = sum(1 for d in faskes_data.values() if d['id_desa'])
    print(f"   With nama_provinsi:  {with_provinsi}")
    print(f"   With nama_kota_kab:  {with_kota_kab}")
    print(f"   With nama_kecamatan: {with_kecamatan}")
    print(f"   With nama_desa:      {with_desa}")
    print(f"   With id_desa:        {with_id_desa}")

    if args.dry_run:
        print("\n   [DRY RUN] Sample data:")
        for nama, data in list(faskes_data.items())[:5]:
            print(f"\n   - {nama}")
            print(f"     jenis_faskes: {data['jenis_faskes']}")
            print(f"     nama_provinsi: {data['nama_provinsi']}")
            print(f"     nama_kota_kab: {data['nama_kota_kab']}")
            print(f"     nama_kecamatan: {data['nama_kecamatan']}")
            print(f"     nama_desa: {data['nama_desa']}")
            print(f"     id_desa: {data['id_desa']}")
        print("\n   Run without --dry-run to update ODK Central")
        return

    # Check credentials
    if not ODK_EMAIL or not ODK_PASSWORD:
        print("\nERROR: ODK credentials not set!")
        print("Set ODK_EMAIL and ODK_PASSWORD environment variables")
        sys.exit(1)

    # Connect to ODK Central
    print(f"\n3. Connecting to ODK Central: {ODK_BASE_URL}")
    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        sys.exit(1)
    print("   Authenticated!")

    # Get entities from ODK
    print(f"\n4. Fetching entities from {DATASET_NAME}...")
    try:
        entities = client.get_entities()
        print(f"   Found {len(entities)} entities")
    except Exception as e:
        print(f"   Error: {e}")
        sys.exit(1)

    # Update entities
    print("\n5. Updating entities...")
    stats = {
        'updated': 0,
        'skipped_no_csv': 0,
        'skipped_no_change': 0,
        'errors': 0
    }

    for entity in entities:
        entity_uuid = entity.get('uuid')
        current_version = entity.get('currentVersion') or {}
        label = current_version.get('label', '')
        current_data = entity.get('data') or {}

        # Look up in faskes data
        if label not in faskes_data:
            stats['skipped_no_csv'] += 1
            continue

        csv_info = faskes_data[label]

        # Check if any field needs update
        fields_to_update = [
            'nama_faskes', 'jenis_faskes',
            'nama_kecamatan', 'nama_desa',
            'nama_provinsi', 'nama_kota_kab', 'id_desa'
        ]

        needs_update = False
        for field in fields_to_update:
            new_val = csv_info.get(field, '')
            current_val = current_data.get(field, '')
            if new_val and new_val != current_val:
                needs_update = True
                break

        if not needs_update:
            stats['skipped_no_change'] += 1
            continue

        # Get current version for update
        entity_detail = client.get_entity(entity_uuid)
        base_version = entity_detail.get('currentVersion', {}).get('version', 1)
        current_data = entity_detail.get('data') or {}

        # Prepare update data - merge with existing
        update_data = dict(current_data)
        for field in fields_to_update:
            new_val = csv_info.get(field, '')
            if new_val:
                update_data[field] = new_val

        # Update entity
        success, error = client.update_entity(entity_uuid, label, update_data, base_version)

        if success:
            stats['updated'] += 1
            print(f"   [{stats['updated']}] Updated: {label}")
            print(f"       jenis={csv_info['jenis_faskes']}, kec={csv_info['nama_kecamatan']}, desa={csv_info['nama_desa']}")
            if csv_info['id_desa']:
                print(f"       id_desa={csv_info['id_desa']}")
        else:
            stats['errors'] += 1
            print(f"   Error updating {label}: {error}")

    # Summary
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)
    print(f"  Updated:          {stats['updated']}")
    print(f"  No CSV data:      {stats['skipped_no_csv']}")
    print(f"  No changes:       {stats['skipped_no_change']}")
    print(f"  Errors:           {stats['errors']}")
    print("=" * 70)


if __name__ == "__main__":
    main()
