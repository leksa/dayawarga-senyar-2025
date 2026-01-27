# Dayawarga Production Environment

## Server Access

### Main Server (API, Frontend, Database)
| Property | Value |
|----------|-------|
| **Server IP** | `103.179.57.203` |
| **SSH User** | `leksa` |
| **SSH Access** | Public key authentication |
| **Hostname** | `dayawarga` |

```bash
ssh leksa@103.179.57.203
```

### Chatbot Server
| Property | Value |
|----------|-------|
| **Server IP** | `103.172.204.21` |
| **SSH User** | `leksa` |
| **SSH Access** | Public key authentication |
| **Hostname** | `aichatsenyar` |

```bash
ssh leksa@103.172.204.21
```

## Deployment Structure

| Path | Description |
|------|-------------|
| `/opt/dayawarga/` | Main deployment directory |
| `/opt/dayawarga/.env` | Production environment variables |
| `/opt/dayawarga/docker-compose.yml` | Docker compose config |

## Docker Containers

| Container | Description | Port |
|-----------|-------------|------|
| `senyar-api` | Go API backend | 8080 |
| `senyar-frontend` | Vue.js main frontend | 80 (via Traefik) |
| `senyar-postgres` | PostgreSQL 16 database | 5432 |
| `senyar-traefik` | Reverse proxy & SSL | 80, 443 |
| `senyar-ghost` | Ghost CMS for stories | 2368 |

## URLs / Subdomains

| Service | URL |
|---------|-----|
| **Main Site (Peta)** | https://dayawarga.com |
| **API** | https://api.dayawarga.com |
| **Admin Portal** | https://admin.dayawarga.com |
| **ODK Central** | https://data.dayawarga.com |
| **Ghost/Stories** | https://stories.dayawarga.com |
| **Authentik** | (check infrastructure/authentik) |

## Database Access

```bash
# Via SSH + Docker
ssh leksa@103.179.57.203
docker exec -it senyar-postgres psql -U senyar -d senyar

# Quick query
ssh leksa@103.179.57.203 "docker exec senyar-postgres psql -U senyar -d senyar -c 'SELECT ...';"
```

## Logs

```bash
# API logs
ssh leksa@103.179.57.203 "docker logs senyar-api --tail 100"

# Frontend logs
ssh leksa@103.179.57.203 "docker logs senyar-frontend --tail 100"

# All containers
ssh leksa@103.179.57.203 "docker ps && docker logs senyar-api --tail 50"
```

## Deployment

### GitHub Actions (Automatic)
- **dayawarga-senyar-2025**: Manual trigger via `gh workflow run "CI/CD Pipeline"`
- **dayawarga-chatbot**: Auto-deploy on push to main

### Manual Deployment
```bash
ssh leksa@103.179.57.203
cd /opt/dayawarga
git pull
docker-compose up -d --build
```

## Related Repositories

| Repo | Description | Local Path |
|------|-------------|------------|
| `leksa/dayawarga-senyar-2025` | Main platform | `/Users/leksa/Development/dayawarga-senyar-2025` |
| `leksa/relawan-chat-llm` | WhatsApp chatbot | `/Users/leksa/Development/dayawarga-chatbot` |

## Environment Variables (Non-Sensitive)

```env
# Database
DB_USER=senyar
DB_NAME=senyar

# ODK Central
ODK_BASE_URL=https://data.dayawarga.com
ODK_PROJECT_ID=3

# S3 Storage
S3_ENABLED=true
S3_ENDPOINT=is3.cloudhost.id

# Scheduler
SCHEDULER_ENABLED=true

# SSL
ACME_EMAIL=admin@dayawarga.com
```

## WhatsApp Chatbot

| Property | Value |
|----------|-------|
| **Server IP** | `103.172.204.21` |
| **Hostname** | `aichatsenyar` |
| **Deploy Path** | `/app/relawan-chat-llm` |
| **Process Manager** | PM2 |
| **Port** | 3001 |
| **Webhook Provider** | Watzap.id |
| **Repository** | `leksa/relawan-chat-llm` |

### Chatbot Commands
```bash
# SSH to chatbot server
ssh leksa@103.172.204.21

# Check status
pm2 list

# View logs
pm2 logs chatbot --lines 50

# Restart
pm2 restart chatbot

# Full restart with npm install
cd /app/relawan-chat-llm && npm install && pm2 restart chatbot
```

## Quick Commands

### Main Server (103.179.57.203)
```bash
# Check all services
ssh leksa@103.179.57.203 "docker ps"

# Restart API
ssh leksa@103.179.57.203 "cd /opt/dayawarga && docker-compose restart senyar-api"

# View API logs
ssh leksa@103.179.57.203 "docker logs senyar-api --tail 100 -f"

# Database query
ssh leksa@103.179.57.203 "docker exec senyar-postgres psql -U senyar -d senyar -c 'SELECT COUNT(*) FROM users;'"
```

### Chatbot Server (103.172.204.21)
```bash
# Check chatbot status
ssh leksa@103.172.204.21 "pm2 list"

# View chatbot logs
ssh leksa@103.172.204.21 "pm2 logs chatbot --lines 50"

# Restart chatbot
ssh leksa@103.172.204.21 "pm2 restart chatbot"

# Full restart with npm install
ssh leksa@103.172.204.21 "cd /app/relawan-chat-llm && npm install && pm2 restart chatbot"
```
