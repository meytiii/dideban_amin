# Findings Verification Matrix — Didebane Amin

**Original Audit Commit**: `ddeb2988c08759e293f0da0a7deff7fdbd3706cd`  
**Remediation HEAD Commit**: `c5e715f199e26cf43ee0ba8551d5f1323fc095fd`  
**Status**: 10/10 FINDINGS VERIFIED AND RESOLVED  

---

| Finding ID | Title | Remediation Code Reference | Verification Test / Evidence | Pre-Fix Status | Final Status |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **DA-SEC-001** | CAPTCHA Answer Exposed in Plaintext in Signed Token | `crypto.go:183-225`, `captcha.go:134-195` | `TestRemediationCaptchaConfidentialityAndIntegrity`, `CAPTCHA_BYPASS_ANALYSIS.md` | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-002** | Missing Session Invalidation on Password Change/Reset | `admin_handlers.go:1685`, `admin_handlers.go:2100` | `TestRemediationSessionRevocationLifecycle`, `TestSessionRevokeAllUserSessions` | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-003** | Denial of Service via Unbounded Pagination Query | `admin_handlers.go:240-260` | `TestRemediationPaginationClampingBounds` (8 boundary conditions) | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-004** | Internal Admin Username Exposure in Public Tracking | `public_handlers.go:465` | `TestPublicTrackingOmitsAdminUsername`, response sanitization | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-005** | Python Stderr Stack Trace Disclosure in Word Export | `admin_handlers.go:3215` | `TestDocxSandboxMaliciousPayload`, subprocess error masking | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-006** | Incident Intake Velocity & Report ID Leakage via SSE | `events.go:14,28,71-120,154` | `TestRemediationSSEBrokerRoleScoping` | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-007** | Missing CSRF Protection on Administrative Logout | `main.go:207`, `admin.js:885` | `TestCSRFMiddleware`, route inspection | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-008** | Insecure PostgreSQL Exposure & Static Credentials | `docker-compose.yml:9,11` | Configuration audit (`127.0.0.1:5432:5432`, env var) | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-009** | Default Superuser Seeded with Static Hash & No 2FA | `docker/init/02_seed.sql:24,27` | Seed SQL audit (`must_change_password=true`, `totp_enforced=true`) | CONFIRMED (Vulnerable) | **REMEDIATED & VERIFIED** |
| **DA-SEC-010** | ClamAV Antivirus Daemon Disabled by Default | `clamav.go:27`, `.env.example:28` | `TestClamAVScanner`, operational mandate documentation | CONFIRMED (Config) | **REMEDIATED & VERIFIED** |
