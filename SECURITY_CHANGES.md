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

## Phase 4: Audit Log Integrity & Access Control

### Verified Findings
1. **Unkeyed SHA-256 Hash Chain**: Previous audit log hash computation used unkeyed SHA-256 over row fields. Anyone with temporary write access to the database could recalculate valid forward hashes after modifying past records.
2. **Silent Genesis Fallback on Database Error**: When reading the latest hash, if an error occurred (`err != nil || h == ""`), the system silently substituted the genesis hash (`000...000`), breaking the cryptographic continuity without alerting operators.
3. **Audit Log Race Conditions Across Instances**: In-process mutexes did not protect against concurrent requests across multiple clustered application instances, allowing audit chain forks under high concurrent write loads.
4. **Missing Audit Detail Field**: Audit log records lacked a structured `detail` column to capture contextual parameters (such as targeted usernames, change summaries, or event payloads), limiting forensics.
5. **Username Normalization & Case/Homoglyph Collision Risks**: Usernames were not strictly normalized or validated against ASCII character sets, leaving room for Unicode homoglyphs, casing confusion, or whitespace spoofing in user management.
6. **Lack of Cryptographic Chain Verification Endpoint & Startup Gate**: No built-in verification existed to audit the integrity of the audit chain from genesis to head on system boot or on demand.

### Code Changes
- **Keyed HMAC-SHA256 Audit Chain (`crypto.go`, `audit.go`)**:
  - Implemented domain-separated audit key derivation: `keyAudit = deriveKey(aesKey, "HERASAT_AUDIT_LOG_CHAIN_HMAC_V1")`.
  - Replaced unkeyed SHA-256 with HMAC-SHA256 incorporating `detail` along with `prev_hash`, `username`, `title`, `action`, `target_report_id`, `ip_address`, and `created_at`.
  - Maintained backwards verification compatibility with pre-Phase-4 unkeyed SHA-256 rows via `computeLegacyAuditHash`.
- **Advisory Locking & Fail-Closed Integrity (`audit.go`)**:
  - Added PostgreSQL transaction-level advisory locking (`SELECT pg_advisory_xact_lock(14021105)`) to serialize all audit writes across all application instances.
  - Required `FOR UPDATE` lock on latest audit row; eliminated silent fallback to genesis hash on database query errors. Genesis hash is permitted ONLY when the table is genuinely empty (`sql.ErrNoRows`).
  - Unified `logAudit` and `logAuditTx` to execute through a shared transaction pipeline (`executeAuditLogTx`), guaranteeing identical semantics.
- **Audit Detail Column & Schema (`schema.sql`, `docker/init/01_schema.sql`, `db.go`, `admin_handlers.go`)**:
  - Added `detail TEXT DEFAULT ''` column to `audit_logs` table and auto-migration in `runMigrations()`.
  - Updated `AuditLogItem` and `GetAuditLogsHandler` to query and return `detail` and `entry_hash`.
  - Provided `logAuditWithDetail` and `logAuditTxWithDetail` for detailed event logging.
- **Least-Privilege Database Roles (`docker/init/03_least_privilege_roles.sql`, `schema.sql`)**:
  - Created least-privilege role setup for `herasat_app` (SELECT, INSERT, UPDATE, DELETE on application tables; strict REVOKE of UPDATE, DELETE, TRUNCATE on `audit_logs`) and `herasat_migrator`.
- **Username Normalization & Homoglyph Protection (`security.go`, `admin_handlers.go`, `admin_auth.go`)**:
  - Added `normalizeAndValidateUsername`: restricts admin usernames to lowercase ASCII alphanumeric characters, hyphens, dots, and underscores (`[a-z0-9_.-]`, length 3-32).
  - Integrated validation into `CreateUserHandler` and normalized usernames to lowercase in `LoginHandler`.
