# Graph Report - Setad_Khabari_HMD  (2026-09-22)

## Corpus Check
- 190 files · ~266,791 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1475 nodes · 2299 edges · 114 communities (79 shown, 35 thin omitted)
- Extraction: 89% EXTRACTED · 11% INFERRED · 0% AMBIGUOUS · INFERRED: 256 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `14949986`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Components
- component-cookbook.md
- Running Tests
- Assets — sourcing canon for icons, logos, illustrations, photography, video
- admin_handlers.go
- macrostructures.md
- Theme — Lumen
- Theme — Hum
- Custom craft — how to hand-build hero artwork
- Custom theme — protocol
- Design flow (default)
- SKILL.md
- Recipes
- Study — extracting design DNA from a screenshot or URL
- admin.js
- Hero enrichment — when, what, and how much
- Theme — Carnival
- Interaction and states
- Project Guidelines
- Major (looks AI-generated)
- app.js
- Slop test — 58 gates + pre-emit self-critique
- Copy
- Critical (ships as slop)
- Typography
- seed_50_reports.go
- SyslogShipper
- contextWithAdmin
- Microinteraction tells
- Archetype index — load ONLY the picks you need
- Theme — Cobalt
- Theme - Grid
- Responsive
- The six axes
- Minor (small taste issues)
- Imagery kit — curated abstract assets, hosted, ready
- adminFetch
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
- Macrostructures
- Floating nav on scroll — the cross-fade morph
- SSEBroker
- Output contract & scope
- Hallmark Design Directives
- mfa.go
- n10-floating-on-scroll-morph.md
- showToast
- no_comments.md
- populateAnalyticsDOM
- NewClamAVScanner
- renderTable
- escapeHTML
- SanitizeAttachment
- fetchAdminUsers
- openForwardModal
- TestCookieSecureCachedResolution
- c2-inline-form-as-cta.md
- f1-bento-grid.md
- ft7-newsletter-first.md
- crypto.go
- h4-stat-led.md
- s5-bottom-anchored.md
- herasat.local/backend
- captcha.go
- crypto_test.go
- workflows/hallmark.md
- c1-outlined-chip.md
- h8-mockup-split-browser-framed.md
- n2-floating-chip.md
- n4-hidden-behind-k.md
- f4-step-sequence.md
- ft5-statement.md
- rules/graphify.md
- workflows/graphify.md
- ft6-letter-close.md
- n6-newspaper-masthead.md
- generate_docx.py
- rules/caveman.md
- workflows/caveman.md
- ContainsVulgarLanguage
- n8-terminal-command.md
- be-ui-file-upload.tsx
- sign-in-card-2.tsx
- Project Guidelines (HMD IT Herasat Portal)
- Impeccable Design Quality & Craft Floor
- Design Taste Frontend Rules
- workflows/impeccable.md
- workflows/taste.md
- tooltip.tsx
- button-1.tsx
- accordion-2.tsx
- h9-custom-illustration-centerpiece.md
- button.tsx
- initAppTooltips
- morph-loading.tsx
- testing.T
- README.md
- t2-logo-wall-hairline.md
- security.go
- main
- admin/confirm.js
- frontend/confirm.js
- public/confirm.js

## God Nodes (most connected - your core abstractions)
1. `getIP()` - 33 edges
2. `adminFetch()` - 25 edges
3. `logAudit()` - 22 edges
4. `viewReport()` - 22 edges
5. `Major (looks AI-generated)` - 20 edges
6. `Slop test — 58 gates + pre-emit self-critique` - 19 edges
7. `SubmitReportHandler()` - 18 edges
8. `SessionManager` - 18 edges
9. `Theme — Hum` - 18 edges
10. `DecryptAES()` - 17 edges

## Surprising Connections (you probably didn't know these)
- `LoginHandler()` --calls--> `VerifyCaptcha()`  [INFERRED]
  admin_auth.go → captcha.go
- `LoginHandler()` --calls--> `VerifyPasswordWithRehash()`  [INFERRED]
  admin_auth.go → crypto.go
