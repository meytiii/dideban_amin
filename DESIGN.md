---
name: Samaneh Amn Setad Khabari Herasat - Hamadan
description: Secure public report portal and administrative intelligence system for the Security Department of Hamadan Education Directorate.
colors:
  primary: "#3b82f6"
  primary-hover: "#2563eb"
  primary-dark: "#1d4ed8"
  neutral-bg: "#0b1120"
  neutral-bg-admin: "#090d16"
  neutral-surface: "rgba(30, 41, 59, 0.72)"
  neutral-surface-solid: "#1e293b"
  neutral-elevated: "rgba(15, 23, 42, 0.65)"
  neutral-input: "rgba(15, 23, 42, 0.6)"
  border-subtle: "rgba(255, 255, 255, 0.08)"
  border-accent: "rgba(59, 130, 246, 0.4)"
  text-main: "#f8fafc"
  text-muted: "#94a3b8"
  text-subtle: "#64748b"
  status-success: "#10b981"
  status-danger: "#f43f5e"
  status-warning: "#f59e0b"
  status-cyan: "#06b6d4"
typography:
  display:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "26px"
    fontWeight: 700
    lineHeight: 1.4
    letterSpacing: "-0.5px"
  headline:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "20px"
    fontWeight: 700
    lineHeight: 1.5
  title:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "16px"
    fontWeight: 600
    lineHeight: 1.6
  body:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "14.5px"
    fontWeight: 400
    lineHeight: 1.7
  label:
    fontFamily: "Vazirmatn, Tahoma, Arial, sans-serif"
    fontSize: "13px"
    fontWeight: 500
    lineHeight: 1.5
rounded:
  sm: "6px"
  md: "10px"
  lg: "16px"
  full: "9999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.text-main}"
    rounded: "{rounded.sm}"
    padding: "14px 24px"
  button-primary-hover:
    backgroundColor: "{colors.primary-hover}"
  button-secondary:
    backgroundColor: "{colors.neutral-surface-solid}"
    textColor: "{colors.text-main}"
    rounded: "{rounded.sm}"
    padding: "14px 22px"
---

# Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان

## Overview

**Creative North Star: "The Sovereign Digital Sentinel - سامانه دیده‌بان امین"**

The visual language of the Herasat Portal bridges high-security institutional trust with ultra-modern digital craftsmanship. Engineered specifically for confidential intelligence and report handling, the interface emits calm authority, strict privacy reassurance, and uncompromising clarity. Deep obsidian and navy tones (`#0b1120`, `#090d16`, `#1e293b`) establish a solid cryptographic foundation, accented with laser-sharp cobalt blue (`#3b82f6`), emerald validation accents (`#10b981`), amber pending warnings (`#f59e0b`), and cyan intelligence badges (`#06b6d4`).

The experience is divided into two distinct operational modes:
1. **Persuade & Reassure (Public Portal):** A frictionless 3-step reporting wizard with transparent anonymity guarantees, drag-and-drop cryptographic evidence upload, instant tracking receipt issuance, and two-way encrypted followup threads.
2. **Operate (Admin Intelligence Console):** High-density, scan-optimized situational awareness, multi-level access control, comprehensive audit trails, automated profanity/vulgarity defense, Jalali date range filtration, and instant print/Word export readiness.

### Key Characteristics:
- **Directional RTL Excellence:** Native Right-to-Left geometry with contextual navigation arrows (Forward = Left `←`, Backward = Right `→`).
- **Standardized Iconography:** Pure Lucide & Tabler SVG glyphs with standardized 2.0px stroke-weight.
- **Glassmorphism & Depth:** Frosted dark slate surfaces with subtle 1px boundary highlights. Heavy backdrop blurs on large scrollable surfaces are strictly forbidden.
- **Micro-tactile Feedback:** Dynamic hover translations, push states, and non-blocking toast notifications.
- **Persian Administrative Calligraphy:** Native Kashida / Tatweel ligature balancing for Persian headings with full screen reader accessibility.

---

## Colors

The color palette centers on a dark-mode institutional system designed for high focus and zero visual fatigue.

### Primary
- **Cobalt Shield Blue** (`#3b82f6`): Primary action triggers, active step indicators, and interactive highlights.
- **Deep Slate Hover** (`#2563eb` / `#1d4ed8`): Focused and active button states.

