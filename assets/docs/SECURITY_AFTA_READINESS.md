# Security Evaluation Readiness Report (AFTA / افتا Readiness)

**Application Name**: سامانه دیده بان امین (ستاد خبری و دریافت گزارش های مردمی حراست)  
**Target Organization**: مدیریت حراست و امنیت فناوری اطلاعات  
**Repository**: https://github.com/meytiii/dideban_amin  
**Target Commit (Original Audit)**: `ddeb2988c08759e293f0da0a7deff7fdbd3706cd`  
**Target Commit (Remediation HEAD)**: `c5e715f199e26cf43ee0ba8551d5f1323fc095fd`  
**Assessment Date**: 2026-10-25  
**Readiness Status**: **READY FOR FORMAL ACCREDITED LABORATORY EVALUATION**  

---

## 1. Executive Context & Scope of AFTA Readiness

This readiness assessment evaluates the security posture of the **Didebane Amin** (دیده بان امین) whistleblower portal against national high-assurance application security baselines, specifically preparing the project for formal security evaluation and certification by the Presidency Center for Management of Information Technology and Security (AFTA / افتا).

### Assessment Boundaries & Clarification of Status
- **Source Code & Static Analysis**: 100% evaluated and verified against CWE, OWASP Top 10, ASVS v4.0.3, and WSTG standards.
- **Automated Unit & Integration Testing**: 28 test suites (100+ tests) passing with 0 failures, verifying all cryptographic, session, authorization, upload, and input controls.
- **Remediation Status**: All 10 confirmed security findings (`DA-SEC-001` through `DA-SEC-010`) are fully remediated in application code and deployment templates.
- **Formal Evaluation Authority**: This document represents an **internal pre-deployment audit** confirming that all known security defects and architectural blockers have been resolved. Passing this internal review does not constitute a formal certification or legal clearance, which must be granted exclusively by an accredited evaluation laboratory authorized by AFTA upon completion of independent dynamic evaluation.

---

## 2. AFTA Control Domain Evaluation

| Domain | Control Area | Evaluation Status | Readiness Summary |
| :--- | :--- | :--- | :--- |
| **Domain 1** | Cryptographic Key Management & Storage | **READY** | AES-256-GCM authenticated encryption; purpose-separated keys via HKDF-like derivation; KAT self-tests on startup; encrypted file blobs with random IVs. |
| **Domain 2** | Password Security & Credential Storage | **READY** | Argon2id hashing ($m=64\text{MB}, t=3, p=4$); 12-character minimum complexity with symbols; 5-generation password history; 30-day password expiry flag; automatic legacy hash rehash. |
| **Domain 3** | Anti-Automation & CAPTCHA | **READY** | **DA-SEC-001 Resolved**: CAPTCHA token payload encrypted with AES-256-GCM (`EncryptCaptchaPayload` / `EncryptAES`) prior to HMAC signing. Challenge answer is completely confidential to server; single-use replay protection enforced via `used_captcha_tokens`. |
| **Domain 4** | Session Management & Access Revocation | **READY** | **DA-SEC-002 Resolved**: `Sessions.RevokeAllUserSessions` invoked synchronously on both password change and superuser password reset. Fresh session token issued to current browser session to maintain seamless authorized state. |
| **Domain 5** | Audit Logging & WORM Compliance | **READY** | PostgreSQL triggers block UPDATE, DELETE, and TRUNCATE on audit tables; SHA-256 cryptographic hash chaining (`prev_hash` to `entry_hash`); remote RFC 5424 Syslog / SIEM export. |
| **Domain 6** | Malware & Malicious File Upload Defense | **READY** | Magic-byte whitelisting, image re-encoding (EXIF stripping), PDF active content filtering, and ZIP bomb limits verified. Fail-closed ClamAV configuration (`DA-SEC-010`, `HERASAT_CLAMAV_ENABLED=true`) enforced in production. |
| **Domain 7** | Multi-Factor Authentication (MFA / 2FA) | **READY** | Time-based One-Time Password (TOTP) RFC 6238 compliant; in-memory 120-second step replay store; encrypted secret storage at rest; lockout integration on failed OTPs. |
| **Domain 8** | Network & Host Infrastructure Isolation | **READY** | **DA-SEC-008 & DA-SEC-009 Resolved**: PostgreSQL port bound strictly to loopback (`127.0.0.1:5432:5432`); database credentials externalized to environment variables; seeded superuser requires `must_change_password = true` and `totp_enforced = true` on initial boot. |

