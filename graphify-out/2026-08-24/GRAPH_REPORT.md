# Graph Report - herasat_backend  (2026-08-24)

## Corpus Check
- 124 files · ~910,436 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 899 nodes · 1654 edges · 84 communities (66 shown, 18 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 126 edges (avg confidence: 0.85)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- conn
- Listener
- database/sql/driver.Value
- SessionManager
- net/http.Request
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
- krb.go

## God Nodes (most connected - your core abstractions)
1. `conn` - 60 edges
2. `Config` - 33 edges
3. `Listener` - 23 edges
4. `ListenerConn` - 22 edges
5. `getIP()` - 20 edges
6. `readBuf` - 20 edges
7. `writeBuf` - 19 edges
8. `rows` - 19 edges
9. `logAudit()` - 18 edges
10. `parseError()` - 18 edges

## Surprising Connections (you probably didn't know these)
- `RateLimitMiddleware()` --calls--> `getIP()`  [INFERRED]
  security.go → admin_auth.go
- `AdminAuthMiddleware()` --calls--> `isCookieSecure()`  [INFERRED]
  admin_handlers.go → security.go
- `DownloadAttachmentHandler()` --calls--> `DecryptFileBytes()`  [INFERRED]
  admin_handlers.go → crypto.go
- `main()` --calls--> `InitCrypto()`  [INFERRED]
  main.go → crypto.go
- `SubmitReportHandler()` --calls--> `EncryptFileBytes()`  [INFERRED]
  public_handlers.go → crypto.go

## Import Cycles
- None detected.

## Communities (84 total, 18 thin omitted)

### Community 0 - "conn"
Cohesion: 0.07
Nodes (14): bufio.Reader, database/sql/driver.Tx, github.com/lib/pq/internal/proto.RequestCode, github.com/lib/pq/internal/proto.ResponseCode, sync/atomic.Bool, readBuf, transactionStatus, writeBuf (+6 more)

### Community 1 - "Listener"
Cohesion: 0.07
Nodes (30): database/sql/driver.Conn, database/sql/driver.Connector, sync.Cond, time.Duration, Dialer, Driver, EventCallbackType, Listener (+22 more)

### Community 2 - "database/sql/driver.Value"
Cohesion: 0.06
Nodes (24): database/sql/driver.Rows, database/sql/driver.Value, database/sql/driver.Valuer, database/sql.Scanner, reflect.Value, ArrayDelimiter, BoolArray, ByteaArray (+16 more)

### Community 3 - "SessionManager"
Cohesion: 0.08
Nodes (21): AdminAuthMiddleware(), SuperuserAuthMiddleware(), InitCrypto(), enforceImmutableAuditLogs(), InitDB(), runMigrations(), net/http.Handler, net/http.HandlerFunc (+13 more)

### Community 4 - "net/http.Request"
Cohesion: 0.09
Nodes (58): generateSessionToken(), GetCurrentAdminHandler(), getIP(), logAudit(), LoginHandler(), LogoutHandler(), ChangeMyPasswordHandler(), CreateUserHandler() (+50 more)

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
Cohesion: 0.08
Nodes (13): reflect.Type, DialerContext, fieldDesc, format, noRows, rows, rowsHeader, safeRetryError (+5 more)

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
Cohesion: 0.13
Nodes (35): applyAuditFiltersAndPagination(), applyFilters(), checkAdminProfile(), deleteReportPermanently(), deleteUser(), escapeHTML(), exportReportsToCSV(), fetchAdminUsers() (+27 more)

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
Cohesion: 0.11
Nodes (25): database/sql.NullBool, github.com/lib/pq/oid.Oid, time.Location, time.Time, parameterStatus, locationCache, timestampParser, appendEncodedText() (+17 more)

### Community 87 - "krb.go"
Cohesion: 0.67
Nodes (3): GSS, NewGSSFunc, RegisterGSSProvider()

## Knowledge Gaps
- **34 isolated node(s):** `ReportSummary`, `AdminUserItem`, `AuditLogItem`, `CaptchaPayload`, `CaptchaResponse` (+29 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `conn` connect `conn` to `Listener`, `database/sql/driver.Value`, `SessionManager`, `Config`, `rows`, `context.Context`, `encode.go`, `krb.go`?**
  _High betweenness centrality (0.100) - this node is a cross-community bridge._
- **Why does `Config` connect `Config` to `conn`, `Listener`, `context.Context`?**
  _High betweenness centrality (0.071) - this node is a cross-community bridge._
- **Why does `ListenerConn` connect `Listener` to `SessionManager`?**
  _High betweenness centrality (0.055) - this node is a cross-community bridge._
- **What connects `ReportSummary`, `AdminUserItem`, `AuditLogItem` to the rest of the system?**
  _34 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `conn` be split into smaller, more focused modules?**
  _Cohesion score 0.07495495495495495 - nodes in this community are weakly interconnected._
- **Should `Listener` be split into smaller, more focused modules?**
  _Cohesion score 0.06682692307692308 - nodes in this community are weakly interconnected._
- **Should `database/sql/driver.Value` be split into smaller, more focused modules?**
  _Cohesion score 0.06110102843315184 - nodes in this community are weakly interconnected._