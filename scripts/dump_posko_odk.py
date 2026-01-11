#!/usr/bin/env python3
"""
Script to dump posko data from Excel to ODK Central with Entity Creation
Usage: python dump_posko_odk.py [--dry-run] [--limit N] [--start N]

This script creates submissions that also create entities in ODK Central.
Each submission includes entity metadata in the meta block to force entity creation.
"""

import os
import sys
import json
import uuid
import argparse
import requests
from datetime import datetime
from typing import Optional, Dict, Any
import pandas as pd

# ========================================
# CONFIGURATION
# ========================================

# ODK Central configuration - load from environment or use defaults
ODK_BASE_URL = os.getenv('ODK_CENTRAL_URL', os.getenv('ODK_BASE_URL', 'https://data.dayawarga.com'))
ODK_PROJECT_ID = os.getenv('ODK_PROJECT_ID', '3')
ODK_FORM_ID = os.getenv('ODK_FORM_ID', 'form_posko_v1')
ODK_EMAIL = os.getenv('ODK_EMAIL', '')
ODK_PASSWORD = os.getenv('ODK_PASSWORD', '')

# Entity configuration - must match XLSForm entities sheet
ENTITY_DATASET = 'posko_entities'  # list_name in entities sheet

# File paths
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(SCRIPT_DIR)
DOCS_DIR = os.path.join(PROJECT_ROOT, 'docs')

DATA_FILE = os.path.join(DOCS_DIR, 'data dump posko_v1.xlsx')
PROVINSI_FILE = os.path.join(DOCS_DIR, 'provinsi.csv')
KOTA_KAB_FILE = os.path.join(DOCS_DIR, 'kota_kab.csv')
KECAMATAN_FILE = os.path.join(DOCS_DIR, 'kecamatan.csv')
DESA_FILE = os.path.join(DOCS_DIR, 'desa.csv')

# ========================================
# VALUE MAPPINGS
# ========================================

STATUS_POSKO_MAP = {
    'Operasional': 'operasional',
    'Non-aktif': 'non_aktif',
    'Evakuasi': 'evakuasi',
    'Persiapan Huntara': 'persiapan_huntara',
}

JENIS_PENGUNGSIAN_MAP = {
    'Lahan pemerintah': 'fasum',
    'Lahan swasta/perorangan': 'swasta',
    'Fasilitas Umum': 'fasum',
    'Bangunan Swasta': 'swasta',
    'Bangunan Pribadi': 'pribadi',
    'Fasilitas Ibadah': 'ibadah',
}

KETERSEDIAAN_AIR_MAP = {
    'Ya': 'cukup',
    'Tidak': 'tidak_ada',
    'Cukup': 'cukup',
    'Terbatas': 'terbatas',
    'Tidak Ada': 'tidak_ada',
}

YN_MAP = {
    'Ya': 'ya',
    'Tidak': 'tidak',
    'yes': 'ya',
    'no': 'tidak',
}

# ========================================
# HELPER FUNCTIONS
# ========================================

def normalize_name(name: str) -> Optional[str]:
    """Normalize location name for lookup"""
    if pd.isna(name):
        return None
    return str(name).upper().strip()


def load_wilayah_lookups():
    """Load wilayah CSV files and create lookup dictionaries"""
    provinsi_df = pd.read_csv(PROVINSI_FILE)
    kota_kab_df = pd.read_csv(KOTA_KAB_FILE)
    kecamatan_df = pd.read_csv(KECAMATAN_FILE)
    desa_df = pd.read_csv(DESA_FILE)

    # Provinsi lookup: nama -> kode
    provinsi_lookup = {normalize_name(row['nama']): str(row['kode'])
                       for _, row in provinsi_df.iterrows()}

    # Kota/Kab lookup: nama -> kode
    kota_lookup = {normalize_name(row['nama']): row['kode']
                   for _, row in kota_kab_df.iterrows()}

    # Kecamatan lookup: (id_kota_kab, nama) -> kode
    kec_lookup = {}
    for _, row in kecamatan_df.iterrows():
        key = (row['id_kota_kab'], normalize_name(row['nama']))
        kec_lookup[key] = row['kode']

    # Desa lookup: (id_kec, nama) -> kode
    # id_kec is float like 11.18.05
    desa_lookup = {}
    for _, row in desa_df.iterrows():
        key = (row['id_kec'], normalize_name(row['nama']))
        desa_lookup[key] = row['kode']

    return {
        'provinsi': provinsi_lookup,
        'kota_kab': kota_lookup,
        'kecamatan': kec_lookup,
        'desa': desa_lookup,
        'kecamatan_df': kecamatan_df,
        'desa_df': desa_df,
    }