### Neutral
- **Obsidian Deep Background (Public)** (`#0b1120`): Canvas background providing high-contrast base for citizen reporting.
- **Obsidian Dark Canvas (Admin)** (`#090d16`): Deeper tactical canvas for administrative intelligence operations.
- **Frosted Slate Surface** (`rgba(30, 41, 59, 0.72)`): Container cards and modal overlays.
- **Elevated Navy Surface** (`#1e293b`): Solid inner components, dropdowns, and drawer panels.
- **Pure Crisp White** (`#f8fafc`): Primary headline and body typography.
- **Muted Platinum** (`#94a3b8`): Subtitles, helper text, and secondary details.
- **Subtle Slate** (`#64748b`): Disabled states, minor timestamps, and placeholder copy.

### Status & Accents
- **Cryptographic Emerald** (`#10b981`): Success notifications, completed wizard steps, and resolved/reviewed status badges.
- **Security Rose** (`#f43f5e`): Destructive actions, file removal, profanity warnings, and validation errors.
- **Warning Amber** (`#f59e0b`): Character count limit warnings, unread status indicators, and pending report alerts.
- **Intelligence Cyan** (`#06b6d4`): Official regional responses, responded report indicators, and secondary intelligence feeds.

### Named Rules
**The Single Accent Rule.** The primary blue accent (`#3b82f6`) is preserved exclusively for interactive forward actions and focal identifiers. Background surfaces never shift hues arbitrarily.

---

## Typography

**Primary Font:** Vazirmatn (with fallback: Tahoma, Arial, sans-serif)
**Direction:** RTL (Right-to-Left) with LTR preservation on cryptographic codes, phone numbers, and CAPTCHAs.

### Hierarchy
- **Display / Portal Header** (Bold 700, 26px, line-height 1.4, tracking -0.5px): Primary system title and security banner.
- **Headline / Section Header** (Bold 700, 20px, line-height 1.5): Step headings, modal headers, and dossier cards.
- **Title / Form Labels** (Medium 500 / Semi-Bold 600, 14.5px-16px): Input field labels and drawer titles.
- **Body / Descriptive Text** (Regular 400, 14.5px, line-height 1.7): Guideline explanations, legal notices, and messages.
- **Label / Micro-Badges** (Medium 500, 12px-13px): Character counters, priority tags, and status chips.

### Persian Tatweel Typography Subsystem
- **Class:** `.persian-heading--tatweel` with `data-tatweel-level` (`portal`, `section`, `card`, `subcard`).
- **Engine:** `window.TatweelTypography` dynamically balances Persian typography using traditional Kashida (`\u0640`) elongation on valid Arabic/Persian ligatures.
- **Screen Reader Duality:** Elements generate a dual DOM structure: `.tatweel-visual` (`aria-hidden="true"`) for rendered presentation, and `.tatweel-sr` (`class="visually-hidden"`) preserving clean unextended text for screen readers and search engines.
- **Responsive Intensity Scaling:** Desktop (max 2 Kashidas), Tablet (max 1 Kashida), Mobile (0 Kashidas - auto-disabled to prevent line wrap breakage).
- **Auto-Dismiss:** Excluded on tables, textareas, inputs, and elements marked `data-tatweel="false"`.

### Named Rules
**The Persian Digit Rule.** Numbers inside general Persian prose and step counters utilize localized Persian numerals (`۰-۹`), while tracking codes (`16-character alphanumeric`) and contact phone numbers enforce explicit LTR formatting.

---

## Layout & Responsive Breakpoints

- **Public Container Model:** Maximum desktop constraint of `1020px` with central alignment and `35px 40px` internal padding.
- **Admin Container Model:** Maximum desktop constraint of `1550px` with central alignment and `30px` internal padding.
- **Responsive Breakpoint Matrix:**
  - `380px`: Ultra-compact mobile. Action buttons compress to icon-first layouts; font sizes scale down 1px.
  - `480px`: Mobile standard. KPI grids collapse to 1 column; modals take full viewport width; Tatweel elongation disables.
  - `640px`: Small tablet / phone landscape. Filters and form grids transition from 2-column to 1-column stack.
  - `768px`: Tablet portrait. KPI cards arrange in 2x2 grid; table responsive containers engage horizontal scroll.
  - `860px / 960px`: Tablet landscape. Header nav tabs wrap gracefully; filter bars align to responsive flex row.
  - `1024px+`: Standard desktop workstation view.
  - `1550px`: Admin dashboard maximum containment boundary.
- **Viewport Stability:** Stable viewport sizing using `min-h-[100dvh]` to eliminate mobile browser address bar layout shifts.
- **RTL Geometry:** Spatial margins and padding flip strictly according to RTL flow (`margin-left` for leading icons, `margin-right` for trailing indicators).

---