- `LoginHandler()` --calls--> `GenerateMFAPendingToken()`  [INFERRED]
  admin_auth.go → mfa.go
- `RateLimitMiddleware()` --calls--> `getIP()`  [INFERRED]
  security.go → admin_auth.go
- `TestGetIPHardening()` --calls--> `getIP()`  [INFERRED]
  test/get_ip_test.go → admin_auth.go

## Import Cycles
- None detected.

## Communities (114 total, 35 thin omitted)

### Community 0 - "Components"
Cohesion: 0.07
Nodes (28): 1. Navigation & Pill Tab Bar, 2. Multi-Step Form Tracker, 3. Glass Category Dropdown (`dropdown-menu`), 4. Centered Smart Evidence Upload (`be-ui-file-upload`), 5. FAQ Themed Accordion (`accordion-2`), 6. Admin 3D Sign-in Card (`sign-in-card-2`), 7. Action Buttons & Directional Arrows, 8. Secure Tracking & Follow-up Thread (+20 more)

### Community 1 - "component-cookbook.md"
Cohesion: 0.03
Nodes (32): C3 · Typographic link, C4 · Sticky bottom bar, F2 · Sticky-scroll stack, F3 · Tabular spec sheet, F5 · Annotated screenshot, F6 · Product card grid, Ft1 · Mast-headed, Ft2 · Inline-rule single line (+24 more)

### Community 2 - "Running Tests"
Cohesion: 0.40
Nodes (4): Option 1: Temporary restore to root, Option 2: Run specific test with root source files, Running Tests, Test Suite

### Community 3 - "Assets — sourcing canon for icons, logos, illustrations, photography, video"
Cohesion: 0.05
Nodes (44): Abstract backgrounds, App mockups / device frames, Assets — sourcing canon for icons, logos, illustrations, photography, video, Avoid, Avoid, Avoid, Avoid, Avoid (+36 more)

### Community 4 - "admin_handlers.go"
Cohesion: 0.06
Nodes (95): generateSessionToken(), GetCurrentAdminHandler(), getIP(), ipInCIDRs(), logAudit(), logAuditTx(), LoginHandler(), LogoutHandler() (+87 more)

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
Cohesion: 0.23
Nodes (3): H7 · Demo Video — Clipped-by-viewport-edge, Preview-block worked examples, `hallmark audit`

### Community 12 - "Recipes"
Cohesion: 0.07
Nodes (27): Accessibility ground truth, Button press, Command palette navigation, Copy-to-clipboard, Drag handle, Dropdown / menu, Form validation, Hard rules for default-on motion (+19 more)

### Community 13 - "Study — extracting design DNA from a screenshot or URL"
Cohesion: 0.08
Nodes (26): After the file is written, Emitting a `design.md` from `study`, Image-mode template, Junk-or-blocked detection, Limits and disclaimers, Refusal — when not to study, Remote URL safety, Source mode — image or URL (+18 more)

### Community 14 - "admin.js"
Cohesion: 0.11
Nodes (14): cancelMFAStep(), checkAdminProfile(), checkMustChangePasswordPolicy(), connectAdminSSE(), handleTimerTick(), loadAdminCaptcha(), markReportHasNewMessage(), resumeSessionTimer() (+6 more)

### Community 15 - "Hero enrichment — when, what, and how much"
Cohesion: 0.09
Nodes (23): Animation discipline (hero specifically), Common mistakes — and the fixes, E1 · Demo Video — Clipped-by-viewport-edge, E2 · Demo Video — Full-bleed muted loop with ghost overlay, E3 · Mock App Screenshot — Browser-framed split, E4 · Mock App Screenshot — Floating no-frame, E5 · Custom Illustration Centerpiece, E6 · Animated Loop — pure CSS / SVG / Motion (+15 more)

### Community 16 - "Theme — Carnival"
Cohesion: 0.09
Nodes (22): Anti-patterns (theme-specific), Axes (diversification), Build hint, Drop 01 · Cold Snap *(default)*, Drop 02 · Citrus Riot, Drop 03 · Diner Sign, Drop 04 · Studio Night, Drop 05 · Aqua Park (+14 more)

