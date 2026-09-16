# Graph Report - herasat_backend  (2026-09-16)

## Corpus Check
- 71 files · ~110,289 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 581 nodes · 1197 edges · 40 communities (26 shown, 14 thin omitted)
- Extraction: 83% EXTRACTED · 17% INFERRED · 0% AMBIGUOUS · INFERRED: 206 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `08379fe4`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- Components
- admin_handlers.go
- admin.js
- Project Guidelines
- app.js
- herasat.local/backend
- rules/graphify.md
- workflows/graphify.md
- generate_docx.py
- rules/caveman.md
- workflows/caveman.md
- ContainsVulgarLanguage
- crypto.go
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
- button.tsx
- initAppTooltips
- morph-loading.tsx
- testing.T
- 🚀 راه‌اندازی سریع (Quickstart)
- security.go
- main
- [2.0.0] - 2026-09-16
- showConfirmDialog
- showConfirmDialog
- showConfirmDialog
- Reporting Security Issues
- SECURITY_CHANGELOG.md

## God Nodes (most connected - your core abstractions)
1. `getIP()` - 29 edges
2. `adminFetch()` - 24 edges
3. `viewReport()` - 21 edges
4. `SessionManager` - 18 edges
5. `logAudit()` - 17 edges
6. `SubmitReportHandler()` - 16 edges
7. `getTestDB()` - 16 edges
8. `DecryptAES()` - 14 edges
9. `showToast()` - 14 edges
10. `escapeHTML()` - 14 edges

## Surprising Connections (you probably didn't know these)
- `LoginHandler()` --calls--> `VerifyCaptcha()`  [INFERRED]
  admin_auth.go → captcha.go
- `LoginHandler()` --calls--> `VerifyPasswordWithRehash()`  [INFERRED]
  admin_auth.go → crypto.go
- `LoginHandler()` --calls--> `isCookieSecure()`  [INFERRED]
  admin_auth.go → security.go
- `RateLimitMiddleware()` --calls--> `getIP()`  [INFERRED]
  security.go → admin_auth.go
- `TestGetIPHardening()` --calls--> `getIP()`  [INFERRED]
  test/get_ip_test.go → admin_auth.go

## Import Cycles
- None detected.

## Communities (40 total, 14 thin omitted)

### Community 0 - "Components"
Cohesion: 0.07
Nodes (28): 1. Navigation & Pill Tab Bar, 2. Multi-Step Form Tracker, 3. Glass Category Dropdown (`dropdown-menu`), 4. Centered Smart Evidence Upload (`be-ui-file-upload`), 5. FAQ Themed Accordion (`accordion-2`), 6. Admin 3D Sign-in Card (`sign-in-card-2`), 7. Action Buttons & Directional Arrows, 8. Secure Tracking & Follow-up Thread (+20 more)

### Community 4 - "admin_handlers.go"
Cohesion: 0.06
Nodes (88): generateSessionToken(), GetCurrentAdminHandler(), getIP(), ipInCIDRs(), logAudit(), logAuditTx(), LoginHandler(), LogoutHandler() (+80 more)

### Community 14 - "admin.js"
Cohesion: 0.06
Nodes (79): adminFetch(), applyAuditFiltersAndPagination(), applyFilters(), checkAdminProfile(), checkMustChangePasswordPolicy(), closeDetailView(), deleteReportPermanently(), deleteUser() (+71 more)

### Community 18 - "Project Guidelines"
Cohesion: 0.40
Nodes (4): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Project Guidelines

### Community 20 - "app.js"
Cohesion: 0.08
Nodes (27): addFiles(), clearDraft(), clearFileInput(), dismissToast(), escapeHtml(), getBeFileIconSvg(), initAppTooltips(), getOrCreateTooltip() (+19 more)

### Community 84 - "generate_docx.py"
Cohesion: 0.50
Nodes (7): format_persian_text_with_dates(), main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph(), set_english_formatting(), set_persian_formatting()

### Community 87 - "ContainsVulgarLanguage"
Cohesion: 0.25
Nodes (12): normalizeForSearch(), cleanRune(), collapseRepeatedRunes(), containsBoundedPhrase(), ContainsVulgarLanguage(), hasVulgarPrefix(), isDiacriticOrInvisible(), isSymbolOrPunct() (+4 more)

