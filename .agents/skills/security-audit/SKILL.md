---
name: security-audit
description: Perform comprehensive security audits on codebases covering frontend SPA, backend API, desktop apps (Electron/Native), AI-integrated systems, and full-stack projects. Generates structured vulnerability reports with severity ratings, fix recommendations, and action plans. Use when the user asks for a security audit, code security review, vulnerability scan, penetration test preparation, security checklist, or mentions security concerns about their codebase. Also triggers on "資安檢查", "security check", "vulnerability assessment", "安全審核". Supports self-updating from latest security intelligence sources.
---

# Security Audit Skill

## Overview

Perform multi-platform security audits producing actionable reports. Supports: Frontend (React/Vue/Angular), Backend (Python/Node/Go/Java), Desktop (Electron/Native), AI-integrated systems, and cross-layer analysis.

## Audit Workflow (6 Stages)

### Stage 1 — Context Building

Understand the project before reading code:

1. Read `package.json` / `pyproject.toml` / `Cargo.toml` for dependencies
2. Map `src/` directory structure (identify main/preload/renderer for Electron)
3. Read build configs (vite, webpack, electron-builder, Dockerfile)
4. Check `.gitignore`, `.env*` files, secrets management (`.teller.yml`, vault)
5. Read CI/CD configs (`.github/workflows/`)

**Decision point:**

```
IF frontend SPA → focus XSS, Token, CSP, Security Headers
IF backend API → focus Injection, Auth, Serialization, DB/Cache
IF desktop (Electron) → focus nodeIntegration, IPC, Preload, Protocol, Auto-update
IF AI-integrated → focus Prompt Injection, Model Access, Data Poisoning, Output Filtering
IF containerized → focus Dockerfile, image supply chain, secrets, runtime, orchestration
IF full-stack → all tracks parallel + cross-layer analysis
```

### Stage 2 — Threat Modeling (STRIDE)

| Threat | Frontend | Backend | Desktop | AI/LLM | Container |
|--------|----------|---------|---------|--------|-----------|
| **S**poofing | Token theft | JWT forgery | Deep link intercept | Prompt identity spoofing | Image substitution |
| **T**ampering | Client data | SQL/NoSQL injection | IPC message tamper | Training data poisoning | Layer injection, runtime config |
| **R**epudiation | Missing audit | Log gaps | No local audit trail | Untracked model decisions | Ephemeral container log loss |
| **I**nfo Disclosure | Source map, console.log | Stack trace, logs | Plaintext store, ASAR | Model memorization leakage | Secrets in layers, env leak |
| **D**oS | No rate limit trigger | Connection pool exhaust | Infinite IPC calls | Prompt-based resource exhaust | Resource limits missing |
| **E**levation | Route guard bypass | RBAC bypass | nodeIntegration→RCE | Jailbreak→unrestricted access | Root container→host escape |

### Stage 3 — Systematic Scanning

Scan from highest to lowest risk. Use `ripgrep` patterns from reference files.

- For scan commands per platform, see [references/scan-commands.md](references/scan-commands.md)
- For complete checklist per phase, see [references/checklist.md](references/checklist.md)

**Priority order:**
1. Hardcoded secrets (all platforms)
2. Dangerous APIs (`eval`, `pickle`, `innerHTML`, `executeJavaScript`)
3. Auth/Token security
4. Input validation & injection
5. Data storage & transmission
6. Platform-specific risks
7. AI/LLM-specific risks
8. Deployment & infrastructure
9. Logging & monitoring

### Stage 4 — Deep Analysis

For each finding, apply this matrix:

```
                    Exploit Difficulty
                 Easy     Medium    Hard
          ┌─────────┬─────────┬─────────┐
    High  │ CRITICAL│  HIGH   │ MEDIUM  │
Impact    ├─────────┼─────────┼─────────┤
    Med   │  HIGH   │ MEDIUM  │  LOW    │
          ├─────────┼─────────┼─────────┤
    Low   │ MEDIUM  │  LOW    │  INFO   │
          └─────────┴─────────┴─────────┘
```

**False positive filters:**
- In test files? → Exclude (unless leaking real secrets)
- In `node_modules` / `.venv`? → Record as dependency risk only
- In migration/seed? → Usually safe (check for hardcoded creds)
- Dev-only? → Downgrade to LOW/INFO
- Fully mitigated? → Downgrade to INFO

**Cross-layer analysis** — check combinations:
- Frontend token storage × Backend token validation
- Desktop IPC exposure × Renderer XSS potential
- AI output trust × downstream system consumption
- Deep link token flow → local storage → API usage
- Container secrets → environment variables → application access
- Container image supply chain → base image CVEs → runtime exploit

### Stage 5 — Report Writing

