# Security Audit Report Template

Use this template when generating audit reports in Stage 5.

---

```markdown
# Security Audit Report

> **Project**: [Project name and description]
> **Date**: [YYYY-MM-DD]
> **Auditor**: AI Security Auditor
> **Methodology**: Security Audit Skill v[X.Y.Z]
> **Scope**: [List all codebases, branches, and config files reviewed]

---

## 1. Executive Summary

### 1.1 Overall Security Rating

**Rating: [🟢 LOW / 🟡 MODERATE / 🟠 MODERATE-HIGH / 🔴 HIGH / ⚫ CRITICAL]**

[1-2 sentence summary of most critical findings and overall posture]

### 1.2 Risk Distribution

| Severity | Frontend 🖥️ | Backend ⚙️ | Desktop 🖳 | AI/LLM 🤖 | Container 🐳 | Total |
|----------|-------------|-----------|-----------|-----------|-------------|-------|
| CRITICAL | - | - | - | - | - | - |
| HIGH     | - | - | - | - | - | - |
| MEDIUM   | - | - | - | - | - | - |
| LOW      | - | - | - | - | - | - |
| INFO     | - | - | - | - | - | - |
| **Total**| - | - | - | - | - | - |

(Remove columns for platforms not in scope)

### 1.3 Risk Categories

| Category | Count | Top Severity |
|----------|-------|-------------|
| Auth & Authorization | - | - |
| Configuration | - | - |
| Data Exposure | - | - |
| Infrastructure | - | - |
| Code Safety | - | - |
| Dependencies | - | - |
| Container | - | - |
| AI/LLM | - | - |

---

## 2. Project Overview

### 2.1 [Platform Name]

| Item | Detail |
|------|--------|
| **Type** | [e.g., React SPA / FastAPI / Electron Desktop] |
| **Framework** | [e.g., React 18.3 + TypeScript 5.5] |
| **Build** | [e.g., Vite 5.4 / electron-builder 26] |
| **Deployment** | [e.g., Nginx / Docker / DMG + Auto-update] |
| **Dependency Count** | [Prod: X, Dev: Y] |

(Repeat for each platform)

---

## 3. [Platform] Findings

### SEC-[XX]-[YYYY]-[NNN] | [SEVERITY] | [Title]

**Category**: [Auth / Config / Data / Infra / Code / Deps / AI]
**CVSS Reference**: [Score] ([Severity])

**Description**:
[What the vulnerability is and why it matters]

**Evidence**:

`[file path]` (line [N]-[M]):
```[language]
[code snippet showing the vulnerability]
```

**Impact**:
- [Bullet points of concrete impacts]

**Fix Recommendation**:
```[language]
[code snippet showing the fix]
```

**Fix Priority**: [🔴 24h / 🟡 1 week / 🟢 1 sprint / ⚪ maintenance]

---

(Repeat for each finding, grouped by platform)

---

## N. Positive Findings (Security Measures in Place)

### [Platform] 🖥️/⚙️/🖳/🤖

| Measure | Description | Rating |
|---------|------------|--------|
| [Name] | [What it does] | ✅ Good / ✅ Excellent / ⚠️ Partial |

---

## N+1. Action Plan

### Sorted by Priority

| Priority | ID | Item | Team | Est. Hours | Status |
|----------|-----|------|------|-----------|--------|
| ⚫ P0 (24h) | SEC-XX-YYYY-NNN | [Description] | [Team] | [Hours] | ⬜ |
| 🔴 P1 (1wk) | SEC-XX-YYYY-NNN | [Description] | [Team] | [Hours] | ⬜ |
| 🟡 P2 (sprint) | SEC-XX-YYYY-NNN | [Description] | [Team] | [Hours] | ⬜ |
| 🟢 P3 (maint) | SEC-XX-YYYY-NNN | [Description] | [Team] | [Hours] | ⬜ |
| ⚪ P4 (plan) | SEC-XX-YYYY-NNN | [Description] | [Team] | [Hours] | ⬜ |

---

## N+2. Appendix

### Audit Tools & Methods

| Method | Scope | Platform |
|--------|-------|---------|
| [Tool/method] | [What it checked] | [🖥️/⚙️/🖳/🤖/🌐] |

### Not Yet Performed (Recommended Follow-up)

| Item | Reason | Recommendation |
|------|--------|---------------|
| [Check] | [Why not done] | [Tool/approach] |

### References

- [Relevant security guides and standards]

---

*Report complete. Next audit: [date or trigger].*
*⚠️ CRITICAL items require immediate attention.*
```
