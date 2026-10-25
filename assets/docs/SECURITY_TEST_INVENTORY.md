# Security Test Inventory — Didebane Amin

**Application Name**: سامانه دیده بان امین (ستاد خبری و دریافت گزارش های مردمی حراست)  
**Assessment Date**: 2026-10-25  
**Target Commit (Original Audit)**: `ddeb2988c08759e293f0da0a7deff7fdbd3706cd`  
**Target Commit (Remediation HEAD)**: `1a85ae1d36c47f3f825eb079fb086d35a35a7afb`  
**Assessment Standards**: OWASP WSTG v4.2, OWASP ASVS v4.0.3, CWE, AFTA Technical Security Requirements  

---

## Test Inventory Status Legend
- **PASS**: Control tested and verified effective against threats.
- **FAIL**: Control evaluated and a confirmed vulnerability or security failure was discovered.
- **NEEDS REVIEW**: Control partially implemented; requires runtime confirmation or environment alignment.
- **NOT TESTED**: Test not performed (e.g. requires external live services or dedicated laboratory equipment).
- **BLOCKED**: Prerequisite, tool, or target environment unavailable.
- **NOT APPLICABLE (N/A)**: Specific threat vector does not apply to this system architecture.

---

## Comprehensive Security Test Matrix

### 1. Cryptography & Secrets Management (CRYPTO)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-CRY-01** | Verify AES-256-GCM authenticated encryption for sensitive fields | Source code review & unit testing in ephemeral environment | **PASS** | `crypto.go`, `TestCryptoFailClosedNeverReturnsPlaintext` | Authenticated encryption with random 12-byte nonces; tamper detection confirmed |
| **TEST-CRY-02** | Verify domain-specific key derivation separation | Manual static audit & self-test verification | **PASS** | `crypto.go:29-58`, `RunCryptographicSelfTests` | Purpose-derived keys (`keyFields`, `keyFilePath`, `keyFileBlob`, `hmacKey`, `keyCaptcha`) via HKDF-like scheme |
| **TEST-CRY-03** | Verify password hashing with Argon2id parameters | Source review & test suite execution | **PASS** | `crypto.go:238-320`, `TestPasswordComplexity` | Argon2id m=64MB, t=3, p=4; automatic rehash from legacy format verified |
| **TEST-CRY-04** | Verify known-answer cryptographic self-tests at server startup | Code trace & test execution | **PASS** | `crypto.go:352-399`, `main.go:140` | KAT validates AES-GCM round-trip, tamper rejection, key derivation, and Argon2id |
| **TEST-CRY-05** | Verify encryption key secrecy and lack of key logging | Codebase search & logging audit | **PASS** | `main.go`, `crypto.go`, `syslog.go` | `HERASAT_AES_KEY` is never written to logs, console, or API responses |
| **TEST-CRY-06** | Verify cryptographic re-encryption of legacy data | Source review & test suite | **PASS** | `migration_v2.go`, `TestMigrationV2` | Re-encryption script cleanly upgrades v1 data to v2 with purpose-derived keys |

---

