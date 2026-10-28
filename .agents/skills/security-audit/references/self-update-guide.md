# Self-Update Guide

This skill is designed to be continuously updated with the latest security intelligence. There are two update mechanisms: **AI-assisted update** (recommended) and **manual update**.

---

## AI-Assisted Self-Update

Run `scripts/self-update.py` to automatically:

1. Fetch latest security advisories from configured sources
2. Generate a summary of new threats, tools, and best practices
3. Suggest additions to the skill's checklist and scan commands
4. Output a diff-style update proposal for human review

```bash
python .cursor/skills/security-audit/scripts/self-update.py
```

### How It Works

1. **Source fetching**: The script fetches RSS/Atom feeds and web pages from security intelligence sources
2. **AI analysis**: The fetched content is formatted as a prompt for the AI agent to analyze
3. **Update proposal**: The agent generates specific additions to reference files
4. **Human review**: Updates are proposed, never auto-applied

### Intelligence Sources

| Source | URL | Focus |
|--------|-----|-------|
| OWASP LLM Top 10 | owasp.org/www-project-top-10-for-large-language-model-applications | AI/LLM vulnerabilities |
| MITRE ATLAS | atlas.mitre.org | Adversarial ML tactics |
| Anthropic Research | anthropic.com/research | AI safety tools & methods |
| Electron Releases | electronjs.org/blog | Desktop app security |
| Node.js Security | nodejs.org/en/blog/vulnerability | Runtime vulnerabilities |
| Python Security | python.org/news/security | Language vulnerabilities |
| GitHub Advisories | github.com/advisories | Dependency vulnerabilities |
| NIST NVD | nvd.nist.gov | CVE database |
| CWE Updates | cwe.mitre.org | Weakness classifications |

### Adding New Sources

Edit `scripts/self-update.py` → `INTELLIGENCE_SOURCES` list to add new URLs.

---

## Manual Update Workflow

When you learn about a new security concern:

### 1. New Vulnerability Type

Add to `references/checklist.md` under the appropriate phase:

```markdown
- [ ] [New check description]
```

### 2. New Scan Pattern

Add to `references/scan-commands.md` under the appropriate platform:

```bash
# [Description of what this finds]
rg "[pattern]" src/
```

### 3. New Tool

Add to SKILL.md → Tool Recommendations table:

```markdown
| [Category] | `[tool-name]` |
```

### 4. New AI Security Concern

Add to `references/ai-security.md` under the appropriate section.

### 5. Version Bump

Update SKILL.md description or `references/changelog.md` with:

```markdown
### vX.Y.Z — YYYY-MM-DD
- **Trigger**: [What triggered this update]
- **Changes**: [What was added/modified]
- **Source**: [Where the information came from]
```

---

## Update Triggers

Run self-update when any of these occur:

| Trigger | Action |
|---------|--------|
| Quarterly schedule | Full update check |
| Major framework release (Electron, React, FastAPI) | Check for security changes |
| High-profile CVE in used dependencies | Add specific scan pattern |
| New AI security tool released | Evaluate and add to tools |
| New attack technique published | Add scan pattern + checklist item |
| After completing an audit | Record lessons learned |
| User requests "update security knowledge" | Full update check |

---

## Continuous Improvement Metrics

Track these to measure skill effectiveness:

| Metric | Target | How to Track |
|--------|--------|-------------|
| True positive rate | > 90% | Count confirmed vs false positive findings |
| Coverage | All active platforms | Check platforms audited vs available |
| Update freshness | < 90 days since last update | Check changelog dates |
| New patterns per audit | 1-3 | Count additions to scan-commands.md |
| Cross-layer findings | At least 1 per audit | Count SEC-FS findings |
