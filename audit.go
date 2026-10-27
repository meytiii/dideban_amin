package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	auditFailures uint64
	auditMu       sync.Mutex
)

const (
	genesisAuditHash    = "0000000000000000000000000000000000000000000000000000000000000000"
	auditAdvisoryLockID = 14021105
)

func computeAuditHash(prevHash, username, title, action string, targetID *int64, ip, detail string, t time.Time) string {
	targetStr := "none"
	if targetID != nil {
		targetStr = strconv.FormatInt(*targetID, 10)
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s", prevHash, username, title, action, targetStr, ip, detail, t.Format(time.RFC3339Nano))
	mac := hmac.New(sha256.New, GetAuditHMACKey())
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}

func computeLegacyAuditHash(prevHash, username, title, action string, targetID *int64, ip string, t time.Time) string {
	targetStr := "none"
	if targetID != nil {
		targetStr = strconv.FormatInt(*targetID, 10)
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", prevHash, username, title, action, targetStr, ip, t.Format(time.RFC3339Nano))
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func getLatestAuditHashLocked(tx *sql.Tx) (string, error) {
	var h sql.NullString
	err := tx.QueryRow("SELECT entry_hash FROM audit_logs ORDER BY id DESC LIMIT 1 FOR UPDATE").Scan(&h)
	if err == sql.ErrNoRows {
		return genesisAuditHash, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to read previous audit entry hash: %w", err)
	}
	if !h.Valid || h.String == "" {
		return genesisAuditHash, nil
	}
	return h.String, nil
}

func executeAuditLogTx(tx *sql.Tx, username, title, action string, targetID *int64, ip, detail string, t time.Time) error {
	if tx == nil {
		return errors.New("audit transaction is nil")
	}

	t = t.UTC().Truncate(time.Microsecond)

	if _, err := tx.Exec("SELECT pg_advisory_xact_lock($1)", auditAdvisoryLockID); err != nil {
		return fmt.Errorf("failed to acquire audit advisory lock: %w", err)
	}

	prevHash, err := getLatestAuditHashLocked(tx)
	if err != nil {
		return err
	}

	entryHash := computeAuditHash(prevHash, username, title, action, targetID, ip, detail, t)

	_, err = tx.Exec(`
		INSERT INTO audit_logs (admin_username, admin_title, action, target_report_id, ip_address, detail, prev_hash, entry_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		username, title, action, targetID, ip, detail, prevHash, entryHash, t)
	if err != nil {
		return fmt.Errorf("failed to insert audit log entry: %w", err)
	}

	return nil
}

func logAuditTxWithDetail(tx *sql.Tx, username, title, action string, targetID *int64, ip, detail string) error {
	auditMu.Lock()
	defer auditMu.Unlock()

	ShipSyslogAudit(username, action, targetID, ip)
	if tx == nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG TX FAILURE: tx is nil (user=%s, action=%s)", username, action)
		return errors.New("tx is nil")
	}

	now := time.Now().UTC()
	err := executeAuditLogTx(tx, username, title, action, targetID, ip, detail, now)
	if err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG TX FAILURE: %v (user=%s, action=%s)", err, username, action)
		return err
	}
	return nil
}

func logAuditWithDetail(username, title, action string, targetID *int64, ip, detail string) error {
	auditMu.Lock()
	defer auditMu.Unlock()

	ShipSyslogAudit(username, action, targetID, ip)
	if DB == nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG FAILURE: DB is nil (user=%s, action=%s)", username, action)
		return errors.New("DB is nil")
	}

	tx, err := DB.Begin()
	if err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG TX BEGIN FAILURE: %v (user=%s, action=%s)", err, username, action)
		return fmt.Errorf("failed to begin audit transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	if err := executeAuditLogTx(tx, username, title, action, targetID, ip, detail, now); err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG FAILURE: %v (user=%s, action=%s)", err, username, action)
		return err
	}

	if err := tx.Commit(); err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG COMMIT FAILURE: %v (user=%s, action=%s)", err, username, action)
		return fmt.Errorf("failed to commit audit transaction: %w", err)
	}

	return nil
}

func logAudit(username, title, action string, targetID *int64, ip string) error {
	return logAuditWithDetail(username, title, action, targetID, ip, "")
}

func logAuditTx(tx *sql.Tx, username, title, action string, targetID *int64, ip string) error {
	return logAuditTxWithDetail(tx, username, title, action, targetID, ip, "")
}

type AuditVerificationResult struct {
	Valid         bool    `json:"valid"`
	VerifiedCount int     `json:"verified_count"`
	FallbackRows  []int64 `json:"fallback_rows,omitempty"`
	BrokenRowID   int64   `json:"broken_row_id,omitempty"`
}

func VerifyAuditChainDetailed(db *sql.DB) (AuditVerificationResult, error) {
	res := AuditVerificationResult{}
	if db == nil {
		return res, errors.New("db is nil")
	}

	rows, err := db.Query(`
		SELECT id, admin_username, admin_title, action, target_report_id, ip_address, COALESCE(detail, ''), prev_hash, entry_hash, created_at
		FROM audit_logs
		ORDER BY id ASC`)
	if err != nil {
		return res, fmt.Errorf("failed to query audit logs for verification: %w", err)
	}
	defer rows.Close()

	expectedPrevHash := genesisAuditHash
	verifiedCount := 0

	for rows.Next() {
		var id int64
		var username, title, action, ip, detail, prevHash, entryHash string
		var targetID sql.NullInt64
		var createdAt time.Time

		if err := rows.Scan(&id, &username, &title, &action, &targetID, &ip, &detail, &prevHash, &entryHash, &createdAt); err != nil {
			res.VerifiedCount = verifiedCount
			res.BrokenRowID = id
			return res, fmt.Errorf("failed to scan audit row %d: %w", id, err)
		}

		if prevHash != expectedPrevHash {
			res.VerifiedCount = verifiedCount
			res.BrokenRowID = id
			return res, fmt.Errorf("audit chain broken at id %d: expected prev_hash %s, got %s", id, expectedPrevHash, prevHash)
		}

		var tidPtr *int64
		if targetID.Valid {
			v := targetID.Int64
			tidPtr = &v
		}

		createdAtTrunc := createdAt.UTC().Truncate(time.Microsecond)
		computedHMAC := computeAuditHash(prevHash, username, title, action, tidPtr, ip, detail, createdAtTrunc)
		if computedHMAC != entryHash {
			fallbackHMAC := computeAuditHash(prevHash, username, title, action, tidPtr, ip, detail, createdAt)
			if fallbackHMAC == entryHash {
				res.FallbackRows = append(res.FallbackRows, id)
			} else {
				legacySHA := computeLegacyAuditHash(prevHash, username, title, action, tidPtr, ip, createdAt)
				if legacySHA == entryHash {
					res.FallbackRows = append(res.FallbackRows, id)
				} else {
					res.VerifiedCount = verifiedCount
					res.BrokenRowID = id
					return res, fmt.Errorf("audit hash mismatch at id %d: computed %s, stored %s", id, computedHMAC, entryHash)
				}
			}
		}

		expectedPrevHash = entryHash
		verifiedCount++
	}

	if err := rows.Err(); err != nil {
		res.VerifiedCount = verifiedCount
		return res, err
	}

	res.Valid = true
	res.VerifiedCount = verifiedCount
	return res, nil
}

func VerifyAuditChain(db *sql.DB) (bool, int, error) {
	res, err := VerifyAuditChainDetailed(db)
	return res.Valid, res.VerifiedCount, err
}

func VerifyAuditChainOnStartup(db *sql.DB) error {
	_, count, err := VerifyAuditChain(db)
	if err != nil {
		log.Printf("[AUDIT ALERT] Audit log chain integrity verification failed: %v", err)
		if os.Getenv("HERASAT_DEV_MODE") != "1" && os.Getenv("HERASAT_AUDIT_STRICT") == "1" {
			return fmt.Errorf("audit log chain integrity failure: %w", err)
		}
		return nil
	}
	log.Printf("[AUDIT] Successfully verified integrity of %d audit log entries in cryptographic chain.", count)
	return nil
}

func VerifyAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	valid, count, err := VerifyAuditChain(DB)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":         "error",
			"valid":          false,
			"verified_count": count,
			"error":          err.Error(),
			"message":        "خطا در اعتبارسنجی زنجیره لاگ‌های امنیتی",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"valid":          valid,
		"verified_count": count,
		"message":        fmt.Sprintf("زنجیره لاگ‌های امنیتی با موفقیت بررسی شد (%d رکورد معتبر).", count),
	})
}