- **Cryptographic Audit Chain Verification (`audit.go`, `main.go`)**:
  - Implemented `VerifyAuditChain`: sequentially validates `prev_hash` chaining and verifies each row's HMAC (or legacy SHA-256).
  - Added `VerifyAuditChainOnStartup` on application boot; fails closed in strict mode (`HERASAT_AUDIT_STRICT=1`).
  - Added authenticated superuser endpoint `GET /api/admin/audit-logs/verify` returning real-time chain verification status.

### Operator Impact
- Audit records are cryptographically tamper-evident with HMAC-SHA256 tied to the server's master key.
- Database administrators can apply `docker/init/03_least_privilege_roles.sql` to restrict the web application user `herasat_app` from deleting or updating audit logs even if SQL injection occurs.
- Superusers can verify the cryptographic chain integrity at any time via `GET /api/admin/audit-logs/verify`.

### Tests & Verification
- `audit_security_test.go`:
  - `TestAuditKeyDerivation`: verifies 32-byte domain-separated HMAC key.
  - `TestAuditChainHMACComputation`: verifies HMAC calculation, 64-character hex length, and tamper-sensitivity to `detail`, `username`, and `prev_hash`.
  - `TestNormalizeAndValidateUsername`: verifies ASCII allowlist, length limits, case normalization, and rejection of spaces, symbols, and Cyrillic homoglyphs.
  - `TestAuditLegacyHashCompatibility`: verifies compatibility with historical SHA-256 records.
  - `TestVerifyAuditLogsHandler`: verifies superuser verification endpoint.
- `audit_fail_test.go`: verified fail-closed rollback when audit logging fails.
- Regression: `go test ./...` passed.

## Phase 5: Cryptography, Data Protection & Safe Operations

### Verified Findings
1. **Plaintext Tracking Codes in Database**: `tracking_code` was stored in plaintext as `VARCHAR(32) UNIQUE NOT NULL`. A database read breach exposed all tracking codes, allowing unauthorized parties to query whistleblower reports and follow-up threads.
2. **Missing Additional Authenticated Data (AAD) Context Binding**: AES-GCM ciphertexts lacked cryptographic binding to database columns and primary keys (`table:column:rowID`), allowing cut-and-paste ciphertext attacks between fields or rows.
3. **Ciphertext Size Exceeding Tracking Code Column Limits**: Encrypted tracking code strings (`v3:1:f:...`) exceeded the legacy `VARCHAR(32)` column constraint.
4. **Single-Admin Report Deletion Risk**: Full report deletion could be executed by a single compromised superuser account without dual authorization, risking destruction of whistleblowing evidence.
5. **No Granular Report Redaction**: Operators had only binary options (keep entire report or delete entire report), forcing full deletions when only specific fields or attachments needed sanitization.
6. **Unencrypted Backups & Process List Credential Leak**: Automated database backups produced unencrypted gzip SQL files on disk. Furthermore, `pg_dump` was invoked with `--dbname <full_conn_str>`, exposing plaintext database passwords to the system process list (`ps` / Task Manager).

### Code Changes
- **Keyed HMAC Tracking Code Lookup & At-Rest Encryption (`crypto.go`, `schema.sql`, `db.go`, `public_handlers.go`, `admin_handlers.go`)**:
  - Implemented `HashTrackingCode`: computes deterministic keyed HMAC-SHA256 (`TRACK:<code_text>`) using `GetAuditHMACKey()`.
  - Added `tracking_code_hash VARCHAR(64)` column and index `idx_reports_tracking_hash` on `reports`.
  - Upgraded `reports.tracking_code` column from `VARCHAR(32)` to `TEXT` to accommodate AES-GCM ciphertexts.
  - In `SubmitReportHandler`: generated tracking codes are hashed (`tracking_code_hash`) and encrypted at rest (`tracking_code`).
  - In `TrackReportHandler` and `SubmitFollowUpHandler`: lookup uses `tracking_code_hash = $1 OR tracking_code = $2`.
  - In `admin_handlers.go`: added `ResolveTrackingCode` helper to decrypt tracking codes for admin views and search queries match against `tracking_code_hash`.
  - In `db.go`: added auto-migration (`migrateTrackingCodes`) to encrypt existing plaintext tracking codes and backfill `tracking_code_hash`.