### Community 17 - "Interaction and states"
Cohesion: 0.10
Nodes (21): Bans, Computation, Contrast discipline, Don't, list, Dropdowns, tooltips, popovers, Focus rings, Forms, Heights and rhythm (+13 more)

### Community 18 - "Project Guidelines"
Cohesion: 0.33
Nodes (5): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Mandatory Code Rule: NO COMMENTS, Project Guidelines

### Community 19 - "Major (looks AI-generated)"
Cohesion: 0.10
Nodes (20): AI-illustration look, Animate-on-scroll on everything, Bounce and elastic easing, Centred everything, Eyebrow on every section, Generic emoji as feature icon, Glassmorphism without purpose, Hover-only affordances (+12 more)

### Community 20 - "app.js"
Cohesion: 0.08
Nodes (36): addFiles(), attemptDismiss(), checkClientVulgarLanguage(), clearDraft(), clearFileInput(), closeGuidelinesDrawer(), dismissToast(), escapeHtml() (+28 more)

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

### Community 26 - "SyslogShipper"
Cohesion: 0.24
Nodes (9): net.Conn, sync.WaitGroup, SyslogMessage, SyslogSeverity, SyslogShipper, InitSyslogShipper(), MapActionToSyslogSeverity(), sanitizeSyslogParam() (+1 more)

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

### Community 36 - "adminFetch"
Cohesion: 0.22
Nodes (18): adminFetch(), deleteReportPermanently(), fetchNotifications(), fetchReports(), fetchWithTimeout(), handleAdminLogout(), makeTextareaAutoExpand(), refresh2FAStatus() (+10 more)

### Community 37 - "Genre — atmospheric"
Cohesion: 0.20
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
Cohesion: 0.22
Nodes (9): Genre — editorial (default), Nav and footer voice, Stamp signature, Themes that belong, Voice, Voice fixtures, What this genre allows, What this genre disallows (+1 more)

### Community 42 - "Layout and space"
Cohesion: 0.22
Nodes (9): Asymmetry techniques, Bans, Depth, Grids, Layout and space, Page-edge clipping, Principles, The spacing scale (+1 more)

### Community 43 - "Motion"
Cohesion: 0.22
Nodes (9): Bans, Durations, Easings, Motion, Page-load orchestration, Principles, Reduced motion, Scroll-linked motion (+1 more)

### Community 44 - "design.md — opt-in portable design system"
Cohesion: 0.20
Nodes (8): After the file is written, CTA — surface the offer in the Step 5 preview, design.md — opt-in portable design system, Format (the tight version), Scope, Triggers (phrase-only — no new verb), Two emission paths (default vs study), Why opt-in (not auto-emit)

### Community 45 - "Export formats"
Cohesion: 0.25
Nodes (7): Export formats, Format 1 — `tokens.css`, Format 2 — Tailwind v4 `@theme`, Format 3 — DTCG `tokens.json`, Format 4 — shadcn/ui CSS variables, Output rule, Token taxonomy — Hallmark's source of truth

### Community 46 - "Colour"
Cohesion: 0.29
Nodes (7): Bans, Colour, Contrast, Dark mode recipe, Palette construction, Principles, Use of the accent

### Community 47 - "Macrostructures"
Cohesion: 0.29
Nodes (7): Diversification rule (mandatory), Hero polish patterns, How to pick, Macrostructures, Nav and footer voice, SaaS page sequence, The 21 macrostructures — index

### Community 48 - "Floating nav on scroll — the cross-fade morph"
Cohesion: 0.33
Nodes (6): Anti-patterns Hallmark refuses, Floating nav on scroll — the cross-fade morph, The four laws — non-negotiable, The property morph (10 properties, one curve), The scroll handler, The structure

### Community 49 - "SSEBroker"
Cohesion: 0.27
Nodes (4): NewSSEBroker(), sync.RWMutex, SSEBroker, SSEClient

### Community 50 - "Output contract & scope"
Cohesion: 0.50
Nodes (3): Output contract, Output contract & scope, Scope and limits

