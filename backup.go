package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
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
	filename := fmt.Sprintf("backup_herasat_%s.hbak", timestamp)
	targetPath := filepath.Join(backupDir, filename)

	var sqlDump bytes.Buffer
	var dumpErr error
	if _, lookErr := exec.LookPath("pg_dump"); lookErr == nil && os.Getenv("HERASAT_DB_CONN") != "" {
		dumpErr = dumpViaPgDump(&sqlDump)
	} else if isDockerContainerRunning("herasat_postgres") {
		dumpErr = dumpViaDocker(&sqlDump)
	} else if db != nil {
		dumpErr = dumpViaNativeSQL(db, &sqlDump)
	} else {
		_, _ = io.WriteString(&sqlDump, "-- HERASAT STANDALONE BACKUP DUMP\n")
	}

	if dumpErr != nil {
		_ = logAudit("SYSTEM", "BACKUP_FAILED", dumpErr.Error(), nil, "127.0.0.1")
		return "", fmt.Errorf("database backup failed: %w", dumpErr)
	}

	var tarBuf bytes.Buffer
	tw := tar.NewWriter(&tarBuf)

	sqlBytes := sqlDump.Bytes()
	sqlHdr := &tar.Header{
		Name:    "database.sql",
		Mode:    0600,
		Size:    int64(len(sqlBytes)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(sqlHdr); err != nil {
		return "", err
	}
	if _, err := tw.Write(sqlBytes); err != nil {
		return "", err
	}

	uploadsDir := "./secure_uploads"
	if entries, err := os.ReadDir(uploadsDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fPath := filepath.Join(uploadsDir, e.Name())
			fData, err := os.ReadFile(fPath)
			if err != nil {
				continue
			}
			info, _ := e.Info()
			modTime := time.Now()
			if info != nil {
				modTime = info.ModTime()
			}
			hdr := &tar.Header{
				Name:    "uploads/" + e.Name(),
				Mode:    0600,
				Size:    int64(len(fData)),
				ModTime: modTime,
			}
			if err := tw.WriteHeader(hdr); err == nil {
				_, _ = tw.Write(fData)
			}
		}
	}
	if err := tw.Close(); err != nil {
		return "", err
	}

	var gzBuf bytes.Buffer
	gzw := gzip.NewWriter(&gzBuf)
	if _, err := gzw.Write(tarBuf.Bytes()); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}

	key := GetBackupKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, gzBuf.Bytes(), []byte("HBAK_HEADER_V1"))

	finalData := make([]byte, 0, 5+len(nonce)+len(ciphertext))
	finalData = append(finalData, []byte("HBAK\x01")...)
	finalData = append(finalData, nonce...)
	finalData = append(finalData, ciphertext...)

	if err := os.WriteFile(targetPath, finalData, 0600); err != nil {
		_ = logAudit("SYSTEM", "BACKUP_FAILED", "WRITE_FILE_ERROR:"+err.Error(), nil, "127.0.0.1")
		return "", fmt.Errorf("failed to write encrypted backup: %w", err)
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

func RestoreBackup(backupFilePath string, db *sql.DB) error {
	data, err := os.ReadFile(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to read backup file: %w", err)
	}

	var decompressedGz []byte

	if len(data) >= 5 && string(data[:5]) == "HBAK\x01" {
		if len(data) < 5+12+16 {
			return errors.New("corrupted backup payload: insufficient length")
		}
		nonce := data[5:17]
		ciphertext := data[17:]

		key := GetBackupKey()
		block, err := aes.NewCipher(key)
		if err != nil {
			return err
		}
		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return err
		}
		decrypted, err := aesgcm.Open(nil, nonce, ciphertext, []byte("HBAK_HEADER_V1"))
		if err != nil {
			return fmt.Errorf("backup decryption failed: authentication tag mismatch or invalid key: %w", err)
		}
		decompressedGz = decrypted
	} else if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		decompressedGz = data
	} else {
		return errors.New("unrecognized backup format or corrupted header")
	}

	gzr, err := gzip.NewReader(bytes.NewReader(decompressedGz))
	if err != nil {
		return fmt.Errorf("failed to initialize gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var sqlContent string
	absUploadDir, _ := filepath.Abs("./secure_uploads")

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if hdr.Name == "database.sql" {
			b, err := io.ReadAll(tr)
			if err == nil {
				sqlContent = string(b)
			}
		} else if strings.HasPrefix(hdr.Name, "uploads/") {
			relName := strings.TrimPrefix(hdr.Name, "uploads/")
			targetPath := filepath.Join(absUploadDir, filepath.Clean(relName))
			if isWithinDirectory(absUploadDir, targetPath) {
				os.MkdirAll(filepath.Dir(targetPath), 0700)
				f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
				if err == nil {
					_, _ = io.Copy(f, tr)
					f.Close()
				}
			}
		}
	}

	if sqlContent != "" && db != nil {
		if _, err := db.Exec(sqlContent); err != nil {
			return fmt.Errorf("failed executing restored SQL: %w", err)
		}
	}

	return nil
}

func parseDBConn(connStr string) (string, string) {
	connStr = strings.TrimSpace(connStr)
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		if u, err := url.Parse(connStr); err == nil {
			if p, ok := u.User.Password(); ok {
				u.User = url.User(u.User.Username())
				return u.String(), p
			}
			return connStr, ""
		}
	}
	parts := strings.Fields(connStr)
	var safeParts []string
	var password string
	for _, p := range parts {
		if strings.HasPrefix(p, "password=") {
			password = strings.TrimPrefix(p, "password=")
			password = strings.Trim(password, "'\"")
		} else {
			safeParts = append(safeParts, p)
		}
	}
	return strings.Join(safeParts, " "), password
}

func dumpViaPgDump(w io.Writer) error {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		return fmt.Errorf("HERASAT_DB_CONN is empty")
	}

	safeConn, password := parseDBConn(connStr)
	cmd := exec.Command("pg_dump", "--dbname", safeConn, "--no-owner", "--no-acl", "--clean", "--if-exists")
	if password != "" {
		cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
	}
	cmd.Stdout = w
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failure: %v - stderr: %s", err, stderr.String())
	}
	return nil
}

func isDockerContainerRunning(containerName string) bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	checkCmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	out, err := checkCmd.Output()
	return err == nil && strings.Contains(string(out), "true")
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
	if db == nil {
		return errors.New("database connection is nil")
	}

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
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "backup_herasat_") {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".sql.gz") && !strings.HasSuffix(entry.Name(), ".hbak") {
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
