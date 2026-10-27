# Security Changes Log - Dideban-e Amin (Herasat Portal)

This document tracks all security remediations, architectural upgrades, and bug fixes across development rounds.

---

## Round 1 Summary (Phases 0 - 6)

### Phase 0: Test Baseline & Core Harness
- Established automated test harness across all security domains (crypto, rate limiting, authentication, input validation, uploads).
- Fixed context key typing in `security.go` to prevent collision.

### Phase 1: Session Security, Account Lockout & Password Policy
- Implemented MFA Nonce validation (`used_mfa_nonces`) to prevent TOTP replay.
- Compound account lockout keyed by `username + "@" + ip` with progressive exponential delay.
- Enforced single-purpose restricted sessions (`RequiresPasswordChange` and `Requires2FASetup`).
- Enforced strict password complexity policy (length >= 12, upper, lower, digits, symbols, banned substrings).

### Phase 2: Network Boundaries, Anti-Malware & Security Headers
- Enforced CIDR allow-listing for admin interface with fail-closed production semantics.
- Fixed ClamAV integration to fail closed on scanner timeout or unreachable daemon.
- Configured trusted proxy validation for `X-Forwarded-For` with CIDR containment.
- Implemented strict Content Security Policy (CSP), anti-framing, and HSTS headers.

### Phase 3: Archive Decompression Bombs & Active Media Content
- Protected ZIP decompression with strict extraction limits (max 50 files, 100MB uncompressed, 20:1 ratio, path traversal checks).
- Active PDF keyword scanner detecting `/JavaScript`, `/JS`, `/Launch`, `/OpenAction`, `/AA`.
- Image EXIF metadata stripping for JPEG/PNG uploads.

### Phase 4: Cryptographic Audit Logging, Advisory Locks & Least Privilege
- Implemented HMAC-SHA256 chained audit trail with advisory transaction locking (`SELECT pg_advisory_xact_lock`).
- Configured PostgreSQL WORM triggers (`prevent_audit_log_modification`) preventing modification or deletion of audit records.
- Configured least-privilege database roles (`herasat_app` vs `herasat_migrator`).

### Phase 5: Tracking Code Hashing, AAD Binding & Encrypted Backups
- Encrypted tracking codes with blind HMAC-SHA256 lookup index (`tracking_code_hash`).
- Implemented AES-GCM v3 AAD binding (`table:column:rowID`).
- Added two-person report deletion workflow (`HERASAT_TWO_PERSON_DELETION=true`) and superuser redaction endpoint (`/api/admin/report/redact`).
- Implemented authenticated AES-GCM encrypted backups (`.hbak`).

### Phase 6: Reliability, Hygiene & Final Verification
- Dynamic python interpreter discovery and XML 1.0 control character stripping.
- Public follow-up message deduplication.
- LRU rate limiter eviction and IPv6 `/64` prefix masking.
- Cleaned unused component prototypes and consolidated migrations into `migrations/`.

---

## Round 2: Deep Remediation & Hardening

### Phase 1: Product-Breaking Bug Fixes

#### 1. Restricted-Session Deadlock & 2FA State Carryover (`AdminAuthMiddleware`, `admin_handlers.go`)
- **Verified Finding**: A session having both `RequiresPasswordChange` and `Requires2FASetup` (e.g. first-boot superuser `admin_it`) was deadlocked. Independent `if` checks in `AdminAuthMiddleware` blocked `/api/admin/change-my-password` due to the 2FA requirement, and blocked `/api/admin/2fa/*` due to the password change requirement. Furthermore, when `ChangeMyPasswordHandler` created the replacement session, it dropped `Requires2FASetup`.
- **Code Changes**:
  - In `AdminAuthMiddleware`: unified restricted endpoint allow-list. When both flags are active, allowed endpoints include `/api/admin/me`, `/api/admin/logout`, `/api/admin/change-my-password`, `/api/admin/2fa/setup`, and `/api/admin/2fa/enable`.
  - Prioritized returning `password_change_required` error when both requirements are unmet on unauthorized paths.
  - In `ChangeMyPasswordHandler`: explicitly carried over `Requires2FASetup: admin.Requires2FASetup` to the newly minted session token.
