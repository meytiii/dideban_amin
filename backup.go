package main

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var backupMutex sync.Mutex

func StartAutoBackup(db *sql.DB) {
	if db == nil {
		return
	}

	enabled := strings.ToLower(strings.TrimSpace(os.Getenv("HERASAT_BACKUP_ENABLED")))
	if enabled == "false" || enabled == "0" {
		log.Println("[INFO] Automated database backup runner is disabled via HERASAT_BACKUP_ENABLED=false")
		return
	}

	intervalHours := 24
	if envInterval := os.Getenv("HERASAT_BACKUP_INTERVAL_HOURS"); envInterval != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(envInterval)); err == nil && parsed > 0 {
			intervalHours = parsed
		}
	}

	log.Printf("[INFO] Automated database backup runner initialized (Interval: %d hours)", intervalHours)

	go func() {
		select {
		case <-time.After(30 * time.Second):
			if backupPath, err := PerformBackup(db); err != nil {
				log.Printf("[ERROR] Initial automated backup failed: %v", err)
			} else {
				log.Printf("[INFO] Initial automated backup generated successfully: %s", backupPath)
			}
		}

		ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if backupPath, err := PerformBackup(db); err != nil {
				log.Printf("[ERROR] Scheduled automated backup failed: %v", err)
			} else {
				log.Printf("[INFO] Scheduled automated backup generated successfully: %s", backupPath)
			}
		}
	}()
}

func PerformBackup(db *sql.DB) (string, error) {
	backupMutex.Lock()
	defer backupMutex.Unlock()

	backupDir := os.Getenv("HERASAT_BACKUP_DIR")
	if strings.TrimSpace(backupDir) == "" {
		backupDir = "./backups"
	}

	if err := os.MkdirAll(backupDir, 0700); err != nil {
		_ = logAudit("SYSTEM", "BACKUP_FAILED", "MKDIR_ERROR:"+err.Error(), nil, "127.0.0.1")
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_herasat_%s.sql.gz", timestamp)
	targetPath := filepath.Join(backupDir, filename)

	f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		_ = logAudit("SYSTEM", "BACKUP_FAILED", "CREATE_FILE_ERROR:"+err.Error(), nil, "127.0.0.1")
		return "", fmt.Errorf("failed to open backup target file: %w", err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	var dumpErr error
	if _, lookErr := exec.LookPath("pg_dump"); lookErr == nil {
		dumpErr = dumpViaPgDump(gzw)
	} else if _, dockerErr := exec.LookPath("docker"); dockerErr == nil {
		dumpErr = dumpViaDocker(gzw)
	} else {
		dumpErr = dumpViaNativeSQL(db, gzw)
	}

	if dumpErr != nil {
		if closeErr := gzw.Close(); closeErr != nil {
			log.Printf("[WARN] Failed closing gzip stream during error unwind: %v", closeErr)
		}
		_ = f.Close()
		_ = os.Remove(targetPath)
		_ = logAudit("SYSTEM", "BACKUP_FAILED", dumpErr.Error(), nil, "127.0.0.1")
		return "", fmt.Errorf("database backup failed: %w", dumpErr)
	}

	if err := gzw.Flush(); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}

	retentionDays := 14
	if envRet := os.Getenv("HERASAT_BACKUP_RETENTION_DAYS"); envRet != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(envRet)); err == nil && parsed > 0 {
			retentionDays = parsed
		}
	}
	rotateOldBackups(backupDir, retentionDays)

	_ = logAudit("SYSTEM", "BACKUP_COMPLETED", filename, nil, "127.0.0.1")
	return targetPath, nil
}

func dumpViaPgDump(w io.Writer) error {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		return fmt.Errorf("HERASAT_DB_CONN is empty")
	}

	cmd := exec.Command("pg_dump", "--dbname", connStr, "--no-owner", "--no-acl", "--clean", "--if-exists")
	cmd.Stdout = w
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failure: %v - stderr: %s", err, stderr.String())
	}
	return nil
}

func dumpViaDocker(w io.Writer) error {
	containerName := "herasat_postgres"
	dbUser := "herasat_user"
	dbName := "herasat_db"

	checkCmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	out, err := checkCmd.Output()
	if err != nil || !strings.Contains(string(out), "true") {
		return fmt.Errorf("container %s not actively running", containerName)
	}

	cmd := exec.Command("docker", "exec", containerName, "pg_dump", "-U", dbUser, "--no-owner", "--no-acl", "--clean", "--if-exists", dbName)
	cmd.Stdout = w
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker exec pg_dump failure: %v - stderr: %s", err, stderr.String())
	}
	return nil
}

func dumpViaNativeSQL(db *sql.DB, w io.Writer) error {
	header := "SET statement_timeout = 0;\nSET client_encoding = 'UTF8';\nBEGIN;\n\n"
	if _, err := io.WriteString(w, header); err != nil {
		return err
	}

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE' ORDER BY table_name")
	if err != nil {
		return fmt.Errorf("failed querying table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tables = append(tables, name)
		}
	}
	_ = rows.Close()

	for _, tbl := range tables {
		colRows, err := db.Query("SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 ORDER BY ordinal_position", tbl)
		if err != nil {
			continue
		}

		var colNames []string
		for colRows.Next() {
			var colName, colType string
			if err := colRows.Scan(&colName, &colType); err == nil {
				colNames = append(colNames, colName)
			}
		}
		_ = colRows.Close()

		if len(colNames) == 0 {
			continue
		}

		qColList := make([]string, len(colNames))
		for i, c := range colNames {
			qColList[i] = fmt.Sprintf("\"%s\"", c)
		}
		colsJoined := strings.Join(qColList, ", ")

		dataRows, err := db.Query(fmt.Sprintf("SELECT %s FROM \"%s\"", colsJoined, tbl))
		if err != nil {
			continue
		}

		scanArgs := make([]interface{}, len(colNames))
		values := make([]interface{}, len(colNames))
		for i := range values {
			scanArgs[i] = &values[i]
		}

		for dataRows.Next() {
			if err := dataRows.Scan(scanArgs...); err != nil {
				continue
			}

			valStrings := make([]string, len(values))
			for i, v := range values {
				if v == nil {
					valStrings[i] = "NULL"
				} else {
					switch val := v.(type) {
					case []byte:
						escaped := strings.ReplaceAll(string(val), "'", "''")
						valStrings[i] = fmt.Sprintf("'%s'", escaped)
					case string:
						escaped := strings.ReplaceAll(val, "'", "''")
						valStrings[i] = fmt.Sprintf("'%s'", escaped)
					case time.Time:
						valStrings[i] = fmt.Sprintf("'%s'", val.Format(time.RFC3339Nano))
					case bool:
						if val {
							valStrings[i] = "TRUE"
						} else {
							valStrings[i] = "FALSE"
						}
					default:
						valStrings[i] = fmt.Sprintf("%v", val)
					}
				}
			}

			insertStmt := fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s);\n", tbl, colsJoined, strings.Join(valStrings, ", "))
			if _, err := io.WriteString(w, insertStmt); err != nil {
				_ = dataRows.Close()
				return err
			}
		}
		_ = dataRows.Close()
	}

	if _, err := io.WriteString(w, "\nCOMMIT;\n"); err != nil {
		return err
	}
	return nil
}

func rotateOldBackups(backupDir string, retentionDays int) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "backup_herasat_") || !strings.HasSuffix(entry.Name(), ".sql.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			target := filepath.Join(backupDir, entry.Name())
			_ = os.Remove(target)
			log.Printf("[INFO] Rotated and removed expired backup archive: %s", entry.Name())
		}
	}
}
