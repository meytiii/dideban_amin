# Graph Report - Setad_Khabari_HMD  (2026-10-27)

## Corpus Check
- 202 files · ~357,182 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1826 nodes · 3225 edges · 108 communities (85 shown, 23 thin omitted)
- Extraction: 85% EXTRACTED · 15% INFERRED · 0% AMBIGUOUS · INFERRED: 494 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `899d00fb`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Components
- component-cookbook.md
- crypto.go
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
- admin_handlers.go
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
- admin.js
- NewClamAVScanner
- Microinteraction tells
- Archetype index — load ONLY the picks you need
- Theme — Cobalt
- Theme - Grid
- Responsive
- The six axes
- Minor (small taste issues)
- Imagery kit — curated abstract assets, hosted, ready
- initActionSearchBar
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
- Round 1 Summary (Phases 0 - 6)
- db.go
- Hallmark Design Directives
- viewReport
- createShamsiDateRangePicker
- initServerSlidingClock
- no_comments.md
- remediation_verification_test.go
- populateAnalyticsDOM
- getTestDB
- Output contract & scope
- database/sql.DB
- Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان
- TestMain
- c1-outlined-chip.md
- c3-typographic-link.md
- herasat.local/backend
- applyFilters
- f6-product-card-grid.md
- workflows/hallmark.md
- n12-banner-retract.md
- adminFetch
- SSEBroker
- metadata.go
- Colors
- rules/graphify.md
- workflows/graphify.md
- s3-sticky-pinned.md
- generate_docx.py
- rules/caveman.md
- workflows/caveman.md
- profanity.go
- SubmitReportHandler
- Typography
- GetHMACKey
- Architectural UI Components & Interaction Standards (Mandatory Rules)
- initAdminCustomSelect
- Impeccable Design Quality & Craft Floor
- Design Taste Frontend Rules
- workflows/impeccable.md
- workflows/taste.md
- admin/js/tatweel.js
- public/js/tatweel.js
- createOtpController
- testing.T
- Dideban-e Amin (دیده‌بان امین)
- Macrostructures
- initAppTooltips
- main
- admin/js/confirm.js
- public/js/confirm.js
- initLoginSlider
- VerifyCaptcha
- f4-step-sequence.md
- commit_message.md
- f5-annotated-screenshot.md
- n11-mega-menu.md
- h9-custom-illustration-centerpiece.md

## God Nodes (most connected - your core abstractions)
1. `adminFromContext()` - 39 edges
2. `getIP()` - 38 edges
3. `Components` - 35 edges
4. `main()` - 30 edges
5. `adminFetch()` - 28 edges
6. `logAudit()` - 26 edges
7. `viewReport()` - 26 edges
8. `InitCrypto()` - 23 edges
9. `SubmitReportHandler()` - 22 edges
10. `showToast()` - 20 edges

## Surprising Connections (you probably didn't know these)
- `init()` --calls--> `HashPassword()`  [INFERRED]
  admin_auth.go → crypto.go
- `LoginHandler()` --calls--> `VerifyCaptcha()`  [INFERRED]
  admin_auth.go → captcha.go
- `LoginHandler()` --calls--> `GenerateMFAPendingToken()`  [INFERRED]
  admin_auth.go → mfa.go
- `getIP()` --calls--> `GetTrustedProxies()`  [INFERRED]
  admin_auth.go → network.go
- `getIP()` --calls--> `ipInCIDRs()`  [INFERRED]
  admin_auth.go → network.go

## Import Cycles
- None detected.

## Communities (108 total, 23 thin omitted)

### Community 0 - "Components"
Cohesion: 0.06
Nodes (35): 10. Persian Jalali Date-Range Popover (`.date-range-popover`), 11. Intelligence Data Table & Bulk Selection Action Bar (`#reportsTable`, `#bulkActionsToolbar`), 12. Official Classified Dossier & Word Document Export (`#report-detail-section`), 13. Accessible Confirmation Dialog System (`frontend/confirm.js`), 14. Native Lightweight Tooltip System (`data-tooltip`), 15. Active Session Countdown & Warning Modal, 16. Animated Dropdown & Select System (`dropdown-01`), 17. Server Datetime Display (`.server-datetime-badge`) (+27 more)