## Elevation & Depth

Surfaces employ tonal layering with subtle 1px border outlines rather than harsh drop shadows.

### Shadow Vocabulary
- **Card Ambient** (`0 20px 40px -15px rgba(0, 0, 0, 0.5)`): Main container elevation over background artwork.
- **Tactile Button Shadow** (`0 4px 12px rgba(0, 0, 0, 0.35)`): Solid elevation for primary action buttons.
- **Accent Glow** (`0 0 20px rgba(59, 130, 246, 0.35)`): Focus states, step tracker pulses, and primary action hovers.
- **Inner Edge Border** (`1px solid rgba(255, 255, 255, 0.08)`): Subtle glass refraction edge on all dark panels.

---

## Shapes & Radii

- **Base Radius:** `6px` (`--radius-sm`) for inputs, buttons, and badges.
- **Panel Radius:** `10px` (`--radius-md`) for info cards, step containers, and dropzones.
- **Container Radius:** `16px` (`--radius-lg`) for the primary card container and dialog modals.
- **Pill Radius:** `9999px` (`--radius-full`) for circular step indicators, status chips, and capsule tab bars.

---

## Components

### 1. Navigation & Pill Tab Bar
- Segmented dual-pill control (`ثبت گزارش جدید` and `پیگیری گزارش`) with sliding spring capsule icon animation (`.tab-icon-pill`) that traverses smoothly from right to left across the active tab.
- Drawer trigger button (`راهنما و سوالات متداول`) for opening the orthogonal-grid FAQ sliding drawer.

### 2. Multi-Step Form Tracker
- 3 distinct stages: `۱. هویت گزارش‌دهنده`, `۲. موضوع و شرح`, `۳. مستندات و ثبت`.
- Active circle pulses with glowing cobalt halo; completed steps illuminate with emerald checkmark (`✓`).

### 3. Glass Category Dropdown (`dropdown-menu`)
- Framer-motion inspired glassmorphic dropdown with Lucide categorized badge icons.
- Rotating ChevronDown (180°), blur-scale entry animation, keyboard Escape support, and auto-dismiss on outside click.

### 4. Centered Smart Evidence Upload (`be-ui-file-upload`)
- Centered drag-and-drop zone with animated cloud icon container and pill browse button.
- Real-time MIME & extension detection rendering dedicated SVG icons (Images, Archives/ZIPs, PDFs, Docs).
- Progress bar simulation (`#3b82f6` -> `#10b981`) and smart metadata chip.

### 5. FAQ Themed Accordion (`accordion-2`)
- Orthogonal background grid overlay (`32px x 32px` line matrix).
- Smooth collapsible items with rotating chevron (180°), dedicated category icons, and high-legibility response copy.

### 6. Admin 3D Sign-in Card (`sign-in-card-2`)
- Interactive 3D perspective card (`1200px`) tracking mouse coordinates (`rotateX`, `rotateY`).
- Quad traveling border light beams (top, right, bottom, left) with pulsing corner glows.
- Frosted crest container and password visibility eye toggling.

### 7. Action Buttons & Directional RTL Arrows
- **Next Step Button (`مرحله بعد`):** Primary blue gradient, trailing arrow pointing Left (`←`) with dynamic `-4px` hover translation.
- **Previous Step Button (`مرحله قبل`):** Secondary slate gradient, leading arrow pointing Right (`→`) with dynamic `+4px` hover translation.
- **Submit Button (`ثبت گزارش`):** Primary gradient with integrated shield-check icon and asynchronous loading spinner.

### 8. Secure Tracking & Follow-up Thread
- 16-character monospace tracking code container with one-click clipboard copy and print receipt capabilities.
- Live follow-up conversation thread rendering reporter remarks vs official security responses.

### 9. Admin KPI Stat Grid & Real-Time Pulse Indicators (`.kpi-grid`)
- High-density situational cards: Total Reports (Blue), Pending Verification (Amber), Reviewed (Emerald), and Responded (Cyan).
- Dynamic pulsating beacon dot (`.kpi-pulse-dot`) on pending cards that triggers upon incoming SSE real-time events.
- Click-to-filter capability mapping directly to corresponding status filters in the main report inbox.

### 10. Persian Jalali Date-Range Popover (`.date-range-popover`)
- Zero-dependency custom Jalali calendar popover with preset shortcut chips (`امروز`, `۷ روز اخیر`, `۳۰ روز اخیر`, `ماه جاری`, `همه`).
- Month-by-month pagination with native Persian month labels and Saturday-to-Friday (`ش-ج`) weekday headings.

