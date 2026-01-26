# Dayawarga Admin Portal

Admin portal untuk manajemen organisasi, grup, dan relawan Dayawarga.

## Tech Stack

- **Vue 3** dengan Composition API + TypeScript
- **Vite** sebagai build tool
- **Pinia** untuk state management
- **shadcn-vue** untuk UI components
- **oidc-client-ts** untuk OIDC authentication
- **Tailwind CSS** untuk styling

## Features

- **Authentication** - OIDC login via Authentik
- **Organizations** - CRUD organisasi relawan
- **Groups** - Manajemen grup dalam organisasi
- **Relawan** - Pendaftaran dan manajemen relawan
- **ODK Integration** - QR code generation untuk ODK Collect
- **WhatsApp Verification** - Toggle akses chatbot untuk relawan

## Development

### Prerequisites

- Node.js 20+
- pnpm

### Setup

```bash
# Install dependencies
pnpm install

# Copy environment file
cp .env.example .env

# Edit .env with your config
# - VITE_OIDC_AUTHORITY: Authentik OIDC URL
# - VITE_OIDC_CLIENT_ID: Client ID from Authentik
# - VITE_API_BASE_URL: Backend API URL

# Start development server
pnpm dev
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `VITE_OIDC_AUTHORITY` | Authentik OIDC endpoint | `http://localhost:9000/application/o/admin-portal/` |
| `VITE_OIDC_CLIENT_ID` | OIDC Client ID | `admin-portal-client` |
| `VITE_API_BASE_URL` | Backend API URL | `http://localhost:8080/api/v1` |

## Production

### Build

```bash
pnpm build
```

Output akan ada di folder `dist/`.

### Docker

```bash
docker build -t dayawarga-admin-portal .
```

### Deployment

Admin Portal di-deploy sebagai bagian dari docker-compose dengan profile `admin-portal`:

```bash
# Di server production
docker compose --profile admin-portal up -d
```

## URLs

| Environment | URL |
|-------------|-----|
| Development | http://localhost:5176 |
| Production | https://admin.dayawarga.com |

## Project Structure

```
src/
├── components/         # Reusable UI components
│   └── ui/            # shadcn-vue components
├── composables/       # Vue composables
├── layouts/           # App layouts
├── router/            # Vue Router config
├── services/          # API client layer
│   ├── api.ts         # Axios instance with auth
│   ├── types.ts       # TypeScript interfaces
│   ├── organizations.ts
│   ├── groups.ts
│   └── relawan.ts
├── stores/            # Pinia stores
│   └── auth.ts        # Auth state + OIDC
└── views/             # Page components
    ├── DashboardView.vue
    ├── OrganizationsView.vue
    ├── GroupsView.vue
    ├── RelawanView.vue
    └── RelawanDetailView.vue
```

## API Integration

Admin Portal berkomunikasi dengan backend API yang dilindungi OIDC:

- Auth header dikirim otomatis via Axios interceptor
- 401 response akan redirect ke login
- Token refresh ditangani oleh oidc-client-ts

## Role-Based Access

| Role | Permissions |
|------|-------------|
| `super_admin` | Full access, manage all organizations |
| `org_admin` | Manage own organization only |
| `group_leader` | Manage own group only |
