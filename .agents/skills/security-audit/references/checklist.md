# Security Audit Checklist (Full Platform)

> Derived from SECURITY_AUDIT_METHODOLOGY v3.0.0
> Markers: 🌐 All platforms, 🖥️ Frontend, ⚙️ Backend, 🖳 Desktop, 🤖 AI/LLM

---

## Phase 0 — Preparation 🌐

- [ ] Identify all repositories in scope
- [ ] Document tech stack (frameworks, languages, versions)
- [ ] Map data flow (user → frontend → API → DB/cache → external services)
- [ ] Identify data sensitivity level (PII, financial, health, education)
- [ ] Review existing security measures (avoid duplicate work)
- [ ] Set up scan environment (ripgrep, audit tools)

---

## Phase 1 — Static Analysis 🌐

### Hardcoded Secrets
- [ ] API keys, tokens, passwords in source code
- [ ] Private keys or certificates
- [ ] Database connection strings with embedded credentials
- [ ] Third-party service credentials

### Dangerous APIs
- [ ] 🖥️ `eval()`, `Function()`, `dangerouslySetInnerHTML`, `innerHTML`, `document.write`
- [ ] ⚙️ `eval()`, `exec()`, `pickle.loads`, `yaml.load` (without SafeLoader), `os.system()`, `__import__`
- [ ] ⚙️ `subprocess.run/call/Popen` with user input
- [ ] 🖳 `executeJavaScript()`, `webContents.loadURL('javascript:...')`
- [ ] 🤖 Unsanitized LLM output rendered as HTML or executed as code

### Input Validation
- [ ] ⚙️ SQL injection via string formatting (`f""`, `.format()`, `%s`)
- [ ] ⚙️ Command injection via unsanitized user input in subprocess
- [ ] 🖥️ URL parameters used without sanitization
- [ ] 🤖 Prompt injection via concatenated user input

---

## Phase 2 — Dependency Security 🌐

- [ ] 🖥️ Run `npm audit` / `pnpm audit` / `yarn audit`
- [ ] ⚙️ Run `pip-audit` / `safety check`
- [ ] 🖳 Run `pnpm audit` (Electron projects)
- [ ] Check for known CVEs in major dependencies
- [ ] Identify unused dependencies (attack surface reduction)
- [ ] Verify dependency lock files are committed
- [ ] 🤖 Check LLM SDK versions for known vulnerabilities

---

## Phase 3 — Auth Security

### Token Management 🌐
- [ ] 🖥️ Token storage mechanism (localStorage vs httpOnly cookie)
- [ ] 🖳 Token storage (electron-store vs safeStorage vs Keychain)
- [ ] ⚙️ JWT secret strength (min 256-bit, from env var)
- [ ] ⚙️ JWT expiration validation enabled (`verify_exp: True`)
- [ ] Token refresh mechanism and race condition handling
- [ ] Logout clears all tokens/sessions

### OAuth/OIDC
- [ ] PKCE used for authorization code flow
- [ ] State parameter validated
- [ ] Redirect URI validated (no open redirect)

### API Security ⚙️
- [ ] All endpoints require authentication (except public ones)
- [ ] API key rotation mechanism exists
- [ ] Rate limiting implemented
- [ ] RBAC / horizontal privilege checks

### Desktop Auth 🖳
- [ ] Deep link token format validated before storage
- [ ] Protocol handler URL validated (no injection)
- [ ] Token not passed via insecure channels

---

## Phase 4 — Data Handling & Transmission

### Storage Security
- [ ] 🖥️ No sensitive data in localStorage (or encrypted)
- [ ] 🖳 Sensitive data uses `safeStorage` API or OS Keychain
- [ ] ⚙️ Database connections use SSL/TLS
- [ ] ⚙️ Redis connections use SSL/TLS
- [ ] ⚙️ Encryption at rest for sensitive columns

