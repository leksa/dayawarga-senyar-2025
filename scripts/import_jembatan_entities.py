#!/usr/bin/env python3
"""
Script to import Jalan/Jembatan entities from GeoJSON to ODK Central
Usage: python import_jembatan_entities.py [--dry-run]
"""

import os
import sys
import json
import uuid
import argparse
import requests
from pathlib import Path
from dotenv import load_dotenv

# Load environment variables
load_dotenv(Path(__file__).parent.parent / ".env")

# Configuration
ODK_BASE_URL = os.getenv("ODK_BASE_URL", "https://data.dayawarga.com")
ODK_PROJECT_ID = int(os.getenv("ODK_PROJECT_ID", "3"))
ODK_EMAIL = os.getenv("ODK_EMAIL", "")
ODK_PASSWORD = os.getenv("ODK_PASSWORD", "")

# Entity list name (must match the dataset name in ODK Central)
ENTITY_LIST_NAME = "jembatan_entities"

# GeoJSON source file
GEOJSON_FILE = Path(__file__).parent.parent / "docs/Jalan_Jembatan_Putus_Akses_Map_Service_PU.geojson"


class ODKCentralClient:
    def __init__(self, base_url: str, email: str, password: str):
        self.base_url = base_url.rstrip('/')
        self.email = email
        self.password = password
        self.session = requests.Session()
        self.token = None

    def authenticate(self) -> bool:
        """Authenticate with ODK Central"""
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

    def get_datasets(self, project_id: int) -> list:
        """Get all datasets (entity lists) in a project"""
        url = f"{self.base_url}/v1/projects/{project_id}/datasets"
        response = self.session.get(url)
        if response.status_code == 200:
            return response.json()
        return []

    def create_dataset(self, project_id: int, name: str) -> dict:
        """Create a new dataset (entity list)"""
        url = f"{self.base_url}/v1/projects/{project_id}/datasets"
        response = self.session.post(url, json={'name': name})
        if response.status_code in [200, 201]:
            return {'success': True, 'data': response.json()}
        return {'success': False, 'error': response.text, 'status': response.status_code}

    def get_dataset_entities(self, project_id: int, dataset_name: str) -> list:
        """Get all entities in a dataset"""
        url = f"{self.base_url}/v1/projects/{project_id}/datasets/{dataset_name}/entities"
        response = self.session.get(url)
        if response.status_code == 200:
            return response.json()
        return []

    def create_entity(self, project_id: int, dataset_name: str, entity_data: dict) -> dict:
        """Create a new entity in a dataset"""
        url = f"{self.base_url}/v1/projects/{project_id}/datasets/{dataset_name}/entities"

        # Format: https://docs.getodk.org/central-api-entity-management/#creating-entities
        payload = {
            "uuid": entity_data.get('uuid', str(uuid.uuid4())),
            "label": entity_data.get('label', ''),
            "data": entity_data.get('data', {})
        }

        response = self.session.post(url, json=payload)
        if response.status_code in [200, 201]:
            return {'success': True, 'data': response.json()}
        return {'success': False, 'error': response.text, 'status': response.status_code}

    def add_dataset_property(self, project_id: int, dataset_name: str, property_name: str) -> dict:
        """Add a property to a dataset schema"""
        url = f"{self.base_url}/v1/projects/{project_id}/datasets/{dataset_name}/properties"
        response = self.session.post(url, json={'name': property_name})
        if response.status_code in [200, 201]:
            return {'success': True}
        # 409 means property already exists, which is fine
        if response.status_code == 409:
            return {'success': True, 'exists': True}
        return {'success': False, 'error': response.text, 'status': response.status_code}


def load_geojson(filepath: Path) -> list:
    """Load and filter GeoJSON for Aceh province only"""
    with open(filepath, 'r') as f:
        data = json.load(f)

    # Filter Aceh only
    aceh_features = [f for f in data['features'] if 'Aceh' in f['properties'].get('provinsi', '')]
    return aceh_features


def feature_to_entity(feature: dict) -> dict:
    """Convert GeoJSON feature to ODK entity format"""
    props = feature['properties']
    coords = feature['geometry']['coordinates']

    entity_uuid = str(uuid.uuid4())

    return {
        'uuid': entity_uuid,
        'label': props.get('nama', ''),
        'data': {
            'nama': entity_uuid,  # Entity ID (was 'name')
            'objectid': str(props.get('objectid', '')),
            'fid2': str(props.get('fid2', '')),
            'nama2': props.get('nama', ''),  # Display name
            'jenis': props.get('jenis', ''),
            'statusjln': props.get('statusjln', ''),
            'nama_provinsi': props.get('provinsi', ''),
            'nama_kabupaten': props.get('kabupaten', ''),
            'keterangan': props.get('keterangan', ''),
            'dampak': props.get('dampak', ''),
            'status1': props.get('status1', ''),
            'penanganan': props.get('penanganan', ''),
            'target_selesai': props.get('target', ''),
            'bailey': props.get('bailey', '').strip(),
            'latitude': str(coords[1]),
            'longitude': str(coords[0]),
            'baseline_sumber': 'BNPB/PU',
        }
    }


