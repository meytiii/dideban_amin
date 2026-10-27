# Security & Reliability Remediation Log

This document tracks all verified security findings, code modifications, operator impacts, and test verifications applied to the *Dideban-e Amin* (Herasat Portal) codebase.

---

## Phase 0: Test Baseline & Context Types

### Verified Findings
- Untyped context string `"admin"` risked key collisions with external middleware and third-party context wrappers.
- Tests located in the `test/` subdirectory could not access unexported backend constructs (`AdminSession`, `getIP`), resulting in test build errors.

### Code Changes
- Relocated all unit test suites from `test/*_test.go` to the repository root `.`.
- Introduced typed context key `adminContextKeyType` and helpers `contextWithAdmin` and `adminFromContext` in `security.go`.
- Updated all admin handler calls, events, and test fixtures to retrieve admin identity via `adminFromContext(r)`.
- Removed stale `test/` directory and consolidated test instructions into root `README.md`.

### Operator Impact
- Pure internal refactor. No configuration changes required.

### Tests & Verification
- `go build ./...`
- `go test ./...`

---

## Phase 1: Authentication, Authorization & Session Hardening

### Verified Findings
1. **Hardcoded Seed Credentials Discrepancy**: `docker/init/02_seed.sql` seeded user `admin_it` with hardcoded Argon2id hash. Discrepancy identified: the seeded hash matched `"Herasat@2026Secure"` rather than the previously documented `@2026Secure`. Hardcoded credentials posed default-credential takeover risk.
2. **Denial-of-Service via Lockout**: Lockout was keyed solely by username, allowing unauthenticated attackers to lock legitimate administrator accounts by sending failed attempts.
3. **MFA Token Client Binding & Replay**: MFA pending tokens were not bound to client IP or User-Agent, and lacked single-use replay enforcement.
4. **TOTP Time Window Skew**: Verification accepted `[-2, 2]` steps (up to 75 seconds skew each direction), expanding the window of opportunity for intercepted codes.
5. **2FA Enforcement Bypass**: Superusers and enforced users could login and bypass 2FA setup without a restricted session boundary.
6. **Forced Password Change Server-Side Gap**: Sessions with `must_change_password` were not blocked at the middleware layer from accessing arbitrary admin API endpoints.
7. **SSE Stream Leaks & Re-Auth Absence**: SSE connections in `AdminEventsHandler` did not re-verify session status on heartbeat, leaking events to revoked or expired admin sessions, and global session revocation broadcast leaked usernames to other clients.
8. **User Enumeration & Timing Attacks**: Missing users did not execute an Argon2id computation, enabling attackers to distinguish valid from invalid usernames via response latency.

### Code Changes
- **Seed & Credential Removal**: Deleted `docker/init/02_seed.sql` and sanitized `README.md`. Implemented `bootstrapInitialAdmin` in `db.go` that generates a cryptographically secure 24-character password upon first boot if no superuser exists, or uses `HERASAT_INITIAL_ADMIN_PASSWORD`.
- **Compound Account Lockout**: Updated `AccountLockoutManager` (`IsLocked`, `RecordFailure`, `Reset`) to key lockouts by compound `(username, ip)`. Added `ProgressiveDelay` (250ms, 500ms, 1s, 2s).
- **Dummy Argon2id Timing Equalization**: Initialized `dummyArgon2Hash` in `admin_auth.go` and executed Argon2id verification for unknown usernames without writing lockout records for non-existent users.
- **Log Scrubbing**: Sanitized and truncated attempted usernames in failed login logs to prevent log injection and credential leakage.
- **MFA Token Hardening**: Bound `MFAPendingPayload` to client IP and User-Agent. Stored used nonces in `used_mfa_nonces` and memory store to ensure single-use replay protection.
- **TOTP Window Restriction**: Reduced TOTP verification tolerance from `[-2, 2]` steps to standard `[-1, 1]` steps (30-second window).
- **2FA & Password Change Restricted Sessions**: Added `Requires2FASetup` and `RequiresPasswordChange` to `AdminSession` and database `admin_sessions`. Enforced in `AdminAuthMiddleware` (403 Forbidden with `2fa_setup_required` or `password_change_required` for unauthorized routes).
- **Password Policy Modernization**: Implemented `ValidatePasswordPolicy` using `utf8.RuneCountInString` (min 12 runes), enforcing uppercase, symbols, and rejecting common defaults (`@2026Secure`, `admin`, `herasat`) and username containment.
- **SSE Stream Security**: Added `SetReadDeadline(time.Time{})` to `AdminEventsHandler`, added periodic session re-validation on every heartbeat, and scoped `BroadcastSessionRevoked` to the specific client session instead of global broadcast.

