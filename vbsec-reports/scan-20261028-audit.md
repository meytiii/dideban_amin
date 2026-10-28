# vbsec Security Scan Report: Setad_Khabari_HMD

- **Date**: 2026-10-28
- **Scope**: Whole Repository
- **Tool**: vbsec (v0.7+) + OWASP + Security-Audit
- **Verdict**: **WARN**

## Executive Summary
Comprehensive security scan of Go backend, frontend assets, database migrations, and container configuration.

## Findings Table

| ID | Severity | Category | File:Line | Summary |
|---|---|---|---|---|
| `VBS-001` | **MEDIUM** | `PATH-TRAVERSAL` | `admin_handlers.go:1294` | Original filename stored without `filepath.Base()`; Zip Slip potential on admin extraction |
| `VBS-002` | **MEDIUM** | `BROKEN-ACCESS-CONTROL` | `events.go:143` | SSE handler queries untyped string context `"admin"` instead of `adminFromContext(r)`; always fails with 401 |
| `VBS-003` | **LOW** | `SECURITY-MISCONFIG` | `Dockerfile:28` | Missing `templates/` and `scripts/` in container image causes DOCX export failure in Docker |
| `VBS-004` | **INFO** | `MISSING-RATE-LIMIT` | `main.go:314` | `/api/security/csp-report` endpoint lacks IP rate limiting |

```json
{
  "verdict": "WARN",
  "summary": "Core security architecture is exceptionally resilient (AFTA compliant cryptography, tamper-evident HMAC audit chains, and strict session management). 2 medium-severity vulnerabilities (Zip Slip potential and SSE context bug) and 1 low-severity container misconfiguration identified.",
  "scope": "whole_repository",
  "files_reviewed": 28,
  "primary_language": "go",
  "specialized_rules_used": ["go", "generic"],
  "mode": "exhaustive_audit",
  "date": "2026-10-28",
  "findings": [
    {
      "file": "admin_handlers.go",
      "line": 1294,
      "rule_id": "PATH-TRAVERSAL",
      "severity": "MEDIUM",
      "issue_summary": "Original attachment filenames stored and re-used in zip header without filepath.Base() sanitization creates potential Zip Slip on admin extraction",
      "fix_summary": "Apply filepath.Base(stripNul(name)) during ingestion and zip entry creation"
    },
    {
      "file": "events.go",
      "line": 143,
      "rule_id": "BROKEN-ACCESS-CONTROL",
      "severity": "MEDIUM",
      "issue_summary": "AdminEventsHandler queries untyped string 'admin' instead of typed adminFromContext(r), causing all SSE connections to fail with 401 Unauthorized",
      "fix_summary": "Replace r.Context().Value(\"admin\") with adminFromContext(r)"
    },
    {
      "file": "Dockerfile",
      "line": 28,
      "rule_id": "VERBOSE-ERROR-DEBUG-MODE",
      "severity": "LOW",
      "issue_summary": "Dockerfile does not copy templates/ and scripts/ into runtime image, causing DOCX generation to fail in container",
      "fix_summary": "Add COPY directives for templates/ and scripts/ in runner stage"
    },
    {
      "file": "main.go",
      "line": 314,
      "rule_id": "MISSING-RATE-LIMIT",
      "severity": "INFO",
      "issue_summary": "CSP reporting endpoint /api/security/csp-report lacks an IP rate limiter",
      "fix_summary": "Wrap CSPReportHandler with RateLimitMiddleware"
    }
  ]
}
```