### Transmission Security
- [ ] All API calls use HTTPS
- [ ] Certificate validation not disabled (`rejectUnauthorized: false`, `verify=False`)
- [ ] CORS configured with specific origins (not `*` in production)
- [ ] ⚙️ No sensitive data in URL query parameters

### Serialization ⚙️
- [ ] No `pickle.loads` on untrusted data (use JSON instead)
- [ ] No `yaml.load` without `SafeLoader`
- [ ] GraphQL depth limiting (if applicable)

---

## Phase 5 — Deployment & Infrastructure

### Web Server 🖥️
- [ ] HTTP Security Headers (CSP, HSTS, X-Frame-Options, X-Content-Type-Options)
- [ ] Source maps not deployed to production (or auto-deleted after Sentry upload)
- [ ] SRI for CDN resources

### Container ⚙️
- [ ] Non-root user in Dockerfile
- [ ] Minimal base image
- [ ] No secrets in Docker layers
- [ ] Health check endpoints

### Desktop Packaging 🖳
- [ ] Code signing (macOS Developer ID, Windows Authenticode)
- [ ] macOS notarization enabled
- [ ] Entitlements follow least privilege
- [ ] Auto-update server uses HTTPS
- [ ] Update packages have signature verification

### Environment Variables 🌐
- [ ] `.env` files in `.gitignore`
- [ ] Secrets managed via vault/SSM (not committed)
- [ ] Production env vars validated at startup

---

## Phase 6 — Frontend Specific 🖥️

- [ ] React auto-escaping relied upon (no `dangerouslySetInnerHTML`)
- [ ] External links have `rel="noopener noreferrer"`
- [ ] Open redirect prevention on `window.location` assignments
- [ ] Console.log does not leak sensitive data in production
- [ ] Route guards protect authenticated pages
- [ ] CSRF protection for state-changing requests

---

## Phase 7 — Backend Specific ⚙️

- [ ] ORM used for database queries (no raw SQL with user input)
- [ ] File upload validation (type, size, filename sanitization)
- [ ] Error responses don't expose internal details
- [ ] Broad exception handlers don't swallow auth errors
- [ ] Background workers validate input
- [ ] Dead letter queue for failed async jobs

---

## Phase 8 — Desktop Specific (Electron) 🖳

### BrowserWindow Security
- [ ] `nodeIntegration: false` (all windows)
- [ ] `contextIsolation: true` (all windows)
- [ ] `sandbox: true` (all windows)
- [ ] `webSecurity: true` (all windows)
- [ ] `allowRunningInsecureContent: false`
- [ ] `enableRemoteModule: false`
- [ ] No override parameters that weaken security

### IPC Security
- [ ] Preload uses `contextBridge.exposeInMainWorld()`
- [ ] IPC channels have whitelist (no raw `ipcRenderer.send` exposed)
- [ ] IPC handlers validate all input
- [ ] No arbitrary key read/write in store handlers

### Preload Security
- [ ] Only necessary APIs exposed (least privilege)
- [ ] `shell.openExternal` validates URL scheme
- [ ] No `fs`, `child_process` exposed to renderer
- [ ] Exposed objects are immutable

### Navigation & Protocol
- [ ] `will-navigate` event restricts navigation targets
- [ ] `setWindowOpenHandler` handles new window requests
- [ ] Custom protocol handlers validate URL format
- [ ] `will-attach-webview` enforces security settings

### Development Mode
- [ ] `ELECTRON_DISABLE_SECURITY_WARNINGS` only in dev
- [ ] `disable-web-security` only in dev
- [ ] `ignore-certificate-errors` only in dev
- [ ] Dev mode detection is reliable (not user-triggerable in prod)

---

## Phase 9 — Container Security 🐳

### Dockerfile
- [ ] Uses minimal base image (`-slim`, `-alpine`, `distroless`)
- [ ] Base image version pinned (no `:latest`)
- [ ] Runs as non-root user (`USER` directive)
- [ ] Multi-stage build (build deps not in production image)
- [ ] No secrets in `ENV`, `ARG`, or `COPY` commands
- [ ] `.dockerignore` excludes `.env`, `.git`, `node_modules`, secrets
- [ ] `HEALTHCHECK` defined
- [ ] No unnecessary packages installed in production stage
- [ ] Uses `COPY` instead of `ADD` (unless tar extraction needed)
- [ ] `npm ci --only=production` or equivalent (no dev dependencies)

