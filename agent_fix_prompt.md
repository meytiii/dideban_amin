# Task: Security and Reliability Remediation for Dideban-e Amin (Herasat Portal)

## Role and context

You are a senior Go security engineer working on a confidential whistleblower reporting portal for a government education department. The stack is Go (standard library, `lib/pq`, `x/crypto`), PostgreSQL 16, and a native HTML/CSS/JS frontend. The system stores sensitive reports from people who may be at personal risk, so anonymity, confidentiality and audit integrity are the highest priorities.

A static code review produced the findings below. Treat each as a hypothesis: confirm it in the code before changing anything, and say so in your report if a finding turns out to be wrong or already fixed.

## Ground rules

1. Before anything else, read `AGENTS.md`, `GEMINI.md`, `DESIGN.md` and every file in `.agents/rules/`. Follow them exactly, including the no-comments rule and the commit message rule.
2. Work in the phases below, in order. Make one commit per phase (or per logical unit within a phase), and keep `go build ./...`, `go vet ./...` and `go test ./...` green after each commit.
3. Every behavior change needs a test. Prefer table-driven tests that cover the failure case as well as the happy path.
4. Database changes must be idempotent migrations that work on both fresh installs and existing databases. Keep `schema.sql` and `docker/init/01_schema.sql` consistent, ideally generated from one source.
5. Keep backward compatibility for existing encrypted data (`v1:`, `v2:` ciphertext prefixes, `HENC\x01` and `HENC\x02` files) and for existing sessions where practical.
6. Security checks must fail closed. The only exception is an explicit development escape hatch (`HERASAT_DEV_MODE=1`), which must log a loud warning.
7. User-facing messages stay in Persian and keep the existing tone. Do not change the JSON contracts that `frontend/` depends on unless you update the frontend in the same commit.
8. Never put real secrets, default passwords or production hostnames in the repository, the README, tests or logs.
9. If a finding requires a product or policy decision (marked **DECISION**), implement the safest configurable default, document the options, and flag it in your final report rather than silently choosing.

## Phase 0: Baseline

- Move all files from `test/` into the repository root. They are `package main` tests that call unexported root functions such as `getIP`, so they cannot compile from a subdirectory. Delete the `test/` directory and move its README content into the root documentation.
- Get `go build`, `go vet`, `gofmt -l` and `go test -race ./...` to a clean baseline. Record which tests fail before you change anything.
- Replace plain-string context keys (`"admin"`) with a typed key and a helper such as `adminFromContext(r)`.

## Phase 1: Authentication and sessions

1. **Enforce 2FA.** The `totp_enforced` column is seeded but never read. If a user has `totp_enforced = true` and TOTP is not enabled, login must return a restricted session that can only reach the 2FA setup endpoints. Enforced users must not be able to disable 2FA. Add a server-side global option (`HERASAT_REQUIRE_2FA`) that enforces TOTP for every admin role.
2. **Enforce forced password change server-side.** In `AdminAuthMiddleware`, when `must_change_password` is true or the password is older than 30 days, reject every API call except `GET /api/admin/me`, `POST /api/admin/change-my-password` and `POST /api/admin/logout`. Return a distinct error code that the frontend can recognize.
3. **Lockout redesign.** Locking by username alone lets anyone with a solved captcha lock out the superuser. Key failures on username and IP together, add progressive delays, and make sure a legitimate login from a known-good IP is not blocked by attacker traffic. Do not create rows for non-existent usernames, or cap them. Run a dummy Argon2 verification for unknown users so response time does not reveal which usernames exist.
4. **TOTP and MFA token.** Narrow the accepted window from ±2 steps to ±1. Make the MFA pending token single-use (record its nonce in the database) and bind it to the client IP or User-Agent. Rate limit and lock out failed password checks in `Enable2FAHandler` and `Disable2FAHandler`.
5. **Password policy.** Count runes rather than bytes, keep the 12-character minimum, require the new password to differ from the username, and reject a small built-in list of common or known-default passwords, including the one printed in the README.
6. **Seed and README.** Remove any credentials from the README. Replace the seeded password with a mechanism that generates a random initial password at first boot, or requires `HERASAT_INITIAL_ADMIN_PASSWORD`, prints it once to stdout, and forces a change. Confirm what password the current seed hash really corresponds to and mention it in your report.
7. **Session event stream.** In `AdminEventsHandler`, re-validate the session on every heartbeat and close the stream when the session is revoked, expired or its role has changed. Broadcast events only to roles and users allowed to see that report, and stop broadcasting usernames of revoked sessions to everyone. Test whether `Server.ReadTimeout` cancels the request context after 10 seconds, and fix it with `http.NewResponseController(w).SetReadDeadline(time.Time{})` if so.
8. Stop logging attempted usernames with `log.Printf`. If you must log them, sanitize and truncate them.

