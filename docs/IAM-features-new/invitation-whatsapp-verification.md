# User Invitation with WhatsApp PIN Verification

## Overview

Flow untuk mengundang user baru ke Dayawarga Admin Portal dengan verifikasi melalui WhatsApp. User tidak perlu Authentik account terlebih dahulu - mereka set password langsung di portal dan diverifikasi via WhatsApp chatbot.

## Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│              USER INVITATION WITH WHATSAPP VERIFICATION                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Admin          Portal           API           Chatbot        WhatsApp      │
│    │              │               │               │               │         │
│    │  1. Create   │               │               │               │         │
│    │  Organization│               │               │               │         │
│    │  + Admin     │               │               │               │         │
│    ├─────────────►│               │               │               │         │
│    │              │               │               │               │         │
│    │              │  2. POST      │               │               │         │
│    │              │  /organizations│              │               │         │
│    │              │  /with-admin  │               │               │         │
│    │              ├──────────────►│               │               │         │
│    │              │               │               │               │         │
│    │              │  3. Create org│               │               │         │
│    │              │     Create user (pending_invitation)         │         │
│    │              │     Generate token                           │         │
│    │              │     Send email with link                     │         │
│    │              │               │               │               │         │
│    │              │◄──────────────┤               │               │         │
│    │              │  invitation_link              │               │         │
│    │◄─────────────┤               │               │               │         │
│                                                                              │
│  ════════════════════════════════════════════════════════════════════════   │
│                                                                              │
│  Invitee                                                                    │
│    │                                                                         │
│    │  4. Click invitation link                                              │
│    │     /invite/accept?token=xxx                                           │
│    ├─────────────►│               │               │               │         │
│    │              │               │               │               │         │
│    │              │  5. GET /invitations/validate?token=xxx      │         │
│    │              ├──────────────►│               │               │         │
│    │              │◄──────────────┤               │               │         │
│    │              │  { email, name }              │               │         │
│    │              │               │               │               │         │
│    │  6. Show page│               │               │               │         │
│    │     "You are invited"                       │               │         │
│    │     [Set Password]                          │               │         │
│    │◄─────────────┤               │               │               │         │
│    │              │               │               │               │         │
│    │  7. Enter password           │               │               │         │
│    ├─────────────►│               │               │               │         │
│    │              │               │               │               │         │
│    │              │  8. POST /invitations/set-password           │         │
│    │              │     { token, password }       │               │         │
│    │              ├──────────────►│               │               │         │
│    │              │               │               │               │         │
│    │              │  9. Validate token            │               │         │
│    │              │     Create Authentik user     │               │         │
│    │              │     Set password              │               │         │
│    │              │     Generate 6-char PIN       │               │         │
│    │              │     Set status = pending_verification         │         │
│    │              │               │               │               │         │
│    │              │◄──────────────┤               │               │         │
│    │              │  { user_id, pin, pin_expires }│               │         │
│    │              │               │               │               │         │
│    │  10. Redirect to /invite/verify             │               │         │
│    │      Show PIN: "X7K2M9"                     │               │         │
│    │      [Send to WhatsApp]                     │               │         │
│    │◄─────────────┤               │               │               │         │
│    │              │               │               │               │         │
│    │  11. Click WhatsApp button   │               │               │         │
│    │      (wa.me/628xxx?text=X7K2M9)             │               │         │
│    ├─────────────────────────────────────────────────────────────►│         │
│    │              │               │               │               │         │
│    │              │               │               │  12. Receive  │         │
│    │              │               │               │      PIN msg  │         │
│    │              │               │◄──────────────┤◄──────────────┤         │
│    │              │               │               │               │         │
│    │              │               │  13. POST /invitations/verify-pin       │
│    │              │               │      { pin, phone }           │         │
│    │              │               │◄──────────────┤               │         │
│    │              │               │               │               │         │
│    │              │               │  14. Validate PIN             │         │
│    │              │               │      Set status = active      │         │
│    │              │               │      Store verification_phone │         │
│    │              │               │      Set verified_at          │         │
│    │              │               │               │               │         │
│    │              │               ├──────────────►│               │         │
│    │              │               │  { success }  │               │         │
│    │              │               │               │               │         │
│    │              │               │               │  15. Send     │         │
│    │              │               │               │  welcome msg  │         │
│    │              │               │               ├──────────────►│         │
│    │              │               │               │               │         │
│    │              │  16. Poll /verification-status (every 3s)    │         │
│    │              ├──────────────►│               │               │         │
│    │              │◄──────────────┤               │               │         │
│    │              │  { verified: true }           │               │         │
│    │              │               │               │               │         │
│    │  17. "Verified!"             │               │               │         │
│    │      Redirect to login       │               │               │         │
│    │◄─────────────┤               │               │               │         │
│    │              │               │               │               │         │
│    │  18. Login via Authentik     │               │               │         │
│    ├─────────────►│               │               │               │         │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

