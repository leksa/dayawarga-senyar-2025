---
sidebar_position: 1
---

# Locations API

Locations API menyediakan endpoints untuk mengelola data posko pengungsi.

## Endpoints

- **GET** `/locations` - Mendapatkan daftar semua lokasi (posko)
- **GET** `/locations/:id` - Mendapatkan detail lokasi spesifik

## Data Model

### Location Object

| Field | Type | Description |
|--------|-------|-------------|
| id | string | Unique identifier (UUID) |
| nama | string | Nama posko |
| type | string | Tipe lokasi (posko, faskes, jembatan) |
| status | string | Status lokasi (active, inactive) |
| longitude | float | Koordinat longitude |
| latitude | float | Koordinat latitude |
| alamat_singkat | string | Alamat ringkas |
| nama_provinsi | string | Nama provinsi |
| nama_kota_kab | string | Nama kota/kabupaten |
| nama_kecamatan | string | Nama kecamatan |
| nama_desa | string | Nama desa/kelurahan |
| jumlah_kk | integer | Jumlah kepala keluarga |
| total_jiwa | integer | Total jumlah jiwa |
| jumlah_perempuan | integer | Jumlah perempuan |
| jumlah_laki | integer | Jumlah laki-laki |
| jumlah_balita | integer | Jumlah balita |
| kebutuhan_air | string | Ketersediaan air |
| baseline_sumber | string | Sumber data baseline |
| updated_at | datetime | Timestamp update terakhir |

## Example Response

```json
{
  "success": true,
  "data": {
    "type": "FeatureCollection",
    "features": [
      {
        "type": "Feature",
        "id": "uuid-string",
        "geometry": {
          "type": "Point",
          "coordinates": [97.1234, 5.5678]
        },
        "properties": {
          "odk_submission_id": "submission-uuid",
          "nama": "Posko Blok V Blang Kandis",
          "type": "posko",
          "status": "active",
          "alamat_singkat": "Blang Kandis, Kabupaten Aceh Timur",
          "nama_provinsi": "Aceh",
          "nama_kota_kab": "Kabupaten Aceh Timur",
          "nama_kecamatan": "Blang Kandis",
          "nama_desa": "Blang Kandis",
          "jumlah_kk": 150,
          "total_jiwa": 750,
          "jumlah_perempuan": 380,
          "jumlah_laki": 370,
          "jumlah_balita": 120,
          "kebutuhan_air": "Tersedia",
          "baseline_sumber": "ODK",
          "updated_at": "2026-01-26T14:00:00Z"
        }
      }
    ]
  },
  "meta": {
    "total": 199,
    "page": 1,
    "limit": 50,
    "timestamp": "2026-01-26T14:00:00Z"
  }
}