## Phase 2: Deployment safety

1. At startup, if `HERASAT_ADMIN_ALLOWED_CIDRS` is empty, refuse to start unless `HERASAT_DEV_MODE=1` or the operator sets an explicit `HERASAT_ADMIN_ALLOW_ANY=true`.
2. If `HERASAT_COOKIE_SECURE=true` is set without TLS (which implies a reverse proxy) and `HERASAT_TRUSTED_PROXIES` is empty, refuse to start with a clear message. Otherwise every client appears as the proxy IP and rate limits, lockouts and the IP whitelist stop working.
3. Parse CIDR environment variables once at startup instead of on every request, and fail on malformed entries rather than silently ignoring them.
4. Log a prominent warning at startup when ClamAV is disabled. **DECISION:** enable it by default in production mode.
5. Reject the placeholder values from `.env.example` for `HERASAT_AES_KEY`. Remove the hard-coded fallback key in `crypto.go`'s `init()` so that no code path can run with a known key. Remove the deprecated `X-XSS-Protection` header and `PreferServerCipherSuites`. Make the Content-Security-Policy and CORS headers consistent with the origin allow-list.
6. Add a Dockerfile and compose service for the app so the documented Docker setup actually runs the whole stack. Pin `python3` plus the `python-docx` dependency in it.

## Phase 3: Uploads and reporter anonymity

1. **Decompression bombs.** Call `image.DecodeConfig` before decoding JPEG or PNG and reject images above a pixel limit (for example 40 megapixels) or with absurd dimensions.
2. **Metadata stripping for every accepted type.** Today only JPEG and PNG are sanitized, so PDF, DOCX, XLSX, MP3, MP4, OGG and WAV keep author names, GPS data and revision history. Choose one approach per type:
   - Sanitize it (rewrite Office files with core and app properties and custom XML removed, rewrite PDFs without the Info dictionary and XMP, strip audio and video tags), or
   - Remove the type from the allow-list.
   **DECISION:** if sanitizing a type is not feasible in pure Go, drop the type and show the reporter a warning that explains why.
3. **Real PDF inspection.** Replace the substring scan with a proper parse (for example `pdfcpu`) that also inspects compressed object streams and hex-encoded names, or reject any PDF with `/OpenAction`, `/AA`, `/JavaScript`, `/Launch`, `/EmbeddedFile` or an encrypted structure.
4. **OOXML and archives.** Inspect DOCX and XLSX containers for `vbaProject.bin`, `externalLink` and `oleObject` parts. Stop accepting top-level `.rar` since it cannot be inspected. Compare ZIP size ratios against what you actually read when you inspect, not only the declared sizes.
5. **Timeouts and memory.** Set `ReadTimeout` high enough for 20 MB over a slow mobile connection on the upload route only (use per-request deadlines with `ResponseController`), while keeping short timeouts elsewhere. Cap the number of files per report (for example 5). Avoid holding multiple full copies of each upload in memory. Call `r.MultipartForm.RemoveAll()` when done.
6. **Do not record reporter IPs.** For anonymous submissions, remove the IP from audit and syslog entries for rejections such as `REJECTED_VULGAR_CONTENT`. **DECISION:** if abuse tracking is required, store a salted, rotating hash instead.
7. **DECISION:** change the profanity filter from hard rejection to flagging the report for staff review, controlled by a setting, so genuine reports quoting abusive language are not discarded.
8. Validate `category` against a server-side allow-list and return a clear validation error instead of a 500 when it is too long.
9. Fix the upload directory boundary check by comparing against `absUploadDir + string(os.PathSeparator)`, in both the download and delete handlers.
10. **Captcha.** Add attempt counting so a token is burned on a wrong answer. Widen the answer space or replace it with a proof-of-work challenge. Make sure failed captchas still count toward the IP rate limit.

## Phase 4: Audit log integrity

