from __future__ import annotations
import argparse
import json
import sys
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Optional
from urllib.request import urlopen, Request
from urllib.error import URLError
from xml.etree import ElementTree
SKILL_DIR = Path(__file__).resolve().parent.parent
REFERENCES_DIR = SKILL_DIR / 'references'
INTELLIGENCE_SOURCES = [{'name': 'OWASP LLM Top 10', 'url': 'https://owasp.org/www-project-top-10-for-large-language-model-applications/', 'type': 'web', 'focus': 'AI/LLM vulnerabilities', 'update_targets': ['ai-security.md', 'checklist.md']}, {'name': 'MITRE ATLAS', 'url': 'https://atlas.mitre.org/', 'type': 'web', 'focus': 'Adversarial ML tactics and techniques', 'update_targets': ['ai-security.md']}, {'name': 'Anthropic Research', 'url': 'https://www.anthropic.com/research', 'type': 'web', 'focus': 'AI safety tools, Constitutional AI, Petri', 'update_targets': ['ai-security.md']}, {'name': 'Electron Security Blog', 'url': 'https://www.electronjs.org/blog', 'type': 'web', 'focus': 'Electron security updates and best practices', 'update_targets': ['electron-security.md', 'checklist.md']}, {'name': 'Node.js Security Releases', 'url': 'https://nodejs.org/en/blog/vulnerability', 'type': 'web', 'focus': 'Node.js runtime vulnerabilities', 'update_targets': ['checklist.md', 'scan-commands.md']}, {'name': 'Python Security', 'url': 'https://www.python.org/news/security/', 'type': 'web', 'focus': 'Python language security advisories', 'update_targets': ['checklist.md', 'scan-commands.md']}, {'name': 'GitHub Advisory Database', 'url': 'https://github.com/advisories?query=type%3Areviewed+severity%3Acritical', 'type': 'web', 'focus': 'Critical dependency vulnerabilities', 'update_targets': ['checklist.md']}, {'name': 'NIST NVD Recent', 'url': 'https://nvd.nist.gov/vuln/search/results?form_type=Basic&results_type=overview&query=electron+OR+react+OR+fastapi&search_type=all', 'type': 'web', 'focus': 'CVEs for project dependencies', 'update_targets': ['checklist.md']}, {'name': 'CWE Top 25', 'url': 'https://cwe.mitre.org/top25/archive/2024/2024_cwe_top25.html', 'type': 'web', 'focus': 'Most dangerous software weaknesses', 'update_targets': ['checklist.md', 'scan-commands.md']}, {'name': 'Docker Security Blog', 'url': 'https://www.docker.com/blog/category/security/', 'type': 'web', 'focus': 'Docker container security updates and best practices', 'update_targets': ['container-security.md', 'checklist.md']}, {'name': 'Kubernetes Security Blog', 'url': 'https://kubernetes.io/blog/', 'type': 'web', 'focus': 'Kubernetes security advisories and features', 'update_targets': ['container-security.md', 'checklist.md']}]

def list_sources():
    print('=' * 70)
    print('Security Intelligence Sources')
    print('=' * 70)
    for i, src in enumerate(INTELLIGENCE_SOURCES, 1):
        print(f"\n{i}. {src['name']}")
        print(f"   URL:     {src['url']}")
        print(f"   Focus:   {src['focus']}")
        print(f"   Updates: {', '.join(src['update_targets'])}")
    print(f'\nTotal: {len(INTELLIGENCE_SOURCES)} sources')

def fetch_url(url: str, timeout: int=15) -> str | None:
    try:
        req = Request(url, headers={'User-Agent': 'SecurityAuditSkill/3.1'})
        with urlopen(req, timeout=timeout) as resp:
            return resp.read().decode('utf-8', errors='replace')[:50000]
    except (URLError, OSError, UnicodeDecodeError) as e:
        print(f'  [WARN] Failed to fetch {url}: {e}', file=sys.stderr)
        return None

