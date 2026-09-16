---
name: hallmark
description: Anti-AI-slop design skill for greenfield pages, audits, redesigns, and design extraction from URLs or screenshots
---

# /hallmark Workflow

1. Execute Pre-flight scan (Step 0) & inspect local `design.md` or tokens.
2. Resolve Design-context gate (Step 1): Audience, Use case, Tone (editorial, modern-minimal, atmospheric, playful).
3. Pick Macrostructure first from named index, rotate theme per diversification rules, and pick nav/footer archetypes.
4. Verify interactive state coverage (all 8 states), OKLCH token lock, mobile responsive discipline (320px–768px), and hero headline sizing.
5. Execute Build and run 58-gate slop test before emitting code.
6. Support explicit subcommands:
   - `hallmark audit <target>`
   - `hallmark redesign <target> [--mood <name>]`
   - `hallmark study <screenshot | URL>`
