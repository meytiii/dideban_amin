# Dideban-e Amin (دیده‌بان امین)

سامانه حراست و دریافت گزارش‌های محرمانه مردمی و درون‌سازمانی (Herasat Whistleblowing Portal)

Pure Go standard library backend, PostgreSQL 16 database, native vanilla HTML/CSS/JS frontend. Zero external frontend framework dependencies; strict fail-closed security architecture.

---

## Architecture & Security Highlights

- **At-Rest Cryptography (v3 AAD)**: AES-256-GCM with Additional Authenticated Data (AAD) cryptographically binding ciphertexts to database coordinates (`table:column:rowID`), preventing cut-and-paste tampering.
- **Blind Tracking Code Indexing**: Whistleblower tracking codes are encrypted at rest; public database queries and follow-up threads utilize keyed HMAC-SHA256 hashes (`TRACK:<code_text>`) without exposing raw tracking codes in database dumps.
- **Tamper-Evident WORM Audit Log Chain**: Keyed HMAC-SHA256 forward-chained audit ledger with transaction-level advisory locking (`pg_advisory_xact_lock(14021105)`), trigger-enforced immutability, least-privilege role separation, and real-time cryptographic verification endpoint (`GET /api/admin/audit-logs/verify`).
- **Two-Person Report Deletion (Dual Control)**: Optional operational workflow (`HERASAT_TWO_PERSON_DELETION=true`) requiring distinct superuser approval within 24 hours to delete reports and attached evidence files.
- **Granular Superuser Redaction**: Dedicated endpoint (`POST /api/admin/report/redact`) allowing targeted redaction of whistleblower identifiers, report description, or specific evidence files while preserving report metadata and immutable audit history.
- **Active Payload Neutralization & File Sanitization**:
  - Image bomb protection via dimension/megapixel bounds.
  - Active PDF parsing rejecting scripts, hex literals, and `/FlateDecode` launch commands.
  - Metadata sanitization for PDFs (`/Info` dictionary), Office OOXML (`docProps/core.xml`), MP3 (ID3 tags), and WAV (RIFF chunks).
  - Malicious archive rejection (decompression ratio limits, nested archive bans, path traversal guards).
- **Hardened Network & Session Controls**:
  - CIDR-based admin portal segmentation and cached trusted reverse proxy validation.
  - Subnet-aware IPv6 `/64` prefix rate limiting with deterministic LRU eviction.
  - Compound `(username, ip)` account lockout and nonce-bound MFA authentication.
  - Authenticated AES-GCM encrypted database and file backup archives (`.hbak`).

---

## Quickstart (Local Docker Setup)

### 1. Prerequisites
- Docker Engine & Docker Compose
- Go 1.22+ (for running the application binary locally)

### 2. Start PostgreSQL & ClamAV via Docker Compose
```bash
docker compose up -d
```
This automatically boots:
- **PostgreSQL 16** on port `5432` with initial schema (`docker/init/01_schema.sql`) and least-privilege role configurations (`docker/init/03_least_privilege_roles.sql`).
- **ClamAV Daemon** on port `3310` with daily virus definitions.

### 3. Configure Environment
Copy `.env.example` to `.env` and set your configuration keys:
```bash
cp .env.example .env
```
Ensure you generate 32 random bytes for `HERASAT_AES_KEY` and `HERASAT_HMAC_KEY`:
```bash
# Example generating 32 random bytes
openssl rand -hex 16
```

### 4. Start the Application
```bash
go run .
```

### 5. Access the Portals
- **Public Whistleblower Portal**: [http://localhost:8080/](http://localhost:8080/)
- **Administrative Cartable**: [http://localhost:8080/admin/](http://localhost:8080/admin/)

#### Initial Superuser Bootstrap
If the database contains no active superusers, the backend automatically provisions `admin` (or `HERASAT_INITIAL_ADMIN_USERNAME`).
If `HERASAT_INITIAL_ADMIN_PASSWORD` is not set in `.env`, a cryptographically random 24-character password is generated and printed once to stdout upon first boot. Password change is enforced immediately on initial login.

---

## Production Deployment

### Database Least-Privilege Roles
Production environments should execute database migrations using `herasat_migrator` and connect the web application using `herasat_app`. The `herasat_app` user has strict `REVOKE` privileges on `UPDATE`, `DELETE`, and `TRUNCATE` against the `audit_logs` table.

```bash
# Apply schema and role permissions
psql -U postgres -d herasat_db -f migrations/001_initial_schema.sql
psql -U postgres -d herasat_db -f migrations/002_least_privilege_roles.sql
```

### Strict Startup Gates
In production (`HERASAT_DEV_MODE=0`), the service enforces fail-closed checks on startup:
1. `HERASAT_AES_KEY` and `HERASAT_HMAC_KEY` must be set, distinct, exactly 32 bytes, and non-placeholder.
2. `HERASAT_ADMIN_ALLOWED_CIDRS` must contain at least one valid CIDR.
3. If `HERASAT_COOKIE_SECURE=true` is set without direct TLS certs, `HERASAT_TRUSTED_PROXIES` must be defined.
4. ClamAV antivirus daemon must respond to `PING` with `PONG`.
5. Cryptographic audit chain must verify from genesis to head (`HERASAT_AUDIT_STRICT=1`).

---

## Backup and Restore

### Automated Backups
Automated encrypted backups run periodically according to `HERASAT_BACKUP_INTERVAL_HOURS` (default: 24h) and are stored in `HERASAT_BACKUP_DIR` (default: `./backups`).
Archives are encrypted with `GetBackupKey()`, authenticated with AES-GCM, and prefixed with magic header `HBAK\x01`.

### Restoring a Backup
```go
err := RestoreBackup("./backups/backup_herasat_20261027_120000.hbak", db)
```

---

## Test Suite Execution

Run the comprehensive test suite covering all security controls, cryptography, parsing, and failure modes:
```bash
go test -v ./...
```
