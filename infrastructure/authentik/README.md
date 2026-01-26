# Authentik Identity Provider Setup

Authentik adalah Identity Provider (IdP) berbasis OIDC yang digunakan untuk single sign-on di Admin Portal dan ODK Central.

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

### Production

| Service | URL |
|---------|-----|
| User Interface | https://auth.dayawarga.com |
| Admin Interface | https://auth.dayawarga.com/if/admin/ |
| OIDC Discovery | https://auth.dayawarga.com/application/o/admin-portal/.well-known/openid-configuration |

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
     - `http://localhost:5176/callback` (dev)
     - `https://admin.dayawarga.com/callback` (prod)
   - Scopes: `openid`, `email`, `profile`

### 2. Buat Application

1. Admin → Applications → Applications → Create
2. Konfigurasi:
   - Name: `Admin Portal`
   - Slug: `admin-portal`
   - Provider: Pilih "Admin Portal OIDC"
   - Launch URL: `http://localhost:5176` atau `https://admin.dayawarga.com`

### 3. Update Frontend Config

Setelah Provider dibuat, update frontend `.env`:

**Local Development:**
```env
VITE_OIDC_AUTHORITY=http://localhost:9000/application/o/admin-portal/
VITE_OIDC_CLIENT_ID=<client_id_dari_provider>
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

**Production:**
```env
VITE_OIDC_AUTHORITY=https://auth.dayawarga.com/application/o/admin-portal/
VITE_OIDC_CLIENT_ID=<client_id_dari_provider>
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

# Reset (CAUTION: deletes all data)
docker compose down -v
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