### Image Supply Chain
- [ ] Image scanning in CI/CD (`trivy`, `grype`, `snyk container`)
- [ ] Critical/High CVEs fail the pipeline
- [ ] Base images from trusted/verified registries only
- [ ] Image signing (`cosign` / Docker Content Trust)
- [ ] SBOM generated per image
- [ ] Rebuild schedule for base image patches

### Runtime
- [ ] No `--privileged` flag
- [ ] Capabilities dropped (`--cap-drop=ALL`, add back selectively)
- [ ] Resource limits set (CPU, memory, PIDs)
- [ ] Read-only root filesystem where possible
- [ ] `--security-opt=no-new-privileges` set
- [ ] No host namespace sharing (PID, IPC, network) unless required
- [ ] Docker socket NOT mounted inside containers

### Secrets
- [ ] No secrets baked into image layers (`docker history` check)
- [ ] Secrets via orchestrator mechanisms (K8s Secrets, Docker Secrets, Vault)
- [ ] Secrets mounted as tmpfs volumes (not env vars when possible)
- [ ] K8s Secrets encrypted at rest (EncryptionConfiguration or external KMS)

### Network & Orchestration
- [ ] Container-to-container traffic restricted (network policies)
- [ ] Ingress/egress rules defined
- [ ] Service mesh mTLS (if applicable)
- [ ] K8s Pod Security Standards applied (Baseline or Restricted)
- [ ] K8s RBAC restricts API access
- [ ] `automountServiceAccountToken: false` where not needed

### Logging & Health
- [ ] Container logs shipped to central logging
- [ ] Health checks defined
- [ ] Runtime anomaly detection configured (Falco, etc.)

---

## Phase 10 — AI/LLM Specific 🤖

### Prompt Injection
- [ ] System prompts isolated from user input
- [ ] Indirect prompt injection defense (external content sanitized)
- [ ] System prompt not extractable
- [ ] Input length limits enforced

### Model Access Control
- [ ] LLM API keys in environment variables
- [ ] Rate limiting on AI endpoints
- [ ] Token budget / spending caps
- [ ] Model selection restricted

### Output Safety
- [ ] LLM output treated as untrusted
- [ ] Generated code sandboxed before execution
- [ ] Content filters on output
- [ ] Multi-stage verification for high-risk actions

### Data Privacy
- [ ] PII scrubbed before LLM API calls
- [ ] Zero-data-retention agreements (if applicable)
- [ ] Model outputs redacted in logs
- [ ] User consent for data sent to LLM providers

### AI Supply Chain
- [ ] LLM provider API versioned and pinned
- [ ] Fallback if provider unavailable
- [ ] Behavior monitoring for model drift
- [ ] Automated red-teaming / behavioral auditing

### AI-Assisted Security
- [ ] AI code scanning in CI/CD (Claude Code Security, etc.)
- [ ] Human review gate for AI-generated patches
- [ ] False positive filtering in AI scan results

---

## Phase 11 — Logging & Monitoring 🌐

- [ ] No sensitive data in logs (tokens, passwords, PII)
- [ ] Structured logging with correlation IDs
- [ ] Error monitoring (Sentry, etc.) doesn't capture sensitive data
- [ ] Audit trail for authentication events
- [ ] Alerting for suspicious patterns
- [ ] 🤖 LLM API call logging (cost tracking, abuse detection)

---

## Phase 12 — Review & Continuous Improvement 🌐

- [ ] Document all findings with SEC-XX-YYYY-NNN format
- [ ] Record false positives for future reference
- [ ] Update scan patterns if new vulnerability types found
- [ ] Run self-update to check latest security intelligence
- [ ] Schedule next audit (quarterly or pre-release)
