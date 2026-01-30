---
---

# Get Location by ID

Mendapatkan detail lengkap dari lokasi spesifik berdasarkan ID (UUID).

## Endpoint

```http
GET /api/v1/locations/:id
```

## Path Parameters

| Parameter | Type | Required | Description |
|-----------|-------|-----------|-------------|
| id | string (UUID) | Yes | Unique identifier lokasi |

## Response

Returns detailed information about a specific location including photos.

### Success (200 OK)

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "odk_submission_id": "submission-uuid",
    "type": "posko",
    "status": "active",
    "baseline_sumber": "ODK",
    "geometry": {
      "type": "Point",
      "coordinates": [97.1234, 5.5678],
      "altitude": 50.5,
      "accuracy": 10.2
    },
    "identitas": {
      "nama": "Posko Blok V Blang Kandis",
      "kontak": "081234567890"
    },
    "alamat": {
      "nama_provinsi": "Aceh",
      "nama_kota_kab": "Kabupaten Aceh Timur",
      "nama_kecamatan": "Blang Kandis",
      "nama_desa": "Blang Kandis"
    },
    "data_pengungsi": {
      "jumlah_kk": 150,
      "total_jiwa": 750,
      "dewasa_perempuan": 280,
      "dewasa_laki": 270,
      "anak_perempuan": 80,
      "anak_laki": 85,
      "balita_perempuan": 20,
      "balita_laki": 15
    },
    "fasilitas": {
      "ketersediaan_air": "Tersedia",
      "kebutuhan_air": 5000,
      "mck": "Ada",
      "dapur_umum": "Ada"
    },
    "komunikasi": {
      "sinyal_seluler": "Tersedia",
      "jaringan_internet": "Tidak ada"
    },
    "akses": {
      "kendaraan_roda_4": "Bisa",
      "kendaraan_roda_2": "Bisa"
    },
    "photos": [
      {
        "type": "foto_utama",
        "filename": "photo1.jpg",
        "url": "/api/v1/photos/photo-id/file"
      }
    ],
    "meta": {
      "submitted_at": "2026-01-26T10:00:00Z",
      "updated_at": "2026-01-26T14:00:00Z",
      "submitter_name": "Ahmad Fauzi"
    }
  }
}
```

## Example Requests

### Get location by ID

```bash
curl http://localhost:8080/api/v1/locations/550e8400-e29b-41d4-a716-446655440000
```

## Error Responses

### Not Found (404)

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Location not found"
  }
}
```

### Invalid ID Format (400)

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid location ID format"
  }
}
```

## Notes

- ID harus valid UUID format (36 characters: 8-4-4-4-12)
- Data JSONB (identitas, alamat, data_pengungsi, dll) dikembalikan sebagai object JSON
- Photos array berisi URL untuk mendapatkan file foto
- Endpoint ini di-cache selama 5 menit