Use finding IDs: `SEC-FE-YYYY-NNN` (frontend), `SEC-BE-YYYY-NNN` (backend), `SEC-DT-YYYY-NNN` (desktop), `SEC-AI-YYYY-NNN` (AI/LLM), `SEC-CT-YYYY-NNN` (container), `SEC-FS-YYYY-NNN` (cross-layer).

For report template, see [references/report-template.md](references/report-template.md)

### Stage 6 — Feedback Loop & Self-Update

After each audit:
1. Record new patterns in execution log (appendix)
2. Update scan commands if new patterns discovered
3. Run `scripts/self-update.py` to check for latest security intelligence
4. Bump version if methodology changed

For self-update mechanism, see [references/self-update-guide.md](references/self-update-guide.md)

---

## Scripts

### `scripts/self-update.py` — Security Intelligence Self-Update

Automatically fetches the latest security advisories from 9 intelligence sources (OWASP, MITRE ATLAS, Anthropic, Electron, Node.js, Python, GitHub, NIST, CWE), then generates a structured update proposal that the AI agent can analyze and apply to the skill's reference files.

**Prerequisites**: Python 3.6+, internet access.

**Commands**:

```bash
# List all configured intelligence sources and their URLs
python3 .cursor/skills/security-audit/scripts/self-update.py --sources

# Preview which sources would be fetched (no network calls)
python3 .cursor/skills/security-audit/scripts/self-update.py --dry-run

# Run full update: fetch sources and print update proposal to stdout
python3 .cursor/skills/security-audit/scripts/self-update.py

# Run full update and save proposal to a file for later review
python3 .cursor/skills/security-audit/scripts/self-update.py --output-dir docs/security-updates/
```

**Workflow**:

1. Script fetches HTML from each source (max 50KB per source, 15s timeout)
2. Generates a structured AI prompt containing fetched content
3. Outputs an "Update Proposal" document (stdout or file)
4. The AI agent reads the proposal and suggests specific additions
5. Human reviews and approves changes to reference files
6. `references/changelog.md` is updated with version bump

**Adding new sources**: Edit `INTELLIGENCE_SOURCES` in the script to add URLs. Each source needs `name`, `url`, `focus` description, and `update_targets` (which reference files it may update).

**When to run**:
- After completing a security audit (Stage 6)
- Quarterly scheduled update
- After a major framework release (Electron, React, FastAPI)
- After a high-profile CVE disclosure
- When user requests "update security knowledge"

---

## Severity Definitions

| Level | Definition | Fix SLA |
|-------|-----------|---------|
| **CRITICAL** | Direct data breach, account takeover, RCE | 24 hours |
| **HIGH** | Exploitable under specific conditions | 1 week |
| **MEDIUM** | Increases attack surface | 1 sprint |
| **LOW** | Best practice violation, no direct threat | Next maintenance |
| **INFO** | Advisory improvement | As resources allow |

---

## Container Security (Phase 10) — Added 2026-02

For projects using Docker, Kubernetes, or other container technologies. See [references/container-security.md](references/container-security.md) for full guide.

### 10.1 Dockerfile Security

- [ ] Uses minimal base image (`-slim`, `-alpine`, `distroless`)
- [ ] Runs as non-root user (`USER` directive present)
- [ ] No secrets in build args or `COPY`/`ADD` commands
- [ ] Multi-stage build separates build-time from runtime dependencies
- [ ] Pinned base image version (no `latest` tag)
- [ ] `.dockerignore` excludes `.env`, `.git`, `node_modules`, secrets

### 10.2 Image Supply Chain

- [ ] Base images from trusted registries only
- [ ] Image scanning in CI/CD (`trivy`, `grype`, `snyk container`)
- [ ] Image signing and verification (Docker Content Trust / cosign)
- [ ] Rebuild schedule for base image CVE patches
- [ ] No unnecessary packages or tools in production image

### 10.3 Runtime Security