def lookup_location(row: pd.Series, lookups: dict) -> dict:
    """Lookup location codes from names"""
    result = {
        'provinsi': None,
        'kota_kab': None,
        'kecamatan': None,
        'desa': None,
    }

    # Provinsi
    prov_name = normalize_name(row.get('sel_provinsi'))
    if prov_name:
        result['provinsi'] = lookups['provinsi'].get(prov_name)

    # Kota/Kab
    kota_name = normalize_name(row.get('sel_kota_kab'))
    if kota_name:
        result['kota_kab'] = lookups['kota_kab'].get(kota_name)

    # Kecamatan - need to lookup by kota_kab + name
    kec_name = normalize_name(row.get('sel_kecamatan'))
    if kec_name and result['kota_kab']:
        # Try with float kota_kab code
        try:
            kota_float = float(result['kota_kab'])
            result['kecamatan'] = lookups['kecamatan'].get((kota_float, kec_name))
        except:
            pass

    # Desa - need to lookup by kecamatan + name
    desa_name = normalize_name(row.get('sel_desa'))
    if desa_name and result['kecamatan']:
        # id_kec format is like "11.18.05"
        try:
            # Convert kecamatan code to float for lookup
            kec_code = result['kecamatan']
            # Try to find matching desa
            desa_df = lookups['desa_df']
            matches = desa_df[
                (desa_df['id_kec'] == kec_code) &
                (desa_df['nama'].str.upper() == desa_name)
            ]
            if len(matches) > 0:
                result['desa'] = matches.iloc[0]['kode']
        except:
            pass

    return result


def safe_int(value, default=0) -> int:
    """Safely convert to int"""
    if pd.isna(value):
        return default
    try:
        return int(float(value))
    except:
        return default


def safe_str(value, default='') -> str:
    """Safely convert to string"""
    if pd.isna(value):
        return default
    return str(value).strip()


def format_date(value) -> str:
    """Format date to ISO format"""
    if pd.isna(value):
        return ''
    try:
        if isinstance(value, datetime):
            return value.strftime('%Y-%m-%d')
        return str(value)[:10]
    except:
        return ''


def format_geopoint(lat, lon) -> str:
    """Format lat/lon to geopoint string"""
    if pd.isna(lat) or pd.isna(lon):
        return ''
    return f"{lat} {lon} 0 0"


def map_value(value, mapping: dict, default='') -> str:
    """Map value using lookup dictionary"""
    if pd.isna(value):
        return default
    str_val = str(value).strip()
    return mapping.get(str_val, default)


# ========================================
# DATA TRANSFORMATION
# ========================================

