#!/usr/bin/env python3
"""
Sync Entity UUIDs between ODK Central and PostgreSQL

This script updates PostgreSQL _entity_id to match ODK Entity UUIDs.
It matches records by nama_posko (location name) since UUIDs are different.

This ensures that when users do mode="update" via ODK Collect,
the sync service can correctly match and update existing records.
"""

import os
import sys
import requests
import psycopg2
from psycopg2.extras import RealDictCursor
from urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# ODK Central configuration
ODK_BASE_URL = os.getenv("ODK_BASE_URL", os.getenv("ODK_CENTRAL_URL", "https://data.dayawarga.com"))
ODK_EMAIL = os.getenv("ODK_EMAIL", "")
ODK_PASSWORD = os.getenv("ODK_PASSWORD", "")
ODK_PROJECT_ID = os.getenv("ODK_PROJECT_ID", "3")
DATASET_NAME = "posko_entities"

# PostgreSQL configuration (uses DB_* from .env)
PG_HOST = os.getenv("DB_HOST", os.getenv("PG_HOST", "localhost"))
PG_PORT = os.getenv("DB_PORT", os.getenv("PG_PORT", "5432"))
PG_DATABASE = os.getenv("DB_NAME", os.getenv("PG_DATABASE", "senyar"))
PG_USER = os.getenv("DB_USER", os.getenv("PG_USER", "senyar"))
PG_PASSWORD = os.getenv("DB_PASSWORD", os.getenv("PG_PASSWORD", ""))

def get_odk_session():
    """Get ODK Central session token"""
    session = requests.Session()
    session.verify = False

    response = session.post(
        f"{ODK_BASE_URL}/v1/sessions",
        json={"email": ODK_EMAIL, "password": ODK_PASSWORD}
    )
    response.raise_for_status()
    token = response.json()["token"]
    session.headers.update({"Authorization": f"Bearer {token}"})
    return session

def get_odk_entities(session):
    """Get all entities from ODK Central"""
    response = session.get(
        f"{ODK_BASE_URL}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities"
    )
    response.raise_for_status()
    return response.json()

def get_pg_connection():
    """Get PostgreSQL connection"""
    return psycopg2.connect(
        host=PG_HOST,
        port=PG_PORT,
        database=PG_DATABASE,
        user=PG_USER,
        password=PG_PASSWORD
    )

def get_pg_locations():
    """Get all locations from PostgreSQL"""
    conn = get_pg_connection()
    cursor = conn.cursor(cursor_factory=RealDictCursor)

    cursor.execute("""
        SELECT id, nama, raw_data->>'_entity_id' as current_entity_id
        FROM locations
        WHERE deleted_at IS NULL
    """)

    locations = cursor.fetchall()
    cursor.close()
    conn.close()

    return locations

def update_pg_entity_id(location_id, new_entity_id):
    """Update _entity_id in PostgreSQL"""
    conn = get_pg_connection()
    cursor = conn.cursor()

    cursor.execute("""
        UPDATE locations
        SET raw_data = raw_data || %s::jsonb,
            updated_at = NOW()
        WHERE id = %s
    """, (f'{{"_entity_id": "{new_entity_id}"}}', str(location_id)))

    conn.commit()
    cursor.close()
    conn.close()

def main():
    print("=" * 60)
    print("Syncing Entity UUIDs: ODK Central → PostgreSQL")
    print("=" * 60)

    # Get ODK session
    print("\n1. Connecting to ODK Central...")
    try:
        session = get_odk_session()
        print("   ✓ Connected")
    except Exception as e:
        print(f"   ✗ Failed: {e}")
        sys.exit(1)

    # Get ODK entities
    print("\n2. Fetching ODK entities...")
    try:
        odk_entities = get_odk_entities(session)
        print(f"   ✓ Found {len(odk_entities)} entities")
    except Exception as e:
        print(f"   ✗ Failed: {e}")
        sys.exit(1)

    # Build lookup by label (nama_posko)
    odk_by_label = {}
    for entity in odk_entities:
        label = entity.get("label", "")
        if label:
            odk_by_label[label] = entity["uuid"]

    print(f"   → {len(odk_by_label)} entities with labels")

    # Get PostgreSQL locations
    print("\n3. Fetching PostgreSQL locations...")
    try:
        pg_locations = get_pg_locations()
        print(f"   ✓ Found {len(pg_locations)} locations")
    except Exception as e:
        print(f"   ✗ Failed: {e}")
        sys.exit(1)

    # Match and update
    print("\n4. Matching and updating UUIDs...")
    updated = 0
    already_match = 0
    no_match = []

    for loc in pg_locations:
        nama = loc["nama"]
        current_id = loc["current_entity_id"]

        if nama in odk_by_label:
            odk_uuid = odk_by_label[nama]

            if current_id == odk_uuid:
                already_match += 1
            else:
                print(f"   → Updating '{nama}'")
                print(f"      Old: {current_id}")
                print(f"      New: {odk_uuid}")
                try:
                    update_pg_entity_id(loc["id"], odk_uuid)
                    updated += 1
                except Exception as e:
                    print(f"      ✗ Error: {e}")
        else:
            no_match.append(nama)

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"  Updated:          {updated}")
    print(f"  Already matching: {already_match}")
    print(f"  No ODK match:     {len(no_match)}")

    if no_match:
        print("\n  Locations without ODK entity match:")
        for nama in no_match[:10]:
            print(f"    - {nama}")
        if len(no_match) > 10:
            print(f"    ... and {len(no_match) - 10} more")

    print("=" * 60)

    if updated > 0:
        print("\n✓ PostgreSQL _entity_id values now match ODK Entity UUIDs")
        print("  Future mode='update' submissions will correctly update existing records")
    elif already_match == len(pg_locations):
        print("\n✓ All UUIDs already match - no updates needed")

if __name__ == "__main__":
    main()
