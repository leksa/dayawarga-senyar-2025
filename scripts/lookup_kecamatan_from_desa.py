#!/usr/bin/env python3
"""
Lookup and fix Kecamatan & Desa names using backbone CSV data.

Strategy:
1. Read DESA from reverse geocoding result
2. Lookup desa name in desa.csv using fuzzy matching
3. Fix DESA name to match official backbone name
4. Lookup id_kec in kecamatan.csv to get kecamatan name
5. Update both KECAMATAN and DESA columns in faskes CSV
"""

import csv
import re
from pathlib import Path

# Paths
DOCS_DIR = Path(__file__).parent.parent / "docs"
FASKES_CSV = DOCS_DIR / "Faskes_Kemenkes_Aceh.csv"
DESA_CSV = DOCS_DIR / "desa.csv"
KECAMATAN_CSV = DOCS_DIR / "kecamatan.csv"


def normalize_name(name):
    """Normalize name for matching - lowercase, remove extra spaces, common variations"""
    if not name:
        return ""
    name = name.lower().strip()
    # Remove common prefixes
    name = re.sub(r'^(desa|gampong|kelurahan|kp\.?|ds\.?)\s+', '', name)
    # Normalize spaces
    name = re.sub(r'\s+', ' ', name)
    # Remove punctuation
    name = re.sub(r'[^\w\s]', '', name)
    return name


def normalize_for_fuzzy(name):
    """More aggressive normalization for fuzzy matching - remove all spaces"""
    if not name:
        return ""
    name = normalize_name(name)
    # Remove all spaces for compound name matching
    name = name.replace(' ', '')
    # Common letter substitutions in Indonesian
    name = name.replace('oe', 'u')  # old spelling
    name = name.replace('dj', 'j')  # old spelling
    name = name.replace('tj', 'c')  # old spelling
    return name


def load_desa_lookup():
    """Load desa.csv and create lookups: normalized_nama -> (id_kec, original_nama)"""
    lookup = {}
    fuzzy_lookup = {}
    with open(DESA_CSV, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            nama = row.get('nama', '')
            id_kec = row.get('id_kec', '')
            if nama and id_kec:
                info = {'id_kec': id_kec, 'nama': nama}
                # Exact normalized lookup
                key = normalize_name(nama)
                lookup[key] = info
                # Fuzzy lookup (no spaces)
                fuzzy_key = normalize_for_fuzzy(nama)
                fuzzy_lookup[fuzzy_key] = info
    return lookup, fuzzy_lookup


def load_kecamatan_lookup():
    """Load kecamatan.csv and create lookup: kode -> nama"""
    lookup = {}
    with open(KECAMATAN_CSV, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        for row in reader:
            kode = row.get('kode', '')
            nama = row.get('nama', '')
            if kode and nama:
                lookup[kode] = nama
    return lookup


def main():
    print("=" * 60)
    print("Lookup Kecamatan from Desa Name")
    print("=" * 60)

    # Load lookups
    print("\n1. Loading lookup tables...")
    desa_lookup, desa_fuzzy_lookup = load_desa_lookup()
    kecamatan_lookup = load_kecamatan_lookup()
    print(f"   Desa entries: {len(desa_lookup)} (exact), {len(desa_fuzzy_lookup)} (fuzzy)")
    print(f"   Kecamatan entries: {len(kecamatan_lookup)}")

    # Read faskes CSV
    print(f"\n2. Reading faskes CSV: {FASKES_CSV}")
    rows = []
    with open(FASKES_CSV, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        rows = list(reader)
    print(f"   Found {len(rows)} faskes")

    # Process each row
    print("\n3. Looking up and fixing kecamatan & desa names...")
    stats = {
        'no_desa': 0,
        'desa_not_found': 0,
        'kec_not_found': 0,
        'desa_fixed': 0,
        'kec_updated': 0
    }

    not_found_desas = []

    for row in rows:
        desa = row.get('DESA', '').strip()

        # Skip if no desa from geocoding
        if not desa:
            stats['no_desa'] += 1
            continue

        # Normalize and lookup - try exact first, then fuzzy
        desa_norm = normalize_name(desa)
        desa_info = desa_lookup.get(desa_norm)

        if not desa_info:
            # Try fuzzy match (no spaces)
            desa_fuzzy = normalize_for_fuzzy(desa)
            desa_info = desa_fuzzy_lookup.get(desa_fuzzy)

        if not desa_info:
            stats['desa_not_found'] += 1
            not_found_desas.append(desa)
            continue

        # Fix desa name to official backbone name
        official_desa = desa_info['nama']
        if row['DESA'] != official_desa:
            row['DESA'] = official_desa
            stats['desa_fixed'] += 1

        # Get kecamatan name from id_kec
        id_kec = desa_info['id_kec']
        kecamatan_nama = kecamatan_lookup.get(id_kec)

        if not kecamatan_nama:
            stats['kec_not_found'] += 1
            continue

        # Update kecamatan
        if row.get('KECAMATAN', '') != kecamatan_nama:
            row['KECAMATAN'] = kecamatan_nama
            stats['kec_updated'] += 1

    # Write back
    print(f"\n4. Writing updated CSV...")
    fieldnames = list(rows[0].keys())
    with open(FASKES_CSV, 'w', encoding='utf-8', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    # Summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"  No Desa data:           {stats['no_desa']}")
    print(f"  Desa not found in CSV:  {stats['desa_not_found']}")
    print(f"  Kecamatan not found:    {stats['kec_not_found']}")
    print(f"  Desa names fixed:       {stats['desa_fixed']}")
    print(f"  Kecamatan updated:      {stats['kec_updated']}")
    print("=" * 60)

    # Show sample of not found desas
    if not_found_desas:
        print(f"\nSample desa names not found in backbone ({len(not_found_desas)} total):")
        unique_desas = list(set(not_found_desas))[:20]
        for d in unique_desas:
            print(f"  - {d}")


if __name__ == "__main__":
    main()
