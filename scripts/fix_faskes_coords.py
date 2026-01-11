#!/usr/bin/env python3
"""
Fix coordinate format in Faskes_Kemenkes_Aceh.csv
- LONGITUDE: 2 digits before decimal (e.g., 97.42379)
- LATITUDE: 1 digit before decimal (e.g., 4.41929)
"""

import csv
import re
from pathlib import Path

INPUT_FILE = Path(__file__).parent.parent / "docs" / "Faskes_Kemenkes_Aceh.csv"
OUTPUT_FILE = INPUT_FILE  # Overwrite the same file


def fix_coordinate(coord_str, num_integer_digits):
    """
    Fix coordinate string format.

    Args:
        coord_str: The coordinate string (might be malformed like "9,742,379")
        num_integer_digits: Number of digits before decimal (2 for LONG, 1 for LAT)

    Returns:
        Properly formatted coordinate string (e.g., "97.42379")
    """
    if not coord_str:
        return coord_str

    coord_str = coord_str.strip()

    # If it's already a valid decimal number, check if it needs fixing
    try:
        val = float(coord_str)
        # Check if it's in reasonable range for Indonesia
        if num_integer_digits == 2 and 90 < val < 110:  # Longitude
            return coord_str
        if num_integer_digits == 1 and -10 < val < 10:  # Latitude
            return coord_str
    except ValueError:
        pass

    # Remove all non-digit characters to get raw digits
    digits_only = re.sub(r'[^\d]', '', coord_str)

    if not digits_only:
        return coord_str

    # Insert decimal point after num_integer_digits
    if len(digits_only) > num_integer_digits:
        integer_part = digits_only[:num_integer_digits]
        decimal_part = digits_only[num_integer_digits:]
        result = f"{integer_part}.{decimal_part}"
    else:
        result = digits_only

    try:
        # Validate the result
        val = float(result)
        return result
    except ValueError:
        return coord_str


def main():
    print(f"Reading: {INPUT_FILE}")

    rows = []
    with open(INPUT_FILE, 'r', encoding='utf-8') as f:
        reader = csv.DictReader(f)
        fieldnames = reader.fieldnames
        rows = list(reader)

    print(f"Found {len(rows)} records")

    # Fix coordinates
    fixed_count = 0
    for row in rows:
        orig_lon = row.get('LONGITUDE', '')
        orig_lat = row.get('LATITUDE', '')

        new_lon = fix_coordinate(orig_lon, 2)  # 2 digits before decimal
        new_lat = fix_coordinate(orig_lat, 1)  # 1 digit before decimal

        if new_lon != orig_lon or new_lat != orig_lat:
            fixed_count += 1
            print(f"  Fixed: {row.get('NAMA', 'Unknown')}")
            print(f"    LON: {orig_lon} -> {new_lon}")
            print(f"    LAT: {orig_lat} -> {new_lat}")

        row['LONGITUDE'] = new_lon
        row['LATITUDE'] = new_lat

    # Write back to the same file
    print(f"\nWriting to: {OUTPUT_FILE}")
    with open(OUTPUT_FILE, 'w', encoding='utf-8', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(rows)

    print(f"\nDone! Fixed {fixed_count} records.")


if __name__ == "__main__":
    main()
