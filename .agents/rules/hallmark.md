# Hallmark Design Directives

## Principles
- **Anti-AI-Slop & Structural Variety:** Ban predictable AI design fingerprints (hero -> 3 cards -> CTA -> footer). Enforce varied macrostructures from the index.
- **Strict Slop Test:** Pass 58-gate slop test on UI emit. Roman headings only, honest copy (no invented metrics), locked tokens.
- **Responsive Stability:** Mandatory mobile viewport verification (320px, 375px, 414px, 768px), `overflow-x: clip`, no text clipping.
- **Interactive State Discipline:** All 8 interactive states implemented (default, hover, focus-visible, active, disabled, loading, error, success).
- **Project Memory & Diversification:** Track builds in `.hallmark/log.json` to prevent repetitive macrostructures and themes across consecutive runs.
