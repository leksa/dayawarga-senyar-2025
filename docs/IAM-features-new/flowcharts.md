# System Flowcharts
# ODK Admin Portal

---

## 1. Authentication Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     AUTHENTICATION FLOW (OIDC)                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  User                   Frontend              Backend           Authentik │
│   │                        │                     │                  │    │
│   │  1. Click Login        │                     │                  │    │
│   ├───────────────────────►│                     │                  │    │
│   │                        │                     │                  │    │
│   │                        │  2. GET /api/auth/login                │    │
│   │                        ├────────────────────►│                  │    │
│   │                        │                     │                  │    │
│   │                        │  3. 302 Redirect to Authentik          │    │
│   │                        │◄────────────────────┤                  │    │
│   │                        │                     │                  │    │
│   │  4. Redirect to Authentik                    │                  │    │
│   │◄───────────────────────┤                     │                  │    │
│   │                        │                     │                  │    │
│   │  5. Enter credentials  │                     │                  │    │
│   ├────────────────────────┼─────────────────────┼─────────────────►│    │
│   │                        │                     │                  │    │
│   │  6. Authentik validates & redirects with code                   │    │
│   │◄───────────────────────┼─────────────────────┼──────────────────┤    │
│   │                        │                     │                  │    │
│   │  7. Callback to /api/auth/callback?code=xxx  │                  │    │
│   ├───────────────────────►│                     │                  │    │
│   │                        ├────────────────────►│                  │    │
│   │                        │                     │                  │    │
│   │                        │  8. Exchange code for tokens           │    │
│   │                        │                     ├─────────────────►│    │
│   │                        │                     │◄─────────────────┤    │
│   │                        │                     │  access_token    │    │
│   │                        │                     │                  │    │
│   │                        │  9. Validate token, get user info      │    │
│   │                        │                     ├─────────────────►│    │
│   │                        │                     │◄─────────────────┤    │
│   │                        │                     │                  │    │
│   │                        │  10. Find or create user in DB         │    │
│   │                        │                     │                  │    │
│   │                        │  11. Create session (JWT cookie)       │    │
│   │                        │◄────────────────────┤                  │    │
│   │                        │                     │                  │    │
│   │  12. Redirect to Dashboard                   │                  │    │
│   │◄───────────────────────┤                     │                  │    │
│   │                        │                     │                  │    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Create Sub-Lembaga Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   CREATE SUB-LEMBAGA FLOW                                │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Admin Lembaga         Frontend              Backend          ODK Central│
│       │                   │                     │                  │    │
│       │  1. Open "Create  │                     │                  │    │
│       │     Sub-Lembaga"  │                     │                  │    │
│       ├──────────────────►│                     │                  │    │
│       │                   │                     │                  │    │
│       │                   │  2. GET /api/odk/projects               │    │
│       │                   ├────────────────────►│                  │    │
│       │                   │                     │                  │    │
│       │                   │  3. Check cache, if stale:              │    │
│       │                   │                     ├─────────────────►│    │
│       │                   │                     │  GET /v1/projects│    │
│       │                   │                     │◄─────────────────┤    │
│       │                   │                     │                  │    │
│       │                   │  4. Filter out already-linked projects  │    │
│       │                   │                     │                  │    │
│       │                   │  5. Return available projects           │    │
│       │                   │◄────────────────────┤                  │    │
│       │                   │                     │                  │    │
│       │  6. Display available                   │                  │    │
│       │     ODK Projects  │                     │                  │    │
│       │◄──────────────────┤                     │                  │    │
│       │                   │                     │                  │    │
│       │  7. Select project│                     │                  │    │
│       │     Enter name    │                     │                  │    │
│       ├──────────────────►│                     │                  │    │
│       │                   │                     │                  │    │
│       │                   │  8. POST /api/lembaga/{id}/sub-lembaga  │    │
│       │                   │     { name, odkProjectId }              │    │
│       │                   ├────────────────────►│                  │    │
│       │                   │                     │                  │    │
│       │                   │  9. Validate:                           │    │
│       │                   │     - User is Admin Lembaga             │    │
│       │                   │     - Project not already linked        │    │
│       │                   │                     │                  │    │
│       │                   │  10. Create sub_lembaga record          │    │
│       │                   │                     │                  │    │
│       │                   │  11. Log audit                          │    │
│       │                   │                     │                  │    │
│       │                   │  12. Return success                     │    │
│       │                   │◄────────────────────┤                  │    │
│       │                   │                     │                  │    │
│       │  13. Show success │                     │                  │    │
│       │      Redirect to  │                     │                  │    │
│       │      member assign│                     │                  │    │
│       │◄──────────────────┤                     │                  │    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Assign Member to Sub-Lembaga Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   ASSIGN MEMBER FLOW                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Admin/PM              Frontend              Backend          ODK Central│
│     │                     │                     │                  │    │
│     │  1. Open "Assign    │                     │                  │    │
│     │     Member" modal   │                     │                  │    │
│     ├────────────────────►│                     │                  │    │
│     │                     │                     │                  │    │
│     │                     │  2. GET /api/lembaga/{id}/members       │    │
│     │                     │     (available members pool)            │    │
│     │                     ├────────────────────►│                  │    │
│     │                     │◄────────────────────┤                  │    │
│     │                     │                     │                  │    │
│     │  3. Show member list│                     │                  │    │
│     │◄────────────────────┤                     │                  │    │
│     │                     │                     │                  │    │
│     │  4. Select member   │                     │                  │    │
│     │     Select role     │                     │                  │    │
│     │     (PM or Relawan) │                     │                  │    │
│     ├────────────────────►│                     │                  │    │
│     │                     │                     │                  │    │
│     │                     │  5. POST /api/sub-lembaga/{id}/members  │    │
│     │                     │     { userId, role }                    │    │
│     │                     ├────────────────────►│                  │    │
│     │                     │                     │                  │    │
│     │                     │  6. Validate:                           │    │
│     │                     │     - User in same lembaga              │    │
│     │                     │     - Not already assigned              │    │
│     │                     │     - PM can only assign Relawan        │    │
│     │                     │                     │                  │    │
│     │                     │  7. Create ODK App User                 │    │
│     │                     │                     ├─────────────────►│    │
│     │                     │                     │ POST /v1/projects│    │
│     │                     │                     │ /{id}/app-users  │    │
│     │                     │                     │◄─────────────────┤    │
│     │                     │                     │ { id, token }    │    │
│     │                     │                     │                  │    │
│     │                     │  8. Assign App User to all forms        │    │
│     │                     │                     ├─────────────────►│    │
│     │                     │                     │ POST /forms/{id}/│    │
│     │                     │                     │ assignments/...  │    │
│     │                     │                     │◄─────────────────┤    │
│     │                     │                     │                  │    │
│     │                     │  9. Generate QR code data               │    │
│     │                     │     (compress + base64)                 │    │
│     │                     │                     │                  │    │
│     │                     │  10. Save sub_lembaga_members record    │    │
│     │                     │      with odk_app_user_id, qr_code      │    │
│     │                     │                     │                  │    │
│     │                     │  11. Log audit                          │    │
│     │                     │                     │                  │    │
│     │                     │  12. Return success + QR data           │    │
│     │                     │◄────────────────────┤                  │    │
│     │                     │                     │                  │    │
│     │  13. Show success   │                     │                  │    │
│     │      Display QR code│                     │                  │    │
│     │◄────────────────────┤                     │                  │    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 4. Remove Member Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   REMOVE MEMBER FLOW                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Admin/PM              Frontend              Backend          ODK Central│
│     │                     │                     │                  │    │
│     │  1. Click "Remove"  │                     │                  │    │
│     │     on member       │                     │                  │    │
│     ├────────────────────►│                     │                  │    │
│     │                     │                     │                  │    │
│     │  2. Confirm dialog  │                     │                  │    │
│     │◄────────────────────┤                     │                  │    │
│     │                     │                     │                  │    │
│     │  3. Confirm         │                     │                  │    │
│     ├────────────────────►│                     │                  │    │
│     │                     │                     │                  │    │
│     │                     │  4. DELETE /api/sub-lembaga/{id}/       │    │
│     │                     │     members/{memberId}                  │    │
│     │                     ├────────────────────►│                  │    │
│     │                     │                     │                  │    │
│     │                     │  5. Validate permissions                │    │
│     │                     │                     │                  │    │
│     │                     │  6. Get odk_app_user_id from record     │    │
│     │                     │                     │                  │    │
│     │                     │  7. Revoke ODK App User                 │    │
│     │                     │                     ├─────────────────►│    │
│     │                     │                     │ DELETE /v1/      │    │
│     │                     │                     │ projects/{id}/   │    │
│     │                     │                     │ app-users/{id}   │    │
│     │                     │                     │◄─────────────────┤    │
│     │                     │                     │                  │    │
│     │                     │  8. Update record:                      │    │
│     │                     │     is_active = false                   │    │
│     │                     │     revoked_at = now                    │    │
│     │                     │     revoked_by = current_user           │    │
│     │                     │                     │                  │    │
│     │                     │  9. Log audit                           │    │
│     │                     │                     │                  │    │
│     │                     │  10. Return success                     │    │
│     │                     │◄────────────────────┤                  │    │
│     │                     │                     │                  │    │
│     │  11. Show success   │                     │                  │    │
│     │      Remove from    │                     │                  │    │
│     │      list           │                     │                  │    │
│     │◄────────────────────┤                     │                  │    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 5. View Own QR Codes Flow (Relawan)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   VIEW OWN QR CODES FLOW                                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  Relawan               Frontend              Backend                     │
│     │                     │                     │                        │
│     │  1. Open "My        │                     │                        │
│     │     Projects" page  │                     │                        │
│     ├────────────────────►│                     │                        │
│     │                     │                     │                        │
│     │                     │  2. GET /api/me/projects                     │
│     │                     ├────────────────────►│                        │
│     │                     │                     │                        │
│     │                     │  3. Query sub_lembaga_members                │
│     │                     │     WHERE user_id = current_user             │
│     │                     │     AND is_active = true                     │
│     │                     │                     │                        │
│     │                     │  4. For each assignment, include:            │
│     │                     │     - sub_lembaga name                       │
│     │                     │     - role                                   │
│     │                     │     - odk_app_user_id (SubmitterID)          │
│     │                     │     - qr_code_data                           │
│     │                     │                     │                        │
│     │                     │  5. Return list                              │
│     │                     │◄────────────────────┤                        │
│     │                     │                     │                        │
│     │  6. Display project │                     │                        │
│     │     cards with:     │                     │                        │
│     │     - Project name  │                     │                        │
│     │     - Role badge    │                     │                        │
│     │     - QR code image │                     │                        │
│     │     - SubmitterID   │                     │                        │
│     │     - Download btn  │                     │                        │
│     │◄────────────────────┤                     │                        │
│     │                     │                     │                        │
│     │  7. Click Download  │                     │                        │
│     ├────────────────────►│                     │                        │
│     │                     │                     │                        │
│     │  8. Generate PNG    │                     │                        │
│     │     from QR data    │                     │                        │
│     │     (client-side)   │                     │                        │
│     │                     │                     │                        │
│     │  9. Download file   │                     │                        │
│     │◄────────────────────┤                     │                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 6. View Submissions Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   VIEW SUBMISSIONS FLOW                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  User                  Frontend              Backend          ODK Central│
│   │                       │                     │                  │    │
│   │  1. Open Submissions  │                     │                  │    │
│   │     page              │                     │                  │    │
│   ├──────────────────────►│                     │                  │    │
│   │                       │                     │                  │    │
│   │                       │  2. GET /api/odk/submissions            │    │
│   │                       │     ?subLembagaId=&dateFrom=&dateTo=    │    │
│   │                       ├────────────────────►│                  │    │
│   │                       │                     │                  │    │
│   │                       │  3. Check user role & permissions       │    │
│   │                       │                     │                  │    │
│   │                       │  4. Determine scope:                    │    │
│   │                       │     Admin: all projects                 │    │
│   │                       │     Admin Lembaga: lembaga projects     │    │
│   │                       │     PM: sub-lembaga only                │    │
│   │                       │     Relawan: own submitterIDs only      │    │
│   │                       │                     │                  │    │
│   │                       │  5. Fetch submissions from ODK          │    │
│   │                       │                     ├─────────────────►│    │
│   │                       │                     │ GET /v1/projects/│    │
│   │                       │                     │ {id}/forms/{id}/ │    │
│   │                       │                     │ submissions      │    │
│   │                       │                     │◄─────────────────┤    │
│   │                       │                     │                  │    │
│   │                       │  6. Map submitterID to user names       │    │
│   │                       │     from sub_lembaga_members table      │    │
│   │                       │                     │                  │    │
│   │                       │  7. Filter by user scope if needed      │    │
│   │                       │                     │                  │    │
│   │                       │  8. Return submissions with names       │    │
│   │                       │◄────────────────────┤                  │    │
│   │                       │                     │                  │    │
│   │  9. Display table     │                     │                  │    │
│   │     with columns:     │                     │                  │    │
│   │     - Date            │                     │                  │    │
│   │     - Form name       │                     │                  │    │
│   │     - Submitter name  │                     │                  │    │
│   │     - SubmitterID     │                     │                  │    │
│   │◄──────────────────────┤                     │                  │    │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 7. Dashboard Statistics Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   DASHBOARD FLOW                                         │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  User                  Frontend              Backend                     │
│   │                       │                     │                        │
│   │  1. Open Dashboard    │                     │                        │
│   ├──────────────────────►│                     │                        │
│   │                       │                     │                        │
│   │                       │  2. GET /api/dashboard/stats               │
│   │                       ├────────────────────►│                        │
│   │                       │                     │                        │
│   │                       │  3. Check user role                        │
│   │                       │                     │                        │
│   │                       │  4. Calculate stats based on role:         │
│   │                       │                     │                        │
│   │                       │     ADMINISTRATOR:                         │
│   │                       │     - Total lembaga                        │
│   │                       │     - Total sub-lembaga                    │
│   │                       │     - Total members                        │
│   │                       │     - Submissions: today/week/month        │
│   │                       │                     │                        │
│   │                       │     ADMIN LEMBAGA:                         │
│   │                       │     - My lembaga sub-lembaga count         │
│   │                       │     - My lembaga member count              │
│   │                       │     - Lembaga submissions                  │
│   │                       │                     │                        │
│   │                       │     PROJECT MANAGER:                       │
│   │                       │     - My sub-lembaga member count          │
│   │                       │     - Sub-lembaga submissions              │
│   │                       │                     │                        │
│   │                       │     RELAWAN:                               │
│   │                       │     - My project count                     │
│   │                       │     - My submission count                  │
│   │                       │                     │                        │
│   │                       │  5. Query recent submissions               │
│   │                       │     (for recent activity list)             │
│   │                       │                     │                        │
│   │                       │  6. Return stats + recent list             │
│   │                       │◄────────────────────┤                        │
│   │                       │                     │                        │
│   │  7. Display:          │                     │                        │
│   │     - Stat cards      │                     │                        │
│   │     - Recent activity │                     │                        │
│   │◄──────────────────────┤                     │                        │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 8. QR Code Generation (Technical)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                   QR CODE GENERATION                                     │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  INPUT (from ODK Central):                                              │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ App User Token: "abc123xyz..."                                   │   │
│  │ Project ID: 5                                                    │   │
│  │ ODK Central URL: "https://odk.example.com"                      │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  STEP 1: Build settings JSON                                            │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ {                                                                │   │
│  │   "general": {                                                   │   │
│  │     "server_url": "https://odk.example.com/v1/key/abc123xyz/    │   │
│  │                    projects/5"                                   │   │
│  │   },                                                             │   │
│  │   "project": {                                                   │   │
│  │     "name": "Survei KTP",                                       │   │
│  │     "icon": "S",                                                │   │
│  │     "color": "#3B82F6"                                          │   │
│  │   },                                                             │   │
│  │   "admin": {}                                                    │   │
│  │ }                                                                │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  STEP 2: Compress with zlib                                             │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ const compressed = zlib.deflateSync(JSON.stringify(settings))   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  STEP 3: Encode to Base64                                               │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ const qrData = compressed.toString('base64')                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  STEP 4: Generate QR Code image                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │ const qr = QRCode.create(qrData, { errorCorrectionLevel: 'H' }) │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│  OUTPUT: QR Code that ODK Collect can scan                              │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```
