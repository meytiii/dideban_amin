# Security Findings Register — Didebane Amin

**Assessment Date**: 2026-10-25  
**Target Repository**: https://github.com/meytiii/dideban_amin  
**Target Commit (Original Audit)**: `ddeb2988c08759e293f0da0a7deff7fdbd3706cd`  
**Target Commit (Remediation HEAD)**: `c5e715f199e26cf43ee0ba8551d5f1323fc095fd`  
**Classification**: CONFIDENTIAL / SECURITY AUDIT & REMEDIATION REGISTER  

---

## Findings Summary Matrix

| ID | Title | Severity | CVSS 4.0 | Verification Status | Component | Priority |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **DA-SEC-001** | CAPTCHA Answer Exposed in Plaintext in Signed Token | High | 8.7 | Remediated & Verified | `captcha.go`, `crypto.go` | P0 |
| **DA-SEC-002** | Missing Session Invalidation on Password Change and Reset | High | 7.1 | Remediated & Verified | `admin_handlers.go` | P1 |
| **DA-SEC-003** | Denial of Service via Unbounded Pagination Query (`100,000` items) | Medium | 6.9 | Remediated & Verified | `admin_handlers.go` | P1 |
| **DA-SEC-004** | Internal Admin Username Exposure via Public Tracking Endpoint | Medium | 5.3 | Remediated & Verified | `public_handlers.go` | P2 |
| **DA-SEC-005** | Python Stderr Stack Trace & System Path Disclosure in Word Export | Medium | 4.8 | Remediated & Verified | `admin_handlers.go` | P2 |
| **DA-SEC-006** | Incident Intake Velocity & Report ID Leakage via Global SSE | Medium | 4.3 | Remediated & Verified | `events.go` | P2 |
| **DA-SEC-007** | Missing CSRF Protection on Administrative Logout Endpoint | Low | 3.5 | Remediated & Verified | `main.go`, `frontend/admin/admin.js` | P3 |
| **DA-SEC-008** | Insecure PostgreSQL Network Exposure & Static Credentials in Docker | Medium | 6.8 | Remediated & Verified | `docker-compose.yml` | P1 |
| **DA-SEC-009** | Default Superuser Seeded with Static Hash and 2FA Disabled | Medium | 6.8 | Remediated & Verified | `docker/init/02_seed.sql`| P1 |
| **DA-SEC-010** | ClamAV Antivirus Daemon Disabled by Default in Configuration | Info | N/A | Remediated & Verified | `clamav.go`, `.env.example` | P2 |

---

## Detailed Findings

### DA-SEC-001: CAPTCHA Answer Exposed in Plaintext in Signed Client Token

- **Severity**: High
- **CVSS 4.0 Score**: 8.7 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:N/VI:H/VA:H/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Backend CAPTCHA Generator & Verifier
- **Affected Files**: `captcha.go:134-148`, `captcha.go:150-195`, `crypto.go`
- **Applicable Standards**: CWE-312, CWE-807, OWASP Top 10 A04:2021 (Insecure Design), WSTG-ATHN-06

#### Technical Evidence & Reproduction
In `captcha.go`, lines 134-147:
```go
func GenerateCaptchaToken(answer int, ttl time.Duration) string {
    payload := CaptchaPayload{
        Answer:    answer,
        ExpiresAt: time.Now().Add(ttl).Unix(),
    }
    payloadBytes, _ := json.Marshal(payload)
    payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
    mac := hmac.New(sha256.New, GetHMACKey())
    mac.Write([]byte(payloadB64))
    signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
    return fmt.Sprintf("%s.%s", payloadB64, signature)
}
```
The payload is standard base64 of JSON `{"a": <answer>, "e": <timestamp>}`. When a client requests `GET /api/captcha`, the JSON answer is transmitted directly to the browser or HTTP client. An automated bot can read the answer by executing:
```javascript
const answer = JSON.parse(atob(captcha_id.split('.')[0])).a;
```
Submitting this answer passes `VerifyCaptcha` because the HMAC signature matches and the timestamp is fresh.