### 11. Intelligence Data Table & Bulk Selection Action Bar (`#reportsTable`, `#bulkActionsToolbar`)
- Striped, RTL-aligned data table with sortable headers (`.sortable-th`), priority chips (`.badge-priority`), and read-state badges.
- Animated sliding bulk action bar appearing when rows are checked, enabling bulk status updates and administrative purges.
- Row state indicators (`.row-new-message`, `.row-regional-response`) for instant triage of escalated cases.

### 12. Official Classified Dossier & Word Document Export (`#report-detail-section`)
- Complete case dossier view presenting tracking metadata, sender classification, incident narration, and evidence attachments.
- Official printable letterhead (`.print-official-header`) with state emblem, classification notice, and print date stamps.
- Templated Word `.docx` generation (`/api/admin/report/export-docx`) conforming to formal administrative documentation standards.
- Layout height collapse rule: `#reports-inbox-section.in-detail-view` collapses container min-height to `0`, seating the dossier directly beneath KPI metric cards without empty space.

### 13. Accessible Confirmation Dialog System (`frontend/confirm.js`)
- Sovereign modal replacement for browser `window.confirm()`.
- Keyboard trap focus (`trapFocus`), Escape key dismissal, and backdrop click cancellation.
- Semantic intent variants: `danger` (destructive operations with rose primary button), `warning` (amber), and `info` (cobalt).
- Solid semi-translucent dark surfaces (`rgba(15, 23, 42, 0.95)`) avoiding expensive blur calculations.

### 14. Native Lightweight Tooltip System (`data-tooltip`)
- Pure CSS/JS attribute-driven tooltips (`[data-tooltip]`) providing contextual guidance on action buttons, table columns, and form fields.
- Automatic viewport edge collision detection and flip positioning to guarantee tooltips never overflow offscreen.

### 15. Active Session Countdown & Warning Modal
- Dedicated session timer widget (`.session-timer-box`) in admin navigation tracking countdown.
- Amber visual state within 2 minutes of expiration, accompanied by an accessible warning modal (`#sessionWarningModal`) providing one-click instant session renewal.

### 16. Animated Dropdown & Select System (`dropdown-01`)
- **Interactive Triggers:** Micro-scaling on hover (`scale(1.01)`) and tap (`scale(0.99)`) with hardware-accelerated `translateZ(0)`. Smooth 180° chevron rotation (`rotate(180deg)`) on active state. Calibrated responsive width with minimum 225px allocation on long text filters (subject and read status) to eliminate Persian label truncation.
- **Surface & Container:** Solid high-performance `#162032` surface with `1.5px` border outline (`rgba(255, 255, 255, 0.12)`), `12px` border-radius, and GPU compositor slide-down animation (`dropdownSlideDown`). Zero backdrop blur to maintain high framerates.
- **Staggered RTL Item Entrance:** Sequential cascade entrance from right (`translateX(16px)` to `0`) via `dropdownItemEnter` with indexed delays (`0.02s` to `0.26s`).
- **Spring Checkmark Feedback:** Dynamic spring bounce indicator (`checkSpring`) on active selected items.
- **Comprehensive Coverage Across Entire Portal:**
  - Public Portal: Report category selector with descriptive subtext helper notes.
  - Admin Reports Inbox: Subject (`filter-type`), Read status (`filter-read-status`), Priority (`filter-priority`), and Report type (`filter-report-type`) with two-way sync to table headers and quick filters.
  - Report Dossier Details: In-line dossier priority dropdown (`detail-priority-select`) with automatic role checking (badge display for regional inspectors, custom interactive select for central admins).
  - Admin Superuser Account Modals: Role permissions (`new_role`, `edit_role`) and account status toggle (`edit_is_active`).
  - Audit Trail Log Filters: Action type (`audit-filter-action`), administrative operator (`audit-filter-user` with dynamic real-time option re-population), and Jalali calendar month (`audit-filter-month`).

### 17. Server Datetime Display (`.server-datetime-badge`)
- Non-distracting, high-legibility server date and clock badge in admin header with tabular numeral sliding digit transitions.
- Tooltip-free and pulsing-indicator-free presentation to ensure a calm, distraction-free environment for security operators.

### 18. Assistant Communication & Chat Language
- **Assistant Chat Language:** ALWAYS speak, respond, and chat in English with user in all conversations and chats, while preserving Persian for client-facing UI strings and portal content.