### Operator Impact
- On initial startup with an empty database, operators must either set `HERASAT_INITIAL_ADMIN_PASSWORD` in the environment or record the generated password output to stdout.
- Environment variable `HERASAT_REQUIRE_2FA=1` can be set to enforce 2FA globally.

### Tests & Verification
- `auth_hardening_test.go`: Added test cases for compound lockout by IP, bound MFA token verification and replay prevention, Unicode rune-based password policy, restricted session middleware enforcement, and initial admin password generation.
- Full regression suite: `go test ./...` passed.

---

## Phase 2: Deployment, Environment & Network Safety

### Verified Findings
1. **Unrestricted Admin Interface Access**: If `HERASAT_ADMIN_ALLOWED_CIDRS` was left empty, the application started without restriction and permitted admin panel and login access from any IP address without explicit operator approval.
2. **Reverse Proxy Spoofing & Header Forgery**: When running behind a reverse proxy (`HERASAT_COOKIE_SECURE=true` without TLS), an empty `HERASAT_TRUSTED_PROXIES` allowed untrusted clients to forge `X-Forwarded-For` or caused all requests to appear as the reverse proxy IP, undermining rate limiting, account lockouts, and IP whitelisting.
3. **Repeated Dynamic CIDR Parsing**: Network CIDR environment variables were parsed on every HTTP request, risking runtime panics, log flooding, or silent dropping of malformed entries.
4. **Silent ClamAV Inactivity**: When ClamAV antivirus daemon was disabled or unreachable, no prominent warning was logged, leaving operators unaware of uninspected file uploads.
5. **Insecure AES Master Key Fallback & Placeholders**: `crypto.go` contained a hardcoded fallback key (`"01234567890123456789012345678901"`) in `init()`, and `.env.example` placeholders were not rejected at startup.
6. **Deprecated Headers & TLS Configurations**: `X-XSS-Protection` header (deprecated and prone to browser-level side-channels) and `PreferServerCipherSuites: true` (deprecated in Go 1.17+) were present in production configurations. CSP was not aligned with `HERASAT_ALLOWED_ORIGINS`.
7. **Incomplete Docker Stack**: `docker-compose.yml` only defined Postgres without an application container or ClamAV daemon.

### Code Changes
- **Startup CIDR Validation & Network Caching (`network.go`)**:
  - Implemented `parseCIDRs` with strict syntax and bitmask validation, failing fast at startup on malformed entries.
  - Pre-parsed and cached `HERASAT_ADMIN_ALLOWED_CIDRS` and `HERASAT_TRUSTED_PROXIES` at startup; avoided per-request parsing.
  - Refused startup if `HERASAT_ADMIN_ALLOWED_CIDRS` is empty unless `HERASAT_ADMIN_ALLOW_ANY=true` or `HERASAT_DEV_MODE=1` is explicitly set.
  - Refused startup if `HERASAT_COOKIE_SECURE=true` without TLS certificates and `HERASAT_TRUSTED_PROXIES` is empty (unless `HERASAT_DEV_MODE=1`).
- **ClamAV Antivirus Daemon Hardening (`clamav.go`)**:
  - Added `Ping()` method using ClamAV's zero-delimited protocol (`zPING\x00` expecting `PONG`).
  - Added `CheckClamAVStartup`: enabled by default in production (`HERASAT_DEV_MODE != "1"`), logs prominent warnings if disabled or unreachable, and enforces fail-closed startup termination if `HERASAT_CLAMAV_REQUIRED=1`.
  - Added fallback host/port resolution for `CLAMAV_URL` / `CLAMAV_PORT`.
- **Cryptographic Master Key Enforcement (`crypto.go`)**:
  - Removed hardcoded fallback key in `crypto.go`'s `init()`.
  - Added `isKnownPlaceholderKey` rejecting known development keys (including `.env.example` placeholders and repeated-byte strings).
  - Configured `TestMain` in `test_main_test.go` to provide a dedicated test key for unit testing.
- **Security Headers & TLS Hardening (`main.go`)**:
  - Removed deprecated `X-XSS-Protection` header and deprecated `PreferServerCipherSuites: true`.
  - Aligned Content-Security-Policy `connect-src` and `form-action` with `HERASAT_ALLOWED_ORIGINS`.
  - Added `Vary: Origin` and enforced 403 Forbidden for preflight `OPTIONS` requests from unauthorized origins.
