#!/usr/bin/env python3
"""
Update ODK Entity Properties from PostgreSQL

This script updates ODK Central entity properties with data from PostgreSQL.
It matches entities by UUID (_entity_id in PostgreSQL = uuid in ODK).
Syncs ALL properties that exist in the entity dataset.
"""

import os
import sys
import json
import subprocess
import requests
from urllib3.exceptions import InsecureRequestWarning
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

# ODK Central configuration
ODK_BASE_URL = os.getenv("ODK_BASE_URL", os.getenv("ODK_CENTRAL_URL", "https://data.dayawarga.com"))
ODK_EMAIL = os.getenv("ODK_EMAIL", "")
ODK_PASSWORD = os.getenv("ODK_PASSWORD", "")
ODK_PROJECT_ID = os.getenv("ODK_PROJECT_ID", "3")
DATASET_NAME = "posko_entities"

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

def get_entity_version(session, entity_uuid):
    """Get current version of entity"""
    response = session.get(
        f"{ODK_BASE_URL}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities/{entity_uuid}"
    )
    if response.status_code == 200:
        return response.json().get("currentVersion", {}).get("version", 1)
    return 1

def update_entity(session, entity_uuid, label, data, base_version):
    """Update entity properties via PATCH"""
    url = f"{ODK_BASE_URL}/v1/projects/{ODK_PROJECT_ID}/datasets/{DATASET_NAME}/entities/{entity_uuid}"

    payload = {
        "label": label,
        "data": data
    }

    headers = {
        "Content-Type": "application/json"
    }

    params = {"baseVersion": base_version}

    response = session.patch(url, json=payload, headers=headers, params=params)

    if response.status_code in [200, 201]:
        return True, None
    else:
        return False, f"Status {response.status_code}: {response.text[:200]}"

def get_pg_data_via_docker():
    """Get ALL location data from PostgreSQL via Docker - includes all entity properties"""
    # Use jsonb concatenation to avoid 100 argument limit in json_build_object
    query = """
    SELECT (
        jsonb_build_object(
            'entity_id', raw_data->>'_entity_id',
            'nama_posko', nama,
            'status_posko', status,
            'id_desa', alamat->>'id_desa',
            'nama_desa', alamat->>'nama_desa',
            'nama_kecamatan', alamat->>'nama_kecamatan',
            'nama_kota_kab', alamat->>'nama_kota_kab',
            'nama_provinsi', alamat->>'nama_provinsi',
            'nama_penanggungjawab', identitas->>'nama_penanggungjawab',
            'contact_penanggungjawab', identitas->>'contact_penanggungjawab',
            'nama_relawan', identitas->>'nama_relawan',
            'contact_relawan', identitas->>'contact_relawan',
            'alamat_dusun', identitas->>'alamat_dusun',
            'institusi', identitas->>'institusi',
            'mulai_tanggal', identitas->>'mulai_tanggal',
            'kota_terdekat', identitas->>'kota_terdekat',
            'baseline_sumber', identitas->>'baseline_sumber'
        ) ||
        jsonb_build_object(
            'total_pengungsi', COALESCE(data_pengungsi->>'total_pengungsi', data_pengungsi->>'total_jiwa'),
            'jumlah_kk', data_pengungsi->>'jumlah_kk',
            'kk_perempuan', data_pengungsi->>'kk_perempuan',
            'kk_anak', data_pengungsi->>'kk_anak',
            'jenis_pengungsian', data_pengungsi->>'jenis_pengungsian',
            'detail_pengungsian', data_pengungsi->>'detail_pengungsian',
            'persen_keterlibatan', data_pengungsi->>'persen_keterlibatan',
            'dewasa_perempuan', data_pengungsi->>'dewasa_perempuan',
            'dewasa_laki', data_pengungsi->>'dewasa_laki',
            'remaja_perempuan', data_pengungsi->>'remaja_perempuan',
            'remaja_laki', data_pengungsi->>'remaja_laki',
            'anak_perempuan', data_pengungsi->>'anak_perempuan',
            'anak_laki', data_pengungsi->>'anak_laki',
            'balita_perempuan', data_pengungsi->>'balita_perempuan',
            'balita_laki', data_pengungsi->>'balita_laki',
            'bayi_perempuan', data_pengungsi->>'bayi_perempuan',
            'bayi_laki', data_pengungsi->>'bayi_laki',
            'lansia', data_pengungsi->>'lansia',
            'ibu_menyusui', data_pengungsi->>'ibu_menyusui',
            'ibu_hamil', data_pengungsi->>'ibu_hamil',
            'remaja_tanpa_ortu', data_pengungsi->>'remaja_tanpa_ortu',
            'anak_tanpa_ortu', data_pengungsi->>'anak_tanpa_ortu',
            'bayi_tanpa_ibu', data_pengungsi->>'bayi_tanpa_ibu',
            'difabel', data_pengungsi->>'difabel',
            'komorbid', data_pengungsi->>'komorbid'
        ) ||
        jsonb_build_object(
            'posko_logistik', fasilitas->>'posko_logistik',
            'posko_faskes', fasilitas->>'posko_faskes',
            'dapur_umum', fasilitas->>'dapur_umum',
            'kapasitas_dapur', fasilitas->>'kapasitas_dapur',
            'ketersediaan_air', fasilitas->>'ketersediaan_air',
            'kebutuhan_air', fasilitas->>'kebutuhan_air',
            'saluran_limbah', fasilitas->>'saluran_limbah',
            'sumber_air', fasilitas->>'sumber_air',
            'toilet_perempuan', fasilitas->>'toilet_perempuan',
            'toilet_laki', fasilitas->>'toilet_laki',
            'toilet_campur', fasilitas->>'toilet_campur',
            'tempat_sampah', fasilitas->>'tempat_sampah',
            'sumber_listrik', fasilitas->>'sumber_listrik',
            'kondisi_penerangan', fasilitas->>'kondisi_penerangan',
            'titik_akses_listrik', fasilitas->>'titik_akses_listrik',
            'posko_kesehatan', fasilitas->>'posko_kesehatan',
            'posko_tenaga_medis', fasilitas->>'posko_tenaga_medis',
            'posko_obat', fasilitas->>'posko_obat',
            'posko_psikososial', fasilitas->>'posko_psikososial',
            'ruang_laktasi', fasilitas->>'ruang_laktasi',
            'layanan_lansia', fasilitas->>'layanan_lansia',
            'layanan_keluarga', fasilitas->>'layanan_keluarga',
            'sekolah_darurat', fasilitas->>'sekolah_darurat',
            'program_pengganti', fasilitas->>'program_pengganti',
            'petugas_keamanan', fasilitas->>'petugas_keamanan',
            'area_interaksi', fasilitas->>'area_interaksi',
            'area_bermain', fasilitas->>'area_bermain'
        ) ||
        jsonb_build_object(
            'ketersediaan_sinyal', komunikasi->>'ketersediaan_sinyal',
            'jaringan_orari', komunikasi->>'jaringan_orari',
            'ketersediaan_internet', komunikasi->>'ketersediaan_internet',
            'jarak_pkm', akses->>'jarak_pkm',
            'jarak_posko_logistik', akses->>'jarak_posko_logistik',
            'nama_faskes_terdekat', akses->>'nama_faskes_terdekat',
            'terisolir', akses->>'terisolir',
            'akses_via', akses->>'akses_via',
            'geometry', CASE WHEN geom IS NOT NULL
                            THEN CONCAT(ST_Y(geom)::text, ' ', ST_X(geom)::text, ' 0 0')
                            ELSE ''
                       END
        )
    )::text as json_data
    FROM locations
    WHERE deleted_at IS NULL
      AND raw_data->>'_entity_id' IS NOT NULL
    """

    result = subprocess.run([
        'docker', 'exec', 'senyar-postgres', 'psql', '-U', 'senyar', '-d', 'senyar',
        '-t', '-A', '-c', query
    ], capture_output=True, text=True)

    if result.returncode != 0:
        print(f"Error querying PostgreSQL: {result.stderr}")
        return {}

    data = {}
    for line in result.stdout.strip().split('\n'):
        if not line:
            continue
        try:
            row = json.loads(line)
            entity_id = row.pop('entity_id', None)
            if entity_id:
                # Convert all values to strings, replace None with empty string
                clean_data = {}
                for k, v in row.items():
                    if v is None:
                        clean_data[k] = ''
                    else:
                        clean_data[k] = str(v)
                data[entity_id] = clean_data
        except json.JSONDecodeError as e:
            print(f"   Warning: Failed to parse JSON: {e}")
            continue

    return data

