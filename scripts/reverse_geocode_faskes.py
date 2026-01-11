#!/usr/bin/env python3
"""
Reverse geocode faskes coordinates to get kecamatan and desa names.
Uses Nominatim (OpenStreetMap) API with rate limiting.
"""

import csv
import time
import requests
import re
from pathlib import Path

# Input/output files
INPUT_FILE = Path(__file__).parent.parent / "docs" / "Faskes_Kemenkes_Aceh.csv"
OUTPUT_FILE = INPUT_FILE  # Overwrite same file with added columns

# Nominatim API settings
NOMINATIM_URL = "https://nominatim.openstreetmap.org/reverse"
HEADERS = {
    "User-Agent": "DayawargaSenyar/1.0 (disaster response mapping)"
}
RATE_LIMIT_DELAY = 1.1  # seconds between requests (Nominatim requires 1 req/sec max)


def parse_coordinate(coord_str):
    """Parse coordinate string that might have comma as decimal separator or be malformed."""
    if not coord_str:
        return None

    # Remove any whitespace
    coord_str = coord_str.strip()

    # Check if it's a normal decimal number
    try:
        val = float(coord_str)
        # Sanity check for Indonesian coordinates
        if -15 < val < 150:  # Reasonable range for Indonesia
            return val
    except ValueError:
        pass

    # Try to fix comma-separated numbers (European format or malformed)
    # Pattern like "9,742,379" should be "97.42379"
    if ',' in coord_str:
        parts = coord_str.split(',')
        if len(parts) == 3:
            # Format like "9,742,379" -> "97.42379"
            try:
                integer_part = parts[0] + parts[1][:1]  # First digit of second part
                decimal_part = parts[1][1:] + parts[2]
                val = float(f"{integer_part}.{decimal_part}")
                if -15 < val < 150:
                    return val
            except (ValueError, IndexError):
                pass
        elif len(parts) == 2:
            # Might be comma as decimal separator
            try:
                val = float(coord_str.replace(',', '.'))
                if -15 < val < 150:
                    return val
            except ValueError:
                pass

    return None


def reverse_geocode(lat, lon):
    """Reverse geocode coordinates using Nominatim API."""
    params = {
        "lat": lat,
        "lon": lon,
        "format": "json",
        "addressdetails": 1,
        "zoom": 18,  # High zoom for detailed address
    }

    try:
        response = requests.get(NOMINATIM_URL, params=params, headers=HEADERS, timeout=10)
        response.raise_for_status()
        data = response.json()

        address = data.get("address", {})

        # Try to extract kecamatan and desa from address components
        # OSM uses different field names: village, suburb, district, subdistrict
        kecamatan = (
            address.get("subdistrict") or
            address.get("district") or
            address.get("city_district") or
            ""
        )

        desa = (
            address.get("village") or
            address.get("suburb") or
            address.get("neighbourhood") or
            address.get("hamlet") or
            ""
        )

        return {
            "kecamatan": kecamatan,
            "desa": desa,
            "display_name": data.get("display_name", ""),
            "raw_address": address,
        }

    except Exception as e:
        print(f"  Error geocoding ({lat}, {lon}): {e}")
        return {"kecamatan": "", "desa": "", "display_name": "", "raw_address": {}}


def main():
    print(f"Reading input file: {INPUT_FILE}")

    rows = []
    with open(INPUT_FILE, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        rows = list(reader)

    print(f"Found {len(rows)} faskes records")

    # Process each row
    results = []
    for i, row in enumerate(rows):
        print(f"\n[{i+1}/{len(rows)}] Processing: {row.get('NAMA', 'Unknown')}")

        # Parse coordinates
        lon = parse_coordinate(row.get("LONGITUDE", ""))
        lat = parse_coordinate(row.get("LATITUDE", ""))

        if lon is None or lat is None:
            print(f"  WARNING: Invalid coordinates - lon={row.get('LONGITUDE')}, lat={row.get('LATITUDE')}")
            row["KECAMATAN"] = ""
            row["DESA"] = ""
            row["GEOCODE_STATUS"] = "invalid_coords"
        else:
            print(f"  Coordinates: {lat}, {lon}")

            # Reverse geocode
            geo_result = reverse_geocode(lat, lon)

            row["KECAMATAN"] = geo_result["kecamatan"]
            row["DESA"] = geo_result["desa"]
            row["GEOCODE_STATUS"] = "ok" if geo_result["kecamatan"] or geo_result["desa"] else "no_data"

            print(f"  Result: Kec. {geo_result['kecamatan']}, Desa {geo_result['desa']}")

            # Rate limiting
            time.sleep(RATE_LIMIT_DELAY)

        results.append(row)

    # Write output
    print(f"\n\nWriting output file: {OUTPUT_FILE}")

    fieldnames = list(rows[0].keys()) + ["KECAMATAN", "DESA", "GEOCODE_STATUS"]
    # Remove duplicates while preserving order
    seen = set()
    fieldnames = [x for x in fieldnames if not (x in seen or seen.add(x))]

    with open(OUTPUT_FILE, 'w', encoding='utf-8', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(results)

    # Summary
    ok_count = sum(1 for r in results if r.get("GEOCODE_STATUS") == "ok")
    no_data_count = sum(1 for r in results if r.get("GEOCODE_STATUS") == "no_data")
    invalid_count = sum(1 for r in results if r.get("GEOCODE_STATUS") == "invalid_coords")

    print(f"\n=== Summary ===")
    print(f"Total: {len(results)}")
    print(f"  OK (got kecamatan/desa): {ok_count}")
    print(f"  No data from geocoder: {no_data_count}")
    print(f"  Invalid coordinates: {invalid_count}")
    print(f"\nOutput saved to: {OUTPUT_FILE}")


if __name__ == "__main__":
    main()