- **Production Containerization (`Dockerfile`, `docker-compose.yml`)**:
  - Created multi-stage `Dockerfile` (Go 1.24 alpine builder, Alpine 3.20 runtime, non-root user `herasat:herasat` UID/GID 10001).
  - Pinned `python3`, `py3-pip`, and `python-docx==1.1.2` in the container.
  - Updated `docker-compose.yml` with full multi-service stack: `app`, `postgres` (with healthcheck), and `clamav`.
- **Configuration Templates**:
  - Updated `.env.example` and `env.example` with all new deployment parameters (`HERASAT_ADMIN_ALLOW_ANY`, `HERASAT_CLAMAV_REQUIRED`, `HERASAT_DEV_MODE`, etc.).

### Operator Impact
- Production deployments must either specify `HERASAT_ADMIN_ALLOWED_CIDRS` or explicitly set `HERASAT_ADMIN_ALLOW_ANY=true`.
- Reverse proxy deployments (`HERASAT_COOKIE_SECURE=true` without local TLS) must declare `HERASAT_TRUSTED_PROXIES` with proxy IP/CIDRs.
- Placeholders such as `CHANGE_ME_TO_EXACTLY_32_BYTES_!` for `HERASAT_AES_KEY` will halt startup with an actionable error.
- ClamAV is enabled by default in production mode; to require daemon availability before serving requests, operators can set `HERASAT_CLAMAV_REQUIRED=1`.

### Tests & Verification
- `deployment_security_test.go`:
  - `TestParseCIDRs`: valid CIDRs, bare IPs, IPv6, and malformed CIDR rejection.
  - `TestAdminAllowedCIDRsStartupRefusal`: startup failure when CIDRs empty, bypass verification via `ALLOW_ANY` or `DEV_MODE`.
  - `TestReverseProxyCookieSecureStartupRefusal`: startup refusal when cookie secure without TLS and empty trusted proxies.
  - `TestIsKnownPlaceholderKey`: rejection of example placeholders and acceptance of valid 32-byte keys.
  - `TestSecureHeadersMiddlewareHardening`: verification of header removals, CSP origin alignment, and CORS preflight rejection.
  - `TestAdminIPWhitelistMiddlewareCached`: verification of admin IP whitelisting using pre-cached CIDR data.
- `clamav_test.go`:
  - `TestClamAVPingAndStartupChecks`: daemon pinging, disabled bypass, and startup failure with `HERASAT_CLAMAV_REQUIRED=1`.
- Regression: `go test ./...` passed.

## Phase 3: Upload Pipeline & Content Security

### Verified Findings
1. **Decompression Bomb Vulnerability in Images**: Previously, image upload resizing parsed pixel streams without decoding headers first, allowing small files to exhaust system RAM upon decompression.
2. **Missing Active Content Inspection in PDFs**: Active script constructs (`/JavaScript`, `/JS`, `/Launch`, `/EmbeddedFiles`, `/EmbeddedFile`, `/OpenAction`, `/AA`, `/SubmitForm`, `/ImportData`, `/RichMedia`, `/GoToR`) as well as encrypted payloads (`/Encrypt`) were not comprehensively checked, including within compressed `/FlateDecode` streams and hex-encoded keywords (`#xx`).
3. **Sensitive Metadata in PDFs and Office Documents**: PDFs retained `/Author`, `/Creator`, `/Producer`, `/CreationDate`, and `/ModDate`. Word/Excel OOXML files retained author details in `docProps/core.xml` and company properties in `docProps/custom.xml`.
4. **Dangerous Archive & OOXML Macro Vectors**: Embedded macro bytecode (`vbaProject.bin`, `.bin`) and external template references (`TargetMode="External"`) could be embedded inside `.docx` or `.xlsx` files. Archive uploads lacked decompression size checking against actual bytes read.
5. **Path Traversal Boundary Vulnerability**: Sibling directory traversal was possible in attachment retrieval and report deletion due to `strings.HasPrefix(absFilePath, absUploadDir)` matching folders sharing the same prefix name without a path separator boundary.
6. **Captcha Brute-Force & Token Reuse**: Captcha tokens did not enforce attempt limits on failed answers, allowing brute-force attempts on the math answer range before expiration.
7. **Whistleblower IP Disclosure in Audit Logs**: Public report submissions and follow-ups logged reporter client IPs in the audit trail, violating anonymity guarantees.

### Code Changes
- **Image Bomb Protection (`metadata.go`)**:
  - Implemented `jpeg.DecodeConfig` and `png.DecodeConfig` checks enforcing max 40 megapixels and max 10,000x10,000 pixel dimensions before full memory allocation.