### Community 88 - "crypto.go"
Cohesion: 0.11
Nodes (33): cryptoRandInt(), GenerateCaptchaToken(), GenerateMathCaptchaSVG(), generateNoiseElements(), GetCaptchaHandler(), renderObfuscatedChars(), VerifyCaptcha(), DecryptAES() (+25 more)

### Community 90 - "be-ui-file-upload.tsx"
Cohesion: 0.09
Nodes (28): clampProgress(), cn(), createFileUploadItem(), EASE_DRAWER, EASE_IN_OUT, EASE_OUT, EASE_OUT_CSS, FAST_TRANSITION (+20 more)

### Community 92 - "Project Guidelines (HMD IT Herasat Portal)"
Cohesion: 0.40
Nodes (4): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Project Guidelines (HMD IT Herasat Portal)

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
Cohesion: 0.07
Nodes (32): testing.T, NewAccountLockoutManager(), NewIPRateLimiter(), NewSessionManager(), TestAuditFailClosed(), TestCaptchaPersistenceTryRedeem(), TestClassificationValidationAndMath(), contextWithAdmin (+24 more)

### Community 111 - "🚀 راه‌اندازی سریع (Quickstart)"
Cohesion: 0.22
Nodes (8): 🚀 راه‌اندازی سریع (Quickstart), سامانه گزارش‌های مردمی و نظارت حراست (Herasat Whistleblowing Platform), ⚙️ متغیرهای محیطی (Environment Variables), 🛡️ معماری امنیتی (Security Architecture), ۱. کپی متغیرهای محیطی, ۲. اجرای پایگاه داده و مهاجرت‌ها, ۳. کامپایل و اجرای برنامه, ۴. اجرای آزمون‌ها (Tests)

### Community 115 - "security.go"
Cohesion: 0.11
Nodes (18): NewUsedCaptchaStore(), database/sql.DB, sync.Mutex, sync.Once, sync.RWMutex, time.Duration, time.Time, AccountLockoutManager (+10 more)

### Community 116 - "main"
Cohesion: 0.10
Nodes (21): AdminAuthMiddleware(), SuperuserAuthMiddleware(), CSRFMiddleware(), isOriginAllowed(), enforceImmutableAuditLogs(), InitDB(), runMigrations(), net/http.Handler (+13 more)

### Community 117 - "[2.0.0] - 2026-09-16"
Cohesion: 0.29
Nodes (6): [2.0.0] - 2026-09-16, Added, Changed, Changelog, Removed, Security

### Community 118 - "showConfirmDialog"
Cohesion: 0.60
Nodes (4): safeEscape(), showConfirmDialog(), cleanup(), onKey()

### Community 119 - "showConfirmDialog"
Cohesion: 0.60
Nodes (4): safeEscape(), showConfirmDialog(), cleanup(), onKey()

### Community 120 - "showConfirmDialog"
Cohesion: 0.60
Nodes (4): safeEscape(), showConfirmDialog(), cleanup(), onKey()

### Community 121 - "Reporting Security Issues"
Cohesion: 0.40
Nodes (4): Disclosure Guidelines, Reporting Security Issues, Scope, Security Policy

## Knowledge Gaps
- **83 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `RecipientUserItem`, `ForwardReportRequest`, `ForwardHistoryItem` (+78 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **14 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `getIP()` connect `admin_handlers.go` to `security.go`, `testing.T`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Why does `main()` connect `main` to `crypto.go`, `security.go`, `testing.T`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **Why does `SubmitReportHandler()` connect `admin_handlers.go` to `crypto.go`, `ContainsVulgarLanguage`?**
  _High betweenness centrality (0.013) - this node is a cross-community bridge._
- **Are the 23 inferred relationships involving `getIP()` (e.g. with `ChangeMyPasswordHandler()` and `ClassifyReportHandler()`) actually correct?**
  _`getIP()` has 23 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `RecipientUserItem` to the rest of the system?**
  _83 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Components` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._
- **Should `admin_handlers.go` be split into smaller, more focused modules?**
  _Cohesion score 0.06030855539971949 - nodes in this community are weakly interconnected._