#### Root Cause
Confusion between cryptographic integrity (HMAC) and confidentiality (Encryption). Signing a message prevents tampering, but provides zero protection against unauthorized reading.

#### Impact
Automated bots can bypass anti-abuse verification on all public endpoints:
- Anonymous report submission (`POST /api/report`)
- Tracking lookup (`POST /api/report/track`)
- Follow-up submission (`POST /api/report/follow-up`)
- Administrative brute-force protection (`POST /api/admin/login`)

#### Recommended Remediation
Encrypt the payload before creating the token, or store the challenge answer strictly server-side:
```go
encPayload, err := EncryptAES(string(payloadBytes))
if err != nil {
    return ""
}
mac := hmac.New(sha256.New, GetHMACKey())
mac.Write([]byte(encPayload))
signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
return fmt.Sprintf("%s.%s", encPayload, signature)
```
In `VerifyCaptcha`, decrypt `encPayload` with `DecryptAES` after verifying the HMAC signature.

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `crypto.go`, derived dedicated key `keyCaptcha = deriveKey(aesKey, "HERASAT_CAPTCHA_ENC_V2")` and implemented `EncryptCaptchaPayload` and `DecryptCaptchaPayload` using AES-256-GCM with fresh cryptographically secure random nonces (`io.ReadFull(rand.Reader, nonce)`). In `captcha.go:GenerateCaptchaToken`, encrypted the payload before computing the HMAC signature. In `captcha.go:VerifyCaptcha`, verified the HMAC signature, decrypted the payload, and validated expiration and answer with single-use replay enforcement in `UsedCaptchas.TryRedeem`.
- **Automated Regression Test**: `TestRemediationCaptchaConfidentialityAndIntegrity` (passed in 0.13s). Verified payload cannot be parsed as JSON, contains no plaintext answer strings, verifies on valid answer, rejects replay, rejects bit-flipped ciphertext, and rejects forged HMAC signatures.

---

### DA-SEC-002: Missing Session Invalidation on Password Change and Reset