### Community 1 - "component-cookbook.md"
Cohesion: 0.02
Nodes (39): C2 · Inline form-as-CTA, C4 · Sticky bottom bar, F1 · Bento grid, F2 · Sticky-scroll stack, F3 · Tabular spec sheet, Ft1 · Mast-headed, Ft2 · Inline-rule single line, Ft3 · Index-style category list (+31 more)

### Community 2 - "crypto.go"
Cohesion: 0.12
Nodes (34): DecryptAES(), DecryptBoundField(), DecryptFileBytes(), DecryptFilePath(), decryptFilePayload(), DecryptWithAAD(), decryptWithKey(), deriveKey() (+26 more)

### Community 3 - "Assets — sourcing canon for icons, logos, illustrations, photography, video"
Cohesion: 0.05
Nodes (44): Abstract backgrounds, App mockups / device frames, Assets — sourcing canon for icons, logos, illustrations, photography, video, Avoid, Avoid, Avoid, Avoid, Avoid (+36 more)

### Community 4 - "mfa.go"
Cohesion: 0.32
Nodes (10): GenerateBase32Secret(), GenerateTOTPCode(), normalizeTOTPCode(), VerifyTOTPCode(), MFAPendingPayload, TestConcurrentTOTPVerificationRace(), TestGenerateTOTPSecret(), TestTOTPGenerationAndVerification() (+2 more)

### Community 5 - "macrostructures.md"
Cohesion: 0.05
Nodes (21): 01 · Bento Grid, 02 · Long Document, 03 · Marquee Hero, 04 · Stat-Led, 05 · Workbench, 06 · Conversational FAQ, 07 · Manifesto, 08 · Photographic (+13 more)

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

### Community 14 - "admin_handlers.go"
Cohesion: 0.05
Nodes (114): generateSessionToken(), GetCurrentAdminHandler(), getIP(), init(), LoginHandler(), LogoutHandler(), canAdminAccessReport(), ChangeMyPasswordHandler() (+106 more)

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

### Community 26 - "admin.js"
Cohesion: 0.08
Nodes (31): check2FAEnforcementNotice(), checkAdminProfile(), checkHostConnection(), checkMustChangePasswordPolicy(), evaluatePasswordStrength(), fetchForwardingRecipients(), handleTimerTick(), hideOfflineBanner() (+23 more)

### Community 27 - "NewClamAVScanner"
Cohesion: 0.23
Nodes (10): CheckClamAVStartup(), NewClamAVScanner(), io.Reader, ClamAVScanner, startMockClamAVServer(), TestClamAVCleanFile(), TestClamAVDisabledBypassesNetwork(), TestClamAVFailClosedWhenUnreachable() (+2 more)

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

### Community 36 - "initActionSearchBar"
Cohesion: 0.14
Nodes (26): applyAdminUsersFilter(), initActionSearchBar(), closeDropdown(), closeSearch(), computeAndRender(), executeItem(), openDropdown(), renderItems() (+18 more)

### Community 37 - "Genre — atmospheric"
Cohesion: 0.18
Nodes (10): Genre — atmospheric, Nav and footer voice, Reference register, Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows (+2 more)

### Community 38 - "Genre — modern-minimal"
Cohesion: 0.20
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
Cohesion: 0.22
Nodes (7): N10 · Floating-on-scroll morph, Anti-patterns Hallmark refuses, Floating nav on scroll — the cross-fade morph, The four laws — non-negotiable, The property morph (10 properties, one curve), The scroll handler, The structure

