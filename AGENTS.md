# Project Guidelines

## Communication Style (Caveman Mode & Language)
- MANDATORY LANGUAGE RULE: ALWAYS communicate, respond, and chat in ENGLISH across all user interactions and conversations without exception.
- Terse smart caveman mode active.
- Drop articles, filler, pleasantries, hedging.
- Direct tool execution without preamble.
- Exact technical terms, file paths, and code blocks preserved.

## Codebase Navigation & Knowledge Graph (Graphify - Primary Reference)
- MANDATORY: Treat `graphify-out/graph.json` and `graphify-out/GRAPH_REPORT.md` as primary architectural reference across all chats.
- For all codebase questions, dependencies, and file relationships, first run `graphify query "<question>"`, `graphify explain "<concept>"`, or consult `graphify-out/`.
- Never do blind exploratory grepping when Graphify can provide the exact subgraph.
- Always run `graphify update .` after creating or modifying code files to keep the graph in sync.

## Frontend Craft & Design System (Impeccable & Design Taste)
- Maintain and follow `DESIGN.md` and `.impeccable/design.json` for all frontend work.
- Use standardized modern icon packs (Lucide, Tabler, Heroicons) with unified stroke-width (2.0 / 2.2).
- RTL-aware interface navigation: Next buttons point left (`←`), Previous buttons point right (`→`).
- Enforce anti-slop principles, responsive stability (`min-h-[100dvh]`, no viewport overflow), single color lock, and tactile micro-interactions.

NO DEPENDANCIES FOR THIS PROJECT... EVERYTHING MUST BE LOCAL FOR SECURITY REASONS (BY EVERYTHING I MEAN ASSETS AND NOT PLUGINS).
PLUGINS OR EXTERNAL LIBRARIES ARE FORBIDDEN.