def main():
    parser = argparse.ArgumentParser(description='Import Jalan/Jembatan entities to ODK Central')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be imported without doing it')
    parser.add_argument('--limit', type=int, default=0, help='Limit number of imports (0 = all)')
    args = parser.parse_args()

    print("=" * 60)
    print("ODK CENTRAL - IMPORT JEMBATAN ENTITIES")
    print("=" * 60)

    # Check credentials
    if not ODK_EMAIL or not ODK_PASSWORD:
        print("\nERROR: ODK credentials not set!")
        print("Set environment variables in .env file:")
        print("  ODK_EMAIL=your-email")
        print("  ODK_PASSWORD=your-password")
        sys.exit(1)

    # Load GeoJSON
    print(f"\nLoading GeoJSON: {GEOJSON_FILE}")
    if not GEOJSON_FILE.exists():
        print(f"ERROR: File not found: {GEOJSON_FILE}")
        sys.exit(1)

    features = load_geojson(GEOJSON_FILE)
    print(f"Found {len(features)} Aceh features")

    # Connect to ODK Central
    print(f"\nConnecting to: {ODK_BASE_URL}")
    print(f"Project ID: {ODK_PROJECT_ID}")

    client = ODKCentralClient(ODK_BASE_URL, ODK_EMAIL, ODK_PASSWORD)

    if not client.authenticate():
        print("Failed to authenticate!")
        sys.exit(1)

    print("Authenticated successfully!")

    # Check if dataset exists
    print(f"\nChecking dataset: {ENTITY_LIST_NAME}")
    datasets = client.get_datasets(ODK_PROJECT_ID)
    dataset_names = [d['name'] for d in datasets]

    if ENTITY_LIST_NAME not in dataset_names:
        print(f"Dataset '{ENTITY_LIST_NAME}' not found. Creating...")
        result = client.create_dataset(ODK_PROJECT_ID, ENTITY_LIST_NAME)
        if not result.get('success'):
            print(f"Failed to create dataset: {result.get('error')}")
            sys.exit(1)
        print("Dataset created successfully!")
    else:
        print(f"Dataset '{ENTITY_LIST_NAME}' exists")

    # Define required properties
    required_properties = [
        'nama', 'objectid', 'fid2', 'nama2', 'jenis', 'statusjln',
        'nama_provinsi', 'nama_kabupaten', 'keterangan', 'dampak',
        'status1', 'penanganan', 'target_selesai', 'bailey',
        'latitude', 'longitude', 'baseline_sumber'
    ]

    # Add properties to dataset schema
    print("\nEnsuring dataset properties...")
    for prop in required_properties:
        result = client.add_dataset_property(ODK_PROJECT_ID, ENTITY_LIST_NAME, prop)
        if result.get('success'):
            if result.get('exists'):
                print(f"  ✓ {prop} (exists)")
            else:
                print(f"  + {prop} (added)")
        else:
            print(f"  ✗ {prop} - {result.get('error')}")

    # Check existing entities
    print("\nChecking existing entities...")
    existing_entities = client.get_dataset_entities(ODK_PROJECT_ID, ENTITY_LIST_NAME)
    existing_labels = {e.get('currentVersion', {}).get('label', '') for e in existing_entities}
    print(f"Existing entities: {len(existing_entities)}")

    # Prepare entities to import
    entities_to_import = []
    for feature in features:
        entity = feature_to_entity(feature)
        # Skip if already exists (by label/name)
        if entity['label'] in existing_labels:
            continue
        entities_to_import.append(entity)

    print(f"New entities to import: {len(entities_to_import)}")

    if len(entities_to_import) == 0:
        print("\nNo new entities to import!")
        return

    # Apply limit
    if args.limit > 0:
        entities_to_import = entities_to_import[:args.limit]
        print(f"Limited to: {len(entities_to_import)} entities")

    # Dry run
    if args.dry_run:
        print("\n[DRY RUN] Would import:")
        for i, entity in enumerate(entities_to_import[:10]):
            print(f"  {i+1}. {entity['label']}")
        if len(entities_to_import) > 10:
            print(f"  ... and {len(entities_to_import) - 10} more")
        print("\nTo import for real, remove --dry-run flag")
        return

    # Import entities
    print("\nImporting entities...")
    success_count = 0
    error_count = 0

    for i, entity in enumerate(entities_to_import):
        result = client.create_entity(ODK_PROJECT_ID, ENTITY_LIST_NAME, entity)

        if result.get('success'):
            success_count += 1
            print(f"✓ [{i+1}/{len(entities_to_import)}] {entity['label'][:50]}...")
        else:
            error_count += 1
            print(f"✗ [{i+1}/{len(entities_to_import)}] {entity['label'][:50]}... - {result.get('error', 'Unknown')[:50]}")

    # Summary
    print("\n" + "=" * 60)
    print(f"SUMMARY: {success_count} imported, {error_count} errors")
    print("=" * 60)


if __name__ == '__main__':
    main()