def main():
    print("=" * 60)
    print("Update ODK Entity Properties from PostgreSQL")
    print("Syncing ALL properties")
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

    # Get PostgreSQL data
    print("\n3. Fetching PostgreSQL data...")
    try:
        pg_data = get_pg_data_via_docker()
        print(f"   ✓ Found {len(pg_data)} locations")
    except Exception as e:
        print(f"   ✗ Failed: {e}")
        sys.exit(1)

    # Update entities
    print("\n4. Updating entities...")
    updated = 0
    skipped = 0
    errors = 0
    no_pg_data = 0

    for entity in odk_entities:
        entity_uuid = entity.get('uuid')
        current_data = entity.get('data') or {}

        # Get data from PostgreSQL
        if entity_uuid not in pg_data:
            no_pg_data += 1
            continue

        loc_data = pg_data[entity_uuid]
        label = loc_data.get('nama_posko', '')

        if not label:
            no_pg_data += 1
            continue

        # Check if there are any missing/empty fields in ODK that PG has data for
        needs_update = False
        for key, pg_value in loc_data.items():
            odk_value = current_data.get(key, '')
            # Update if ODK is empty but PG has data
            if not odk_value and pg_value:
                needs_update = True
                break

        # Also update if label is different
        if entity.get('label') != label:
            needs_update = True

        if not needs_update:
            skipped += 1
            continue

        # Get current version
        base_version = get_entity_version(session, entity_uuid)

        # Update entity with all data from PostgreSQL
        success, error = update_entity(session, entity_uuid, label, loc_data, base_version)

        if success:
            updated += 1
            print(f"   ✓ Updated: {label}")
        else:
            errors += 1
            print(f"   ✗ Error updating {entity_uuid}: {error}")

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"  Updated:          {updated}")
    print(f"  Already complete: {skipped}")
    print(f"  No PG data:       {no_pg_data}")
    print(f"  Errors:           {errors}")
    print("=" * 60)

if __name__ == "__main__":
    main()