### Community 52 - "mfa.go"
Cohesion: 0.28
Nodes (12): GetHMACKey(), GenerateBase32Secret(), GenerateMFAPendingToken(), GenerateTOTPCode(), ValidateMFAPendingToken(), VerifyTOTPCode(), MFAPendingPayload, TestConcurrentTOTPVerificationRace() (+4 more)

### Community 54 - "showToast"
Cohesion: 0.22
Nodes (14): exportReportsToCSV(), fetchActiveSessions(), fetchSystemStatus(), formatJalaliDateTime(), loadAdvSubView(), parseUserAgent(), renderSessionsTable(), renderSessionsTableSkeleton() (+6 more)

### Community 56 - "populateAnalyticsDOM"
Cohesion: 0.15
Nodes (7): populateAnalyticsDOM(), renderAnalyticsDashboard(), renderCategoryPieChart(), renderClassificationPieChart(), renderClientSideAnalytics(), renderMonthlyChart(), renderPrioritiesSegmentedBar()

### Community 57 - "NewClamAVScanner"
Cohesion: 0.24
Nodes (8): NewClamAVScanner(), io.Reader, ClamAVScanner, startMockClamAVServer(), TestClamAVCleanFile(), TestClamAVDisabledBypassesNetwork(), TestClamAVFailClosedWhenUnreachable(), TestClamAVMalwareFound()

### Community 58 - "renderTable"
Cohesion: 0.23
Nodes (12): applyFilters(), calculatePagination(), closeDetailView(), getMatchSnippet(), getScrollBehavior(), handleKPICardFilter(), highlightSearchTerm(), makePersianRegex() (+4 more)

### Community 59 - "escapeHTML"
Cohesion: 0.27
Nodes (10): applyAuditFiltersAndPagination(), escapeHTML(), fetchAuditLogs(), getJalaliMonth(), openForwardHistoryModal(), populateAuditUsersDropdown(), renderAuditTable(), renderAuditTableSkeleton() (+2 more)

### Community 60 - "SanitizeAttachment"
Cohesion: 0.33
Nodes (8): SanitizeAttachment(), SanitizeJPEG(), SanitizePNG(), createTestJPEGWithEXIF(), createTestPNG(), TestMetadataSanitizationJPEG(), TestMetadataSanitizationPNG(), TestSanitizeAttachmentPassthrough()

### Community 61 - "fetchAdminUsers"
Cohesion: 0.33
Nodes (9): deleteUser(), fetchAdminUsers(), liftUserLockout(), openEditUserModal(), openResetPwdModal(), openShortcutsModal(), promoteTemporarily(), renderUsersTable() (+1 more)

### Community 62 - "openForwardModal"
Cohesion: 0.29
Nodes (8): fetchForwardingRecipients(), indexReportForSearch(), normalizePersianForSearch(), openForwardModal(), parseTimestamp(), renderRecipientList(), renderRecipientListSkeleton(), updateSelectedRecipientsCounter()

### Community 63 - "TestCookieSecureCachedResolution"
Cohesion: 0.50
Nodes (3): InitCookieSecurity(), TestCookieSecureCachedResolution(), TestCookieSecureStartupFatalWithoutTLSOrFlag()

### Community 67 - "crypto.go"
Cohesion: 0.17
Nodes (17): DecryptFileBytes(), DecryptFilePath(), decryptFilePayload(), decryptWithKey(), deriveKey(), EncryptAES(), EncryptFileBytes(), EncryptFilePath() (+9 more)

### Community 71 - "captcha.go"
Cohesion: 0.21
Nodes (14): cryptoRandInt(), GenerateCaptchaToken(), GenerateMathCaptchaSVG(), generateNoiseElements(), GetCaptchaHandler(), renderObfuscatedChars(), VerifyCaptcha(), GetLegacyHMACKey() (+6 more)

### Community 72 - "crypto_test.go"
Cohesion: 0.40
Nodes (5): VerifyPassword(), VerifyPasswordWithRehash(), TestArgon2idDualFormatAndRehash(), TestDeriveKeyDeterministic(), TestValidatePasswordComplexity()