### 2. Authentication & Session Management (AUTH)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-AUT-01** | Verify brute-force lockout after 5 consecutive failed attempts | Source review & unit test execution | **PASS** | `security.go:687-817`, `TestAccountLockout` | 15-minute temporary lockout enforced and persisted in PostgreSQL |
| **TEST-AUT-02** | Verify single concurrent session enforcement per account | Source review & unit test | **PASS** | `security.go:86-150`, `TestSessionManagement` | Older sessions kicked with `CONCURRENT_LOGIN` when new session created |
| **TEST-AUT-03** | Verify session cookie security flags (`HttpOnly`, `SameSite`, `Secure`) | Source review & test suite | **PASS** | `admin_auth.go:152-160`, `TestCookieSecure` | `HttpOnly=true`, `SameSite=Lax`, `Secure` enforced when `HERASAT_COOKIE_SECURE=true` or TLS active |
| **TEST-AUT-04** | Verify session token entropy and storage hashing | Static code inspection | **PASS** | `admin_auth.go:27-31`, `security.go:57-60` | 32-byte `crypto/rand` tokens; stored in database as SHA-256 hash |
| **TEST-AUT-05** | Verify session invalidation on password change and superuser reset | Unit & integration test execution | **PASS** | `admin_handlers.go:1685`, `admin_handlers.go:2100`, `TestRemediationSessionRevocationLifecycle` | **REMEDIATED (DA-SEC-002)**: `RevokeAllUserSessions` called on password reset and change; fresh session issued to current browser |
| **TEST-AUT-06** | Verify dormant account auto-lockout after 90 days of inactivity | Source code inspection | **PASS** | `admin_auth.go:97-102` | Accounts inactive > 90 days are automatically disabled upon login attempt |
| **TEST-AUT-07** | Verify TOTP two-factor authentication replay protection | Code trace & test execution | **PASS** | `mfa.go:36-60`, `mfa.go:111-144` | In-memory 120s step replay store prevents reuse of submitted 6-digit codes |
| **TEST-AUT-08** | Verify TOTP secret encryption at rest | Database query & handler inspection | **PASS** | `mfa.go:409-416`, `schema.sql:55` | Base32 TOTP secret is encrypted with AES-256-GCM before DB insertion |

---

### 3. Authorization & Access Control (AUTHZ)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-ATZ-01** | Verify role separation (`superuser`, `edare_kol`, `navahi_manategh`) | Middleware inspection & unit testing | **PASS** | `admin_handlers.go:95-104`, `TestRoleDefinitions` | Superuser-only endpoints strictly enforced via `SuperuserAuthMiddleware` |
| **TEST-ATZ-02** | Verify object-level access control on report details | Source code inspection | **PASS** | `admin_handlers.go:106-121`, `TestForwarding` | Regional officers restricted strictly to reports forwarded to their username |
| **TEST-ATZ-03** | Verify attachment download access authorization | Source code inspection | **PASS** | `admin_handlers.go:1044-1047` | `canAdminAccessReport` strictly verified before reading or serving attachments |
| **TEST-ATZ-04** | Verify network IP whitelisting for admin portal | Source review & test suite | **PASS** | `middleware.go:39-63`, `TestGetIPHardening` | `HERASAT_ADMIN_ALLOWED_CIDRS` enforced on `/admin/` and all admin API handlers |
| **TEST-ATZ-05** | Verify real-time SSE event broadcast access scoping | Unit & integration test execution | **PASS** | `events.go:71-120`, `TestRemediationSSEBrokerRoleScoping` | **REMEDIATED (DA-SEC-006)**: `new_report` events restricted strictly to `superuser` and `edare_kol` roles |

---

### 4. Input Validation, Abuse Prevention & CAPTCHA (INPUT)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-INP-01** | Verify CAPTCHA answer confidentiality in client token | Unit test & token property test | **PASS** | `captcha.go:134-195`, `crypto.go:183-225`, `TestRemediationCaptchaConfidentialityAndIntegrity` | **REMEDIATED (DA-SEC-001)**: Payload encrypted with AES-256-GCM; unreadable without server key |
| **TEST-INP-02** | Verify CAPTCHA replay prevention / token reuse | Source review & test suite | **PASS** | `captcha.go:67-111`, `TestCaptchaReplayPrevention` | Token signature inserted into `used_captcha_tokens` table with conflict rejection |
| **TEST-INP-03** | Verify rate limiting on report submission, login, and tracking | Source review & test suite | **PASS** | `main.go:179-185`, `TestIPRateLimiterBoundedCapacity` | Token bucket per IP with automatic stale visitor memory eviction |
| **TEST-INP-04** | Verify request body size limits on JSON and multipart endpoints | Source review & test suite | **PASS** | `main.go:188`, `TestLimitBodyMiddlewareRejectsOversizedPayload` | 256KB cap on JSON endpoints; 20MB cap on multipart file uploads |
| **TEST-INP-05** | Verify profanity and vulgarity legal compliance filter | Source review & corpus testing | **PASS** | `profanity.go`, `TestContainsVulgarLanguageCorpus` | Extensive 106-case test corpus passing with Persian, Finglish, and obfuscated terms |
| **TEST-INP-06** | Verify tracking code format validation and sanitization | Source review | **PASS** | `security.go:861-871`, `public_handlers.go:408` | `isValidTrackingCode` restricts to 32 chars and alphanumeric `[A-Z0-9]` |
| **TEST-INP-07** | Verify pagination upper bounds on administrative cartable | Unit test execution across bounds | **PASS** | `admin_handlers.go:240-260`, `TestRemediationPaginationClampingBounds` | **REMEDIATED (DA-SEC-003)**: `pageSize` hard clamped to max 100, unbounded query disabled |
| **TEST-INP-08** | Verify public tracking endpoint strips internal admin usernames | Unit test & response inspection | **PASS** | `public_handlers.go:465`, `TestPublicTrackingOmitsAdminUsername` | **REMEDIATED (DA-SEC-004)**: `TrackReportHandler` zeroes out `admin_username` in messages for anonymous reporters |