def transform_row(row: pd.Series, lookups: dict, row_idx: int) -> Dict[str, Any]:
    """Transform a single row from Excel to ODK submission format"""

    # Lookup location codes
    locations = lookup_location(row, lookups)

    # Generate unique UUIDs - same UUID for both instance and entity
    # This ensures the entity_id matches the submission for proper linking
    entity_uuid = str(uuid.uuid4())
    instance_id = f"uuid:{entity_uuid}"

    # Get posko name for entity label
    nama_posko = safe_str(row.get('nama_posko'))

    # Build submission data
    submission = {
        # Meta - includes entity metadata for ODK Central to create entity
        'meta': {
            'instanceID': instance_id,
            # Entity block - this triggers entity creation in ODK Central
            'entity': {
                'dataset': ENTITY_DATASET,
                'id': entity_uuid,  # Entity UUID (without 'uuid:' prefix)
                'create': '1',  # Force entity creation
                'label': nama_posko,  # Entity label (calc_nama_posko equivalent)
            }
        },

        # Store entity_id for reference in calc fields
        '_entity_uuid': entity_uuid,

        # Location selection (using codes)
        'sel_provinsi': locations['provinsi'] or '11',  # Default to Aceh
        'sel_kota_kab': locations['kota_kab'] or '',
        'sel_kecamatan': locations['kecamatan'] or '',
        'sel_desa': locations['desa'] or '',

        # Mode - always 'baru' for new submissions
        'mode': 'baru',

        # Basic info
        'nama_posko': safe_str(row.get('nama_posko')),

        # Identitas group
        'grp_identitas': {
            'status_posko': map_value(row.get('grp_identitas-status_posko'), STATUS_POSKO_MAP, 'operasional'),
            'nama_penanggungjawab': safe_str(row.get('grp_identitas-nama_penanggungjawab')),
            'contact_penanggungjawab': safe_str(row.get('grp_identitas-contact_penanggungjawab')),
            'institusi': safe_str(row.get('grp_identitas-institusi')),
            'nama_relawan': safe_str(row.get('grp_identitas-nama_relawan')),
            'contact_relawan': safe_str(row.get('grp_identitas-contact_relawan')),
            'alamat_dusun': safe_str(row.get('grp_identitas-alamat_dusun')),
            'mulai_tanggal': format_date(row.get('grp_identitas-mulai_tanggal')),
            'kota_terdekat': safe_str(row.get('grp_identitas-kota_terdekat')),
            'koordinat': format_geopoint(
                row.get('grp_identitas-koordinat-Latitude'),
                row.get('grp_identitas-koordinat-Longitude')
            ),
        },

        # Akses group
        'grp_akses': {
            'terisolir': '',
            'akses_via': '',
            'jarak_pkm': safe_int(row.get('grp_akses-jarak_pkm')) if pd.notna(row.get('grp_akses-jarak_pkm')) else '',
            'jarak_posko_logistik': '',
            'nama_faskes_terdekat': safe_str(row.get('grp_akses-nama_faskes_terdekat')),
        },

        # Pengungsian group
        'grp_pengungsian': {
            'jenis_pengungsian': map_value(row.get('grp_pengungsian-jenis_pengungsian'), JENIS_PENGUNGSIAN_MAP, 'fasum'),
            'detail_pengungsian': safe_str(row.get('grp_pengungsian-detail_pengungsian')),
            'total_pengungsi': safe_int(row.get('grp_pengungsian-total_pengungsi')),
            'persen_keterlibatan': safe_str(row.get('grp_pengungsian-persen_keterlibatan')),
        },

        # Demografi group
        'grp_demografi': {
            'jumlah_kk': safe_int(row.get('grp_data_pengungsi-jumlah_kk')),
            'kk_perempuan': safe_int(row.get('grp_data_pengungsi-kk_perempuan')),
            'kk_anak': safe_int(row.get('grp_data_pengungsi-kk_anak')),
            'dewasa_perempuan': safe_int(row.get('grp_data_pengungsi-dewasa_perempuan')),
            'dewasa_laki': safe_int(row.get('grp_data_pengungsi-dewasa_laki')),
            'remaja_perempuan': safe_int(row.get('grp_data_pengungsi-remaja_perempuan')),
            'remaja_laki': safe_int(row.get('grp_data_pengungsi-remaja_laki')),
            'anak_perempuan': safe_int(row.get('grp_data_pengungsi-anak_perempuan')),
            'anak_laki': safe_int(row.get('grp_data_pengungsi-anak_laki')),
            'balita_perempuan': safe_int(row.get('grp_data_pengungsi-balita_perempuan')),
            'balita_laki': safe_int(row.get('grp_data_pengungsi-balita_laki')),
            'bayi_perempuan': safe_int(row.get('grp_data_pengungsi-bayi_perempuan')),
            'bayi_laki': safe_int(row.get('grp_data_pengungsi-bayi_laki')),
            'lansia': safe_int(row.get('grp_data_pengungsi-lansia')),
            'ibu_menyusui': safe_int(row.get('grp_data_pengungsi-ibu_menyusui')),
            'ibu_hamil': safe_int(row.get('grp_data_pengungsi-ibu_hamil')),
            'anak_tanpa_ortu': safe_int(row.get('grp_data_pengungsi-anak_tanpa_ortu')),
            'bayi_tanpa_ibu': safe_int(row.get('grp_data_pengungsi-bayi_tanpa_ibu')),
            'difabel': safe_int(row.get('grp_data_pengungsi-difabel')),
            'komorbid': safe_int(row.get('grp_data_pengungsi-komorbid')),
        },

        # Fasilitas group
        'grp_fasilitas': {
            'dapur_umum': '',
            'ketersediaan_air': map_value(row.get('grp_fasilitas-ketersediaan_air'), KETERSEDIAAN_AIR_MAP, ''),
            'saluran_limbah': map_value(row.get('grp_fasilitas-note_sanitasi'), YN_MAP, ''),
            'sumber_air': '',
            'toilet_perempuan': '',
            'toilet_laki': '',
            'toilet_campur': '',
            'tempat_sampah': '',
            'sumber_listrik': '',
            'kondisi_penerangan': '',
            # Health facilities
            'posko_kesehatan': '',
            'posko_obat': '',
            'posko_psikososial': '',
            'ruang_laktasi': '',
            'layanan_lansia': '',
            'layanan_keluarga': '',
            # Education
            'sekolah_darurat': '',
            'program_pengganti': '',
            # Security
            'petugas_keamanan': '',
            # Area
            'area_interaksi': '',
            'area_bermain': '',
        },

        # Komunikasi group
        'grp_komunikasi': {
            'ketersediaan_sinyal': '',
            'jaringan_orari': '',
            'ketersediaan_internet': '',
        },

        # Baseline/Sumber data
        'grp_baseline': {
            'baseline_sumber': 'BNPB',
        },
    }

    return submission


