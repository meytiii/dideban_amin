# Graph Report - herasat_backend  (2026-08-25)

## Corpus Check
- 126 files · ~980,816 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 953 nodes · 1780 edges · 85 communities (68 shown, 17 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 139 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- conn
- Listener
- database/sql/driver.Value
- SessionManager
- admin_handlers.go
- Config
- Client
- cpu_arm64.go
- rows
- block
- queryResult
- doinit
- README.md
- context.Context
- admin.js
- XOF
- syscall_darwin_x86_gc.go
- Error
- deprecated.go
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
- captcha.go
- generate_docx.py

## God Nodes (most connected - your core abstractions)
1. `conn` - 60 edges
2. `Config` - 33 edges
3. `getIP()` - 23 edges
4. `Listener` - 23 edges
5. `ListenerConn` - 22 edges
6. `logAudit()` - 21 edges
7. `readBuf` - 20 edges
8. `writeBuf` - 19 edges
9. `rows` - 19 edges
10. `parseError()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `RateLimitMiddleware()` --calls--> `getIP()`  [INFERRED]
  security.go → admin_auth.go
- `AdminAuthMiddleware()` --calls--> `isCookieSecure()`  [INFERRED]
  admin_handlers.go → security.go
- `DownloadAttachmentHandler()` --calls--> `DecryptFileBytes()`  [INFERRED]
  admin_handlers.go → crypto.go
- `GenerateCaptchaToken()` --calls--> `GetHMACKey()`  [INFERRED]
  captcha.go → crypto.go
- `main()` --calls--> `InitCrypto()`  [INFERRED]
  main.go → crypto.go

## Import Cycles
- None detected.

## Communities (85 total, 17 thin omitted)

### Community 0 - "conn"
Cohesion: 0.07
Nodes (18): bufio.Reader, database/sql/driver.Tx, github.com/lib/pq/internal/proto.RequestCode, github.com/lib/pq/internal/proto.ResponseCode, sync/atomic.Bool, GSS, NewGSSFunc, readBuf (+10 more)

### Community 1 - "Listener"
Cohesion: 0.06
Nodes (33): database/sql/driver.Conn, database/sql/driver.Connector, sync.Cond, time.Duration, Dialer, DialerContext, Driver, EventCallbackType (+25 more)

### Community 2 - "database/sql/driver.Value"
Cohesion: 0.06
Nodes (24): database/sql/driver.Rows, database/sql/driver.Value, database/sql/driver.Valuer, database/sql.Scanner, reflect.Value, ArrayDelimiter, BoolArray, ByteaArray (+16 more)

### Community 3 - "SessionManager"
Cohesion: 0.08
Nodes (20): AdminAuthMiddleware(), SuperuserAuthMiddleware(), InitCrypto(), enforceImmutableAuditLogs(), InitDB(), runMigrations(), net/http.Handler, net/http.HandlerFunc (+12 more)

### Community 4 - "admin_handlers.go"
Cohesion: 0.07
Nodes (70): generateSessionToken(), GetCurrentAdminHandler(), getIP(), logAudit(), LoginHandler(), LogoutHandler(), AdminSendMessageHandler(), ChangeMyPasswordHandler() (+62 more)

### Community 5 - "Config"
Cohesion: 0.11
Nodes (25): crypto/tls.Config, database/sql/driver.Driver, net/netip.Addr, Config, ConfigMultihost, LoadBalanceHosts, ProtocolVersion, SSLMode (+17 more)

### Community 6 - "Client"
Cohesion: 0.08
Nodes (17): hash.Hash, Client, NewClient(), appendUint64(), checkSum(), consumeUint64(), digest, digest (+9 more)

### Community 7 - "cpu_arm64.go"
Cohesion: 0.08
Nodes (26): aarch64SysctlCPUID, sysctlNode, syscall.Errno, archInit(), extractBits(), parseARM64SVERegister(), parseARM64SystemRegisters(), readARM64Registers() (+18 more)

### Community 8 - "rows"
Cohesion: 0.10
Nodes (10): reflect.Type, fieldDesc, format, noRows, rows, rowsHeader, decideColumnFormats(), parsePortalRowDescribe() (+2 more)

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

### Community 13 - "context.Context"
Cohesion: 0.11
Nodes (14): context.Context, database/sql/driver.NamedValue, database/sql/driver.Result, database/sql/driver.Stmt, database/sql/driver.TxOptions, net.Conn, net.Dialer, copyin (+6 more)

### Community 14 - "admin.js"
Cohesion: 0.09
Nodes (48): applyAuditFiltersAndPagination(), applyFilters(), checkAdminProfile(), checkMustChangePasswordPolicy(), deleteReportPermanently(), deleteUser(), escapeHTML(), exportReportsToCSV() (+40 more)

### Community 15 - "XOF"
Cohesion: 0.19
Nodes (6): XOF, io.Reader, io.Writer, digest, digest, NewXOF()

### Community 16 - "syscall_darwin_x86_gc.go"
Cohesion: 0.29
Nodes (9): _C_int, darwinKernelVersionCheck(), darwinSupportsAVX512(), darwinOSRelease(), darwinSysctlEnabled(), Errno, syscall_syscall6(), sysctl() (+1 more)

### Community 17 - "Error"
Cohesion: 0.24
Nodes (5): github.com/lib/pq/pqerror.Code, expandTab(), conn, Error, posToLine()

### Community 18 - "deprecated.go"
Cohesion: 0.31
Nodes (7): bytes.Buffer, PGError, CopyIn(), CopyInSchema(), makeStmt(), ParseURL(), BufferQuoteIdentifier()

### Community 19 - "blake2bAVX2_amd64.go"
Cohesion: 0.27
Nodes (6): hashBlocksGeneric(), hashBlocks(), hashBlocks(), hashBlocksAVX(), hashBlocksAVX2(), hashBlocksSSE4()

### Community 20 - "app.js"
Cohesion: 0.13
Nodes (7): clearFileInput(), dismissToast(), handleSelectedFile(), restoreDraft(), saveDraft(), scheduleDraftSave(), showToast()

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
Cohesion: 0.10
Nodes (26): database/sql.NullBool, github.com/lib/pq/oid.Oid, sync.Mutex, time.Location, time.Time, parameterStatus, locationCache, timestampParser (+18 more)

### Community 83 - "captcha.go"
Cohesion: 0.42
Nodes (8): cryptoRandInt(), GenerateCaptchaToken(), GenerateMathCaptchaSVG(), generateNoiseElements(), GetCaptchaHandler(), renderObfuscatedChars(), CaptchaPayload, CaptchaResponse

### Community 84 - "generate_docx.py"
Cohesion: 0.80
Nodes (4): main(), populate_docx(), replace_in_table(), replace_placeholders_in_paragraph()

## Knowledge Gaps
- **33 isolated node(s):** `AdminUserItem`, `AuditLogItem`, `CaptchaPayload`, `CaptchaResponse`, `herasat.local/backend` (+28 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **17 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `conn` connect `conn` to `Listener`, `database/sql/driver.Value`, `Config`, `context.Context`, `encode.go`?**
  _High betweenness centrality (0.090) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `conn`, `Listener`, `context.Context`?**
  _High betweenness centrality (0.066) - this node is a cross-community bridge._
- **Why does `ListenerConn` connect `Listener` to `encode.go`?**
  _High betweenness centrality (0.052) - this node is a cross-community bridge._
- **Are the 19 inferred relationships involving `getIP()` (e.g. with `AdminSendMessageHandler()` and `ChangeMyPasswordHandler()`) actually correct?**
  _`getIP()` has 19 INFERRED edges - model-reasoned connections that need verification._
- **What connects `AdminUserItem`, `AuditLogItem`, `CaptchaPayload` to the rest of the system?**
  _33 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `conn` be split into smaller, more focused modules?**
  _Cohesion score 0.0653417645287564 - nodes in this community are weakly interconnected._
- **Should `Listener` be split into smaller, more focused modules?**
  _Cohesion score 0.06086956521739131 - nodes in this community are weakly interconnected._