### Community 49 - "Round 1 Summary (Phases 0 - 6)"
Cohesion: 0.06
Nodes (33): 1. AAD Bound Encryption (`v3:<kid>:b:`) & Key Ring Rotation (`crypto.go`, `migration_v2.go`), 1. Multipart Memory Spooling & Concurrency Limiting (`public_handlers.go`), 1. Non-Enumerating Login Timing & Status Parity (`admin_auth.go`), 1. Restricted-Session Deadlock & 2FA State Carryover (`AdminAuthMiddleware`, `admin_handlers.go`), 2. Audit Chain Microsecond Truncation & Verification Compatibility (`audit.go`), 2. Blind Indexing for Reporter Tracking Codes (`public_handlers.go`, `db.go`), 2. Direct Stream ClamAV Scanning & Sniffing (`public_handlers.go`, `clamav.go`), 2. Sliding Lockout Decay (`security.go`) (+25 more)

### Community 50 - "db.go"
Cohesion: 0.26
Nodes (11): bootstrapInitialAdmin(), enforceImmutableAuditLogs(), generateInitialPassword(), InitDB(), migrateTrackingCodes(), runMigrations(), VerifyAppRolePrivileges(), VerifyWORMTriggers() (+3 more)

### Community 52 - "viewReport"
Cohesion: 0.13
Nodes (31): closeDetailView(), connectAdminSSE(), createReportStatusBadge(), deleteReportPermanently(), dismissToast(), exportReportsToCSV(), fetchNotifications(), fetchReports() (+23 more)

### Community 53 - "createShamsiDateRangePicker"
Cohesion: 0.27
Nodes (13): createShamsiDateRangePicker(), closePopover(), commitDateRange(), openPopover(), renderCalendar(), renderCalendarHighlights(), resetRange(), getJalaliDayEndMs() (+5 more)

### Community 54 - "initServerSlidingClock"
Cohesion: 0.47
Nodes (6): fetchServerTime(), getServerNow(), initServerSlidingClock(), tick(), updateDate(), updateDigit()

### Community 56 - "remediation_verification_test.go"
Cohesion: 0.47
Nodes (5): parseInteger(), parsePositiveInt(), TestPublicTrackingOmitsAdminUsername(), TestRemediationPaginationClampingBounds(), TestRemediationSessionRevocationLifecycle()

### Community 59 - "populateAnalyticsDOM"
Cohesion: 0.15
Nodes (7): populateAnalyticsDOM(), renderAnalyticsDashboard(), renderCategoryPieChart(), renderClassificationPieChart(), renderClientSideAnalytics(), renderMonthlyChart(), renderPrioritiesSegmentedBar()

### Community 60 - "getTestDB"
Cohesion: 0.18
Nodes (13): NewIPRateLimiter(), NewSessionManager(), TestCaptchaPersistenceTryRedeem(), TestAuditChainConcurrentAndTamperDetection(), TestChangeMyPasswordPreservesRequires2FA(), TestInspectOOXMLFixtures(), TestSessionDeadlockAndPermutations(), TestIPRateLimiterBoundedCapacity() (+5 more)

### Community 62 - "Output contract & scope"
Cohesion: 0.50
Nodes (3): Output contract, Output contract & scope, Scope and limits

### Community 64 - "database/sql.DB"
Cohesion: 0.07
Nodes (39): dumpViaPgDumpToFile(), executeRestoredSQLGz(), getPgDumpPath(), getPsqlPath(), parseDBConn(), PerformBackup(), RestoreBackup(), restoreLegacyV1Backup() (+31 more)

### Community 65 - "Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان"
Cohesion: 0.15
Nodes (12): Design System: سامانه دیده‌بان امین - ستاد خبری و دریافت گزارش‌های مردمی حراست همدان, Do:, Do's and Don'ts, Don't:, Elevation & Depth, Hardware Acceleration & Anti-Lag Performance Rules, Key Characteristics:, Layout & Responsive Breakpoints (+4 more)