### 19. Network Connection Monitoring & Immediate Session Revocation
- **Public & Admin Parity:** Both public and admin interfaces feature real-time connectivity banners fixed at `top: 0` (`.offline-banner`).
  - Offline Disconnect: `ارتباط با سرور قطع شده است. لطفاً اتصال اینترنت خود را بررسی کنید.` (Crimson red gradient `#dc2626` to `#991b1b` with offline Wi-Fi SVG icon).
  - Online Restored: `ارتباط با سرور برقرار شد.` (Emerald green gradient `#059669` to `#047857` with Wi-Fi SVG icon, auto-dismissing after 3 seconds).
- **Mandatory Admin Security Revocation:** Unlike the public reporting form where user draft inputs are preserved, dropping network connection in the administrative intelligence portal immediately revokes the active session cookie (`herasat_admin_session`), wipes session-level security tokens (`herasat_csrf`, `2fa_prompt_shown`), terminates SSE streams and active background polling intervals, closes all open modals, and immediately forces view transition back to `#login-section` with an explanatory security termination notice.
- **Pre-Submit Connectivity Guard:** Login and MFA verification forms verify connectivity prior to dispatching network credentials, providing instant feedback without hanging timeouts.

---

## Hardware Acceleration & Anti-Lag Performance Rules

- **Zero Backdrop-Filter on Large Surfaces:** Never use `backdrop-filter: blur(...)` across full cards, hero panels, or scrollable surfaces. Real-time multi-pass Gaussian blur over 1000px surfaces drops framerates to 15-20fps on low-end laptops and mobile phones. Use calibrated dark opaque or semi-translucent surfaces (`#162032`, `#1e293b`) with 1px border outlines.
- **Fixed Layer Isolation:** All fullscreen `position: fixed` backgrounds (`body::before`) must enforce `transform: translateZ(0); will-change: transform;` to permanently separate background rasterization from scrolling content.
- **Compositor-Only Animations:** Only animate `transform` (with `translateZ(0)`) and `opacity`. Never animate `box-shadow`, `filter`, or layout dimensions (`margin`, `padding`, `width`, `height`) on hover.
- **CSS Containment:** Apply `contain: layout style;` and `content-visibility: auto; contain-intrinsic-size: ...;` to major views (`#landingSection`, `#submitSection`, `#trackingSection`, `#dashboard-section`, `#report-detail-section`) to prevent layout recalculation cascades.
- **Asset Dimension Discipline:** Never load multi-megapixel images for small UI badges or logos. Scale assets to 2x retina max (e.g. 400x400 for 200x200 display) to prevent GPU memory texture thrashing.

---

## Do's and Don'ts

### Do:
- **Do** maintain standard 2.0px stroke-width on all SVG glyphs.
- **Do** point forward navigation arrows to the Left (`←`) and backward arrows to the Right (`→`) in RTL mode.
- **Do** provide clear feedback cycles for every async action (spinner + disabled state).
- **Do** preserve LTR direction on tracking codes, phone numbers, and CAPTCHA fields.
- **Do** isolate fixed background elements to their own GPU compositor layer (`translateZ(0)`).
- **Do** enforce CSS containment (`contain: layout style`) on modular page views.
- **Do** wrap decorative elongated Persian headings in `.persian-tatweel-wrapper` with an accompanying `.tatweel-sr` accessible element for screen readers.
- **Do** use `showConfirmDialog()` from `frontend/confirm.js` for all destructive user confirmations.

### Don't:
- **Don't** mix multiple icon sets or introduce emoji icons inside critical UI buttons.
- **Don't** use pure black `#000000` for dark mode surfaces—use layered obsidian/slate tones (`#0b1120`, `#090d16`, `#1e293b`).
- **Don't** allow form inputs or buttons to wrap text onto multiple lines at desktop breakpoints.
- **Don't** bypass validation checks or profanity screening before submission.
- **Don't** use `backdrop-filter: blur(...)` on large content containers or scrollable cards.
- **Don't** transition expensive `box-shadow` blurs on hover; animate `transform` instead.
- **Don't** put comments in codes.

NO DEPENDANCIES FOR THIS PROJECT... EVERYTHING MUST BE LOCAL FOR SECURITY REASONS (BY EVERYTHING I MEAN ASSETS AND NOT PLUGINS).
PLUGINS OR EXTERNAL LIBRARIES ARE FORBIDDEN.

## Mandatory Task Completion Rule: Commit Summary & Description
- Whenever a task is done, ALWAYS write a commit summary (≤50 characters, Conventional Commits format) and a clear, descriptive commit description fit for the task done so the user can commit to the GitHub repository.
- GitHub Repository: https://github.com/meytiii/dideban_amin