---

### 5. File Uploads & Malware Defense (UPLOAD)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-UPL-01** | Verify ClamAV integration fails closed when unreachable | Source review & unit test | **PASS** | `clamav.go:50-129`, `TestClamAVFailClosedWhenUnreachable` | HTTP 503 returned and files wiped from disk if ClamAV daemon is unreachable |
| **TEST-UPL-02** | Verify ClamAV malware detection handling | Source review & unit test | **PASS** | `clamav.go:117-126`, `TestClamAVMalwareFound` | HTTP 400 returned and `REJECTED_MALICIOUS_FILE` logged on malware finding |
| **TEST-UPL-03** | Verify binary magic bytes validation against filename extensions | Source review & test suite | **PASS** | `metadata.go:83-147`, `TestMetadataSanitization` | Strict whitelist: PDF, PNG, JPG, ZIP, DOCX, XLSX, RAR, TXT, MP3, OGG, WAV, MP4 |
| **TEST-UPL-04** | Verify JPEG/PNG image re-encoding to strip EXIF & metadata | Source review & test suite | **PASS** | `metadata.go:14-43`, `TestMetadataSanitizationJPEG/PNG` | Decodes and re-encodes pixel buffer; strips EXIF, GPS, and trailing payloads |
| **TEST-UPL-05** | Verify PDF active content inspection | Source code review | **PASS** | `metadata.go:59-81`, `public_handlers.go:255` | Rejects PDFs with `/JavaScript`, `/JS`, `/Launch`, `/EmbeddedFiles`, `/SubmitForm` |
| **TEST-UPL-06** | Verify ZIP bomb and directory traversal prevention | Source review & test suite | **PASS** | `public_handlers.go:35-77`, `TestInspectZipArchive` | Max 50MB uncompressed, max 1000 entries, no passwords, no `..`, ratio <= 20x |
| **TEST-UPL-07** | Verify uploaded evidence file encryption at rest | Source code inspection | **PASS** | `public_handlers.go:308-323`, `crypto.go:176-196` | Uploads encrypted with AES-256-GCM (`HENC\x02`) and stored with `0600` permissions |
| **TEST-UPL-08** | Verify directory traversal prevention on evidence download | Source code inspection | **PASS** | `admin_handlers.go:1075-1081` | Clean path verified with `strings.HasPrefix(absFilePath, absUploadDir)` |

---

### 6. Database & Audit Integrity (DB)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-DAT-01** | Verify SQL injection resistance across all database queries | Comprehensive static code audit | **PASS** | `admin_handlers.go`, `public_handlers.go`, `db.go` | All dynamic values bound using parameterized placeholders (`$1`, `$2`, etc.) |
| **TEST-DAT-02** | Verify immutable audit log constraints (WORM compliance) | SQL schema & trigger inspection | **PASS** | `schema.sql:145-165`, `db.go:405-440` | PostgreSQL triggers raise exception on UPDATE, DELETE, or TRUNCATE on `audit_logs` |
| **TEST-DAT-03** | Verify audit log cryptographic hash chaining | Source code inspection | **PASS** | `admin_auth.go:319-370` | SHA-256 hash chains `prev_hash` to `entry_hash` with timestamp and user details |
| **TEST-DAT-04** | Verify remote Syslog / SIEM RFC 5424 export | Source review & unit test | **PASS** | `syslog.go`, `TestSyslogFormatRFC5424SanitizesParams` | Asynchronous RFC 5424 structured syslog export with UDP, TCP, and TLS options |