### Community 71 - "applyFilters"
Cohesion: 0.13
Nodes (20): applyAuditFiltersAndPagination(), applyFilters(), buildPaginationHTML(), getJalaliDateKeyFromDate(), getMatchSnippet(), getScrollBehavior(), getVisiblePaginationPages(), handleKPICardFilter() (+12 more)

### Community 75 - "adminFetch"
Cohesion: 0.12
Nodes (29): adminFetch(), cancelMFAStep(), deleteUser(), escapeHTML(), fetchActiveSessions(), fetchAdminUsers(), fetchAuditLogs(), fetchSystemStatus() (+21 more)

### Community 76 - "SSEBroker"
Cohesion: 0.16
Nodes (8): NewSSEBroker(), sync.RWMutex, sseBroadcastMsg, SSEBroker, SSEClient, TestRemediationSSEBrokerRoleScoping(), TestSSEBrokerClientLifecycleAndBroadcast(), TestSSESlowClientDoesNotBlockBroker()

### Community 78 - "metadata.go"
Cohesion: 0.10
Nodes (31): decodePDFNameToken(), InspectOOXML(), isPDFDelimiter(), SanitizeAttachment(), SanitizeJPEG(), SanitizeMP3(), SanitizeOOXML(), SanitizePDF() (+23 more)

### Community 79 - "Colors"
Cohesion: 0.40
Nodes (5): Colors, Named Rules, Neutral, Primary, Status & Accents

### Community 84 - "generate_docx.py"
Cohesion: 0.50
Nodes (7): format_persian_text_with_dates(), main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph(), set_english_formatting(), set_persian_formatting()

### Community 87 - "profanity.go"
Cohesion: 0.20
Nodes (17): normalizeForSearch(), cleanRune(), collapseRepeatedRunes(), containsBoundedPhrase(), ContainsVulgarLanguage(), hasVulgarLatinSuffix(), hasVulgarPersianSuffix(), init() (+9 more)

### Community 88 - "SubmitReportHandler"
Cohesion: 0.23
Nodes (12): generateTrackingCode(), inspectZipArchive(), isValidReportCategory(), SubmitReportHandler(), validateMagicBytes(), PublicReportStatus, ReportMessageItem, TestDownloadResponseControllerDeadline() (+4 more)

### Community 89 - "Typography"
Cohesion: 0.50
Nodes (4): Hierarchy, Named Rules, Persian Tatweel Typography Subsystem, Typography

### Community 91 - "GetHMACKey"
Cohesion: 0.21
Nodes (11): GetHMACKey(), ValidatePasswordComplexity(), GenerateMFAPendingToken(), ValidateMFAPendingToken(), TestCompoundAccountLockoutByIP(), TestInitialAdminPasswordGenerator(), TestMFAPendingTokenClientBindingAndReplay(), TestPasswordPolicyUnicodeAndRestrictions() (+3 more)

### Community 92 - "Architectural UI Components & Interaction Standards (Mandatory Rules)"
Cohesion: 0.09
Nodes (22): 10. Header Action Buttons (Icon-Only Pattern), 11. Animated Floating Label Input Architecture, 12. Modern Refined Badge Architecture (`badge-2`, `.badge-2`), 1. Form Validation & Persian Error Mandate, 2. Bespoke Top-Level Alert Architecture (`IconAlert`), 3. Dropdown Anti-Slop Selection Overlay Standard, 4. Universal Password Visibility Toggle, 5. 6-Cell OTP Square Input Architecture (+14 more)

### Community 93 - "initAdminCustomSelect"
Cohesion: 0.47
Nodes (5): initAdminCustomFilters(), initAdminCustomSelect(), closeMenu(), openMenu(), toggleMenu()

### Community 103 - "admin/js/tatweel.js"
Cohesion: 0.38
Nodes (9): applyWrapperToNode(), canConnect(), enhanceElement(), enhanceText(), enhanceWord(), getViewportIntensity(), init(), recheckOverflow() (+1 more)

