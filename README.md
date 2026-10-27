# Dideban-e Amin (دیده‌بان امین)

Confidential whistleblower and intelligence reporting portal engineered for government education departments. Built with Go standard library, PostgreSQL 16, native vanilla CSS/JS interface, and zero external dependencies.

---

## Core Security Architecture

1. **Cryptographic Anonymity & AAD Field Binding**:
   - Report fields (descriptions, reporter identities, phone numbers, notes) encrypted with authenticated AES-256-GCM bound to context (`v3:<kid>:b:<base64>`) using Additional Authenticated Data (`table:column:rowID`).
   - Ciphertexts cannot be swapped, relocated across rows, or replayed across columns.
   - Multi-key ring support (`HERASAT_KEY_ID`, `HERASAT_AES_KEY_PREVIOUS`) for zero-downtime key rotation.

2. **Blind Indexing for Reporter Tracking**:
   - Public tracking codes are blinded using HMAC-SHA256 (`tracking_code_hash`).
   - Plaintext tracking codes are never stored, logged, or queried directly.

3. **Cryptographic WORM Audit Trail**:
   - Cryptographically linked HMAC-SHA256 audit log verified at microsecond precision.
   - PostgreSQL WORM triggers (`prevent_audit_log_modification`) block all `UPDATE`, `DELETE`, and `TRUNCATE` operations on audit tables.
   - Post-commit remote RFC 5424 syslog transmission to SIEM collectors.

4. **Upload Defense & Deep Metadata Sanitization**:
   - Streamed chunk-based upload processing with 1MB memory ceiling.
   - Concurrency-limited upload pipeline with HTTP 503 load shedding.
   - Deep metadata sanitization for PDF, JPEG, PNG, DOCX, XLSX, MP3, and WAV attachments.
   - Token-aware PDF active content scanner and OOXML macro detection.
   - Direct streaming to ClamAV antivirus daemon (`zINSTREAM`).
   - ZIP bomb protection via aggregate archive compression ratio analysis.

5. **Authentication & Access Control**:
   - Argon2id password hashing with constant-time dummy verification for non-existent users (prevents timing-based username enumeration).
   - Compound sliding lockout keyed by username and client IP with progressive exponential delays.
   - Database-persisted TOTP replay prevention (`used_mfa_nonces`).
   - Network segmentation via CIDR allow-listing for administrative endpoints.
   - Scoped Server-Sent Events (SSE) broadcasts restricted to authorized regional administrators.

---

## Environment Configuration

Configure the environment using `.env`:

```bash
cp .env.example .env
```

Key variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `HERASAT_DB_CONN` | PostgreSQL connection string | `host=127.0.0.1 port=5432 ...` |
| `HERASAT_AES_KEY` | Primary 32-byte master encryption key | Required |
| `HERASAT_HMAC_KEY` | Primary 32-byte HMAC key | Required |
| `HERASAT_KEY_ID` | Active key identifier for key ring | `k1` |
| `HERASAT_AES_KEY_PREVIOUS` | Comma-separated previous keys (`kid:hex_or_raw`) | Optional |
| `HERASAT_ADMIN_ALLOWED_CIDRS` | Allowed administrative IP ranges | `127.0.0.1/32` |
| `HERASAT_AUDIT_STRICT` | Strict audit chain verification on startup | `1` |
| `HERASAT_CLAMAV_ENABLED` | Enable ClamAV antivirus scanning | `true` |
| `HERASAT_BACKUP_ENABLED` | Enable automated encrypted backups | `true` |

---

## Operational Commands & CLI

### 1. Database Migrations & Startup
```bash
go run .
```

### 2. Manual Backup
Creates an encrypted backup (`HBAK\x02` format) containing PostgreSQL dump and evidence attachments:
```bash
herasat backup [optional_output_path]
```

### 3. Database Restore
Restores an encrypted archive into the target database:
```bash
herasat restore <path_to_backup.hbak>
```

### 4. Cryptographic Key Re-encryption
Re-encrypts all database records under the active master key:
```bash
herasat reencrypt
```

---

## Testing & Verification

Run the full verification suite against PostgreSQL:

```bash
export HERASAT_DB_CONN="host=127.0.0.1 port=5432 user=herasat_user password=... dbname=herasat_db sslmode=disable"
export HERASAT_REQUIRE_DB_TESTS=1
go test -v ./...
```
