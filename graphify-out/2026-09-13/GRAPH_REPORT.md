# Graph Report - herasat_backend  (2026-09-13)

## Corpus Check
- 151 files · ~318,602 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1174 nodes · 2064 edges · 108 communities (81 shown, 27 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 152 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Components
- Listener
- database/sql/driver.Value
- writeBuf
- admin_handlers.go
- Config
- Client
- cpu_arm64.go
- rows
- block
- queryResult
- doinit
- README.md
- parseError
- admin.js
- XOF
- syscall_darwin_x86_gc.go
- Error
- Project Guidelines
- blake2bAVX2_amd64.go
- app.js
- proto.go
- syscall_aix_ppc64_gc.go
- Code
- cpu.go
- cpu_linux_riscv64.go
- cpu_gccgo_x86.go
- syscall_darwin_arm64_gc.go
- CheckPermissions
- CHANGELOG.md
- pqutil.go
- path.go
- cpu_loong64.go
- cpu_x86.go
- As
- As
- pgpass.go
- cpu_aix.go
- cpu_linux_arm.go
- cpu_linux_loong64.go
- cpu_linux_mips64x.go
- cpu_linux_ppc64x.go
- .Get
- oid/doc.go
- herasat.local/backend
- rules/graphify.md
- workflows/graphify.md
- encode.go
- conn
- generate_docx.py
- rules/caveman.md
- workflows/caveman.md
- readBuf
- context.Context
- package.json
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
- krb.go
- button.tsx
- initAppTooltips
- morph-loading.tsx
- ContainsVulgarLanguage

## God Nodes (most connected - your core abstractions)
1. `conn` - 60 edges
2. `Config` - 33 edges
3. `getIP()` - 24 edges
4. `Listener` - 23 edges
5. `logAudit()` - 22 edges
6. `ListenerConn` - 22 edges
7. `readBuf` - 20 edges
8. `writeBuf` - 19 edges
9. `rows` - 19 edges
10. `parseError()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `RateLimitMiddleware()` --calls--> `getIP()`  [INFERRED]
  security.go → admin_auth.go
- `DownloadAttachmentHandler()` --calls--> `DecryptFileBytes()`  [INFERRED]
  admin_handlers.go → crypto.go
- `SubmitReportHandler()` --calls--> `EncryptFileBytes()`  [INFERRED]
  public_handlers.go → crypto.go
- `TestValidatePasswordComplexity()` --calls--> `ValidatePasswordComplexity()`  [INFERRED]
  crypto_test.go → crypto.go
- `main()` --calls--> `NewIPRateLimiter()`  [INFERRED]
  main.go → security.go

## Import Cycles
- None detected.

## Communities (108 total, 27 thin omitted)

### Community 0 - "Components"
Cohesion: 0.07
Nodes (28): 1. Navigation & Pill Tab Bar, 2. Multi-Step Form Tracker, 3. Glass Category Dropdown (`dropdown-menu`), 4. Centered Smart Evidence Upload (`be-ui-file-upload`), 5. FAQ Themed Accordion (`accordion-2`), 6. Admin 3D Sign-in Card (`sign-in-card-2`), 7. Action Buttons & Directional Arrows, 8. Secure Tracking & Follow-up Thread (+20 more)

### Community 1 - "Listener"
Cohesion: 0.05
Nodes (36): database/sql/driver.Conn, database/sql/driver.Connector, sync.Cond, sync.Mutex, time.Duration, UsedCaptchaStore, Dialer, DialerContext (+28 more)

### Community 2 - "database/sql/driver.Value"
Cohesion: 0.06
Nodes (24): database/sql/driver.Rows, database/sql/driver.Value, database/sql/driver.Valuer, database/sql.Scanner, reflect.Value, ArrayDelimiter, BoolArray, ByteaArray (+16 more)

### Community 3 - "writeBuf"
Cohesion: 0.15
Nodes (3): github.com/lib/pq/internal/proto.RequestCode, writeBuf, Connector

### Community 4 - "admin_handlers.go"
Cohesion: 0.06
Nodes (90): generateSessionToken(), GetCurrentAdminHandler(), getIP(), logAudit(), LoginHandler(), LogoutHandler(), AdminAuthMiddleware(), AdminSendMessageHandler() (+82 more)

### Community 5 - "Config"
Cohesion: 0.11
Nodes (25): crypto/tls.Config, database/sql/driver.Driver, net/netip.Addr, Config, ConfigMultihost, LoadBalanceHosts, ProtocolVersion, SSLMode (+17 more)

### Community 6 - "Client"
Cohesion: 0.06
Nodes (24): bytes.Buffer, hash.Hash, PGError, Client, CopyIn(), CopyInSchema(), makeStmt(), ParseURL() (+16 more)

### Community 7 - "cpu_arm64.go"
Cohesion: 0.08
Nodes (26): aarch64SysctlCPUID, sysctlNode, syscall.Errno, archInit(), extractBits(), parseARM64SVERegister(), parseARM64SystemRegisters(), readARM64Registers() (+18 more)

### Community 8 - "rows"
Cohesion: 0.11
Nodes (9): reflect.Type, fieldDesc, format, noRows, rows, rowsHeader, decideColumnFormats(), parsePortalRowDescribe() (+1 more)

### Community 9 - "block"
Cohesion: 0.11
Nodes (28): block, t00, t01, t02, t03, t04, t05, t06 (+20 more)

### Community 10 - "queryResult"
Cohesion: 0.11
Nodes (20): facility, facilityList, function, queryResult, kimdQuery(), klmdQuery(), kmaQuery(), kmcQuery() (+12 more)

### Community 11 - "doinit"
Cohesion: 0.10
Nodes (12): bigEndian, byteOrder, littleEndian, hostByteOrder(), archInit(), doinit(), isSet(), linuxKernelCanEmulateCPUID() (+4 more)

### Community 12 - "README.md"
Cohesion: 0.20
Nodes (9): Authentication, Bulk imports with `COPY [..] FROM STDIN`, bytea with copy, LastInsertId, NOTICE errors, Protocol debug output, Running tests, timestamps (+1 more)

### Community 13 - "parseError"
Cohesion: 0.27
Nodes (3): transactionStatus, stmt, parseError()

### Community 14 - "admin.js"
Cohesion: 0.07
Nodes (54): applyAuditFiltersAndPagination(), applyFilters(), checkAdminProfile(), checkMustChangePasswordPolicy(), deleteReportPermanently(), deleteUser(), escapeHTML(), exportReportsToCSV() (+46 more)

### Community 15 - "XOF"
Cohesion: 0.19
Nodes (6): XOF, io.Reader, io.Writer, digest, digest, NewXOF()

### Community 16 - "syscall_darwin_x86_gc.go"
Cohesion: 0.29
Nodes (9): _C_int, darwinKernelVersionCheck(), darwinSupportsAVX512(), darwinOSRelease(), darwinSysctlEnabled(), Errno, syscall_syscall6(), sysctl() (+1 more)

### Community 17 - "Error"
Cohesion: 0.24
Nodes (5): github.com/lib/pq/pqerror.Code, expandTab(), conn, Error, posToLine()

### Community 18 - "Project Guidelines"
Cohesion: 0.40
Nodes (4): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Project Guidelines

### Community 19 - "blake2bAVX2_amd64.go"
Cohesion: 0.27
Nodes (6): hashBlocksGeneric(), hashBlocks(), hashBlocks(), hashBlocksAVX(), hashBlocksAVX2(), hashBlocksSSE4()

### Community 20 - "app.js"
Cohesion: 0.09
Nodes (23): addFiles(), clearDraft(), clearFileInput(), dismissToast(), escapeHtml(), getBeFileIconSvg(), initAppTooltips(), getOrCreateTooltip() (+15 more)

### Community 21 - "proto.go"
Cohesion: 0.29
Nodes (3): AuthCode, RequestCode, ResponseCode

### Community 22 - "syscall_aix_ppc64_gc.go"
Cohesion: 0.53
Nodes (5): syscallFunc, callgetsystemcfg(), errno, rawSyscall6(), syscall6()

### Community 24 - "cpu.go"
Cohesion: 0.50
Nodes (4): CacheLinePad, option, init(), processOptions()

### Community 25 - "cpu_linux_riscv64.go"
Cohesion: 0.70
Nodes (4): riscvHWProbePairs, doinit(), isSet(), riscvHWProbe()

### Community 28 - "cpu_gccgo_x86.go"
Cohesion: 0.60
Nodes (4): cpuid(), gccgoGetCpuidCount(), gccgoXgetbv(), xgetbv()

### Community 29 - "syscall_darwin_arm64_gc.go"
Cohesion: 0.60
Nodes (4): darwinSysctlEnabled(), Errno, syscall_syscall6(), sysctlbyname()

### Community 30 - "CheckPermissions"
Cohesion: 0.67
Nodes (3): os.FileInfo, CheckPermissions(), SSLKeyPermissions()

### Community 31 - "CHANGELOG.md"
Cohesion: 0.40
Nodes (4): Features, Features, Fixes, Fixes

### Community 34 - "cpu_loong64.go"
Cohesion: 0.83
Nodes (3): cfgIsSet(), get_cpucfg(), initOptions()

### Community 82 - "encode.go"
Cohesion: 0.06
Nodes (36): database/sql.NullBool, github.com/lib/pq/oid.Oid, sync.RWMutex, time.Location, time.Time, accountFailureInfo, AccountLockoutManager, ActiveSessionItem (+28 more)

### Community 83 - "conn"
Cohesion: 0.22
Nodes (5): bufio.Reader, database/sql/driver.Tx, sync/atomic.Bool, Error, conn

### Community 84 - "generate_docx.py"
Cohesion: 0.50
Nodes (7): format_persian_text_with_dates(), main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph(), set_english_formatting(), set_persian_formatting()

### Community 87 - "readBuf"
Cohesion: 0.18
Nodes (4): github.com/lib/pq/internal/proto.ResponseCode, readBuf, parseStatementRowDescribe(), recvNotification()

### Community 88 - "context.Context"
Cohesion: 0.11
Nodes (14): context.Context, database/sql/driver.NamedValue, database/sql/driver.Result, database/sql/driver.Stmt, database/sql/driver.TxOptions, net.Conn, net.Dialer, copyin (+6 more)

### Community 89 - "package.json"
Cohesion: 0.05
Nodes (36): class-variance-authority, clsx, framer-motion, lucide-react, author, dependencies, class-variance-authority, clsx (+28 more)

### Community 90 - "be-ui-file-upload.tsx"
Cohesion: 0.09
Nodes (28): clampProgress(), cn(), createFileUploadItem(), EASE_DRAWER, EASE_IN_OUT, EASE_OUT, EASE_OUT_CSS, FAST_TRANSITION (+20 more)

### Community 92 - "Project Guidelines (HMD IT Herasat Portal)"
Cohesion: 0.40
Nodes (4): Codebase Navigation & Knowledge Graph (Graphify - Primary Reference), Communication Style (Caveman Mode & Language), Frontend Craft & Design System (Impeccable & Design Taste), Project Guidelines (HMD IT Herasat Portal)

### Community 100 - "button-1.tsx"
Cohesion: 0.50
Nodes (3): Button(), ButtonArrowProps, buttonVariants

### Community 103 - "krb.go"
Cohesion: 0.67
Nodes (3): GSS, NewGSSFunc, RegisterGSSProvider()

### Community 104 - "button.tsx"
Cohesion: 0.23
Nodes (7): Button, ButtonProps, buttonVariants, cn(), ButtonWithIconProps, DropdownMenuProps, DropdownOption

### Community 106 - "initAppTooltips"
Cohesion: 0.52
Nodes (6): initAppTooltips(), getOrCreateTooltip(), hide(), isMobileOrTouch(), positionTooltip(), show()

### Community 108 - "ContainsVulgarLanguage"
Cohesion: 0.20
Nodes (13): TestClassificationValidationAndMath(), TestValidatePasswordComplexity(), testing.T, cleanRune(), collapseRepeatedRunes(), containsBoundedPhrase(), ContainsVulgarLanguage(), hasVulgarPrefix() (+5 more)

## Knowledge Gaps
- **112 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `CaptchaPayload`, `CaptchaResponse`, `AccordionItemProps` (+107 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **27 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `conn` connect `conn` to `Listener`, `database/sql/driver.Value`, `writeBuf`, `Config`, `krb.go`, `parseError`, `encode.go`, `readBuf`, `context.Context`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `context.Context`, `Listener`, `conn`, `writeBuf`?**
  _High betweenness centrality (0.057) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `getIP()` (e.g. with `AdminSendMessageHandler()` and `ChangeMyPasswordHandler()`) actually correct?**
  _`getIP()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `CaptchaPayload` to the rest of the system?**
  _112 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Components` be split into smaller, more focused modules?**
  _Cohesion score 0.06896551724137931 - nodes in this community are weakly interconnected._
- **Should `Listener` be split into smaller, more focused modules?**
  _Cohesion score 0.05160662122687439 - nodes in this community are weakly interconnected._
- **Should `database/sql/driver.Value` be split into smaller, more focused modules?**
  _Cohesion score 0.06019871420222092 - nodes in this community are weakly interconnected._