---

### 7. Browser & HTTP Security (HTTP)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-SEC-01** | Verify Content Security Policy (CSP) headers | Source review | **PASS** | `main.go:35` | `script-src 'self'`, `object-src 'none'`, `frame-ancestors 'none'` active |
| **TEST-SEC-02** | Verify Clickjacking defense (`X-Frame-Options`) | Source review | **PASS** | `main.go:39` | `X-Frame-Options: DENY` sent on all HTTP responses |
| **TEST-SEC-03** | Verify MIME sniffing defense (`X-Content-Type-Options`) | Source review | **PASS** | `main.go:38` | `nosniff` header sent on all HTTP responses |
| **TEST-SEC-04** | Verify CSRF token enforcement on administrative state changes | Source review & test suite | **PASS** | `csrf.go:24-48`, `TestCSRFMiddleware` | Double-submit cookie with constant-time token comparison |
| **TEST-SEC-05** | Verify CSRF enforcement on logout handler | Routing review & test execution | **PASS** | `main.go:207`, `admin.js:885`, `TestCSRFMiddleware` | **REMEDIATED (DA-SEC-007)**: `LogoutHandler` wrapped in `CSRFMiddleware`; `adminFetch` used |
| **TEST-SEC-06** | Verify XSS output encoding in frontend client scripts | Static audit of JavaScript files | **PASS** | `app.js:1102`, `admin.js:4035` | HTML entities consistently escaped via `escapeHtml`/`escapeHTML` |
| **TEST-SEC-07** | Verify Host header injection defense | Source review | **PASS** | `main.go:16-33` | `HERASAT_ALLOWED_HOSTS` validates and rejects unrecognized host headers |
| **TEST-SEC-08** | Verify error message sanitization in dossier export | Source inspection & error testing | **PASS** | `admin_handlers.go:3215`, `TestDocxSandboxMaliciousPayload` | **REMEDIATED (DA-SEC-005)**: Subprocess stderr logged internally; generic Persian error returned |

---

### 8. Deployment & Infrastructure Security (INFRA)

| Test ID | Objective | Method | Status | Evidence Reference | Notes / Limitations |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **TEST-INF-01** | Verify PostgreSQL network exposure in container setup | Configuration review | **PASS** | `docker-compose.yml:11` | **REMEDIATED (DA-SEC-008)**: Port 5432 bound strictly to `127.0.0.1:5432:5432` |
| **TEST-INF-02** | Verify default database credentials in container setup | Configuration review | **PASS** | `docker-compose.yml:9` | **REMEDIATED (DA-SEC-008)**: Parameterized via `${HERASAT_DB_PASSWORD}` |
| **TEST-INF-03** | Verify initial superuser password and MFA policy | Configuration review | **PASS** | `docker/init/02_seed.sql:24,27` | **REMEDIATED (DA-SEC-009)**: `must_change_password=true` and `totp_enforced=true` enforced |
| **TEST-INF-04** | Verify production TLS enforcement safeguards | Source review & unit test | **PASS** | `main.go:118-120`, `TestCookieSecureStartupFatalWithoutTLSOrFlag` | Server refuses to boot without TLS or explicit trusted proxy flag |
| **TEST-INF-05** | Verify active ClamAV daemon configuration | Configuration review & fail-closed test | **PASS** | `clamav.go:27`, `.env.example:28`, `TestClamAVFailClosedWhenUnreachable` | **REMEDIATED (DA-SEC-010)**: Documented mandatory setting; fail-closed behavior verified |
| **TEST-INF-06** | Full authenticated dynamic vulnerability penetration test | Dynamic runtime penetration testing & active HTTP verification | **PASS** | `assets/docs/SECURITY_PENETRATION_TEST_REPORT.md`, `assets/docs/security-evidence/RUNTIME_PENTEST_EVIDENCE.md` | 19 automated & manual dynamic test cases executed against live staging server; all 16 controls verified effective |
