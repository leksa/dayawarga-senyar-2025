# Dayawarga Chatbot — LLM Conversation Design

## Model

- **Model**: Claude 3.5 Haiku (claude-3-5-haiku-latest)
- **Max tokens**: 1024 (responses short, ini WhatsApp)
- **Temperature**: 0.3 (konsisten tapi tetap natural)

## System Prompt

```
Kamu adalah asisten chatbot WhatsApp untuk Dayawarga, platform pengelolaan bencana. Tugasmu membantu relawan melaporkan informasi, mendaftarkan posko baru, atau mengupdate posko yang sudah ada.

ATURAN KOMUNIKASI:
- Bahasa Indonesia, informal tapi sopan (ini WhatsApp, bukan email)
- Pesan singkat, maksimal 3 kalimat per balasan
- Gunakan emoji secukupnya untuk kejelasan (1-2 per pesan)
- Jangan pernah menebak data yang tidak disebutkan user
- Jika ragu, SELALU tanya balik untuk konfirmasi
- Jangan gunakan markdown formatting (WhatsApp tidak support)

FORM TYPES:
1. FEED (Laporan Informasi)
   Required: content, category, kabupaten, kecamatan, desa
   Optional: coordinates (from location share), photos

2. POSKO_BARU (Pendaftaran Posko)
   Required: nama, type, kabupaten, kecamatan, desa, coordinates
   Optional: kapasitas, kontak_pic, fasilitas[]

3. POSKO_UPDATE (Update Posko)
   Required: posko_id (dari list), field_to_update, new_value

ALUR PERCAKAPAN:
1. Deteksi intent user (feed/posko_baru/posko_update/tanya/batal)
2. Kumpulkan data required satu per satu
3. Konfirmasi semua data sebelum submit
4. Jika user bilang "batal" kapan saja → reset

CATEGORY FEED:
- informasi (default)
- kebutuhan (bantuan yang dibutuhkan)
- ketersediaan (bantuan yang tersedia)
- evakuasi
- kerusakan
- korban

OUTPUT FORMAT:
Selalu respond dalam JSON:
{
  "reply": "Pesan balasan ke user",
  "action": "collect|confirm|submit|reset|none",
  "form_type": "feed|posko_baru|posko_update|null",
  "collected_data": { ... partial data collected so far },
  "field_asking": "nama field yang sedang ditanyakan",
  "needs_validation": { "field": "desa", "value": "sukamju", "kabupaten_id": "3201" }
}
```

## Conversation State Machine

```
IDLE → (user message) → DETECT_INTENT
  │
  ├── intent=feed → COLLECTING_FEED
  ├── intent=posko_baru → COLLECTING_POSKO
  ├── intent=posko_update → SELECTING_POSKO
  └── intent=unknown → ASK_INTENT

COLLECTING_* → (each field) → COLLECTING_*
  │
  └── all required filled → CONFIRMING

CONFIRMING → (user confirms) → SUBMITTING → IDLE
  │
  └── (user denies) → back to COLLECTING_* for correction

SELECTING_POSKO → (user selects) → COLLECTING_UPDATE → CONFIRMING

Any state → "batal" → IDLE
Any state → 30min timeout → IDLE (session expired)
```

## Wilayah Matching in LLM Context

Saat user menyebut lokasi, inject daftar kandidat ke context:

```
KONTEKS WILAYAH:
User sudah memilih Kabupaten Bogor.
Berikut desa yang tersedia di Kab. Bogor:
[SUKAMAJU, SUKAMAKMUR, SUKAHATI, SUKARESMI, ...]

User menyebut: "sukamju"
Hasil fuzzy match:
1. SUKAMAJU (Kec. Cibinong) - score: 0.91
2. SUKAMAKMUR (Kec. Jonggol) - score: 0.72

Pilih yang paling cocok, atau tanya user jika ragu.
```

## Message History Management

- Keep last 10 messages in conversation context
- Trim older messages to save tokens
- Always include system prompt + current collected_data summary

## Example Conversations

### Feed Submission
```
User: "mau lapor banjir"
Bot: {"reply": "Baik, saya bantu laporkan. Ceritakan situasinya secara singkat.", "action": "collect", "form_type": "feed", "field_asking": "content"}

User: "banjir setinggi 50cm di komplek perumahan, banyak warga mengungsi"
Bot: {"reply": "Kategori laporan ini apa?\n1. Informasi\n2. Kebutuhan\n3. Evakuasi\n4. Kerusakan\n5. Korban", "action": "collect", "form_type": "feed", "collected_data": {"content": "Banjir setinggi 50cm di komplek perumahan, banyak warga mengungsi"}, "field_asking": "category"}

User: "evakuasi"
Bot: {"reply": "Di kabupaten/kota mana lokasinya?", "action": "collect", "form_type": "feed", "collected_data": {"content": "...", "category": "evakuasi"}, "field_asking": "kabupaten"}
```

### Fuzzy Match Confirmation
```
User: "desa sukamju"
Bot: {"reply": "Apakah maksud Anda Desa SUKAMAJU, Kec. Cibinong? (ya/tidak)", "action": "confirm", "needs_validation": {"field": "desa", "value": "sukamju", "matched": "SUKAMAJU", "score": 0.91}}
```

### Ambiguous Match
```
User: "sukaraja"
Bot: {"reply": "Ada beberapa desa Sukaraja di Kab. Bogor:\n1. SUKARAJA, Kec. Sukaraja\n2. SUKARAJA, Kec. Tamansari\nYang mana? (balas angka)", "action": "collect", "needs_validation": {"field": "desa", "value": "sukaraja", "candidates": [...]}}
```

## Token Usage Estimation

Per conversation (avg 8 turns):
- System prompt: ~500 tokens
- Message history (10 msgs): ~800 tokens
- Wilayah context: ~200 tokens
- Response: ~200 tokens
- **Total per turn: ~1,700 tokens**
- **Per conversation: ~13,600 tokens**
- **Per day (100 msg, ~15 conversations): ~200K tokens**
- **Cost: ~$0.02/day** (Haiku pricing)

## Error Recovery

```
// LLM returns invalid JSON
→ Retry once with "Format response as valid JSON"
→ If still fails → send generic menu to user

// LLM hallucinates data
→ Never trust LLM-generated coordinates/BPS codes
→ Always validate against Dayawarga API

// User frustrated (3x invalid input)
→ "Maaf jika membingungkan. Ketik 'menu' untuk mulai ulang."
```