1. **Least-privilege DB roles.** Create a migration or owner role and a separate application role. The application role gets only `SELECT, INSERT` on `audit_logs` and no ability to alter triggers or the table. Update `docker-compose.yml` and the init SQL accordingly. The application must not connect as the Postgres superuser.
2. **Keyed hash chain.** Compute `entry_hash` with HMAC-SHA256 using a key derived from the master key, instead of plain SHA-256. If the previous hash cannot be read, fail the operation. Never fall back to the genesis hash.
3. **Concurrency.** Take `pg_advisory_xact_lock` inside the insert transaction so the chain is correct with more than one app instance. `logAudit` and `logAuditTx` should share one implementation.
4. **Field lengths.** Variable data (usernames, IPs, error text) currently goes into `action VARCHAR(64)`, so long values make inserts fail silently. Add a `detail TEXT` column, keep `action` as a fixed enum-like constant, and move the variable part into `detail`. Include `detail` in the hash. Validate and limit usernames on creation (length, allowed character set, Unicode normalization to block look-alike names).
5. Make `logAudit` return its error to callers that matter. If a failed audit write is ignored for a security-relevant action, treat it as an error.
6. Add a startup verification of the whole chain (or the last N entries), and an admin-visible status endpoint that reports chain validity.
7. Fail startup if the WORM triggers cannot be installed or verified, unless `HERASAT_DEV_MODE=1`. Migration errors must be fatal rather than logged warnings.

## Phase 5: Data protection

1. **Tracking codes.** Stop storing the bearer tracking code in plaintext. Add an HMAC lookup column (keyed with a derived key) and store the code encrypted if admins need to see it. Migrate existing rows. Keep the public tracking endpoint working with the same code format.
2. **Encryption hygiene.** Add a key ID to the ciphertext format (`v3:<kid>:...`) so keys can be rotated. Use AES-GCM additional authenticated data that binds the ciphertext to its table, column and row ID so values cannot be swapped between rows. Keep decrypt-compat with older prefixes. Add a re-encryption command that migrates old data, and document the rotation procedure.
3. **Backups.**
   - Encrypt backup files (for example with `age` or an AES-GCM stream using a derived key).
   - Do not pass the DB password on the command line. Use `PGPASSFILE` or the `PGPASSWORD` environment variable for the child process.
   - Include `secure_uploads/` in the backup set.
   - Remove the hand-rolled SQL dump fallback or make it fail loudly. A partial backup must never be reported as success.
   - Do not shell out to `docker exec` from the application.
   - Add a tested restore script and verify checksums.
4. **DECISION: separation of duties.** The IT superuser can currently read and delete every report. Add a configurable two-person rule for report deletion (a second superuser must approve), and an option that prevents the technical superuser role from reading report content. Document the default you chose.

## Phase 6: Reliability and cleanliness

1. The DOCX export hard-codes the `python` binary. Resolve it with `exec.LookPath` (`python3` first, then `python`) or an environment variable, and fail with a clear error. Strip XML-illegal control characters before passing text to `python-docx`.
2. Use `ON CONFLICT` or transactions consistently in follow-up and notification inserts so partial writes cannot leave inconsistent state. Add a cap on follow-up messages per report and per time window.
3. Make rate limiter eviction deterministic (evict the least recently seen visitors), and key IPv6 clients by /64.
4. Remove dead code and clutter after confirming it is unused: `components/ui/*.tsx`, `lib/utils.ts`, the duplicate `frontend/confirm.js`, and the `graphify-out/` directory. **Ask before deleting** anything under `.agents/` or `.impeccable/`.
5. Replace the duplicated schema definitions with a single versioned migrations folder, applied at startup.

## Verification

- `go build ./...`, `go vet ./...`, `gofmt -l .`, `staticcheck ./...` and `go test -race ./...` must all pass.
- Add integration tests (a test Postgres via docker-compose is fine) covering: 2FA enforcement, forced password change, lockout fairness, audit chain verification and tamper detection, tracking code hashing and migration, and backup and restore.
- Add upload tests with real fixtures: an oversized-dimension PNG and JPEG, a PDF with JavaScript in a compressed object stream, a DOCX with author metadata, a macro-enabled OOXML file, and a ZIP bomb.
- Run a manual end-to-end check of submit, track, reply and admin review flows, and confirm the frontend still works.

## Deliverables

1. The commits described above, with clear messages that follow the repository's commit message rule.
2. A `SECURITY_CHANGES.md` that lists, for each finding: what you verified, what you changed, any behavior or configuration changes operators must know about (new environment variables, migration steps), and tests added.
3. A final report with three sections: **Fixed**, **Not fixed or needs a decision** (with your recommendation), and **New risks introduced or remaining**. Include any finding you judged to be a false positive and why.
4. An updated `.env.example` and README that document every new setting, with no real credentials.