# ========================================
# ODK CENTRAL API
# ========================================

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

    def create_submission(self, project_id: str, form_id: str, submission_data: dict) -> dict:
        """Create a new submission"""
        url = f"{self.base_url}/v1/projects/{project_id}/forms/{form_id}/submissions"

        # ODK Central expects XML, but also supports JSON for some endpoints
        # We'll use the JSON submission endpoint
        headers = {
            'Content-Type': 'application/json',
        }

        response = self.session.post(url, json=submission_data, headers=headers)

        if response.status_code in [200, 201]:
            return {'success': True, 'data': response.json()}
        else:
            return {'success': False, 'error': response.text, 'status': response.status_code}


def build_xml_submission(submission_data: dict, form_id: str) -> str:
    """Build XML submission for ODK Central with entity creation support"""

    def escape_xml(value) -> str:
        """Escape XML special characters"""
        str_val = str(value)
        str_val = str_val.replace('&', '&amp;')
        str_val = str_val.replace('<', '&lt;')
        str_val = str_val.replace('>', '&gt;')
        str_val = str_val.replace('"', '&quot;')
        return str_val

    def dict_to_xml(d: dict, indent: int = 2, skip_keys: set = None) -> str:
        if skip_keys is None:
            skip_keys = set()
        lines = []
        prefix = ' ' * indent
        for key, value in d.items():
            if key in skip_keys:
                continue
            if isinstance(value, dict):
                lines.append(f"{prefix}<{key}>")
                lines.append(dict_to_xml(value, indent + 2))
                lines.append(f"{prefix}</{key}>")
            elif value is not None and value != '':
                lines.append(f"{prefix}<{key}>{escape_xml(value)}</{key}>")
        return '\n'.join(lines)

    meta = submission_data.get('meta', {})
    instance_id = meta.get('instanceID', f'uuid:{uuid.uuid4()}')
    entity_data = meta.get('entity', {})

    # Build entity XML block if entity data exists
    entity_xml = ''
    if entity_data:
        entity_id = entity_data.get('id', '')
        entity_dataset = entity_data.get('dataset', ENTITY_DATASET)
        entity_create = entity_data.get('create', '1')
        entity_label = escape_xml(entity_data.get('label', ''))

        # Entity element with proper attributes for ODK Central
        # Format: <entity dataset="name" id="uuid" create="1"><label>...</label></entity>
        entity_xml = f'''
    <entity dataset="{entity_dataset}" id="{entity_id}" create="{entity_create}">
      <label>{entity_label}</label>
    </entity>'''

    # Skip internal keys when converting to XML
    skip_keys = {'meta', '_entity_uuid'}

    xml = f'''<?xml version="1.0" encoding="UTF-8"?>
<data id="{form_id}" version="2025011002"
      xmlns:entities="http://www.opendatakit.org/xforms/entities"
      entities:entities-version="2024.1.0">
  <meta>
    <instanceID>{instance_id}</instanceID>{entity_xml}
  </meta>
{dict_to_xml({k: v for k, v in submission_data.items()}, 2, skip_keys)}
</data>'''

    return xml


# ========================================
# MAIN
# ========================================

