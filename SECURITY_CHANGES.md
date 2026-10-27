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
