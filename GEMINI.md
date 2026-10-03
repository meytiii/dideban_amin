# Project Guidelines (HMD IT Herasat Portal)

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

## UI Performance & Hardware Acceleration Discipline (Anti-Lag Rules)
- FORBIDDEN: Never use `backdrop-filter: blur(...)` on large surfaces, hero cards, or scrollable containers. Blurring large viewports drops framerates to 15-20fps on low-end hardware and mobile devices. Use calibrated solid/semi-translucent surfaces (`#162032`, `#1e293b`) with subtle 1px border outlines.
- ALWAYS isolate fixed backgrounds (`body::before`) to their own GPU compositor layer using `transform: translateZ(0); will-change: transform;` to eliminate scroll repaints.
- Compositor-Only Transitions: Only animate `transform` (with `translateZ(0)`) and `opacity`. Never transition `box-shadow`, `filter`, or layout dimensions (`margin`, `padding`, `width`, `height`) on hover.
- CSS Containment: Always apply `contain: layout style;` or `content-visibility: auto; contain-intrinsic-size: ...;` to major views and sections to prevent layout recalculation cascades.
- Asset Dimension Discipline: Never load multi-megapixel textures for small UI badges or logos. Scale assets to 2x retina max (e.g. 400x400 for 200x200 display) to prevent GPU memory texture thrashing.

NO DEPENDANCIES FOR THIS PROJECT... EVERYTHING MUST BE LOCAL FOR SECURITY REASONS (BY EVERYTHING I MEAN ASSETS AND NOT PLUGINS).
PLUGINS OR EXTERNAL LIBRARIES ARE FORBIDDEN.

## Mandatory Code Rule: NO COMMENTS
- NEVER write comments in any code (Go, JavaScript, TypeScript, Python, SQL, HTML, CSS, etc.) ever.
- All code files must remain 100% comment-free at all times.

## Mandatory Task Completion Rule: Commit Summary & Description
- Whenever a task is done, ALWAYS write a commit summary (≤50 characters, Conventional Commits format) and a clear, descriptive commit description fit for the task done so the user can commit to the GitHub repository.
- GitHub Repository: https://github.com/meytiii/dideban_amin