- **Active PDF Inspection & Stripping (`metadata.go`)**:
  - Implemented `ScanPDFForActiveContent`: rejects encrypted PDFs (`/Encrypt`), scans uncompressed and decompressing `/FlateDecode` object streams for dangerous keywords, and decodes hex-encoded PDF name literals (`#xx`).
  - Added `SanitizePDF`: strips `/Author`, `/Creator`, `/Producer`, `/CreationDate`, and `/ModDate` fields from `/Info` dictionaries.
- **OOXML Security & Sanitization (`metadata.go`)**:
  - Added `InspectOOXML`: rejects macro streams (`vbaProject.bin`, `.bin`, `externalLink`, `oleObject`) and external template references.
  - Added `SanitizeOOXML`: replaces `docProps/core.xml` and `docProps/app.xml` with sanitized minimal XML without author or creation timestamps, and deletes `docProps/custom.xml`.
- **Audio File Sanitization (`metadata.go`)**:
  - Added `SanitizeMP3`: strips ID3v2 headers and ID3v1 trailing tags.
  - Added `SanitizeWAV`: parses RIFF chunks, retaining only valid `fmt ` and `data` chunks while dropping metadata chunks such as `LIST INFO`.
- **Archive Security (`public_handlers.go`)**:
  - Added `inspectZipArchive`: verifies actual uncompressed bytes do not exceed 50MB, bounds maximum archive entries to 1000, rejects encrypted zip entries, bans nested dangerous file extensions, and aborts on suspicious decompression ratios (>20:1 above 512KB).
- **Public Handler Protections (`public_handlers.go`)**:
  - Added `http.ResponseController` read deadline (5 minutes) for multi-part file uploads.
  - Added `defer r.MultipartForm.RemoveAll()` for guaranteed disk cleanup of multipart temporary files.
  - Enforced max 5 attachments limit and 64-character category allow-list validation.
  - Content moderation: supported `HERASAT_PROFANITY_ACTION=flag` (default) which elevates report priority to high while preserving whistleblower evidence, and `HERASAT_PROFANITY_ACTION=reject` for strict blocking.
  - Scrubbed client IP from public reporter audit logs to pass `"anonymous"`.
- **Path Traversal Boundary Hardening (`security.go`, `admin_handlers.go`)**:
  - Added `isWithinDirectory` ensuring paths are within the parent directory with proper path separator boundaries.
  - Updated attachment download and report deletion handlers to use `isWithinDirectory`.
- **Captcha Brute-Force Hardening (`captcha.go`)**:
  - Expanded math operands (`[25, 95]` and `[10, 45]`) to expand the search space.
  - Added `RecordFailure` and attempt tracking; burns captcha token after repeated failed attempts.
- **Frontend File Picker Alignment (`frontend/public/index.html`)**:
  - Removed `.rar` and `application/x-rar-compressed` from file input accept attributes.

### Operator Impact
- Uploaded files are strictly validated against magic bytes; unsupported file types (`.rar`, `.mp4`, `.ogg`) and active PDFs or macro-enabled documents will be rejected with clear Persian error messages.
- Operators can control vulgarity handling via `HERASAT_PROFANITY_ACTION=flag` (default) or `HERASAT_PROFANITY_ACTION=reject`.
- Public reporter logs record `"anonymous"` for IP addresses to guarantee whistleblower safety.

### Tests & Verification
- `uploads_security_test.go`:
  - `TestJPEGDecompressionBomb` and `TestPNGDecompressionBomb`: verified rejection of oversized dimensions.
  - `TestPDFActiveContentRejection`: verified rejection of encrypted PDFs, direct JavaScript, hex-encoded keywords, and `/FlateDecode` streams with launch commands.
  - `TestPDFMetadataSanitization`: verified stripping of Author and Creator metadata.
  - `TestOOXMLMacroAndTemplateRejection`: verified rejection of `vbaProject.bin`, external templates, and sanitization of `docProps/core.xml`.
  - `TestZipArchiveInspection`: verified rejection of nested executables, encrypted entries, and zip bomb ratios.
  - `TestAudioSanitization`: verified removal of ID3v2/ID3v1 tags in MP3 and non-audio chunks in WAV.
  - `TestValidateBinaryMagicBytes`: verified allowed and rejected file formats.
  - `TestCaptchaBruteForceProtection` and `TestCaptchaSingleUseRedemption`: verified failure burning and replay prevention.
  - `TestIsWithinDirectory`: verified directory boundary containment and rejection of sibling directory traversal.
  - `TestCategoryValidationAllowList`: verified category allow-list validation.
- Regression: `go test ./...` passed.