def generate_update_prompt(fetched_content: Dict[str, str]) -> str:
    today = datetime.now(timezone.utc).strftime('%Y-%m-%d')
    prompt = f'# Security Intelligence Update Analysis — {today}\n\nYou are updating the Security Audit Skill with the latest security intelligence.\nBelow is content fetched from security intelligence sources.\n\n## Your Task\n\n1. Identify NEW threats, vulnerabilities, tools, or best practices not yet in the skill\n2. For each finding, specify which reference file to update and what to add\n3. Format output as structured update proposals\n\n## Current Skill Coverage (do NOT duplicate these)\n\n- Frontend: React, Vue, Angular — XSS, CSP, Token, Headers\n- Backend: Python (FastAPI), Node.js, Go — Injection, Auth, Serialization\n- Desktop: Electron — nodeIntegration, IPC, Preload, Protocol, Auto-update\n- AI/LLM: Prompt Injection, Model Access, Output Safety, Data Privacy, Supply Chain\n- Cross-layer: Token flow, CORS, Error chain\n\n## Fetched Intelligence Content\n\n'
    for source_name, content in fetched_content.items():
        if content:
            truncated = content[:5000]
            prompt += f'### Source: {source_name}\n\n```\n{truncated}\n```\n\n'
    prompt += '\n## Output Format\n\nFor each new finding, output:\n\n```\n### UPDATE PROPOSAL [N]\n- **Source**: [Source name]\n- **Target file**: [reference file to update]\n- **Section**: [Where to add in that file]\n- **Type**: [New checklist item / New scan command / New tool / New threat / Updated info]\n- **Content**:\n  [Exact markdown to add]\n- **Rationale**: [Why this is important and new]\n```\n\nIf no updates are needed, output:\n```\n### NO UPDATES NEEDED\nAll current skill content is up to date with fetched intelligence.\n```\n'
    return prompt

def run_update(dry_run: bool=False, output_dir: Optional[str]=None):
    print(f'Security Audit Skill — Self-Update')
    print(f"Date: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M UTC')}")
    print(f'Skill directory: {SKILL_DIR}')
    print(f"{'=' * 50}")
    if dry_run:
        print('\n[DRY RUN] Would fetch from these sources:')
        for src in INTELLIGENCE_SOURCES:
            print(f"  - {src['name']}: {src['url']}")
        return
    fetched = {}
    for src in INTELLIGENCE_SOURCES:
        print(f"\nFetching: {src['name']}...")
        content = fetch_url(src['url'])
        if content:
            fetched[src['name']] = content
            print(f'  [OK] {len(content)} chars')
        else:
            print(f'  [SKIP] Failed to fetch')
    if not fetched:
        print('\n[ERROR] No sources could be fetched. Check network connectivity.')
        sys.exit(1)
    print(f"\n{'=' * 50}")
    print(f'Successfully fetched {len(fetched)}/{len(INTELLIGENCE_SOURCES)} sources')
    print(f"{'=' * 50}")
    prompt = generate_update_prompt(fetched)
    output_path = None
    if output_dir:
        os.makedirs(output_dir, exist_ok=True)
        today = datetime.now(timezone.utc).strftime('%Y-%m-%d')
        output_path = os.path.join(output_dir, f'update-proposal-{today}.md')
    output_content = f"# Security Intelligence Update Proposal\n\n> Generated: {datetime.now(timezone.utc).strftime('%Y-%m-%d %H:%M UTC')}\n> Sources fetched: {len(fetched)}/{len(INTELLIGENCE_SOURCES)}\n\n## Instructions for AI Agent\n\nCopy the following prompt into a conversation with the AI agent to generate\nupdate proposals. The agent will analyze the fetched content and suggest\nspecific additions to the skill's reference files.\n\n---\n\n{prompt}\n\n---\n\n## Next Steps\n\n1. Review the AI-generated update proposals\n2. Apply approved changes to reference files\n3. Update `references/changelog.md` with version bump\n4. Test the updated skill on a real project\n"
    if output_path:
        with open(output_path, 'w', encoding='utf-8') as f:
            f.write(output_content)
        print(f'\nUpdate proposal written to: {output_path}')
        print('Next: Feed this to the AI agent for analysis and skill updates.')
    else:
        print('\n' + output_content)

def main():
    parser = argparse.ArgumentParser(description='Security Audit Skill — Self-Update Script')
    parser.add_argument('--sources', action='store_true', help='List intelligence sources')
    parser.add_argument('--output-dir', help='Directory for update proposals')
    parser.add_argument('--dry-run', action='store_true', help='Show without fetching')
    args = parser.parse_args()
    if args.sources:
        list_sources()
    else:
        run_update(dry_run=args.dry_run, output_dir=args.output_dir)
if __name__ == '__main__':
    main()