### Community 84 - "generate_docx.py"
Cohesion: 0.50
Nodes (7): format_persian_text_with_dates(), main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph(), set_english_formatting(), set_persian_formatting()

### Community 87 - "ContainsVulgarLanguage"
Cohesion: 0.25
Nodes (12): normalizeForSearch(), cleanRune(), collapseRepeatedRunes(), containsBoundedPhrase(), ContainsVulgarLanguage(), hasVulgarPrefix(), isDiacriticOrInvisible(), isSymbolOrPunct() (+4 more)

### Community 90 - "be-ui-file-upload.tsx"
Cohesion: 0.09
Nodes (28): clampProgress(), cn(), createFileUploadItem(), EASE_DRAWER, EASE_IN_OUT, EASE_OUT, EASE_OUT_CSS, FAST_TRANSITION (+20 more)

### Community 92 - "Project Guidelines (HMD IT Herasat Portal)"
Cohesion: 0.33
Nodes (5): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Mandatory Code Rule: NO COMMENTS, Project Guidelines (HMD IT Herasat Portal)

### Community 100 - "button-1.tsx"
Cohesion: 0.50
Nodes (3): Button(), ButtonArrowProps, buttonVariants

### Community 104 - "button.tsx"
Cohesion: 0.23
Nodes (7): Button, ButtonProps, buttonVariants, cn(), ButtonWithIconProps, DropdownMenuProps, DropdownOption

### Community 106 - "initAppTooltips"
Cohesion: 0.52
Nodes (7): initAppTooltips(), convertTitles(), getOrCreateTooltip(), hide(), isMobileOrTouch(), positionTooltip(), show()

### Community 108 - "testing.T"
Cohesion: 0.11
Nodes (22): testing.T, TestAuditFailClosed(), TestClassificationValidationAndMath(), TestDBCredentialsRequiredAtStartup(), TestForwardHistoryItemWithRegionalResponse(), TestForwardHistorySerialization(), TestForwardingRequestValidation(), TestNavahiManateghClassificationForbidden() (+14 more)

### Community 115 - "security.go"
Cohesion: 0.08
Nodes (30): NewUsedCaptchaStore(), database/sql.DB, sync.Mutex, sync.Once, time.Duration, time.Ticker, time.Time, hashToken() (+22 more)

### Community 116 - "main"
Cohesion: 0.11
Nodes (18): AdminAuthMiddleware(), SuperuserAuthMiddleware(), CSRFMiddleware(), isOriginAllowed(), enforceImmutableAuditLogs(), InitDB(), runMigrations(), net/http.Handler (+10 more)

### Community 118 - "admin/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 119 - "frontend/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

### Community 120 - "public/confirm.js"
Cohesion: 0.60
Nodes (4): getOrCreateDialog(), promoteTemporarily(), showConfirmDialog(), trapFocus()

## Knowledge Gaps
- **676 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `RecipientUserItem`, `ForwardReportRequest`, `ForwardHistoryItem` (+671 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **35 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Anti-patterns — the named tells` connect `Minor (small taste issues)` to `Major (looks AI-generated)`, `SKILL.md`, `Microinteraction tells`, `Critical (ships as slop)`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Why does `Assets — sourcing canon for icons, logos, illustrations, photography, video` connect `Assets — sourcing canon for icons, logos, illustrations, photography, video` to `SKILL.md`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `Custom theme — protocol` connect `Custom theme — protocol` to `SKILL.md`?**
  _High betweenness centrality (0.020) - this node is a cross-community bridge._
- **Are the 27 inferred relationships involving `getIP()` (e.g. with `ChangeMyPasswordHandler()` and `ClassifyReportHandler()`) actually correct?**
  _`getIP()` has 27 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `RecipientUserItem` to the rest of the system?**
  _676 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Components` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._
- **Should `component-cookbook.md` be split into smaller, more focused modules?**
  _Cohesion score 0.028985507246376812 - nodes in this community are weakly interconnected._