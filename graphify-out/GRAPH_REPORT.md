# Graph Report - Setad_Khabari_HMD  (2026-10-25)

## Corpus Check
- 241 files · ~403,482 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 2091 nodes · 3128 edges · 177 communities (137 shown, 40 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 290 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `7f94b354`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Components
- component-cookbook.md
- Running Tests
- Assets — sourcing canon for icons, logos, illustrations, photography, video
- mfa.go
- macrostructures.md
- Theme — Lumen
- Theme — Hum
- Custom craft — how to hand-build hero artwork
- Custom theme — protocol
- Design flow (default)
- SKILL.md
- Recipes
- Study — extracting design DNA from a screenshot or URL
- getIP
- Hero enrichment — when, what, and how much
- Theme — Carnival
- Interaction and states
- Architectural UI Components & Interaction Standards (Mandatory Rules)
- Major (looks AI-generated)
- app.js
- Slop test — 58 gates + pre-emit self-critique
- Copy
- Critical (ships as slop)
- Typography
- seed_50_reports.go
- ۲. تشریح سناریوهای آزمون نفوذ و حملات شبیه سازی شده
- NewClamAVScanner
- Microinteraction tells
- Archetype index — load ONLY the picks you need
- Theme — Cobalt
- Theme - Grid
- Responsive
- The six axes
- Minor (small taste issues)
- Imagery kit — curated abstract assets, hosted, ready
- admin/tatweel.js
- Genre — atmospheric
- Genre — modern-minimal
- Genre — playful
- § Multi-page flow — design.md first, then redesign
- Genre — editorial (default)
- Layout and space
- Motion
- design.md — opt-in portable design system
- Export formats
- Colour
- SyslogShipper
- Floating nav on scroll — the cross-fade morph
- SSEBroker
- action-search-bar.tsx
- Hallmark Design Directives
- initActionSearchBar
- pagination.tsx
- public/tatweel.js
- no_comments.md
- otp-input.tsx
- Comprehensive Documentation Audit & System Verification Report
- LoginHandler
- populateAnalyticsDOM
- sliding-number.tsx
- calendar.tsx
- Output contract & scope
- admin_handlers.go
- security.go
- Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان
- initServerSlidingClock
- crypto.go
- password-strength.tsx
- input.tsx
- herasat.local/backend
- Comprehensive Pre-Deployment Security Assessment & Vulnerability Audit
- label.tsx
- workflows/hallmark.md
- badge-2.tsx
- applyFilters
- toggle.tsx
- forceLogoutOnConnectionDrop
- metadata.go
- Colors
- rules/graphify.md
- workflows/graphify.md
- c2-inline-form-as-cta.md
- SubmitReportHandler
- generate_docx.py
- rules/caveman.md
- workflows/caveman.md
- profanity.go
- Authorized Dynamic Penetration Testing & Vulnerability Assessment Report
- Typography
- be-ui-file-upload.tsx
- sign-in-card-2.tsx
- Architectural UI Components & Interaction Standards (Mandatory Rules)
- Impeccable Design Quality & Craft Floor
- Design Taste Frontend Rules
- workflows/impeccable.md
- workflows/taste.md
- demo.tsx
- tooltip.tsx
- button-1.tsx
- accordion-2.tsx
- adminFetch
- button.tsx
- net/http.Request
- initAppTooltips
- morph-loading.tsx
- testing.T
- expandable-search-bar.tsx
- Running Locally with Docker (No Local PostgreSQL Installation Required)
- 2. Detailed Technical Evidence Captures
- openEditUserModal
- fetchReports
- main
- database/sql.DB
- admin/confirm.js
- frontend/confirm.js
- public/confirm.js
- 2. Implemented Code Changes
- popover.tsx
- createOtpController
- admin.js
- file-card-collections.tsx
- 2. Completed Fixes Detail
- createShamsiDateRangePicker
- f4-step-sequence.md
- commit_message.md
- Comprehensive Security Test Matrix
- 3. Resolved Pre-Evaluation Blockers (Verification Log)
- escapeHTML
- n11-mega-menu.md
- file-card-collections-demo.tsx
- initLoginSlider
- Technical Evidence: CAPTCHA Token Answer Leakage Analysis
- DA-SEC-001: CAPTCHA Answer Exposed in Plaintext in Signed Client Token
- DA-SEC-002: Missing Session Invalidation on Password Change and Reset
- DA-SEC-003: Denial of Service via Unbounded Pagination Query (`page_size=all`)
- DA-SEC-004: Internal Admin Username Exposure via Public Tracking Endpoint
- DA-SEC-005: Python Stderr Stack Trace & System Path Disclosure in Word Export
- DA-SEC-006: Incident Intake Velocity & Report ID Leakage via Global SSE
- DA-SEC-007: Missing CSRF Protection on Administrative Logout Endpoint
- DA-SEC-008: Insecure PostgreSQL Network Exposure & Static Credentials in Docker
- DA-SEC-009: Default Superuser Seeded with Static Hash and 2FA Disabled
- DA-SEC-010: ClamAV Antivirus Daemon Disabled by Default in Configuration
- ft7-newsletter-first.md
- Detailed Findings
- c4-sticky-bottom-bar.md
- f3-tabular-spec-sheet.md
- f5-annotated-screenshot.md
- getTestDB
- h9-custom-illustration-centerpiece.md
- s1-left-margin-numbered.md
- FINDINGS_VERIFICATION_MATRIX.md
- remediation_verification_test.go
- captcha.go
- s2-hanging.md
- NewIPRateLimiter
- contextWithAdmin
- generate_defense_pdf.py
- n12-banner-retract.md

## God Nodes (most connected - your core abstractions)
1. `getIP()` - 35 edges
2. `Components` - 35 edges
3. `adminFetch()` - 28 edges
4. `logAudit()` - 26 edges
5. `viewReport()` - 26 edges
6. `showToast()` - 20 edges
7. `initActionSearchBar()` - 20 edges
8. `main()` - 20 edges
9. `Major (looks AI-generated)` - 20 edges
10. `fetchReports()` - 19 edges

## Surprising Connections (you probably didn't know these)
- `SubmitReportHandler()` --calls--> `ScanPDFForActiveContent()`  [INFERRED]
  public_handlers.go → metadata.go
- `TestInputSanitizationAndLimits()` --calls--> `truncateRunes()`  [INFERRED]
  test/input_limits_test.go → security.go
- `VerifyMFAHandler()` --calls--> `generateSessionToken()`  [INFERRED]
  mfa.go → admin_auth.go
- `LoginHandler()` --calls--> `VerifyCaptcha()`  [INFERRED]
  admin_auth.go → captcha.go
- `LoginHandler()` --calls--> `GenerateMFAPendingToken()`  [INFERRED]
  admin_auth.go → mfa.go

## Import Cycles
- None detected.

## Communities (177 total, 40 thin omitted)

### Community 0 - "Components"
Cohesion: 0.06
Nodes (35): 10. Persian Jalali Date-Range Popover (`.date-range-popover`), 11. Intelligence Data Table & Bulk Selection Action Bar (`#reportsTable`, `#bulkActionsToolbar`), 12. Official Classified Dossier & Word Document Export (`#report-detail-section`), 13. Accessible Confirmation Dialog System (`frontend/confirm.js`), 14. Native Lightweight Tooltip System (`data-tooltip`), 15. Active Session Countdown & Warning Modal, 16. Animated Dropdown & Select System (`dropdown-01`), 17. Server Datetime Display (`.server-datetime-badge`) (+27 more)

### Community 1 - "component-cookbook.md"
Cohesion: 0.03
Nodes (38): C1 · Outlined chip, C3 · Typographic link, F1 · Bento grid, F2 · Sticky-scroll stack, F6 · Product card grid, Ft1 · Mast-headed, Ft2 · Inline-rule single line, Ft3 · Index-style category list (+30 more)

### Community 2 - "Running Tests"
Cohesion: 0.40
Nodes (4): Option 1: Temporary restore to root, Option 2: Run specific test with root source files, Running Tests, Test Suite

### Community 3 - "Assets — sourcing canon for icons, logos, illustrations, photography, video"
Cohesion: 0.05
Nodes (44): Abstract backgrounds, App mockups / device frames, Assets — sourcing canon for icons, logos, illustrations, photography, video, Avoid, Avoid, Avoid, Avoid, Avoid (+36 more)

### Community 4 - "mfa.go"
Cohesion: 0.21
Nodes (17): GetHMACKey(), Disable2FAHandler(), Enable2FAHandler(), GenerateBase32Secret(), GenerateMFAPendingToken(), GenerateTOTPCode(), normalizeTOTPCode(), Setup2FAHandler() (+9 more)

### Community 5 - "macrostructures.md"
Cohesion: 0.04
Nodes (28): 01 · Bento Grid, 02 · Long Document, 03 · Marquee Hero, 04 · Stat-Led, 05 · Workbench, 06 · Conversational FAQ, 07 · Manifesto, 08 · Photographic (+20 more)

### Community 6 - "Theme — Lumen"
Cohesion: 0.06
Nodes (33): 1. Hand-built apparatus (no orb), 2. Lowercase headline with verb-landmark (no italics), 3. Mono eyebrow above every section, 4. Blueprint grid background (4% opacity), 5. Meter strip (below the hero), 6. Hairline cards with inner emission, 7. Three-stat row with Instrument Serif numerals + tabular-nums, Anti-patterns (theme-specific) (+25 more)

### Community 7 - "Theme — Hum"
Cohesion: 0.06
Nodes (32): 1. The button system — three variations, the press is the feedback, 2. Multi-accent section bands, 3. Color-shift card grid, 4. Streak / counter tick-up, 5. Character moment — one small reacting mark, 6. Big rounded everything, 7. Star-burst micro-celebration on success, Anti-patterns (theme-specific) (+24 more)

### Community 8 - "Custom craft — how to hand-build hero artwork"
Cohesion: 0.07
Nodes (30): A hand-built SVG with declarative animation, A worked example — the bakery loaf as a single div, Animation choices for SVG in 2026, Anti-patterns of CSS art, Anti-patterns of declarative animation, Anti-patterns of generated stills, Anti-patterns of hand-built SVG, Anti-patterns of Three.js / WebGL (+22 more)

### Community 9 - "Custom theme — protocol"
Cohesion: 0.07
Nodes (29): § A · The one follow-up question, B.1 · Anchor accent first, B.2 · Paper, B.3 · Ink, B.4 · Supporting greys, B.5 · Focus, B.6 · Accent-ink (overlay text colour on accent), B.7 · Verification (+21 more)

### Community 10 - "Design flow (default)"
Cohesion: 0.07
Nodes (28): 0. Pre-flight scan, 1. Design-context gate, 2.5. Check project memory, 2.6. Theme route — studied-DNA, catalog, or custom, 2. Pick a macrostructure FIRST, 3. Load the visual ruleset, 4. Decide on hero enrichment, 5. Preview (+20 more)

### Community 11 - "SKILL.md"
Cohesion: 0.21
Nodes (3): H7 · Demo Video — Clipped-by-viewport-edge, Preview-block worked examples, `hallmark audit`

### Community 12 - "Recipes"
Cohesion: 0.07
Nodes (27): Accessibility ground truth, Button press, Command palette navigation, Copy-to-clipboard, Drag handle, Dropdown / menu, Form validation, Hard rules for default-on motion (+19 more)

### Community 13 - "Study — extracting design DNA from a screenshot or URL"
Cohesion: 0.08
Nodes (26): After the file is written, Emitting a `design.md` from `study`, Image-mode template, Junk-or-blocked detection, Limits and disclaimers, Refusal — when not to study, Remote URL safety, Source mode — image or URL (+18 more)

### Community 14 - "getIP"
Cohesion: 0.14
Nodes (29): computeAuditHash(), GetCurrentAdminHandler(), getIP(), getLatestAuditHash(), getLatestAuditHashTx(), ipInCIDRs(), logAudit(), logAuditTx() (+21 more)

### Community 15 - "Hero enrichment — when, what, and how much"
Cohesion: 0.09
Nodes (23): Animation discipline (hero specifically), Common mistakes — and the fixes, E1 · Demo Video — Clipped-by-viewport-edge, E2 · Demo Video — Full-bleed muted loop with ghost overlay, E3 · Mock App Screenshot — Browser-framed split, E4 · Mock App Screenshot — Floating no-frame, E5 · Custom Illustration Centerpiece, E6 · Animated Loop — pure CSS / SVG / Motion (+15 more)

### Community 16 - "Theme — Carnival"
Cohesion: 0.09
Nodes (22): Anti-patterns (theme-specific), Axes (diversification), Build hint, Drop 01 · Cold Snap *(default)*, Drop 02 · Citrus Riot, Drop 03 · Diner Sign, Drop 04 · Studio Night, Drop 05 · Aqua Park (+14 more)

### Community 17 - "Interaction and states"
Cohesion: 0.10
Nodes (21): Bans, Computation, Contrast discipline, Don't, list, Dropdowns, tooltips, popovers, Focus rings, Forms, Heights and rhythm (+13 more)

### Community 18 - "Architectural UI Components & Interaction Standards (Mandatory Rules)"
Cohesion: 0.09
Nodes (22): 10. Header Action Buttons (Icon-Only Pattern), 11. Animated Floating Label Input Architecture, 12. Modern Refined Badge Architecture (`badge-2`, `.badge-2`), 1. Form Validation & Persian Error Mandate, 2. Bespoke Top-Level Alert Architecture (`IconAlert`), 3. Dropdown Anti-Slop Selection Overlay Standard, 4. Universal Password Visibility Toggle, 5. 6-Cell OTP Square Input Architecture (+14 more)

### Community 19 - "Major (looks AI-generated)"
Cohesion: 0.10
Nodes (20): AI-illustration look, Animate-on-scroll on everything, Bounce and elastic easing, Centred everything, Eyebrow on every section, Generic emoji as feature icon, Glassmorphism without purpose, Hover-only affordances (+12 more)

### Community 20 - "app.js"
Cohesion: 0.06
Nodes (44): addFiles(), checkClientVulgarLanguage(), checkHostConnection(), clearDraft(), clearFileInput(), dismissToast(), escapeHtml(), _fastCheckClientVulgar() (+36 more)

### Community 21 - "Slop test — 58 gates + pre-emit self-critique"
Cohesion: 0.11
Nodes (19): Contrast & readability, Diversification gates, Hero enrichment gates, Honest copy · no fabricated content, Implementation gates, Input-state gate, Layout-safety gates, Microinteractions (+11 more)

### Community 22 - "Copy"
Cohesion: 0.11
Nodes (17): Austere, Banned opening lines (anti-patterns), Brutalist, Buttons, Copy, Editorial, Empty states, Error messages (+9 more)

### Community 23 - "Critical (ships as slop)"
Cohesion: 0.12
Nodes (17): Aurora-blob background, Card-in-card, Critical (ships as slop), Default-attractor sameness, Floating-orb decoration, Full-viewport centred hero, Inter-everywhere, Lazy-loaded LCP (+9 more)

### Community 24 - "Typography"
Cohesion: 0.12
Nodes (17): Banned defaults, Bans, Body text rules, Free body faces, Free display faces, Free mono / outlier faces, Headings rules, Hero headline sizing — match size to copy length (+9 more)

### Community 25 - "seed_50_reports.go"
Cohesion: 0.70
Nodes (4): deriveKey(), encryptWithKey(), genTrackingCode(), main()

### Community 26 - "۲. تشریح سناریوهای آزمون نفوذ و حملات شبیه سازی شده"
Cohesion: 0.06
Nodes (32): نتایج کلیدی ارزیابی:, گزارش جامع نتایج آزمون نفوذپذیری پویا و سازوکارهای دفاعی سامانه دیده بان امین, ۱. ارزیابی هدرهای امنیتی مرورگر (PENTEST-HDR-01), ۱. معماری رمزنگاری پیشرفته (AES-256-GCM و تفکیک کلیدها), ۱. چکیده مدیریتی و نتیجه کلان آزمون, ۱۱. حمله بهره برداری از نشست پس از سرقت و تغییر رمز (PENTEST-AUT-02), ۱۲ و ۱۳. حمله ارتقای سطح دسترسی عمودی (PENTEST-RBAC-01 & 02 - Vertical Privilege Escalation), ۱۴. حمله دسترسی غیرمجاز افقی به پرونده سایر نواحی (PENTEST-BOLA-01 - IDOR / BOLA) (+24 more)

### Community 27 - "NewClamAVScanner"
Cohesion: 0.24
Nodes (8): NewClamAVScanner(), io.Reader, ClamAVScanner, startMockClamAVServer(), TestClamAVCleanFile(), TestClamAVDisabledBypassesNetwork(), TestClamAVFailClosedWhenUnreachable(), TestClamAVMalwareFound()

### Community 28 - "Microinteraction tells"
Cohesion: 0.14
Nodes (14): Animated hover gradients, Auto-rotating carousels with no pause, Bouncy overshoot easings on UI, Celebratory success toasts, Confirmation dialogs for reversible actions, Cursor follower dots, Focus rings that animate in, Microinteraction tells (+6 more)

### Community 29 - "Archetype index — load ONLY the picks you need"
Cohesion: 0.14
Nodes (14): Archetype index — load ONLY the picks you need, Component cookbook, CTAs / signups, Feature blocks, Footers, Heroes, Mobile collapse — per archetype, Navigation (+6 more)

### Community 30 - "Theme — Cobalt"
Cohesion: 0.14
Nodes (13): Anti-patterns, Axes (diversification), Build hint, How Cobalt differs from neighbouring themes, Macrostructure affinity, Macrostructure rejection, Motion, Reference register (+5 more)

### Community 31 - "Theme - Grid"
Cohesion: 0.14
Nodes (13): Axes (diversification), Build hint, Do-nots (this theme's own failure modes), How Grid differs from its neighbours, Material, Motion, Reference register, The marks kit (+5 more)

### Community 32 - "Responsive"
Cohesion: 0.15
Nodes (13): Bans, Breakpoints, Clickable text — never wraps, Fluid scaling, Images, Internationalisation, Mobile — non-negotiable, Pointer and hover queries (+5 more)

### Community 33 - "The six axes"
Cohesion: 0.15
Nodes (13): 1. Section-heading placement, 2. Body composition, 3. Divider language, 4. Button voice, 5. Image treatment, 6. Reveal pattern, Anti-patterns of structural sameness, Domain → trio (offer these three; never default) (+5 more)

### Community 34 - "Minor (small taste issues)"
Cohesion: 0.18
Nodes (11): 100vw widths, Anti-patterns — the named tells, Double-hyphen dashes, Every section padded the same, How `hallmark audit` should report, Minor (small taste issues), Placeholder names, Startup-cliché product names (+3 more)

### Community 35 - "Imagery kit — curated abstract assets, hosted, ready"
Cohesion: 0.18
Nodes (11): Anti-patterns, Categories, Decoration, Empty state, Generation pipeline (out-of-band, one-time), Imagery kit — curated abstract assets, hosted, ready, Layered hero composition (the masterclass move), Manifest (placeholder until generation pass ships) (+3 more)

### Community 36 - "admin/tatweel.js"
Cohesion: 0.38
Nodes (9): applyWrapperToNode(), canConnect(), enhanceElement(), enhanceText(), enhanceWord(), getViewportIntensity(), init(), recheckOverflow() (+1 more)

### Community 37 - "Genre — atmospheric"
Cohesion: 0.20
Nodes (10): Genre — atmospheric, Nav and footer voice, Reference register, Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows (+2 more)

### Community 38 - "Genre — modern-minimal"
Cohesion: 0.18
Nodes (10): Genre — modern-minimal, Nav and footer voice, Reference register (for the LLM, not credited to anyone), Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows (+2 more)

### Community 39 - "Genre — playful"
Cohesion: 0.20
Nodes (10): Genre — playful, Nav and footer voice, Reference register, Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows (+2 more)

### Community 40 - "§ Multi-page flow — design.md first, then redesign"
Cohesion: 0.20
Nodes (10): 1. Read the project, then pause, 2. Produce `design.md` at the project root, 3. Redesign each page reading from `design.md`, 4. Diversification rule — INVERTED for multi-page, 5. When to amend `design.md` instead of overriding, `hallmark redesign`, § Multi-page flow — design.md first, then redesign, Non-destructive implementation rule (+2 more)

### Community 41 - "Genre — editorial (default)"
Cohesion: 0.20
Nodes (9): Genre — editorial (default), Nav and footer voice, Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows, What this genre disallows (+1 more)

### Community 42 - "Layout and space"
Cohesion: 0.22
Nodes (9): Asymmetry techniques, Bans, Depth, Grids, Layout and space, Page-edge clipping, Principles, The spacing scale (+1 more)

### Community 43 - "Motion"
Cohesion: 0.22
Nodes (9): Bans, Durations, Easings, Motion, Page-load orchestration, Principles, Reduced motion, Scroll-linked motion (+1 more)

### Community 44 - "design.md — opt-in portable design system"
Cohesion: 0.25
Nodes (8): After the file is written, CTA — surface the offer in the Step 5 preview, design.md — opt-in portable design system, Format (the tight version), Scope, Triggers (phrase-only — no new verb), Two emission paths (default vs study), Why opt-in (not auto-emit)

### Community 45 - "Export formats"
Cohesion: 0.25
Nodes (7): Export formats, Format 1 — `tokens.css`, Format 2 — Tailwind v4 `@theme`, Format 3 — DTCG `tokens.json`, Format 4 — shadcn/ui CSS variables, Output rule, Token taxonomy — Hallmark's source of truth

### Community 46 - "Colour"
Cohesion: 0.29
Nodes (7): Bans, Colour, Contrast, Dark mode recipe, Palette construction, Principles, Use of the accent

### Community 47 - "SyslogShipper"
Cohesion: 0.24
Nodes (9): net.Conn, sync.WaitGroup, SyslogMessage, SyslogSeverity, SyslogShipper, InitSyslogShipper(), MapActionToSyslogSeverity(), sanitizeSyslogParam() (+1 more)

### Community 48 - "Floating nav on scroll — the cross-fade morph"
Cohesion: 0.33
Nodes (6): Anti-patterns Hallmark refuses, Floating nav on scroll — the cross-fade morph, The four laws — non-negotiable, The property morph (10 properties, one curve), The scroll handler, The structure

### Community 49 - "SSEBroker"
Cohesion: 0.27
Nodes (4): sync.RWMutex, sseBroadcastMsg, SSEBroker, SSEClient

### Community 50 - "action-search-bar.tsx"
Cohesion: 0.40
Nodes (5): Action, ActionSearchBar(), allActions, SearchResult, useDebounce()

### Community 52 - "initActionSearchBar"
Cohesion: 0.15
Nodes (25): initActionSearchBar(), closeDropdown(), closeSearch(), computeAndRender(), executeItem(), openDropdown(), renderItems(), initAuditActionSearchBar() (+17 more)

### Community 53 - "pagination.tsx"
Cohesion: 0.21
Nodes (12): Pagination, PaginationEllipsis, PaginationEllipsisProps, PaginationItem, PaginationItemProps, paginationItemVariants, PaginationNavProps, paginationNavVariants (+4 more)

### Community 54 - "public/tatweel.js"
Cohesion: 0.38
Nodes (9): applyWrapperToNode(), canConnect(), enhanceElement(), enhanceText(), enhanceWord(), getViewportIntensity(), init(), recheckOverflow() (+1 more)

### Community 56 - "otp-input.tsx"
Cohesion: 0.17
Nodes (12): ALLOW, CROSSFADE, EASE, OtpCellProps, OtpInput(), OtpInputHandle, OtpInputProps, OtpMode (+4 more)

### Community 57 - "Comprehensive Documentation Audit & System Verification Report"
Cohesion: 0.18
Nodes (10): 1. Executive Summary & Verification Verdict, 2. Repository Subsystems & Files Systematically Inspected, 3. Test Suite & Empirical Verification Procedures, 4. Major Chapters Completed in Documentation Deliverables, 5. Important Findings, Known Limitations & Unverified Assumptions, 6. Exact Output Paths of All Deliverables, Comprehensive Documentation Audit & System Verification Report, Known Architectural Limitations: (+2 more)

### Community 58 - "LoginHandler"
Cohesion: 0.29
Nodes (12): generateSessionToken(), LoginHandler(), ChangeMyPasswordHandler(), CreateUserHandler(), ResetUserPasswordHandler(), HashPassword(), RunCryptographicSelfTests(), ValidatePasswordComplexity() (+4 more)

### Community 59 - "populateAnalyticsDOM"
Cohesion: 0.15
Nodes (7): populateAnalyticsDOM(), renderAnalyticsDashboard(), renderCategoryPieChart(), renderClassificationPieChart(), renderClientSideAnalytics(), renderMonthlyChart(), renderPrioritiesSegmentedBar()

### Community 62 - "Output contract & scope"
Cohesion: 0.50
Nodes (3): Output contract, Output contract & scope, Scope and limits

### Community 63 - "admin_handlers.go"
Cohesion: 0.11
Nodes (26): ExportReportDocxHandler(), formatJalaliTime(), getRestrictedSubprocessEnv(), gregorianToJalali(), AdminNotificationItem, AdminNotificationsResponse, AdminUserItem, AnalyticsResponse (+18 more)

### Community 64 - "security.go"
Cohesion: 0.18
Nodes (11): time.Time, hashToken(), RateLimitMiddleware(), truncateRunes(), ActiveSessionItem, AdminSession, cachedAccountFailure, cachedSession (+3 more)

### Community 65 - "Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان"
Cohesion: 0.15
Nodes (12): Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان, Do:, Do's and Don'ts, Don't:, Elevation & Depth, Hardware Acceleration & Anti-Lag Performance Rules, Key Characteristics:, Layout & Responsive Breakpoints (+4 more)

### Community 66 - "initServerSlidingClock"
Cohesion: 0.47
Nodes (6): fetchServerTime(), getServerNow(), initServerSlidingClock(), tick(), updateDate(), updateDigit()

### Community 67 - "crypto.go"
Cohesion: 0.17
Nodes (21): DecryptAES(), DecryptFileBytes(), DecryptFilePath(), decryptFilePayload(), decryptWithKey(), deriveKey(), EncryptAES(), EncryptFileBytes() (+13 more)

### Community 68 - "password-strength.tsx"
Cohesion: 0.15
Nodes (14): CELL, CROSSFADE, defaultLabels, defaultPasswordRules, EvaluatedRule, INSTANT, PasswordRule, PasswordStrength() (+6 more)

### Community 71 - "Comprehensive Pre-Deployment Security Assessment & Vulnerability Audit"
Cohesion: 0.07
Nodes (29): 10. Prioritized Remediation Roadmap, 11. Retesting & Quality Verification Sign-Off, 1. Executive Summary, 2.1 Objectives, 2.2 Scope of Assessment, 2.3 Explicit Exclusions, 2. Assessment Objectives, Scope & Exclusions, 3. Assessment Environment & Tooling (+21 more)

### Community 74 - "badge-2.tsx"
Cohesion: 0.28
Nodes (7): Badge(), BadgeButton(), BadgeButtonProps, badgeButtonVariants, BadgeDotProps, BadgeProps, badgeVariants

### Community 75 - "applyFilters"
Cohesion: 0.23
Nodes (12): applyAuditFiltersAndPagination(), applyFilters(), buildPaginationHTML(), closeDetailView(), getJalaliDateKeyFromDate(), getVisiblePaginationPages(), indexReportForSearch(), normalizePersianForSearch() (+4 more)

### Community 77 - "forceLogoutOnConnectionDrop"
Cohesion: 0.27
Nodes (10): cancelMFAStep(), checkHostConnection(), forceLogoutOnConnectionDrop(), hideOfflineBanner(), loadAdminCaptcha(), pingHost(), showOfflineBanner(), startRecoveryPolling() (+2 more)

### Community 78 - "metadata.go"
Cohesion: 0.29
Nodes (9): SanitizeAttachment(), SanitizeJPEG(), SanitizePNG(), ScanPDFForActiveContent(), createTestJPEGWithEXIF(), createTestPNG(), TestMetadataSanitizationJPEG(), TestMetadataSanitizationPNG() (+1 more)

### Community 79 - "Colors"
Cohesion: 0.40
Nodes (5): Colors, Named Rules, Neutral, Primary, Status & Accents

### Community 83 - "SubmitReportHandler"
Cohesion: 0.17
Nodes (18): RespondToReportHandler(), SubmitRegionalResponseHandler(), UpdateInternalNotesHandler(), VerifyCaptcha(), DecryptCaptchaPayload(), GetLegacyHMACKey(), ValidateBinaryMagicBytes(), generateTrackingCode() (+10 more)

### Community 84 - "generate_docx.py"
Cohesion: 0.50
Nodes (7): format_persian_text_with_dates(), main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph(), set_english_formatting(), set_persian_formatting()

### Community 87 - "profanity.go"
Cohesion: 0.20
Nodes (17): normalizeForSearch(), cleanRune(), collapseRepeatedRunes(), containsBoundedPhrase(), ContainsVulgarLanguage(), hasVulgarLatinSuffix(), hasVulgarPersianSuffix(), init() (+9 more)

### Community 88 - "Authorized Dynamic Penetration Testing & Vulnerability Assessment Report"
Cohesion: 0.09
Nodes (22): 1.1 Target Preconditions & Authorization, 1.2 Dedicated Test Accounts & Role Separation, 1. Testing Scope & Runtime Environment, 2. Testing Methodology & OWASP WSTG Mapping, 3. Dynamic Runtime Test Execution & Findings Matrix, 4.1 DA-SEC-001 Verification: CAPTCHA Token Confidentiality & Tampering, 4.2 DA-SEC-002 Verification: Immediate Session Revocation on Password Change, 4.3 DA-SEC-003 Verification: Resource Exhaustion Prevention in Cartable Pagination (+14 more)

### Community 89 - "Typography"
Cohesion: 0.50
Nodes (4): Hierarchy, Named Rules, Persian Tatweel Typography Subsystem, Typography

### Community 90 - "be-ui-file-upload.tsx"
Cohesion: 0.09
Nodes (28): clampProgress(), cn(), createFileUploadItem(), EASE_DRAWER, EASE_IN_OUT, EASE_OUT, EASE_OUT_CSS, FAST_TRANSITION (+20 more)

### Community 92 - "Architectural UI Components & Interaction Standards (Mandatory Rules)"
Cohesion: 0.09
Nodes (22): 10. Header Action Buttons (Icon-Only Pattern), 11. Animated Floating Label Input Architecture, 12. Modern Refined Badge Architecture (`badge-2`, `.badge-2`), 1. Form Validation & Persian Error Mandate, 2. Bespoke Top-Level Alert Architecture (`IconAlert`), 3. Dropdown Anti-Slop Selection Overlay Standard, 4. Universal Password Visibility Toggle, 5. 6-Cell OTP Square Input Architecture (+14 more)

### Community 100 - "button-1.tsx"
Cohesion: 0.50
Nodes (3): Button(), ButtonArrowProps, buttonVariants

### Community 103 - "adminFetch"
Cohesion: 0.14
Nodes (27): adminFetch(), createReportStatusBadge(), deleteReportPermanently(), dismissToast(), exportReportsToCSV(), fetchSystemStatus(), fetchWithTimeout(), formatJalaliDateTime() (+19 more)

### Community 104 - "button.tsx"
Cohesion: 0.24
Nodes (6): Button, ButtonProps, buttonVariants, ButtonWithIconProps, DropdownMenuProps, DropdownOption

### Community 105 - "net/http.Request"
Cohesion: 0.16
Nodes (20): GetActiveSessionsHandler(), GetAnalyticsHandler(), GetAuditLogsHandler(), GetForwardingHistoryHandler(), GetForwardingRecipientsHandler(), GetNotificationsHandler(), GetReportsHandler(), GetServerTimeHandler() (+12 more)

### Community 106 - "initAppTooltips"
Cohesion: 0.52
Nodes (7): initAppTooltips(), convertTitles(), getOrCreateTooltip(), hide(), isMobileOrTouch(), positionTooltip(), show()

### Community 108 - "testing.T"
Cohesion: 0.11
Nodes (23): testing.T, InitCookieSecurity(), TestClassificationValidationAndMath(), TestCookieSecureCachedResolution(), TestCookieSecureStartupFatalWithoutTLSOrFlag(), TestDBCredentialsRequiredAtStartup(), TestForwardHistoryItemWithRegionalResponse(), TestForwardHistorySerialization() (+15 more)

### Community 111 - "Running Locally with Docker (No Local PostgreSQL Installation Required)"
Cohesion: 0.22
Nodes (8): 1. Prerequisites, 2. Start PostgreSQL via Docker Compose, 3. Start the Backend Application, 4. Access the Services, 5. Stopping the Database, Default Administrative Credentials:, Didebane Amin (دیده‌بان امین), Running Locally with Docker (No Local PostgreSQL Installation Required)

### Community 113 - "2. Detailed Technical Evidence Captures"
Cohesion: 0.10
Nodes (20): 1. Dynamic Execution Matrix, 2. Detailed Technical Evidence Captures, Case 11: PENTEST-AUT-02 — Session Revocation on Password Rotation (DA-SEC-002), Case 14: PENTEST-BOLA-01 — Horizontal Object Access Boundary (BOLA / IDOR), Case 15: PENTEST-SSE-01 — Real-Time Notification Scoping (DA-SEC-006), Case 16: PENTEST-DOS-01 — Administrative Pagination Clamping (DA-SEC-003), Case 17: PENTEST-DOS-02 — Request Body Upper Bound Enforcement, Case 18: PENTEST-UPL-01 — Binary Magic Byte Validation (Executable in PNG) (+12 more)

### Community 114 - "openEditUserModal"
Cohesion: 0.29
Nodes (10): applyAdminUsersFilter(), deleteUser(), fetchAdminUsers(), liftUserLockout(), loadAdvSubView(), openEditUserModal(), renderAdminUsersPage(), renderUsersTable() (+2 more)

### Community 115 - "fetchReports"
Cohesion: 0.27
Nodes (12): connectAdminSSE(), fetchNotifications(), fetchReports(), markReportHasNewMessage(), renderNotificationsList(), renderReportsTableSkeleton(), renewSessionInteractive(), scheduleReportsRefresh() (+4 more)

### Community 116 - "main"
Cohesion: 0.11
Nodes (22): AdminAuthMiddleware(), SuperuserAuthMiddleware(), CSRFMiddleware(), isOriginAllowed(), enforceImmutableAuditLogs(), InitDB(), runMigrations(), compress/gzip.Writer (+14 more)

### Community 117 - "database/sql.DB"
Cohesion: 0.16
Nodes (9): NewUsedCaptchaStore(), database/sql.DB, sync.Mutex, sync.Once, time.Duration, time.Ticker, AccountLockoutManager, TOTPReplayStore (+1 more)

### Community 118 - "admin/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 119 - "frontend/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 120 - "public/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 121 - "2. Implemented Code Changes"
Cohesion: 0.11
Nodes (18): 1. Executive Summary, 2. Implemented Code Changes, 3. Test Suite Execution & Verification Evidence, 4. Operational & Deployment Actions Required, 5. Residual Risk Assessment, A. CAPTCHA Confidentiality & Integrity (`crypto.go`, `captcha.go`), B. Session Invalidation on Password Change & Reset (`admin_handlers.go`), C. Server-Side Pagination Bounds & Clamping (`admin_handlers.go`) (+10 more)

### Community 123 - "createOtpController"
Cohesion: 0.28
Nodes (8): createOtpController(), clearAll(), showError(), syncValue(), init2FAModalOtpControllers(), initMfaOtpInput(), showMFAStep(), toAsciiDigits()

### Community 125 - "admin.js"
Cohesion: 0.08
Nodes (30): check2FAEnforcementNotice(), checkAdminProfile(), checkMustChangePasswordPolicy(), evaluatePasswordStrength(), fetchForwardingRecipients(), getScrollBehavior(), handleKPICardFilter(), handleTimerTick() (+22 more)

### Community 126 - "file-card-collections.tsx"
Cohesion: 0.33
Nodes (3): colorBannerMap, FileCardProps, FormatFileProps

### Community 127 - "2. Completed Fixes Detail"
Cohesion: 0.12
Nodes (15): 1. Status of Planned Actions, 2. Completed Fixes Detail, 3. Operational Deployment Checklist (Pre-Flight), Action P0-1: CAPTCHA Encryption & Tamper Resistance, Action P1-1: Complete Session Revocation, Action P1-2: Server-Side Bounds on Administrative Cartable, Action P1-3: Docker & PostgreSQL Security, Action P1-4: Seeded Superuser Security Policy (+7 more)

### Community 128 - "createShamsiDateRangePicker"
Cohesion: 0.27
Nodes (13): createShamsiDateRangePicker(), closePopover(), commitDateRange(), openPopover(), renderCalendar(), renderCalendarHighlights(), resetRange(), getJalaliDayEndMs() (+5 more)

### Community 133 - "Comprehensive Security Test Matrix"
Cohesion: 0.17
Nodes (11): 1. Cryptography & Secrets Management (CRYPTO), 2. Authentication & Session Management (AUTH), 3. Authorization & Access Control (AUTHZ), 4. Input Validation, Abuse Prevention & CAPTCHA (INPUT), 5. File Uploads & Malware Defense (UPLOAD), 6. Database & Audit Integrity (DB), 7. Browser & HTTP Security (HTTP), 8. Deployment & Infrastructure Security (INFRA) (+3 more)

### Community 134 - "3. Resolved Pre-Evaluation Blockers (Verification Log)"
Cohesion: 0.18
Nodes (10): 1. Executive Context & Scope of AFTA Readiness, 2. AFTA Control Domain Evaluation, 3. Resolved Pre-Evaluation Blockers (Verification Log), 4. Evaluation Pre-Flight Steps for Testing Laboratory, Assessment Boundaries & Clarification of Status, Blocker 1: CAPTCHA Token Answer Leakage (`DA-SEC-001`) — RESOLVED & VERIFIED, Blocker 2: Orphan Sessions on Credential Modification (`DA-SEC-002`) — RESOLVED & VERIFIED, Blocker 3: Cartable Unbounded Pagination (`DA-SEC-003`) — RESOLVED & VERIFIED (+2 more)

### Community 135 - "escapeHTML"
Cohesion: 0.17
Nodes (15): escapeHTML(), fetchActiveSessions(), fetchAuditLogs(), initAdminCustomFilters(), initAdminCustomSelect(), closeMenu(), openMenu(), toggleMenu() (+7 more)

### Community 138 - "initLoginSlider"
Cohesion: 0.70
Nodes (5): initLoginSlider(), goToSlide(), nextSlide(), startTimer(), stopTimer()

### Community 144 - "Technical Evidence: CAPTCHA Token Answer Leakage Analysis"
Cohesion: 0.25
Nodes (7): Client-Side Extraction Proof, Execution Path Trace, Overview, Remediation Resolution & Verification (DA-SEC-001), Remediation Strategy, Root Cause, Technical Evidence: CAPTCHA Token Answer Leakage Analysis

### Community 145 - "DA-SEC-001: CAPTCHA Answer Exposed in Plaintext in Signed Client Token"
Cohesion: 0.33
Nodes (6): DA-SEC-001: CAPTCHA Answer Exposed in Plaintext in Signed Client Token, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 147 - "DA-SEC-002: Missing Session Invalidation on Password Change and Reset"
Cohesion: 0.33
Nodes (6): DA-SEC-002: Missing Session Invalidation on Password Change and Reset, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 149 - "DA-SEC-003: Denial of Service via Unbounded Pagination Query (`page_size=all`)"
Cohesion: 0.33
Nodes (6): DA-SEC-003: Denial of Service via Unbounded Pagination Query (`page_size=all`), Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 150 - "DA-SEC-004: Internal Admin Username Exposure via Public Tracking Endpoint"
Cohesion: 0.33
Nodes (6): DA-SEC-004: Internal Admin Username Exposure via Public Tracking Endpoint, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 151 - "DA-SEC-005: Python Stderr Stack Trace & System Path Disclosure in Word Export"
Cohesion: 0.33
Nodes (6): DA-SEC-005: Python Stderr Stack Trace & System Path Disclosure in Word Export, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 156 - "DA-SEC-006: Incident Intake Velocity & Report ID Leakage via Global SSE"
Cohesion: 0.33
Nodes (6): DA-SEC-006: Incident Intake Velocity & Report ID Leakage via Global SSE, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 157 - "DA-SEC-007: Missing CSRF Protection on Administrative Logout Endpoint"
Cohesion: 0.33
Nodes (6): DA-SEC-007: Missing CSRF Protection on Administrative Logout Endpoint, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 158 - "DA-SEC-008: Insecure PostgreSQL Network Exposure & Static Credentials in Docker"
Cohesion: 0.33
Nodes (6): DA-SEC-008: Insecure PostgreSQL Network Exposure & Static Credentials in Docker, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 159 - "DA-SEC-009: Default Superuser Seeded with Static Hash and 2FA Disabled"
Cohesion: 0.33
Nodes (6): DA-SEC-009: Default Superuser Seeded with Static Hash and 2FA Disabled, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 160 - "DA-SEC-010: ClamAV Antivirus Daemon Disabled by Default in Configuration"
Cohesion: 0.33
Nodes (6): DA-SEC-010: ClamAV Antivirus Daemon Disabled by Default in Configuration, Impact, Recommended Remediation, Remediation & Retest Verification, Root Cause, Technical Evidence & Reproduction

### Community 162 - "Detailed Findings"
Cohesion: 0.50
Nodes (3): Detailed Findings, Findings Summary Matrix, Security Findings Register — Didebane Amin

### Community 166 - "getTestDB"
Cohesion: 0.16
Nodes (11): NewAccountLockoutManager(), NewSessionManager(), TestAuditFailClosed(), TestCaptchaPersistenceTryRedeem(), TestAccountLockoutConcurrentFailures(), TestSecurityStoresInitOrder(), getTestDB(), TestSessionManagementAndRateLimiter() (+3 more)

### Community 170 - "remediation_verification_test.go"
Cohesion: 0.19
Nodes (10): NewSSEBroker(), parseInteger(), parsePositiveInt(), TestPublicTrackingOmitsAdminUsername(), TestRemediationPaginationClampingBounds(), TestRemediationSessionRevocationLifecycle(), TestRemediationSSEBrokerRoleScoping(), TestAdminEventsHandlerStreamAndContextDone() (+2 more)

### Community 171 - "captcha.go"
Cohesion: 0.29
Nodes (11): cryptoRandInt(), GenerateCaptchaToken(), GenerateMathCaptchaSVG(), generateNoiseElements(), GetCaptchaHandler(), renderObfuscatedChars(), EncryptCaptchaPayload(), CaptchaPayload (+3 more)

### Community 173 - "NewIPRateLimiter"
Cohesion: 0.33
Nodes (5): inspectZipArchive(), NewIPRateLimiter(), TestCaptchaReplayPrevention(), TestInspectZipArchive(), TestIPRateLimiterBoundedCapacity()

## Knowledge Gaps
- **979 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `RecipientUserItem`, `ForwardReportRequest`, `ForwardHistoryItem` (+974 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **40 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Anti-patterns — the named tells` connect `Minor (small taste issues)` to `Major (looks AI-generated)`, `SKILL.md`, `Microinteraction tells`, `Critical (ships as slop)`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `Assets — sourcing canon for icons, logos, illustrations, photography, video` connect `Assets — sourcing canon for icons, logos, illustrations, photography, video` to `SKILL.md`?**
  _High betweenness centrality (0.015) - this node is a cross-community bridge._
- **Why does `Major (looks AI-generated)` connect `Major (looks AI-generated)` to `Minor (small taste issues)`?**
  _High betweenness centrality (0.009) - this node is a cross-community bridge._
- **Are the 29 inferred relationships involving `getIP()` (e.g. with `ChangeMyPasswordHandler()` and `ClassifyReportHandler()`) actually correct?**
  _`getIP()` has 29 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `RecipientUserItem` to the rest of the system?**
  _979 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Components` be split into smaller, more focused modules?**
  _Cohesion score 0.05714285714285714 - nodes in this community are weakly interconnected._
- **Should `component-cookbook.md` be split into smaller, more focused modules?**
  _Cohesion score 0.02531645569620253 - nodes in this community are weakly interconnected._