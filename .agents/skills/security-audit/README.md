# Security Audit Skill

> **Version**: 3.2.0 | **Language**: [繁體中文](README-TW.md)

A comprehensive AI agent skill for performing multi-platform security audits on codebases. Designed for Cursor IDE — the AI agent automatically activates this skill when a security review is requested.

## What It Does

Performs structured, repeatable security audits across **6 platform types**, producing actionable vulnerability reports with severity ratings, fix recommendations, and prioritized action plans.

| Platform | Examples | Key Focus Areas |
|----------|----------|----------------|
| **Frontend** | React, Vue, Angular | XSS, Token storage, CSP, Security Headers |
| **Backend** | Python (FastAPI), Node.js, Go, Java | Injection, Auth, Serialization, DB/Cache security |
| **Desktop** | Electron, Native apps | nodeIntegration, IPC, Preload, Auto-update, Code Signing |
| **Container** | Docker, Kubernetes | Dockerfile, Image supply chain, Runtime hardening, Secrets |
| **AI/LLM** | ChatGPT integrations, LangChain | Prompt Injection, Output safety, Model access, Data privacy |
| **Cross-layer** | Full-stack combinations | Token flow, CORS alignment, Error chain, Attack path analysis |

## How It Works

The skill follows a **6-stage audit workflow**:

```
Stage 1          Stage 2          Stage 3          Stage 4          Stage 5          Stage 6
Context     ──▶  Threat      ──▶  Systematic  ──▶  Deep        ──▶  Report     ──▶  Feedback
Building         Modeling         Scanning         Analysis         Writing          Loop
```

1. **Context Building** — Understand the project's tech stack, architecture, data flow, and existing security measures
2. **Threat Modeling** — Apply STRIDE framework per platform to identify likely attack vectors
3. **Systematic Scanning** — Run ripgrep patterns and audit tools, prioritized from highest to lowest risk
4. **Deep Analysis** — Evaluate each finding using impact × exploit difficulty matrix, filter false positives, perform cross-layer analysis
5. **Report Writing** — Generate structured report with `SEC-XX-YYYY-NNN` finding IDs, evidence, fix code, and action plan
6. **Feedback Loop** — Record lessons learned, update scan patterns, run self-update for latest intelligence

## Directory Structure

```
security-audit/
├── SKILL.md                             # Core instructions for the AI agent (< 500 lines)
├── README.md                            # This file (English)
├── README-TW.md                         # 繁體中文版說明
├── references/
│   ├── checklist.md                     # Full phase-by-phase checklist (Phase 0–12)
│   ├── scan-commands.md                 # ripgrep scan patterns per platform
│   ├── report-template.md              # Report structure template
│   ├── ai-security.md                  # AI/LLM security deep dive
│   ├── electron-security.md            # Electron-specific security guide
│   ├── container-security.md           # Container/Docker/K8s security guide
│   ├── self-update-guide.md            # Self-update mechanism documentation
│   └── changelog.md                    # Version history
└── scripts/
    └── self-update.py                  # Auto-fetch security intelligence & generate update proposals
```

## Quick Start

### Trigger the Skill

Simply ask the AI agent in Cursor:

```
"Please perform a security audit on this project"
"請對這個專案做資安檢查"
"Run a vulnerability scan on the codebase"
```

The agent will automatically follow the 6-stage workflow and produce a report.

### Self-Update

Keep the skill's knowledge current with the latest security intelligence:

```bash
# List all intelligence sources (11 sources: OWASP, MITRE, Anthropic, Docker, K8s, etc.)
python3 .cursor/skills/security-audit/scripts/self-update.py --sources

# Preview what would be fetched
python3 .cursor/skills/security-audit/scripts/self-update.py --dry-run

# Fetch latest intelligence and generate update proposal
python3 .cursor/skills/security-audit/scripts/self-update.py

# Save proposal to file for review
python3 .cursor/skills/security-audit/scripts/self-update.py --output-dir docs/security-updates/
```

The script fetches from 11 sources, generates a structured AI prompt, and the agent proposes specific changes to reference files. All updates require human approval — nothing is auto-applied.

## Severity Levels

| Level | Definition | Fix SLA |
|-------|-----------|---------|
| **CRITICAL** | Direct data breach, account takeover, or RCE | 24 hours |
| **HIGH** | Exploitable under specific conditions | 1 week |
| **MEDIUM** | Increases attack surface or reduces defense | 1 sprint |
| **LOW** | Best practice not followed, no direct threat | Next maintenance |
| **INFO** | Advisory improvement suggestion | As resources allow |

## Finding ID Format

| Prefix | Platform |
|--------|----------|
| `SEC-FE-YYYY-NNN` | Frontend |
| `SEC-BE-YYYY-NNN` | Backend |
| `SEC-DT-YYYY-NNN` | Desktop (Electron/Native) |
| `SEC-CT-YYYY-NNN` | Container (Docker/K8s) |
| `SEC-AI-YYYY-NNN` | AI/LLM |
| `SEC-FS-YYYY-NNN` | Cross-layer (Full-Stack) |

## Tools Ecosystem

| Category | Tools |
|----------|-------|
| Secret scanning | `gitleaks`, `trufflehog` |
| Dependency audit | `npm audit`, `pip-audit`, `safety`, `snyk` |
| Static analysis | `semgrep`, `bandit`, `eslint-plugin-security` |
| Container security | `trivy`, `grype`, `hadolint`, `dockle`, `Falco` |
| K8s security | `kubesec`, `kube-bench`, `polaris` |
| Electron audit | `electronegativity`, ASAR inspection |
| AI/LLM testing | `Petri`, `garak`, `promptfoo` |
| DAST | `OWASP ZAP`, `Burp Suite` |

## References & Standards

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [OWASP LLM Top 10](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
- [OWASP API Security Top 10](https://owasp.org/API-Security/)
- [Electron Security Checklist](https://www.electronjs.org/docs/latest/tutorial/security)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [NIST SP 800-190: Container Security](https://csrc.nist.gov/publications/detail/sp/800-190/final)
- [MITRE ATLAS](https://atlas.mitre.org/)
- [STRIDE Threat Model](https://learn.microsoft.com/en-us/azure/security/develop/threat-modeling-tool-threats)

## Version History

See [references/changelog.md](references/changelog.md) for full details.

| Version | Date | Highlights |
|---------|------|-----------|
| v3.2.0 | 2026-02-25 | Container/Docker/K8s security, script usage docs |
| v3.1.0 | 2026-02-25 | AI/LLM security, self-update mechanism |
| v3.0.0 | 2026-02-25 | Desktop/Electron security |
| v2.0.0 | 2026-02-25 | Backend API security |
| v1.0.0 | 2026-02-25 | Initial release (Frontend) |

## License

Internal project tool. Distributed with the repository.