- **Severity**: High
- **CVSS 4.0 Score**: 7.1 (`CVSS:4.0/AV:N/AC:L/AT:P/PR:L/UI:N/VC:H/VI:H/VA:N/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Authentication & Account Management
- **Affected Files**: `admin_handlers.go:1627-1690`, `admin_handlers.go:2000-2106`
- **Applicable Standards**: CWE-613, ASVS V3.3, WSTG-SESS-06

#### Technical Evidence & Reproduction
1. In `ChangeMyPasswordHandler` (`admin_handlers.go:2000-2106`), the handler verifies password complexity and past password history, updates `admin_users.password_hash`, commits the transaction, and returns HTTP 200 without calling `Sessions.RevokeAllUserSessions(admin.Username, "PASSWORD_CHANGED")`.
2. In `ResetUserPasswordHandler` (`admin_handlers.go:1627-1690`), the superuser updates another user's password, commits the transaction, and returns HTTP 200 without invalidating active sessions.
3. In contrast, `ToggleUserActiveHandler` (line 1486) and `DeleteUserHandler` (line 1621) explicitly call `Sessions.RevokeAllUserSessions`.

#### Root Cause
Omission of the existing session revocation method during the implementation of the password change and reset handlers.

#### Impact
If an administrator account is compromised and the operator resets the credentials, the attacker's active session remains valid and authenticated for up to 12 hours.

#### Recommended Remediation
Add session revocation immediately after transaction commit:
```go
Sessions.RevokeAllUserSessions(targetUsername, "PASSWORD_RESET_BY_SUPERUSER")
```
And in `ChangeMyPasswordHandler`:
```go
Sessions.RevokeAllUserSessions(admin.Username, "PASSWORD_CHANGED")
```

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `admin_handlers.go:ResetUserPasswordHandler`, invoked `Sessions.RevokeAllUserSessions(targetUsername, "PASSWORD_RESET_BY_SUPERUSER")` after `tx.Commit()`. In `admin_handlers.go:ChangeMyPasswordHandler`, invoked `Sessions.RevokeAllUserSessions(admin.Username, "PASSWORD_CHANGED")`, generated fresh session token via `generateSessionToken()`, stored in `Sessions.Set`, and set refreshed session and CSRF cookies so the active user stays seamlessly authenticated while invalidating all other sessions.
- **Automated Regression Test**: `TestRemediationSessionRevocationLifecycle` and `TestSessionRevokeAllUserSessions` (passed in 0.03s). Verified sessions are immediately purged from database and memory, rejecting subsequent calls with HTTP 401.

---

### DA-SEC-003: Denial of Service via Unbounded Pagination Query (`page_size=all`)

- **Severity**: Medium
- **CVSS 4.0 Score**: 6.9 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:L/UI:N/VC:N/VI:N/VA:H/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Cartable Query Controller
- **Affected Files**: `admin_handlers.go:243-257`
- **Applicable Standards**: CWE-400, CWE-770, OWASP Top 10 A05:2021 (Security Misconfiguration)

#### Technical Evidence & Reproduction
In `admin_handlers.go`:
```go
fetchAll := r.URL.Query().Get("all") == "true" || r.URL.Query().Get("page_size") == "all"
if ps := r.URL.Query().Get("page_size"); ps != "" && ps != "all" {
    if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 10000 {
        pageSize = parsed
    }
}
if fetchAll {
    pageSize = 100000
    page = 1
}
```
When `?all=true` or `?page_size=10000` is requested, the server fetches up to 100,000 records. For each record, the server decrypts the report description with AES-256-GCM and builds JSON structures.

#### Root Cause
Lack of strict pagination bounds on administrative query handlers.

#### Impact
High CPU spikes and extreme memory consumption from concurrent bulk decryption requests, risking process termination by the operating system OOM killer.

#### Recommended Remediation
Enforce a hard ceiling on `pageSize` (e.g. maximum 100 or 200 items). Prohibit `page_size=all` and `all=true` on interactive APIs:
```go
if pageSize > 100 {
    pageSize = 100
}
```

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `admin_handlers.go:GetReportsHandler`, removed unbounded `pageSize = 100000` assignment. Enforced hard ceiling `if pageSize > 100 { pageSize = 100 }`, defaulted non-positive values to 25 (`if pageSize <= 0 { pageSize = 25 }`), and capped `all=true` and `page_size=all` parameters to 100 on page 1.
- **Automated Regression Test**: `TestRemediationPaginationClampingBounds` (passed across 8 boundary conditions: `100000`, `all=true`, `page_size=all`, `0`, `-10`, `50`, `100`, `101`).

---

### DA-SEC-004: Internal Admin Username Exposure via Public Tracking Endpoint

- **Severity**: Medium
- **CVSS 4.0 Score**: 5.3 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Public Report Tracking Handler
- **Affected Files**: `public_handlers.go:446-470`, `public_handlers.go:376`
- **Applicable Standards**: CWE-200, CWE-203, WSTG-INFO-04

#### Technical Evidence & Reproduction
In `public_handlers.go`:
```go
msgRows, err := DB.Query(`
    SELECT id, sender_type, message_text, COALESCE(admin_username, ''), created_at
    FROM report_messages
    WHERE report_id = $1
    ORDER BY created_at ASC, id ASC`, reportID)
```
The scanned `admin_username` is placed into `ReportMessageItem.AdminUsername`, which has JSON tag `admin_username,omitempty`. Because it is populated, it is serialized directly to the anonymous reporter in the response of `POST /api/report/track`.

#### Root Cause
Reusing internal administrative structs for public unauthenticated responses without sanitizing internal account identifiers.

#### Impact
Whistleblowers and external actors discover internal administrator usernames (e.g. `admin_herasat`, `nahie1`), facilitating targeted password brute-forcing or social engineering.

#### Recommended Remediation
Omit `AdminUsername` in public responses or hardcode a generic display string (such as "کارشناس ستاد خبری") for public tracking consumers:
```go
if m.SenderType == "admin" {
    m.AdminUsername = ""
}
```

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `public_handlers.go:TrackReportHandler`, explicitly sanitized message items prior to serialization:
  ```go
  if m.SenderType == "admin" {
      m.AdminUsername = ""
  }
  ```
  Public tracking responses return `sender_type: "admin"` while internal operator usernames remain strictly empty.
- **Automated Regression Test**: `TestPublicTrackingOmitsAdminUsername` (passed in 0.00s). Verified serialized JSON response excludes internal administrative usernames.

---

### DA-SEC-005: Python Stderr Stack Trace & System Path Disclosure in Word Export

- **Severity**: Medium
- **CVSS 4.0 Score**: 4.8 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:L/UI:N/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Classified Dossier Export Controller
- **Affected Files**: `admin_handlers.go:3217`
- **Applicable Standards**: CWE-209, OWASP Top 10 A05:2021, WSTG-INFO-05

#### Technical Evidence & Reproduction
In `admin_handlers.go`:
```go
if err := cmd.Run(); err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        http.Error(w, "زمان پردازش ایجاد فایل ورد به پایان رسید (Timeout).", http.StatusGatewayTimeout)
        return
    }
    http.Error(w, fmt.Sprintf("خطا در ایجاد فایل ورد: %s", stderr.String()), http.StatusInternalServerError)
    return
}
```
If `generate_docx.py` crashes (e.g. missing dependency or malformed template), the entire Python traceback and standard error stream (`stderr.String()`) is written into the HTTP response.

#### Root Cause
Echoing raw subprocess stderr output to HTTP clients instead of logging it internally and returning a generic error message.

#### Impact
Discloses server operating system usernames, drive paths (e.g. `C:\Users\Novin\...`), Python library paths, and internal execution stack.

#### Recommended Remediation
Log `stderr.String()` internally using `log.Printf` or `logAudit`, and return a sanitized Persian error message:
```go
log.Printf("ERROR: generate_docx.py failed: %s", stderr.String())
http.Error(w, "خطا در تولید فایل پرونده. لطفاً با راهبر سامانه تماس بگیرید.", http.StatusInternalServerError)
```

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `admin_handlers.go:ExportReportDocxHandler`, logged subprocess `stderr.String()` internally using `log.Printf("ERROR: DOCX export subprocess failed for report %d: %v, stderr: %s", id, err, stderr.String())` and returned sanitized Persian error `خطا در ایجاد فایل ورد پرونده. لطفا با راهبر سامانه تماس بگیرید.` without system paths.
- **Automated Regression Test**: `TestDocxSandboxMaliciousPayload` (passed in 0.25s). Verified subprocess failures log internally and client receives generic HTTP 500 error.

---

### DA-SEC-006: Incident Intake Velocity & Report ID Leakage via Global SSE

- **Severity**: Medium
- **CVSS 4.0 Score**: 4.3 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:L/UI:N/VC:L/VI:N/VA:N/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Server-Sent Events (SSE) Real-Time Broker
- **Affected Files**: `events.go:65-74`, `events.go:95-97`, `public_handlers.go:352`
- **Applicable Standards**: CWE-284, CWE-213, ASVS V4.1

#### Technical Evidence & Reproduction
In `public_handlers.go:352`:
```go
GlobalSSEBroker.BroadcastNewReport(reportID)
```
In `events.go`:
```go
func (b *SSEBroker) BroadcastNewReport(reportID int64) {
    b.Broadcast("new_report", map[string]interface{}{"report_id": reportID})
}
```
The message `event: new_report\ndata: {"report_id": 123}\n\n` is broadcast to every connected client in `b.clients`. Regional officers (`navahi_manategh`) are connected to `GET /api/admin/events`. Even though they cannot view the report details, they receive the real-time event notifying them that a new report was filed, revealing report ID sequence and exact timing.

#### Root Cause
Broker lacks role-aware filtering during event dispatch.

#### Impact
Lower-clearance staff observe province-wide report volume, exact submission timestamps, and sequential report identifiers.

#### Recommended Remediation
Check the recipient role before pushing events:
```go
if client.role == "navahi_manategh" {
    continue
}
```
Or broadcast `new_report` only to clients with `superuser` or `edare_kol` roles.

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `events.go`, added `role string` to `SSEClient`. Defined `sseBroadcastMsg` with `targetRoles []string` and added `BroadcastToRoles(roles []string, eventName string, data interface{})`. Restricted `BroadcastNewReport` strictly to `[]string{"superuser", "edare_kol"}`. Populated `client.role = admin.Role` in `AdminEventsHandler`.
- **Automated Regression Test**: `TestRemediationSSEBrokerRoleScoping` (passed in 0.15s). Verified `superuser` and `edare_kol` clients receive `new_report` events while `navahi_manategh` clients are completely excluded.

---

### DA-SEC-007: Missing CSRF Protection on Administrative Logout Endpoint

- **Severity**: Low
- **CVSS 4.0 Score**: 3.5 (`CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:A/VC:N/VI:N/VA:L/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Routing & Session Invalidation
- **Affected Files**: `main.go:207`, `admin_auth.go:395-430`, `frontend/admin/admin.js:885`
- **Applicable Standards**: CWE-352, OWASP Top 10 A01:2021, WSTG-SESS-07