### Community 105 - "public/js/tatweel.js"
Cohesion: 0.38
Nodes (9): applyWrapperToNode(), canConnect(), enhanceElement(), enhanceText(), enhanceWord(), getViewportIntensity(), init(), recheckOverflow() (+1 more)

### Community 106 - "createOtpController"
Cohesion: 0.28
Nodes (8): createOtpController(), clearAll(), showError(), syncValue(), init2FAModalOtpControllers(), initMfaOtpInput(), showMFAStep(), toAsciiDigits()

### Community 108 - "testing.T"
Cohesion: 0.06
Nodes (40): isKnownPlaceholderKey(), context.Context, testing.T, contextWithAdmin(), normalizeAndValidateUsername(), TestAuditFailClosed(), TestAuditChainHMACComputation(), TestAuditKeyDerivation() (+32 more)

### Community 111 - "Dideban-e Amin (دیده‌بان امین)"
Cohesion: 0.20
Nodes (9): 1. Database Migrations & Startup, 2. Manual Backup, 3. Database Restore, 4. Cryptographic Key Re-encryption, Core Security Architecture, Dideban-e Amin (دیده‌بان امین), Environment Configuration, Operational Commands & CLI (+1 more)

### Community 113 - "Macrostructures"
Cohesion: 0.29
Nodes (7): Diversification rule (mandatory), Hero polish patterns, How to pick, Macrostructures, Nav and footer voice, SaaS page sequence, The 21 macrostructures — index

### Community 114 - "initAppTooltips"
Cohesion: 0.67
Nodes (6): initAppTooltips(), getOrCreateTooltip(), hide(), isMobileOrTouch(), positionTooltip(), show()

### Community 116 - "main"
Cohesion: 0.07
Nodes (43): AdminAuthMiddleware(), SuperuserAuthMiddleware(), CSRFMiddleware(), isOriginAllowed(), compress/gzip.Writer, net/http.Handler, net/http.HandlerFunc, net.IP (+35 more)

### Community 118 - "admin/js/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 120 - "public/js/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 121 - "initLoginSlider"
Cohesion: 0.70
Nodes (5): initLoginSlider(), goToSlide(), nextSlide(), startTimer(), stopTimer()

### Community 123 - "VerifyCaptcha"
Cohesion: 0.13
Nodes (20): cryptoRandInt(), GenerateCaptchaToken(), GenerateMathCaptchaSVG(), generateNoiseElements(), GetCaptchaHandler(), renderObfuscatedChars(), VerifyCaptcha(), DecryptCaptchaPayload() (+12 more)

## Knowledge Gaps
- **745 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `RecipientUserItem`, `ForwardReportRequest`, `ForwardHistoryItem` (+740 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **23 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Anti-patterns — the named tells` connect `Minor (small taste issues)` to `Major (looks AI-generated)`, `SKILL.md`, `Microinteraction tells`, `Critical (ships as slop)`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `Hallmark` connect `Design flow (default)` to `SKILL.md`?**
  _High betweenness centrality (0.016) - this node is a cross-community bridge._
- **Why does `Microinteractions` connect `Recipes` to `SKILL.md`?**
  _High betweenness centrality (0.012) - this node is a cross-community bridge._
- **Are the 36 inferred relationships involving `adminFromContext()` (e.g. with `GetCurrentAdminHandler()` and `ChangeMyPasswordHandler()`) actually correct?**
  _`adminFromContext()` has 36 INFERRED edges - model-reasoned connections that need verification._
- **Are the 34 inferred relationships involving `getIP()` (e.g. with `GetTrustedProxies()` and `ipInCIDRs()`) actually correct?**
  _`getIP()` has 34 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `RecipientUserItem` to the rest of the system?**
  _745 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Components` be split into smaller, more focused modules?**
  _Cohesion score 0.05714285714285714 - nodes in this community are weakly interconnected._