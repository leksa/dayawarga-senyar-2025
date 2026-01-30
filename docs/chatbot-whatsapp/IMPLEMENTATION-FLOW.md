# Dayawarga Chatbot — Implementation Flow

## Request Lifecycle

```
Watzap Webhook POST /webhook
    │
    ▼
┌─ Parse & Validate ──────────────────────┐
│  1. Parse JSON body                      │
│  2. Check type === "incoming_chat"       │
│  3. Extract: from, message, type,        │
│     message_id, latitude, longitude      │
│  4. Idempotency: skip if message_id seen │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Auth Check ─────────────────────────────┐
│  1. Lookup phone in wa_relawan (SQLite)  │
│  2. If not found → "Maaf, Anda belum    │
│     terdaftar sebagai relawan."          │
│  3. If inactive → reject                │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Session Load ───────────────────────────┐
│  1. Get/create session from Map          │
│  2. Check TTL (30min)                    │
│  3. If expired → reset to IDLE           │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Handle by Message Type ─────────────────┐
│  text → process as conversation          │
│  location → store coords in session,     │
│             continue collecting           │
│  image → store media_url in session      │
└──────────────────────┬──────────────────┘
                       ▼
┌─ LLM Processing ────────────────────────┐
│  1. Build messages array:                │
│     - system prompt                      │
│     - last 10 conversation messages      │
│     - current collected_data summary     │
│  2. Call Claude Haiku                    │
│  3. Parse JSON response                  │
│  4. Handle action:                       │
│     - collect → update session           │
│     - confirm → show summary             │
│     - submit → call Dayawarga API        │
│     - reset → clear session              │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Wilayah Validation (if needed) ────────┐
│  1. Get needs_validation from LLM resp   │
│  2. Fetch desa list from cache/API       │
│  3. Jaro-Winkler fuzzy match             │
│  4. Score > 0.88 → auto-confirm         │
│  5. Score 0.65-0.88 → ask user          │
│  6. Score < 0.65 → retry                │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Submission (if action=submit) ──────────┐
│  1. Validate all required fields         │
│  2. Call Dayawarga API:                  │
│     - POST /api/feeds (feed)             │
│     - POST /api/locations (posko baru)   │
│     - PUT /api/locations/:id (update)    │
│  3. On success → confirm to user         │
│  4. On error → retry once, then report   │
└──────────────────────┬──────────────────┘
                       ▼
┌─ Send Reply ─────────────────────────────┐
│  1. Call Watzap send_message             │
│  2. Log outgoing message to SQLite       │
│  3. Update session state                 │
└──────────────────────────────────────────┘
```

## File-by-File Implementation

### `src/index.ts` — Entry point
- Create Express app
- Register routes: POST /webhook, GET /health
- Start session cleanup interval
- Init SQLite database
- Log startup info

### `src/config.ts` — Environment config
- WATZAP_API_KEY, WATZAP_NUMBER_KEY
- ANTHROPIC_API_KEY
- DAYAWARGA_API_URL, DAYAWARGA_API_KEY
- PORT (default 3001)
- NODE_ENV

### `src/webhook.ts` — Webhook handler
- Parse Watzap incoming payload
- Validate message type
- Route to conversation handler
- Handle errors gracefully

### `src/session.ts` — Session manager
- In-memory Map<string, Session>
- get/set/delete/cleanup methods
- TTL check on every access
- setInterval cleanup every 5 min

### `src/conversation.ts` — Conversation orchestrator
- Build LLM context from session
- Call Claude Haiku
- Parse JSON response
- Route actions (collect, confirm, submit, reset)
- Handle validation triggers

### `src/watzap.ts` — Watzap API client
- sendText(phone, message)
- sendImage(phone, imageUrl, caption)
- Retry on failure (1 attempt)

### `src/dayawarga-api.ts` — Dayawarga REST client
- getWilayah(level, parentId)
- getLocations(search)
- createLocation(data)
- updateLocation(id, data)
- createFeed(data)
- Auth via X-API-Key header

### `src/wilayah.ts` — Wilayah matching
- fetchAndCacheDesa(kabupatenId) → SQLite cache
- normalizeDesaName(input) → strip prefix, expand abbreviations
- jaroWinkler(a, b) → similarity score
- findDesaMatch(input, candidates) → MatchResult

### `src/db.ts` — SQLite initialization
- Create tables if not exist
- Expose db instance
- Helper: logConversation(), getRelawan(), cacheWilayah()

### `src/prompts.ts` — LLM prompts
- SYSTEM_PROMPT constant
- buildMessages(session, userMessage) → messages array
- Wilayah context injection

## Deployment (pm2)

### ecosystem.config.js
```javascript
module.exports = {
  apps: [{
    name: 'dayawarga-chatbot',
    script: 'node_modules/.bin/tsx',
    args: 'src/index.ts',
    env: {
      NODE_ENV: 'production'
    },
    max_memory_restart: '200M',
    error_file: './logs/error.log',
    out_file: './logs/out.log',
    time: true
  }]
};
```

### VPS Setup
```bash
# 1. Clone repo
git clone <repo> /opt/dayawarga-chatbot
cd /opt/dayawarga-chatbot

# 2. Install deps
npm install

# 3. Setup env
cp .env.example .env
# Edit .env with real values

# 4. Start with pm2
pm2 start ecosystem.config.js
pm2 save
pm2 startup

# 5. Reverse proxy (Caddy)
# /etc/caddy/Caddyfile
# chatbot.yourdomain.com {
#   reverse_proxy localhost:3001
# }
```

### Watzap Webhook Setup
```bash
curl -X POST https://api.watzap.id/v1/set_webhook \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "YOUR-API-KEY",
    "number_key": "YOUR-NUMBER-KEY",
    "endpoint_url": "https://chatbot.yourdomain.com/webhook"
  }'
```