- [ ] Read-only root filesystem (`--read-only`)
- [ ] No `--privileged` flag
- [ ] Dropped capabilities (`--cap-drop=ALL`, add back only what's needed)
- [ ] Resource limits set (CPU, memory, PIDs)
- [ ] No host network mode (`--network=host`) unless required
- [ ] No host PID/IPC namespace sharing
- [ ] Seccomp / AppArmor profiles applied

### 10.4 Secrets Management

- [ ] Secrets via orchestrator secrets (K8s Secrets, Docker Secrets), not env vars when possible
- [ ] No secrets baked into image layers
- [ ] Secrets mounted as tmpfs volumes (not persisted)
- [ ] External secrets manager integration (Vault, AWS Secrets Manager, SSM)

### 10.5 Network & Orchestration

- [ ] Container-to-container traffic restricted (network policies)
- [ ] Ingress/Egress rules defined
- [ ] Service mesh mTLS (if applicable)
- [ ] Pod Security Standards / Pod Security Policies (K8s)
- [ ] RBAC for container orchestrator API access

### 10.6 Logging & Monitoring

- [ ] Container logs shipped to central logging (no local-only)
- [ ] Health checks defined (`HEALTHCHECK` in Dockerfile)
- [ ] Runtime anomaly detection (Falco, Sysdig)
- [ ] Container image audit trail

---

## AI/LLM Security (Phase 11) — Added 2026-02

For systems integrating LLM/AI capabilities, additional checks apply.

### 11.1 Prompt Injection

- [ ] User input separated from system prompts (no concatenation)
- [ ] Indirect prompt injection defense (content from external sources sanitized)
- [ ] System prompt not extractable by users
- [ ] Input length limits enforced

### 11.2 Model Access Control

- [ ] API keys for LLM providers in env vars (not hardcoded)
- [ ] Rate limiting on AI endpoints (cost + abuse prevention)
- [ ] Model selection restricted (users cannot switch to unrestricted models)
- [ ] Token budget / spending caps enforced

### 11.3 Output Safety

- [ ] LLM output treated as untrusted (sanitized before rendering)
- [ ] Code generated by LLM sandboxed before execution
- [ ] Content filters on model output (Constitutional Classifiers / guardrails)
- [ ] Multi-stage verification for high-risk actions (code deployment, data modification)

### 11.4 Data Privacy & Training

- [ ] User data not sent to LLM providers without consent
- [ ] PII scrubbed before LLM API calls
- [ ] Zero-data-retention agreements with providers (if applicable)
- [ ] Model outputs not stored in logs without redaction

### 11.5 AI Supply Chain

- [ ] LLM provider API versioned and pinned
- [ ] Fallback mechanism if provider unavailable
- [ ] Model behavior monitoring for drift or degradation
- [ ] Automated red-teaming / behavioral auditing (e.g., Petri-style tools)

### 11.6 AI-Assisted Security Tools Integration

- [ ] Using AI-powered code scanning (Claude Code Security, GitHub Copilot Security, etc.)
- [ ] Multi-stage verification pipeline for AI-found vulnerabilities (reduce false positives)
- [ ] AI scan results integrated into CI/CD pipeline
- [ ] Human review gate for AI-generated patches

---

## Tool Recommendations

| Category | Tools |
|----------|-------|
| Secret scanning | `gitleaks`, `trufflehog` |
| Frontend deps | `npm audit`, `pnpm audit`, `yarn audit`, `snyk` |
| Backend deps (Python) | `pip-audit`, `safety`, `bandit` |
| Backend deps (Node) | `npm audit`, `snyk` |
| Electron audit | `electronegativity`, ASAR inspection |
| Static analysis | `semgrep`, `eslint-plugin-security`, `pylint` |
| Container image scan | `trivy`, `grype`, `snyk container` |
| Container runtime | `Falco`, `Sysdig` |
| Container build lint | `hadolint`, `dockle` |
| K8s security | `kubesec`, `kube-bench`, `polaris` |
| Image signing | `cosign` (Sigstore), Docker Content Trust |
| DAST | `OWASP ZAP`, `Burp Suite` |
| AI/LLM testing | `Petri` (Anthropic), `garak`, `promptfoo` |
| Code search | `ripgrep (rg)` |
| HTTP headers | `securityheaders.com`, `observatory.mozilla.org` |
| Binary signing | `codesign --verify`, `spctl --assess` (macOS) |

---

## Reference Files

| File | Contents | When to read |
|------|----------|-------------|
| [references/checklist.md](references/checklist.md) | Full phase-by-phase checklist (Phase 0–11) | During Stage 3 scanning |
| [references/scan-commands.md](references/scan-commands.md) | ripgrep patterns per platform | During Stage 3 scanning |
| [references/report-template.md](references/report-template.md) | Report structure template | During Stage 5 report writing |
| [references/ai-security.md](references/ai-security.md) | AI/LLM security deep dive | When auditing AI-integrated systems |
| [references/electron-security.md](references/electron-security.md) | Electron-specific security guide | When auditing Electron apps |
| [references/container-security.md](references/container-security.md) | Container/Docker/K8s security guide | When auditing containerized apps |
| [references/self-update-guide.md](references/self-update-guide.md) | How to self-update this skill | During Stage 6 or on user request |
| [references/changelog.md](references/changelog.md) | Version history and changes | On demand |

---

## Quick Start Example

```
User: "請對這個專案做資安檢查"

1. Identify project type(s) from package.json / pyproject.toml / Dockerfile
2. Follow Stage 1–6 workflow
3. Generate report using SEC-XX-YYYY-NNN format
4. Provide action plan sorted by priority (P0→P4)
5. Record execution instance for future improvement
```
