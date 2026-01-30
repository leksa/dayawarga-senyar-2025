# Authentik Identity Provider Setup

Authentik adalah Identity Provider (IdP) berbasis OIDC yang digunakan untuk single sign-on di Admin Portal dan ODK Central.

**Versi**: 2025.6.1

## Quick Start

### 1. Generate Secrets

```bash
# Generate secret key
openssl rand -base64 60

# Generate PostgreSQL password
openssl rand -base64 32
```

### 2. Setup Environment

```bash
# Copy example env
cp .env.example .env

# Edit .env dan isi AUTHENTIK_SECRET_KEY dan AUTHENTIK_POSTGRES_PASSWORD
```

### 3. Start Authentik

```bash
docker compose up -d
```

### 4. Initial Setup

1. Buka http://localhost:9000/if/flow/initial-setup/
2. Buat akun admin pertama
3. Login ke Admin Interface: http://localhost:9000/if/admin/

## URLs

### Local Development

| Service | URL |
|---------|-----|
| User Interface | http://localhost:9000 |
| Admin Interface | http://localhost:9000/if/admin/ |
| Initial Setup | http://localhost:9000/if/flow/initial-setup/ |
| Password Recovery | http://localhost:9000/if/flow/recovery/ |

### Production

| Service | URL |
|---------|-----|
| User Interface | https://auth.dayawarga.com |
| Admin Interface | https://auth.dayawarga.com/if/admin/ |
| Password Recovery | https://auth.dayawarga.com/if/flow/recovery/ |
| OIDC Discovery | https://auth.dayawarga.com/application/o/admin-portal/.well-known/openid-configuration |

## Custom Branding (Dayawarga Theme)

### 1. Upload Logo & Set Branding

1. Buka Admin Interface: http://localhost:9000/if/admin/
2. Pergi ke **System** → **Brands**
3. Edit brand **authentik-default** atau buat baru:
   - **Domain**: `localhost` (atau domain production)
   - **Branding title**: `Dayawarga`
   - **Logo**: Upload `branding/logo.png` atau URL: `/static/dist/assets/icons/icon.png`
   - **Favicon**: Upload `branding/logo.png`

### 2. Apply Custom CSS (Authentik 2025.4+)

1. Di halaman Brand yang sama, scroll ke **Branding settings**
2. Cari field **Custom CSS**
3. Copy-paste isi dari file `branding/custom.css`
4. Klik **Update**

### 3. Set Default Flow Background (Optional)

1. Di halaman Brand, cari **Default flow background**
2. Upload background image atau biarkan kosong untuk gradient dari CSS

### Theme Files

```
infrastructure/authentik/branding/
├── logo.png          # Logo Dayawarga
└── custom.css        # Custom CSS theme (dark teal theme)
```

### Preview Setelah Setup