- **Ciphertext Context Binding & v3 AAD Encryption (`crypto.go`)**:
  - Added `EncryptWithAAD` and `DecryptWithAAD` for AES-GCM with Additional Authenticated Data.
  - Added `EncryptBoundField` and `DecryptBoundField` binding context as `table:column:rowID`.
  - Upgraded field encryption to prefix `v3:1:f:` and file path encryption to `v3:1:p:`, with backward-compatible decryption of `v2:`, `v1:`, and raw ciphertexts.
  - Upgraded `EncryptFileBytes` to header `HENC\x03`, preserving backward-compatible decryption for `HENC\x02` and `HENC\x01`.
- **Two-Person Report Deletion Workflow (`admin_handlers.go`, `schema.sql`, `db.go`)**:
  - Added `pending_deletions` table tracking `report_id`, `requested_by`, `reason`, `status`, and timestamps.
  - In `DeleteReportHandler`: when `HERASAT_TWO_PERSON_DELETION=true`, deletion requests return HTTP 202 Accepted. A second distinct superuser must approve the request within 24 hours to trigger execution. Requests auto-expire after 24 hours. Self-approval is strictly forbidden (HTTP 403).
- **Superuser Report Redaction Endpoint (`admin_handlers.go`, `main.go`)**:
  - Added `POST /api/admin/report/redact` (superuser only).
  - Enables granular redaction of `reporter_name`, `reporter_phone`, `description`, or specified attachment IDs.
  - Replaces redacted content with `[حذف امنیتی توسط <username> در <date>]`, physically deletes attachments from disk, and logs a comprehensive audit entry with redacted field names and operator reason.
- **Encrypted Backups & Safe Connection Strings (`backup.go`)**:
  - Added `GetBackupKey()` derived from master key via `deriveKey(aesKey, "HERASAT_BACKUP_ENCRYPTION_V1")`.
  - `PerformBackup`: creates a unified archive containing `database.sql` and `./secure_uploads/`, compresses with gzip, encrypts with AES-GCM using `GetBackupKey()` with AAD `HBAK_HEADER_V1`, prepends magic header `HBAK\x01`, and outputs `.hbak` files.
  - `RestoreBackup`: verifies `HBAK\x01` header and AES-GCM authentication tag, extracts archive entries safely (with path traversal containment), and restores database SQL.
  - `dumpViaPgDump`: parses connection strings to extract credentials and passes `PGPASSWORD` via environment variable, eliminating passwords from process table arguments.
  - Added `isDockerContainerRunning` and fallback logic to guarantee dump reliability across environments.

### Operator Impact
- Report tracking codes are encrypted in the database; index lookups operate via blind HMAC hashing, preventing correlation or compromise from database dumps.
- Operators can enforce two-person authorization for report deletions by setting `HERASAT_TWO_PERSON_DELETION=true`.
- Backups are encrypted with authenticated AES-GCM (`.hbak`); manual tampering or invalid keys are detected and rejected.
- Superusers can redact sensitive whistleblower information while maintaining the report shell and audit history.

### Tests & Verification
- `data_protection_test.go`:
  - `TestV3CiphertextWithAAD`: verified encryption/decryption with AAD and verified failure on mismatched AAD (cut-and-paste attack resistance) and tampered ciphertext.
  - `TestBoundFieldHelper`: verified field-level binding with `table:column:rowID`.
  - `TestTrackingCodeHMACAndResolution`: verified deterministic HMAC generation, distinct hashes for distinct codes, and resolution of encrypted codes.
  - `TestEncryptedBackupAndTamperDetection`: verified backup creation with `HBAK\x01` header, successful restore, and rejection of tampered backups.
  - `TestTwoPersonDeletionStatusValidation`: verified 24-hour expiration check and dual-control checks.
  - `TestRedactionMarkerFormat`: verified redaction marker format and roundtrip encryption.
- `crypto_test.go`: updated assertions for `v3:1:f:`, `v3:1:p:`, and `HENC\x03`.
- Full regression: `go test ./...` passed (100%).




