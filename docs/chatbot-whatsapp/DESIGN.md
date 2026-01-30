# Dayawarga Chatbot WhatsApp — Technical Design

## Overview

Chatbot WhatsApp untuk relawan bencana Dayawarga. Menerima laporan informasi (feed), pendaftaran posko baru, dan update posko via percakapan natural language di WhatsApp.

## Architecture

```
┌──────────┐         ┌────────────────────────────────────┐
│  Relawan │  WA     │     CHATBOT SERVICE (VPS Baru)     │
│  (Phone) │◄───────►│                                    │
└──────────┘         │  Node.js + TypeScript (tsx)        │
                     │  ├── Express (HTTP server)         │
                     │  ├── In-memory Map (sessions)      │
                     │  ├── SQLite (cache + relawan)      │
                     │  ├── Claude Haiku (LLM)            │
                     │  ├── Jaro-Winkler (fuzzy match)    │
                     │  └── Watzap.id (WA gateway)        │
                     └────────────────┬───────────────────┘
                                      │ REST API
                                      │ (API Key auth)
                                      ▼
                     ┌────────────────────────────────────┐
                     │     DAYAWARGA API (Existing VPS)    │
                     │  PostgreSQL:                        │
                     │  ├── locations (posko)              │
                     │  ├── information_feeds              │
                     │  ├── faskes                         │
                     │  └── wilayah (provinsi→desa, BPS)   │
                     └────────────────────────────────────┘
```

## Stack

| Component | Technology | Justification |
|-----------|-----------|---------------|
| Runtime | Node.js 20 LTS + TypeScript | AI agent effectiveness, ecosystem |
| Runner | tsx (no build step) | Fast iteration |
| HTTP | Express | Simple, 3 routes only |
| Session | In-memory Map + TTL 30min | 100 msg/day, no Redis needed |
| Local DB | better-sqlite3 | Cache geocoding, relawan whitelist |
| LLM | Claude Haiku (Anthropic SDK) | Cost-effective, fast, accurate |
| WA API | Watzap.id | Provider lokal, webhook-based |
| Process | pm2 | Auto-restart, logging |
| Matching | Jaro-Winkler (custom) | Fuzzy desa name matching |

## Constraints

- **100 messages/day** — no need for Redis, BullMQ, or heavy infra
- **3-day deadline** — minimal viable, ship fast
- **1 developer + 2 AI agents** — prioritize simplicity
- **Single VPS** — pm2, no Docker required for MVP

## Data Flow

### 1. Feed Submission
```
User WA → "Saya mau lapor informasi"
Bot → "Apa informasi yang ingin dilaporkan?"
User → "Banjir di desa Sukamaju, air setinggi 1 meter"
Bot → "Di kabupaten/kota mana?"
User → "Kab Bogor"
Bot → "Di kecamatan/desa mana? (Atau kirim lokasi)"
User → shares location / types "sukamju cibinong"
Bot → fuzzy match → "Desa SUKAMAJU, Kec. Cibinong. Benar? (ya/tidak)"
User → "ya"
Bot → POST /api/feeds → "Laporan berhasil dikirim!"
```

### 2. Posko Baru
```
User → "Daftarkan posko baru"
Bot → "Apa nama posko?"
User → "Posko Masjid Al-Ikhlas"
Bot → "Di kabupaten/kota mana?"
... (collect: nama, lokasi, kapasitas, kontak, fasilitas)
Bot → POST /api/locations → "Posko berhasil didaftarkan!"
```

### 3. Posko Update
```
User → "Update posko"
Bot → GET /api/locations?search=... → "Posko mana yang ingin diupdate?"
User → pilih dari list
Bot → "Apa yang ingin diupdate? (pengungsi/fasilitas/status)"
... (collect update data)
Bot → PUT /api/locations/:id → "Posko berhasil diupdate!"
```

## Wilayah Matching Strategy

```
User sebut desa
    │
    ▼
┌─ Tanya kabupaten/kota dulu ─────────────┐
│  → Narrow 83K desa ke ~200-500          │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Code: Fuzzy Match (Jaro-Winkler) ─────┐
│  Normalize → score vs candidate list    │
│  > 0.88 → konfirmasi ke user           │
│  0.65-0.88 → LLM kasih pilihan         │
│  < 0.65 → minta ketik ulang            │
└──────────────────────┬──────────────────┘
                       ▼
┌─ LLM: Konfirmasi / Tanya Balik ────────┐
│  "Maksud Anda SUKAMAJU, Kec Cibinong?" │
│  Selalu konfirmasi sebelum submit       │
└─────────────────────────────────────────┘
```

## Session Management

- Key: phone number
- Value: `{ state, formType, collectedData, lastActivity, messageHistory }`
- TTL: 30 minutes inactivity → auto-reset
- Cleanup: setInterval every 5 minutes

## Error Handling

- LLM gagal parse → fallback menu: "Pilih: 1. Feed, 2. Posko Baru, 3. Posko Update"
- Dayawarga API error → "Maaf, terjadi gangguan. Coba lagi nanti."
- 3x gagal berturut-turut → reset conversation, mulai dari awal
- Watzap webhook retry → idempotency check via message_id

## Security

- Relawan whitelist: hanya nomor terdaftar yang bisa submit
- API Key: header `X-API-Key` untuk komunikasi chatbot ↔ Dayawarga API
- Rate limit: max 10 msg/menit per user (anti-spam)

## Watzap.id API Reference

**Base URL**: `https://api.watzap.id/v1`

| Action | Endpoint | Body |
|--------|----------|------|
| Send Text | POST `/send_message` | `{api_key, number_key, phone_no, message}` |
| Send Image | POST `/send_image_url` | `{api_key, number_key, phone_no, url, message}` |
| Set Webhook | POST `/set_webhook` | `{api_key, number_key, endpoint_url}` |

**Webhook Incoming Format:**
```json
{
  "type": "incoming_chat",
  "data": {
    "chat_id": "xxx",
    "message_id": "xxx",
    "name": "User",
    "from": "628xxxx",
    "message": "text content",
    "media_url": "URL if media",
    "type": "text|image|location",
    "latitude": "...",
    "longitude": "..."
  }
}
```

## Database Schema (Chatbot - SQLite)

```sql
-- Relawan whitelist
CREATE TABLE wa_relawan (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  phone VARCHAR(20) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  organization VARCHAR(255),
  is_active BOOLEAN DEFAULT 1,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Conversation logs (audit trail)
CREATE TABLE wa_conversations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  phone VARCHAR(20) NOT NULL,
  message_id VARCHAR(255) UNIQUE,
  direction VARCHAR(10) NOT NULL, -- 'incoming' | 'outgoing'
  message TEXT,
  metadata JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Wilayah cache (fetched from Dayawarga API)
CREATE TABLE wilayah_cache (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kabupaten_id VARCHAR(20),
  data JSON NOT NULL, -- cached desa list
  fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Dayawarga API Endpoints (Required)

Chatbot butuh endpoint ini di Dayawarga API:

| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/wilayah?level=kabupaten` | List semua kabupaten |
| GET | `/api/wilayah?parent_id=X&level=desa` | List desa dalam kabupaten |
| GET | `/api/locations?search=X` | Cari posko by nama |
| POST | `/api/locations` | Buat posko baru |
| PUT | `/api/locations/:id` | Update posko |
| POST | `/api/feeds` | Submit feed/informasi |
| GET | `/api/faskes?kabupaten_id=X` | List faskes |

**Auth**: Header `X-API-Key: <shared-secret>`