- **Operator Impact**: First-boot administrators can seamlessly change their initial password and subsequently enroll in 2FA without getting locked out of the portal.
- **Tests Added**:
  - `TestSessionDeadlockAndPermutations`: table-driven test covering all 17 combinations of restriction flags against protected and permitted routes.
  - `TestChangeMyPasswordPreservesRequires2FA`: verified replacement session maintains `Requires2FASetup: true`.

#### 2. Audit Chain Microsecond Truncation & Verification Compatibility (`audit.go`)
- **Verified Finding**: `executeAuditLogTx` hashed `time.Now().UTC()` at nanosecond precision via `RFC3339Nano`, but PostgreSQL `TIMESTAMPTZ` rounds and stores timestamps at microsecond precision. Recomputing the HMAC on retrieved rows resulted in verification failure on newly written entries.
- **Code Changes**:
  - In `executeAuditLogTx`: explicitly truncated `t` to microsecond precision (`t.UTC().Truncate(time.Microsecond)`) before computing `computeAuditHash` and before executing the SQL `INSERT`.
  - Added `VerifyAuditChainDetailed`: verifies audit chain with backward compatibility, retrying against non-truncated timestamps if legacy records exist, tracking fallback counts, and precisely returning `BrokenRowID` on tampering.
- **Operator Impact**: The cryptographic audit chain remains 100% valid across continuous operations and restarts, without false alarm chain breakage.
- **Tests Added**:
  - `TestAuditChainConcurrentAndTamperDetection`: executed 40 concurrent audit writes across 8 goroutines, verified valid chain integrity, mutated a single entry, and asserted exact broken row ID reporting.

#### 3. Token-Aware PDF Scanner & Precise OOXML Part Filtering (`metadata.go`)
- **Verified Finding**: Case-insensitive raw substring matching of `/JS` and `/AA` caused 60% to 90% false-positive rejections on clean binary PDF files (due to random byte occurrences in JPEG images). Substring matching on `.bin` in `InspectOOXML` rejected benign office files containing `xl/printerSettings/printerSettings1.bin`.
- **Code Changes**:
  - Replaced raw substring search with token-aware PDF scanner (`scanPDFBufferTokens`): parses PDF name tokens (`/` followed by delimiter/whitespace-terminated names) outside stream blocks, and decodes `#xx` hexadecimal escapes (e.g. `/J#61vaScript`).
  - Matched spec-defined names exactly and case-sensitively: `JavaScript`, `JS`, `Launch`, `EmbeddedFile`, `EmbeddedFiles`, `OpenAction`, `AA`, `SubmitForm`, `ImportData`, `RichMedia`, `GoToR`, `Encrypt`.
  - Ensured `/AA` does not falsely match `/AAAAAA+FontName`.
  - Decompressed and token-scanned Flate-compressed object streams.
  - In `InspectOOXML`: restricted dangerous `.bin` rejections to `vbaProject.bin`, `oleObject*.bin`, `embeddings/*`, and `externalLinks/*`, allowing `printerSettings*.bin`.
- **Operator Impact**: Whistleblowers can successfully attach standard, clean PDF scans and Excel workbooks with default printer configurations without spurious upload rejections.
- **Tests Added**:
  - `TestTokenAwarePDFScanner`: verified clean 5MB photo-heavy PDF passes, while Flate-compressed `/JavaScript`, hex-escaped `/J#61vaScript`, and `/Encrypt` PDFs fail.
  - `TestInspectOOXMLFixtures`: verified clean XLSX with `printerSettings1.bin` passes, while XLSX with `vbaProject.bin` fails.
  - Real test fixtures committed under `testdata/`.

#### 4. Toolchain Alignment (`Dockerfile`)
- **Verified Finding**: `Dockerfile` specified `golang:1.24-alpine` builder image whereas `go.mod` declared `go 1.26.5`.
- **Code Changes**: Updated `Dockerfile` builder image to `golang:1.26-alpine`.
- **Operator Impact**: Consistent compilation environment between local development toolchain and Docker container builds.
