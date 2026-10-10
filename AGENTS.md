# Project Guidelines

## Communication Style (Caveman Mode & Language)
- MANDATORY LANGUAGE RULE: ALWAYS communicate, respond, and chat in ENGLISH across all user interactions and conversations without exception.
- Terse smart caveman mode active (/caveman).
- Drop articles, filler, pleasantries, hedging.
- Do NOT sugarcoat responses. Be direct, clear, objective.
- Direct tool execution without preamble.
- Exact technical terms, file paths, and code blocks preserved.

## Zero Assumptions & Interview Mandate (/grill-me)
- FORBIDDEN: Never assume anything or leave doubts. If requirement ambiguous, underspecified, or too general, stop and interview user via `/grill-me` (one question at a time using `ask_question`).
- Pre-task Walkthrough: Before executing any task (after interview alignment), write detailed walkthrough artifact. Wait for user to review and proceed before touching code.

## Codebase Navigation & Knowledge Graph (Graphify - Primary Reference)
- MANDATORY FIRST STEP: For every task, observe and consult `graphify-out/graph.json`, `graphify-out/GRAPH_REPORT.md`, or run `graphify query "<question>"` to preserve token budget and avoid blind file scanning.
- Never do blind exploratory grepping when Graphify can provide the exact subgraph.
- Always run `graphify update .` after creating or modifying code files to keep graph in sync.

## Frontend Craft, Design System & Anti-AI-Slop Discipline
- Maintain and follow `DESIGN.md` and `.impeccable/design.json` for all frontend work.
- Always apply design skills: `/impeccable`, `/hallmark`, `/taste`, `/ui-ux-pro-max`, `/ux-ui-agent-skills`, `/high-end-visual-design` to guarantee bespoke, high-craft interfaces.
- BANNED: Generic AI templates, cookie-cutter components, and repetitive boilerplate slop. Interface must be unique, following `DESIGN.md` specifications or user reference samples.
- Humanized UI Text: Always apply `/humanizer` to all visible frontend copy, strings, labels, and notices. No UI text should ever feel AI-generated.
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

## Engineering Standards & Step Verification
- Verify each step systematically.
- Ensure final code is 100% functional, robust, and adheres to strict senior-level engineering standards.

## Mandatory Task Completion & Auto-Commit Rule
- Whenever a task is done, ALWAYS write a commit summary (≤50 characters, Conventional Commits format) and a clear, descriptive commit description fit for the task done.
- Automatically commit changes to the main branch of the repository: https://github.com/meytiii/dideban_amin.