## API Endpoints

### Public Endpoints (No Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/invitations/validate?token=` | Validate invitation token, return user info |
| POST | `/api/v1/invitations/set-password` | Set password, create Authentik user, generate PIN |
| GET | `/api/v1/invitations/verification-status/:user_id` | Poll verification status |
| POST | `/api/v1/invitations/regenerate-pin/:user_id` | Generate new PIN if expired |
| POST | `/api/v1/invitations/verify-pin` | Verify PIN (called by chatbot) |

### Request/Response Examples

#### Validate Token
```bash
GET /api/v1/invitations/validate?token=abc123...

Response:
{
  "success": true,
  "data": {
    "email": "user@example.com",
    "name": "John Doe",
    "role": "member"
  }
}
```

#### Set Password
```bash
POST /api/v1/invitations/set-password
{
  "token": "abc123...",
  "password": "SecurePassword123!"
}

Response:
{
  "success": true,
  "data": {
    "user_id": "uuid-here",
    "email": "user@example.com",
    "pin": "X7K2M9",
    "pin_expires": "2026-01-27T20:00:00Z"
  }
}
```

#### Verify PIN (Chatbot)
```bash
POST /api/v1/invitations/verify-pin
{
  "pin": "X7K2M9",
  "phone": "628123456789"
}

Response:
{
  "success": true,
  "data": {
    "user_id": "uuid-here",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active"
  },
  "message": "User verified successfully"
}
```

## Database Schema

### Users Table Additions

```sql
ALTER TABLE users
    ADD COLUMN verification_pin VARCHAR(6),
    ADD COLUMN verification_pin_expires_at TIMESTAMPTZ,
    ADD COLUMN verification_phone VARCHAR(20),
    ADD COLUMN verified_at TIMESTAMPTZ;

CREATE UNIQUE INDEX idx_users_verification_pin
    ON users(verification_pin) WHERE verification_pin IS NOT NULL;
```

### User Status Flow

```
pending_invitation → pending_verification → active
       ↑                    ↑                  ↑
  Created by admin    Password set,      PIN verified
                      PIN generated       via WhatsApp
```

## PIN Generation

- 6 characters alphanumeric
- Uses uppercase letters A-Z (excluding confusing: I, O)
- Uses numbers 2-9 (excluding confusing: 0, 1)
- Valid for 15 minutes
- Can be regenerated if expired

```go
func generatePIN() string {
    const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
    result := make([]byte, 6)
    rand.Read(result)
    for i := range result {
        result[i] = chars[int(result[i])%len(chars)]
    }
    return string(result)
}
```

## Chatbot Integration

The chatbot (`dayawarga-chatbot`) intercepts 6-character alphanumeric messages and attempts PIN verification:

```typescript
// webhook.ts
function isInvitationPIN(message: string): boolean {
  if (message.length !== 6) return false;
  const pinPattern = /^[A-Z2-9]{6}$/i;
  return pinPattern.test(message);
}

// In handleWebhook:
if (message_body && isInvitationPIN(message_body.trim())) {
  const pin = message_body.trim().toUpperCase();
  const result = await verifyInvitationPIN(pin, phone);
  
  if (result.success) {
    await sendText(phone, `Selamat datang di Dayawarga, ${result.name}! ...`);
    return;
  }
  
  if (result.error === 'invalid PIN' || result.error === 'PIN has expired') {
    await sendText(phone, `PIN tidak valid atau sudah kedaluwarsa...`);
    return;
  }
}
```

## Frontend Pages

### `/invite/accept` - Accept Invitation Page
- Validates token on load
- Shows "You are invited to join Dayawarga"
- Password form (min 8 chars, with confirm)
- Submit → redirects to verify page

### `/invite/verify` - PIN Verification Page
- Displays 6-digit PIN prominently
- Copy PIN button
- WhatsApp button (opens wa.me link)
- Polls verification status every 3 seconds
- Shows success animation when verified
- Regenerate PIN button if expired
- Auto-redirect to login when verified

## Environment Variables

### API (.env)
```env
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=api
SMTP_PASSWORD=xxx
APP_BASE_URL=https://portal.dayawarga.com
```

### Admin Portal (.env)
```env
VITE_API_BASE_URL=https://api.dayawarga.com/api/v1
VITE_WHATSAPP_NUMBER=6281234567890
```

## Security Considerations

1. **PIN Expiration**: 15 minutes prevents replay attacks
2. **Single Use**: PIN is cleared after successful verification
3. **Phone Binding**: Phone number is recorded for audit
4. **Rate Limiting**: Standard API rate limits apply
5. **Token Security**: Invitation tokens are 64-char hex strings