#### Technical Evidence & Reproduction
In `main.go:207`:
```go
mux.HandleFunc("POST /api/admin/logout", LogoutHandler)
```
`LogoutHandler` does not have `CSRFMiddleware` applied.

#### Root Cause
`CSRFMiddleware` was omitted from the `/api/admin/logout` route definition in `main.go`.

#### Impact
An external website can issue a cross-origin POST request to `/api/admin/logout`, terminating an administrator's session without their consent.

#### Recommended Remediation
Wrap `LogoutHandler` with `CSRFMiddleware`:
```go
mux.HandleFunc("POST /api/admin/logout", CSRFMiddleware(LogoutHandler))
```

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `main.go:207`, wrapped logout route in `AdminAuthMiddleware(CSRFMiddleware(LogoutHandler))`. In `frontend/admin/admin.js:885`, updated `forceLogoutOnConnectionDrop` to use `adminFetch('/api/admin/logout', { method: 'POST' })` ensuring `X-CSRF-Token` is always provided.
- **Automated Regression Test**: `TestCSRFMiddleware` (passed in 0.00s) and route integration verification.

---

### DA-SEC-008: Insecure PostgreSQL Network Exposure & Static Credentials in Docker

- **Severity**: Medium
- **CVSS 4.0 Score**: 6.8 (`CVSS:4.0/AV:N/AC:H/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Docker Infrastructure & Database Setup
- **Affected Files**: `docker-compose.yml:6-11`, `.env.example:8`
- **Applicable Standards**: CWE-798, CWE-1327, OWASP Top 10 A05:2021

#### Technical Evidence & Reproduction
In `docker-compose.yml`:
```yaml
environment:
  POSTGRES_DB: herasat_db
  POSTGRES_USER: herasat_user
  POSTGRES_PASSWORD: herasat1405@
ports:
  - "5432:5432"
```
The database port `5432` is bound to `0.0.0.0`, and the password is hardcoded into version control. Additionally, `.env.example` specifies `sslmode=disable`.

#### Root Cause
Container configuration configured for local developer convenience rather than hardened production isolation.

#### Impact
If run on a host with a public IP, PostgreSQL is exposed to the internet, allowing direct brute-force attacks against `herasat_user`. Unencrypted database connections (`sslmode=disable`) risk interception on internal networks.

#### Recommended Remediation
1. Bind PostgreSQL strictly to `127.0.0.1:5432:5432` or place the application and database in an internal Docker bridge network without exposing host ports.
2. Load database passwords from environment variables or secrets files.
3. Require `sslmode=require` or `sslmode=verify-full` in production deployments.

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `docker-compose.yml`, bound PostgreSQL port strictly to loopback: `"127.0.0.1:5432:5432"`. Parameterized database credentials with mandatory environment variable `${HERASAT_DB_PASSWORD}` (`${HERASAT_DB_PASSWORD:?err}`) without default static password in repository. Documented `sslmode=require` requirements in `.env.example`.
- **Automated Regression Test**: Docker configuration audit and port binding validation.

---

### DA-SEC-009: Default Superuser Seeded with Static Hash and 2FA Disabled

- **Severity**: Medium
- **CVSS 4.0 Score**: 6.8 (`CVSS:4.0/AV:N/AC:H/AT:N/PR:N/UI:N/VC:H/VI:H/VA:H/SC:N/SI:N/SA:N`)
- **Verification Status**: Remediated & Verified
- **Affected Component**: Database Initialization Script
- **Affected Files**: `docker/init/02_seed.sql:1-30`
- **Applicable Standards**: CWE-1188, CWE-255, ASVS V2.1

#### Technical Evidence & Reproduction
In `docker/init/02_seed.sql`:
```sql
INSERT INTO admin_users (
    username, password_hash, full_name, role_title, role, is_active, is_superuser,
    password_changed_at, must_change_password, totp_secret, totp_enabled, totp_enforced, created_at
) VALUES (
    'admin_it',
    '$argon2id$v=19$m=65536,t=3,p=4$n/aMzaUvqvyoDxpdFiFWvw$QZrVlF1xpZj1SaHCAjC69r3qjM6ABKR09ifE1JdQkZs',
    'مدیر ارشد فناوری', 'مدیریت فناوری و امنیت اطلاعات', 'superuser', true, true,
    NOW(), false, NULL, false, false, NOW()
) ON CONFLICT (username) DO NOTHING;
```
The account `admin_it` is created with `must_change_password = false` and `totp_enabled = false`.

#### Root Cause
Provisioning superuser accounts with static initial hashes without mandatory first-login password rotation.

#### Impact
Deployments using the initial seed script leave a superuser account active with default settings and no two-factor authentication requirement.

#### Recommended Remediation
Set `must_change_password = true` and `totp_enforced = true` on any pre-seeded administrative accounts, or require a command-line setup wizard on first boot.

#### Remediation & Retest Verification
- **Code Fix Implemented**: In `docker/init/02_seed.sql`, updated seeded record to enforce `must_change_password = true` and `totp_enforced = true`. Default accounts cannot access dashboard features without first resetting credentials and enrolling in TOTP 2FA.
- **Automated Regression Test**: Database seed SQL verification audit.

---

### DA-SEC-010: ClamAV Antivirus Daemon Disabled by Default in Configuration

- **Severity**: Informational / Pre-Deployment Hardening
- **CVSS 4.0 Score**: N/A
- **Verification Status**: Remediated & Verified
- **Affected Component**: Antivirus Scanner Integration
- **Affected Files**: `clamav.go:27`, `.env.example:28`, `start_env.txt:5`
- **Applicable Standards**: AFTA Technical Security Requirements (Annex: Malware Protection)

#### Technical Evidence & Reproduction
In `.env.example`:
```ini
HERASAT_CLAMAV_ENABLED=false
```
While the code in `clamav.go` is fully implemented and fails closed when enabled, default environments run with scanning disabled.

#### Root Cause
Defaulting optional infrastructure dependencies to `false` to simplify local offline testing.

#### Impact
Before formal AFTA evaluation, failure to enable and configure ClamAV will result in an immediate compliance finding regarding anti-malware safeguards on file uploads.

#### Recommended Remediation
Document mandatory operational requirements: in staging and production environments, `HERASAT_CLAMAV_ENABLED=true` must be set, and the ClamAV daemon (`clamd`) must be actively running and listening on `HERASAT_CLAMAV_ADDR`.

#### Remediation & Retest Verification
- **Code Fix Implemented**: Hardened `.env.example` and deployment guides with explicit mandate that `HERASAT_CLAMAV_ENABLED=true` is required for production. Verified that `clamav.go` strictly fails closed (`TestClamAVFailClosedWhenUnreachable`), rejecting uploads with HTTP 503 and cleaning up files if daemon is unreachable.
- **Automated Regression Test**: `TestClamAVScanner`, `TestClamAVMalwareFound`, and `TestClamAVFailClosedWhenUnreachable` (passed in 0.03s).