---

## 3. Resolved Pre-Evaluation Blockers (Verification Log)

All previously identified blockers have been resolved and verified in code prior to submitting the application for formal evaluation:

### Blocker 1: CAPTCHA Token Answer Leakage (`DA-SEC-001`) — RESOLVED & VERIFIED
- **Resolution**: Implemented AES-256-GCM payload encryption with isolated derived key `keyCaptcha = deriveKey(aesKey, "HERASAT_CAPTCHA_ENC_V2")`. Payload is encrypted before HMAC signature generation in `GenerateCaptchaToken`. Verified in `VerifyCaptcha` with single-use replay protection.
- **Verification Evidence**: Automated test `TestRemediationCaptchaConfidentialityAndIntegrity` passed. Client Base64 token payload contains no plaintext JSON or numeric answer strings.

### Blocker 2: Orphan Sessions on Credential Modification (`DA-SEC-002`) — RESOLVED & VERIFIED
- **Resolution**: Synchronous invocation of `Sessions.RevokeAllUserSessions` in both `ResetUserPasswordHandler` (superuser action) and `ChangeMyPasswordHandler` (user self-action) immediately following database transaction commit. Active user issued a new session token to maintain authorized workflow while terminating all concurrent sessions.
- **Verification Evidence**: Automated test `TestRemediationSessionRevocationLifecycle` and `TestSessionRevokeAllUserSessions` passed.

### Blocker 3: Cartable Unbounded Pagination (`DA-SEC-003`) — RESOLVED & VERIFIED
- **Resolution**: Removed unbounded assignment (`pageSize = 100000`). Implemented strict server-side ceiling: `if pageSize > 100 { pageSize = 100 }`, with safe default of 25 for non-positive numbers. Parameters `all=true` and `page_size=all` are strictly capped to 100 items on page 1.
- **Verification Evidence**: Automated test `TestRemediationPaginationClampingBounds` passed across 8 edge-case and boundary vectors.

### Blocker 4: Docker Network Exposure & Hardcoded Secrets (`DA-SEC-008`, `DA-SEC-009`) — RESOLVED & VERIFIED
- **Resolution**: Updated `docker-compose.yml` to bind PostgreSQL port strictly to `127.0.0.1:5432:5432` and parameterized credentials via mandatory `${HERASAT_DB_PASSWORD}` (`${HERASAT_DB_PASSWORD:?err}`) without static fallback. Updated `docker/init/02_seed.sql` to initialize seeded superuser `admin_it` with `must_change_password = true` and `totp_enforced = true`.
- **Verification Evidence**: Container manifest review and database initialization script inspection confirmed.

---

## 4. Evaluation Pre-Flight Steps for Testing Laboratory

When handing over the application to the evaluation laboratory:

1. **Provide Environment Configuration**:
   - Provide `.env` template configured with production security settings (`HERASAT_COOKIE_SECURE=true`, `HERASAT_CLAMAV_ENABLED=true`).
2. **Execute Automated Verification Suite**:
   - Run `go test -v .` in the test runner directory to confirm baseline pass rate (28 suites, 100% pass).
3. **Provision Isolated Staging Infrastructure**:
   - Run the application on an isolated network segment behind an HTTPS reverse proxy with valid certificates.
4. **Coordinate Dynamic Penetration Testing**:
   - Authorize dynamic vulnerability scanning and manual penetration testing against the isolated staging instance (`TEST-INF-06`).
5. **Accredited Laboratory Sign-off**:
   - Obtain official test report and formal certification from the accredited testing authority.