- Background gradient gelap (#1a1f2e → #0f1219)
- Logo Dayawarga di header
- Input fields dengan border teal
- Tombol login gradient teal (#4DB6AC → #2B7A9E)
- Font Plus Jakarta Sans

## Password Recovery Setup

### 1. Setup SMTP untuk Email

Update `docker-compose.yml` environment untuk Authentik server dan worker:

```yaml
environment:
  # ... existing config ...
  AUTHENTIK_EMAIL__HOST: live.smtp.mailtrap.io
  AUTHENTIK_EMAIL__PORT: 587
  AUTHENTIK_EMAIL__USERNAME: ${SMTP_USERNAME}
  AUTHENTIK_EMAIL__PASSWORD: ${SMTP_PASSWORD}
  AUTHENTIK_EMAIL__USE_TLS: true
  AUTHENTIK_EMAIL__FROM: noreply@dayawarga.com
```

Atau setup via Admin UI:
1. **System** → **Settings** → scroll ke **Email**
2. Isi SMTP settings

### 2. Verify Recovery Flow

1. Pergi ke **Flows and Stages** → **Flows**
2. Pastikan ada flow dengan designation **Recovery**
3. Jika tidak ada, buat flow baru atau import default recovery flow

### 3. Set Recovery Flow di Brand

1. **System** → **Brands** → Edit brand
2. Di **Default flows**, set **Recovery flow** ke recovery flow yang ada

### 4. Test Recovery

1. Buka http://localhost:9000/if/flow/recovery/
2. Masukkan email
3. Cek inbox untuk link reset password

## OIDC Provider Setup

Setelah Authentik berjalan, buat OIDC Provider untuk Admin Portal:

### 1. Buat Provider

1. Admin → Applications → Providers → Create
2. Pilih "OAuth2/OpenID Provider"
3. Konfigurasi:
   - Name: `Admin Portal OIDC`
   - Authorization flow: default-provider-authorization-implicit-consent
   - Client type: Confidential
   - Client ID: (auto-generated, catat ini)
   - Client Secret: (auto-generated, catat ini)
   - Redirect URIs:
     - `http://localhost:5173/callback` (dev)
     - `https://admin.dayawarga.com/callback` (prod)
   - Scopes: `openid`, `email`, `profile`

### 2. Buat Application

1. Admin → Applications → Applications → Create
2. Konfigurasi:
   - Name: `Admin Portal`
   - Slug: `admin-portal`
   - Provider: Pilih "Admin Portal OIDC"
   - Launch URL: `http://localhost:5173` atau `https://admin.dayawarga.com`

### 3. Update Frontend Config

Setelah Provider dibuat, update frontend `.env`:

**Local Development:**
```env
VITE_OIDC_AUTHORITY=http://localhost:9000/application/o/admin-portal/
VITE_OIDC_CLIENT_ID=<client_id_dari_provider>
VITE_AUTHENTIK_BASE_URL=http://localhost:9000
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

**Production:**
```env
VITE_OIDC_AUTHORITY=https://auth.dayawarga.com/application/o/admin-portal/
VITE_OIDC_CLIENT_ID=<client_id_dari_provider>
VITE_AUTHENTIK_BASE_URL=https://auth.dayawarga.com
VITE_API_BASE_URL=https://api.dayawarga.com/api/v1
```

## ODK Central OIDC Integration

Untuk mengintegrasikan ODK Central dengan Authentik:

### 1. Buat Provider untuk ODK

1. Admin → Providers → Create OAuth2/OpenID Provider
2. Name: `ODK Central OIDC`
3. Redirect URIs: `https://data.dayawarga.com/v1/oidc/callback`

### 2. Konfigurasi ODK Central

Tambahkan environment variables ke ODK Central server:

```env
OIDC_ENABLED=true
OIDC_ISSUER_URL=https://auth.dayawarga.com/application/o/odk-central/
OIDC_CLIENT_ID=<client_id>
OIDC_CLIENT_SECRET=<client_secret>
```

## Production Deployment

Authentik dijalankan sebagai bagian dari docker-compose dengan profile `admin-portal`:

```bash
# Di server production
cd /opt/dayawarga
docker compose --profile admin-portal up -d
```

Required environment variables di `.env`:
```env
AUTHENTIK_SECRET_KEY=<generate: openssl rand -base64 60>
AUTHENTIK_POSTGRES_PASSWORD=<strong password>
ADMIN_PORTAL_OIDC_CLIENT_ID=<from Authentik admin>

# SMTP for password recovery
SMTP_USERNAME=<mailtrap username>
SMTP_PASSWORD=<mailtrap password>
```

## Management Commands

```bash
# Start
docker compose up -d

# Stop
docker compose down

# View logs
docker compose logs -f

# Restart
docker compose restart

# Upgrade (pull new image)
docker compose pull && docker compose up -d

# Reset (CAUTION: deletes all data)
docker compose down -v
```

## Upgrading Authentik

### From 2024.x to 2025.x

1. Backup database:
   ```bash
   docker compose exec authentik-postgres pg_dump -U authentik authentik > backup.sql
   ```

2. Update image version di `docker-compose.yml`:
   ```yaml
   image: ghcr.io/goauthentik/server:2025.6.1
   ```

3. Pull dan restart:
   ```bash
   docker compose pull
   docker compose up -d
   ```

4. Check logs for migration:
   ```bash
   docker compose logs -f authentik-server
   ```

## Troubleshooting

### Port Already in Use

Ubah port di `.env`:
```env
AUTHENTIK_PORT_HTTP=9001
AUTHENTIK_PORT_HTTPS=9444
```

### Cannot Access Admin Interface

Cek container status:
```bash
docker compose ps
docker compose logs authentik-server
```

### Database Connection Error

Pastikan PostgreSQL healthy:
```bash
docker compose logs authentik-postgres
```

### Custom CSS Not Applied

1. Pastikan versi Authentik >= 2025.4.0
2. Clear browser cache (Ctrl+Shift+R)
3. Cek CSS syntax errors di browser console

### Password Recovery Email Not Sent

1. Cek SMTP config di **System** → **Settings** → **Email**
2. Test email: **System** → **Settings** → scroll ke Email → **Test Email**
3. Cek logs: `docker compose logs authentik-worker`

---

## Production Server Configuration Reference

> **Note**: All sensitive values (passwords, tokens, IPs) are stored in `.env` file on the server.
> Never commit actual credentials to version control.

### Server Access

| Item | Value |
|------|-------|
| Server IP | See `.env` or internal documentation |
| SSH User | See internal documentation |
| Project Path | `/opt/dayawarga` |

### URLs

| Service | URL |
|---------|-----|
| Main Frontend | https://dayawarga.com |
| Admin Portal | https://admin.dayawarga.com |
| API | https://api.dayawarga.com |
| Authentik SSO | https://auth.dayawarga.com |
| ODK Central | https://data.dayawarga.com |

### OIDC Configuration (Admin Portal)

| Item | Value |
|------|-------|
| Issuer URL | `https://auth.dayawarga.com/application/o/admin-portal/` |
| Client ID | See Authentik Admin UI or `.env` |
| Discovery URL | `https://auth.dayawarga.com/application/o/admin-portal/.well-known/openid-configuration` |

### SMTP Configuration

| Item | Value |
|------|-------|
| Host | See `.env` |
| Port | `587` |
| Username | See `.env` |
| Password | See `.env` |
| From Email | `noreply@dayawarga.com` |
| TLS | Required (STARTTLS on port 587) |

### Authentik API Token

For programmatic user management, generate a token in Authentik Admin UI:
1. Go to **Directory** → **Tokens and App passwords**
2. Create a new token with appropriate permissions

Example API call:
```bash
curl -X POST "https://auth.dayawarga.com/api/v3/core/users/" \
  -H "Authorization: Bearer <your-api-token>" \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "name": "New User", "email": "user@example.com", "is_active": true}'
```

### Database Access

```bash
# Connect to PostgreSQL (Dayawarga DB)
docker exec -it senyar-postgres psql -U senyar -d senyar

# Connect to PostgreSQL (Authentik DB)
docker exec -it authentik-postgres psql -U authentik -d authentik
```

### Common Commands

```bash
# Navigate to project
cd /opt/dayawarga

# View logs
docker logs senyar-api 2>&1 | tail -50
docker logs senyar-admin-portal 2>&1 | tail -50
docker logs authentik-server 2>&1 | tail -50

# Restart services
docker compose up -d api
docker compose --profile admin-portal up -d

# Rebuild and restart API
docker compose build api && docker compose up -d api

# Check running containers
docker compose ps
docker compose --profile admin-portal ps

# Pull latest and restart
git pull && docker compose build api && docker compose up -d api
```

### Environment Variables (.env on server)

Key variables in `/opt/dayawarga/.env`:

```env
# Database
DB_USER=senyar
DB_PASSWORD=<password>
DB_NAME=senyar

# ODK Central
ODK_BASE_URL=https://data.dayawarga.com
ODK_EMAIL=<odk-admin-email>
ODK_PASSWORD=<odk-admin-password>
ODK_PROJECT_ID=<project-id>

# OIDC for API
OIDC_ISSUER_URL=https://auth.dayawarga.com/application/o/admin-portal/
OIDC_CLIENT_ID=<from-authentik-admin>

# Admin Portal OIDC
ADMIN_PORTAL_OIDC_CLIENT_ID=<from-authentik-admin>

# SMTP
SMTP_HOST=<smtp-host>
SMTP_PORT=587
SMTP_USERNAME=<smtp-username>
SMTP_PASSWORD=<smtp-password>
SMTP_FROM=noreply@dayawarga.com

# Authentik
AUTHENTIK_SECRET_KEY=<generate: openssl rand -base64 60>
AUTHENTIK_POSTGRES_PASSWORD=<strong-password>
```

### Authentik Admin Account

| Item | Value |
|------|-------|
| Username | `akadmin` |
| Email | Set during initial setup |
| Password | Set during initial setup (change immediately) |
| Role in Dayawarga DB | `super_admin` |

---

## Webhook: Auto-Sync Users to ODK Central

When a new user is created in Authentik, a webhook automatically creates the corresponding ODK Web User.

### How It Works

1. **User Created in Authentik** → Triggers `model_created` event for User model
2. **Event Matcher Policy** (`odk-user-sync-match-user-created`) → Matches the event
3. **Notification Rule** (`odk-user-sync-rule`) → Sends notification via webhook transport
4. **Webhook Transport** (`odk-user-sync-webhook`) → POSTs to API endpoint
5. **API Handler** → Creates ODK Web User with the same email

### Configuration Summary

| Component | Name | Purpose |
|-----------|------|---------|
| Webhook Mapping | `odk-user-sync-mapping` | Formats payload with user data and secret |
| Notification Transport | `odk-user-sync-webhook` | Sends POST to API webhook endpoint |
| Event Matcher Policy | `odk-user-sync-match-user-created` | Matches `model_created` + `User` model |
| Notification Rule | `odk-user-sync-rule` | Binds policy to transport |

### Webhook Payload Format

```json
{
  "event_type": "user_created",
  "secret": "<AUTHENTIK_WEBHOOK_SECRET>",
  "user": {
    "pk": 123,
    "username": "newuser",
    "email": "newuser@example.com",
    "name": "New User",
    "is_active": true
  }
}
```

### API Endpoint

```
POST https://api.dayawarga.com/api/v1/webhooks/authentik
```

### Environment Variables

Add to `.env`:
```env
AUTHENTIK_WEBHOOK_SECRET=<generate with: openssl rand -hex 32>
```

### Important Notes

- **Manual Project Assignment**: The webhook only creates ODK Web User. Admin must manually assign user to ODK project to get App User + QR code.
- **Auto-Approval**: Users are created with password (immediately active, no email verification in ODK).
- **Secret in Body**: Authentik webhooks don't support custom headers, so the secret is included in the payload body.

### Troubleshooting

1. **Check Authentik Logs**:
   ```bash
   docker compose logs authentik-worker | grep -i webhook
   ```

2. **Check API Logs**:
   ```bash
   docker logs senyar-api 2>&1 | grep -i webhook
   ```

3. **Test Webhook Manually**:
   ```bash
   curl -X POST "https://api.dayawarga.com/api/v1/webhooks/authentik" \
     -H "Content-Type: application/json" \
     -d '{
       "event_type": "user_created",
       "secret": "<your-secret>",
       "user": {
         "pk": 999,
         "username": "testuser",
         "email": "test@example.com",
         "name": "Test User",
         "is_active": true
       }
     }'
   ```

4. **Health Check**:
   ```bash
   curl "https://api.dayawarga.com/api/v1/webhooks/authentik/health"
   ```
