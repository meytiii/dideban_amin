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

## Architectural UI Components & Interaction Standards (Mandatory Rules)

### 1. Form Validation & Persian Error Mandate
- FORBIDDEN: Default browser English error popups or native bubbles. Forms must use `novalidate` or override default reporting.
- Dynamic Persian Field Extraction: Automatically extract field name from nearest `<label>` or `placeholder`, strip trailing asterisks (`*`) or colons (`:`), and format error as: `لطفاً فیلد «{نام فیلد}» را تکمیل نمایید.`
- Visual & Motion Feedback: Failed inputs immediately receive `.input-error` and `.input-error-shake` (horizontal shake animation). Input error styling MUST clear dynamically as soon as the user interacts (`input` or `change` event).
- Persian Toasts: Display errors via non-blocking Persian toast notifications (`showToast(msg, 'error')`).

### 2. Bespoke Top-Level Alert Architecture (`IconAlert`)
- FORBIDDEN: Generic Bootstrap-style or plain colored alert banners.
- Mandatory Structure: All in-page feedback and modal alerts must use the bespoke `IconAlert` component featuring:
  - Solid/translucent dark slate background (`rgba(..., 0.12)`).
  - Subtle 1px boundary border matching alert intent tone (`rgba(..., 0.28)`).
  - 12px rounded corners (`rounded-xl` / `--radius-md`).
  - Dedicated circular/pill icon badge containing a standardized Lucide SVG (stroke-width: 2.0 / 2.2).
  - High-contrast bold title (`.icon-alert-title`) + clean message body (`.icon-alert-desc`).
  - Clean dismiss button (`.icon-alert-close`) with smooth fade-out.

### 3. Dropdown Anti-Slop Selection Overlay Standard
- BANNED: Vertical colored border strips on selected items (e.g. `border-right: 3px solid #38bdf8`), which is cookie-cutter AI slop.
- Mandatory Selected Style: Selected items in dropdowns must use a subtle dark-tech background overlay (`rgba(255, 255, 255, 0.08)` or subtle accent tint `rgba(59, 130, 246, 0.12)`), an optional subtle inset highlight (`box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.06)`), and an animated spring checkmark indicator (`checkSpring`). Hover states must remain smooth without causing layout shift.

### 4. Universal Password Visibility Toggle
- Reusable Container: All password inputs across public portal and admin panel must be wrapped in `.password-wrapper` with relative positioning.
- Toggle Button: Positioned absolute inside the input padding (`left: 10px` in RTL) using `.btn-toggle-pw`.
- Dual Lucide SVGs: Contains both Eye (`.eye-icon-show`) and EyeOff (`.eye-icon-hide.d-none`).
- Event Delegation & Modal Reset: Must use global document-level delegation (`$(document).on('click', '.btn-toggle-pw', ...)`). Whenever modals or forms reset or close, input types must automatically revert to `type="password"`, showing `.eye-icon-show` and hiding `.eye-icon-hide`.

### 5. 6-Cell OTP Square Input Architecture
- Reusable Container: Standardized `.otp-input-container` with 6 individual `.otp-cell` inputs (48x48px on desktop, 42x42px on mobile) split into 3+3 cells with a central `.otp-divider`.
- Interaction Architecture: Auto-advancing on single digit entry, backspace navigation to previous cell, left/right arrow navigation, and complete 6-digit paste handling distributing numbers across cells with auto-focus on the 6th cell.
- State Classes: Discrete `.is-error` (red border + shake) and `.is-success` (emerald pulse).
- Component Parity: Same visual design and controller logic applied to both 2FA login verification and 2FA activation modal.

### 6. Status Badge Single Source of Truth & Public-Admin Parity
- FORBIDDEN: Color discrepancies between public report tracking and admin cartable.
- Mandatory Palette Lock:
  - `read` / "بررسی شده": Intelligence Cyan (`#06b6d4` / `#2de4da`), bg `rgba(6, 182, 212, 0.14)`, border `rgba(6, 182, 212, 0.35)`.
  - `new` / "جدید" (در نوبت بررسی): Warm Amber (`#EAA65D`).
  - `in_progress` / "ارجاع شده": Electric Sky Blue (`#008AF5`).
  - `responded` / "پاسخ داده شده": Cryptographic Emerald (`#10b981` / `#57BC6C`).
  - `rejected` / "رد شده": Security Rose (`#D57463` / `#f43f5e`).
  - `archived` / "بایگانی شده": Slate Zinc (`#777777`).

### 7. Precision Error Routing & Shake Scoping
- In multi-field asynchronous forms (such as 2FA setup requiring both password confirmation and OTP code), errors must be strictly scoped to the exact failing field.
- If password verification fails, shake ONLY the password input (`.input-error` and `.input-error-shake`). NEVER inappropriately shake or wipe unrelated inputs (like OTP cells) when the password caused the error.

### 8. Contextual Action Buttons & Data Integrity
- Action buttons tied to conditional data (such as clipboard copy buttons next to reporter phone numbers) must NEVER be displayed if the underlying data is missing or empty.
- If field is empty or anonymous, the action button is strictly hidden (`display: none`), displaying only the neutral placeholder badge (e.g. `بدون شماره تماس`).

### 9. Modal Viewport Ergonomics & Responsive Heights
- High-occupancy modal lists (such as the expert referral selection modal) must never be constrained to miniature heights that force premature scrolling.
- Minimum comfortable height: `max-height: min(450px, 48vh)` or up to `90vh` modal constraint, allowing 7–8 items to be comfortably visible at a glance.
- Mobile Safeguards: Media query `@media (max-height: 700px)` dynamically adjusts modal heights to prevent viewport overflow on small laptops and landscape devices.

### 10. Header Action Buttons (Icon-Only Pattern)
- Topbar utility triggers (such as the notification bell `#btnNotifBell`) must remain clean, icon-only buttons (36x36px or 40x40px) with centered Lucide SVG glyphs. Text labels are omitted, and context is provided via accessible `data-tooltip`.

### 11. Animated Floating Label Input Architecture
- Mandatory Pattern: All text inputs and textareas across public and admin interfaces must be wrapped in `.floating-group` with `.floating-label`.
- Zero Collision Mandate: Text fields using floating labels must strictly set `placeholder=" "` (single space) and NEVER define or dynamically inject inline placeholder strings. The animated label itself acts as the placeholder in resting state, floating up on focus or value presence.

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