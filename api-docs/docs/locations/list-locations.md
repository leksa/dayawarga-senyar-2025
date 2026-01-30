---
---

# Get All Locations

Mendapatkan daftar semua lokasi (posko pengungsi) dengan dukungan pagination dan filtering.

## Endpoint

```http
GET /api/v1/locations
```

## Query Parameters

| Parameter | Type | Required | Default | Description |
|-----------|-------|-----------|-------------|
| page | integer | No | 1 | Nomor halaman (min: 1) |
| limit | integer | No | 50 | Jumlah item per halaman (min: 1, max: 200) |
| type | string | No | - | Filter berdasarkan tipe lokasi (posko, faskes, jembatan) |
| status | string | No | - | Filter berdasarkan status (active, inactive) |
| search | string | No | - | Pencarian berdasarkan nama lokasi |
| bbox | string | No | - | Filter berdasarkan bounding box (format: minLng,minLat,maxLng,maxLat) |

## Response

Returns a GeoJSON FeatureCollection of locations.

### Success (200 OK)

```json
{
  "success": true,
  "data": {
    "type": "FeatureCollection",
    "features": [
      {
        "type": "Feature",
        "id": "550e8400-e29b-41d4-a716-446655440000",
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
```

## Example Requests

### Get first page (default)

```bash
curl http://localhost:8080/api/v1/locations
```

### Get page 2 with 20 items

```bash
curl http://localhost:8080/api/v1/locations?page=2&limit=20
```

### Filter by type

```bash
curl http://localhost:8080/api/v1/locations?type=posko
```

### Filter by status

```bash
curl http://localhost:8080/api/v1/locations?status=active
```

### Search by name

```bash
curl http://localhost:8080/api/v1/locations?search=Blang
```

### Filter by bounding box

```bash
curl http://localhost:8080/api/v1/locations?bbox=97.0,5.5,98.0,6.0
```

## Notes

- Result dikembalikan dalam format GeoJSON untuk kemudahan integrasi dengan peta
- Endpoint ini di-cache selama 5 menit
- Gunakan parameter `?nocache=true` untuk mem-bypass cache
- Default limit adalah 50, maximum limit adalah 200