def main():
    parser = argparse.ArgumentParser(description='Dump posko data to ODK Central')
    parser.add_argument('--dry-run', action='store_true', help='Print submissions without sending')
    parser.add_argument('--limit', type=int, default=0, help='Limit number of records (0 = all)')
    parser.add_argument('--start', type=int, default=0, help='Start from record N')
    parser.add_argument('--output-json', type=str, help='Output transformed data to JSON file')
    parser.add_argument('--output-xml', type=str, help='Output sample XML to file')
    args = parser.parse_args()

    print("=" * 60)
    print("POSKO DATA DUMP TO ODK CENTRAL")
    print("=" * 60)

    # Check credentials
    if not args.dry_run and (not ODK_EMAIL or not ODK_PASSWORD):
        print("\nERROR: ODK credentials not set!")
        print("Set environment variables:")
        print("  export ODK_EMAIL='your-email'")
        print("  export ODK_PASSWORD='your-password'")
        print("\nOr use --dry-run to test without sending")
        sys.exit(1)

    # Load data
    print(f"\nLoading data from: {DATA_FILE}")
    df = pd.read_excel(DATA_FILE)
    print(f"Total records: {len(df)}")

    # Load wilayah lookups
    print("\nLoading wilayah lookups...")
    lookups = load_wilayah_lookups()
    print(f"  Provinsi: {len(lookups['provinsi'])} entries")
    print(f"  Kota/Kab: {len(lookups['kota_kab'])} entries")
    print(f"  Kecamatan: {len(lookups['kecamatan'])} entries")
    print(f"  Desa: {len(lookups['desa'])} entries")

    # Determine range
    start_idx = args.start
    end_idx = len(df) if args.limit == 0 else min(start_idx + args.limit, len(df))

    print(f"\nProcessing records {start_idx} to {end_idx - 1}")

    # Transform data
    submissions = []
    for idx in range(start_idx, end_idx):
        row = df.iloc[idx]
        submission = transform_row(row, lookups, idx)
        submissions.append(submission)

        if args.dry_run and idx < start_idx + 3:
            entity_info = submission.get('meta', {}).get('entity', {})
            print(f"\n--- Record {idx}: {submission.get('nama_posko')} ---")
            print(f"Entity ID: {entity_info.get('id')}")
            print(f"Entity Dataset: {entity_info.get('dataset')}")
            print(f"Entity Label: {entity_info.get('label')}")
            print(f"Location: {submission.get('sel_provinsi')}/{submission.get('sel_kota_kab')}/{submission.get('sel_kecamatan')}/{submission.get('sel_desa')}")
            print(f"Status: {submission.get('grp_identitas', {}).get('status_posko')}")
            print(f"Total Pengungsi: {submission.get('grp_pengungsian', {}).get('total_pengungsi')}")
            print(f"Sumber Data: {submission.get('grp_baseline', {}).get('baseline_sumber')}")

    # Output to JSON if requested
    if args.output_json:
        with open(args.output_json, 'w') as f:
            json.dump(submissions, f, indent=2, default=str)
        print(f"\nSaved {len(submissions)} submissions to {args.output_json}")

    # Output sample XML if requested
    if args.output_xml and len(submissions) > 0:
        xml = build_xml_submission(submissions[0], ODK_FORM_ID)
        with open(args.output_xml, 'w') as f:
            f.write(xml)
        print(f"\nSaved sample XML to {args.output_xml}")

    # Dry run stops here
    if args.dry_run:
        print(f"\n[DRY RUN] Would submit {len(submissions)} records")
        print("\nTo submit for real, remove --dry-run flag")
        return

    # Connect to ODK Central
    print(f"\nConnecting to ODK Central: {ODK_BASE_URL}")
    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Submit records
    success_count = 0
    error_count = 0

    for idx, submission in enumerate(submissions):
        # Build XML
        xml = build_xml_submission(submission, ODK_FORM_ID)

        # Submit
        url = f"{ODK_BASE_URL}/v1/projects/{ODK_PROJECT_ID}/forms/{ODK_FORM_ID}/submissions"
        headers = {
            'Content-Type': 'application/xml',
        }

        response = client.session.post(url, data=xml.encode('utf-8'), headers=headers)

        if response.status_code in [200, 201]:
            success_count += 1
            print(f"✓ [{idx + start_idx}] {submission.get('nama_posko')}")
        else:
            error_count += 1
            print(f"✗ [{idx + start_idx}] {submission.get('nama_posko')}: {response.status_code} - {response.text[:100]}")

    # Summary
    print("\n" + "=" * 60)
    print(f"SUMMARY: {success_count} success, {error_count} errors")
    print("=" * 60)


if __name__ == '__main__':
    main()
