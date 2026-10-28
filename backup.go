package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/json"
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

const (
	backupMagicV1   = "HBAK\x01"
	backupMagicV2   = "HBAK\x02"
	backupChunkSize = 1024 * 1024
)

type BackupManifest struct {
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files"`
	ChunkSize int       `json:"chunk_size"`
}

func getPgDumpPath() string {
	if p := strings.TrimSpace(os.Getenv("HERASAT_PG_DUMP_PATH")); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath("pg_dump"); err == nil {
		return p
	}
	return ""
}

func getPsqlPath() string {
	if p := strings.TrimSpace(os.Getenv("HERASAT_PSQL_PATH")); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath("psql"); err == nil {
		return p
	}
	return ""
}

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

func dumpViaPgDumpToFile(targetFilePath string) error {
	dockerContainer := strings.TrimSpace(os.Getenv("HERASAT_DOCKER_CONTAINER"))
	pgDumpPath := getPgDumpPath()
	if pgDumpPath == "" && dockerContainer == "" {
		return errors.New("pg_dump executable not found: automated streaming backup requires PostgreSQL client utilities")
	}

	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		return fmt.Errorf("HERASAT_DB_CONN is empty")
	}

	safeConn, password := parseDBConn(connStr)
	var cmd *exec.Cmd
	if pgDumpPath != "" {
		cmd = exec.Command(pgDumpPath, "--dbname", safeConn, "--no-owner", "--no-acl", "--clean", "--if-exists")
		if password != "" {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
		}
	} else {
		cmd = exec.Command("docker", "exec", "-e", "PGPASSWORD", dockerContainer, "pg_dump", "--dbname", safeConn, "--no-owner", "--no-acl", "--clean", "--if-exists")
		if password != "" {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
		}
	}

	outFile, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gzw := gzip.NewWriter(outFile)
	defer gzw.Close()

	cmd.Stdout = gzw
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pg_dump failure: %v - stderr: %s", err, stderr.String())
	}

	if err := gzw.Close(); err != nil {
		return err
	}
	return outFile.Close()
}

func PerformBackup(db *sql.DB, customTarget ...string) (string, error) {
	backupMutex.Lock()
	defer backupMutex.Unlock()

	backupDir := os.Getenv("HERASAT_BACKUP_DIR")
	if strings.TrimSpace(backupDir) == "" {
		backupDir = "./backups"
	}

	if err := os.MkdirAll(backupDir, 0700); err != nil {
		_ = logAuditWithDetail("SYSTEM", "BackupEngine", "BACKUP_FAILED", nil, "127.0.0.1", "MKDIR_ERROR: "+err.Error())
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_herasat_%s.hbak", timestamp)
	targetPath := filepath.Join(backupDir, filename)
	if len(customTarget) > 0 && strings.TrimSpace(customTarget[0]) != "" {
		targetPath = strings.TrimSpace(customTarget[0])
		if dir := filepath.Dir(targetPath); dir != "" {
			_ = os.MkdirAll(dir, 0700)
		}
	}

	tmpDir, err := os.MkdirTemp("", "herasat_backup_*")
	if err != nil {
		_ = logAuditWithDetail("SYSTEM", "BackupEngine", "BACKUP_FAILED", nil, "127.0.0.1", "TMPDIR_ERROR: "+err.Error())
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	dbDumpPath := filepath.Join(tmpDir, "database.sql.gz")
	if err := dumpViaPgDumpToFile(dbDumpPath); err != nil {
		_ = logAuditWithDetail("SYSTEM", "BackupEngine", "BACKUP_FAILED", nil, "127.0.0.1", "PG_DUMP_ERROR: "+err.Error())
		return "", err
	}

	var fileList []string
	fileList = append(fileList, "database.sql.gz")

	uploadsDir := "./secure_uploads"
	var uploadFiles []string
	if entries, err := os.ReadDir(uploadsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				uploadFiles = append(uploadFiles, e.Name())
				fileList = append(fileList, "uploads/"+e.Name())
			}
		}
	}

	manifest := BackupManifest{
		Version:   "2.0",
		CreatedAt: time.Now().UTC(),
		Files:     fileList,
		ChunkSize: backupChunkSize,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
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

	manifestNonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, manifestNonce); err != nil {
		return "", err
	}
	manifestCipher := aesgcm.Seal(nil, manifestNonce, manifestBytes, []byte("MANIFEST"))

	tmpTarget := targetPath + ".tmp"
	outFile, err := os.OpenFile(tmpTarget, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		_ = logAuditWithDetail("SYSTEM", "BackupEngine", "BACKUP_FAILED", nil, "127.0.0.1", "WRITE_FILE_ERROR: "+err.Error())
		return "", err
	}
	defer func() {
		outFile.Close()
		_ = os.Remove(tmpTarget)
	}()

	if _, err := outFile.Write([]byte(backupMagicV2)); err != nil {
		return "", err
	}

	var mLenBuf [4]byte
	binary.BigEndian.PutUint32(mLenBuf[:], uint32(len(manifestCipher)))
	if _, err := outFile.Write(mLenBuf[:]); err != nil {
		return "", err
	}
	if _, err := outFile.Write(manifestNonce); err != nil {
		return "", err
	}
	if _, err := outFile.Write(manifestCipher); err != nil {
		return "", err
	}

	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)

		dbInfo, statErr := os.Stat(dbDumpPath)
		if statErr == nil {
			dbHdr := &tar.Header{
				Name:    "database.sql.gz",
				Mode:    0600,
				Size:    dbInfo.Size(),
				ModTime: dbInfo.ModTime(),
			}
			if err := tw.WriteHeader(dbHdr); err == nil {
				if f, err := os.Open(dbDumpPath); err == nil {
					_, _ = io.Copy(tw, f)
					f.Close()
				}
			}
		}

		for _, name := range uploadFiles {
			p := filepath.Join(uploadsDir, name)
			if fi, err := os.Stat(p); err == nil {
				hdr := &tar.Header{
					Name:    "uploads/" + name,
					Mode:    0600,
					Size:    fi.Size(),
					ModTime: fi.ModTime(),
				}
				if err := tw.WriteHeader(hdr); err == nil {
					if uf, err := os.Open(p); err == nil {
						_, _ = io.Copy(tw, uf)
						uf.Close()
					}
				}
			}
		}

		_ = tw.Close()
		_ = pw.Close()
	}()

	chunkBuf := make([]byte, backupChunkSize)
	chunkIndex := 0

	for {
		n, readErr := io.ReadFull(pr, chunkBuf)
		if n > 0 {
			chunkData := chunkBuf[:n]
			isFinal := byte(0x00)
			if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				isFinal = 0x01
			}

			chunkNonce := make([]byte, aesgcm.NonceSize())
			if _, err := io.ReadFull(rand.Reader, chunkNonce); err != nil {
				pr.Close()
				return "", err
			}

			aad := []byte(fmt.Sprintf("CHUNK:%d", chunkIndex))
			encryptedChunk := aesgcm.Seal(nil, chunkNonce, chunkData, aad)

			var headerBuf [5]byte
			binary.BigEndian.PutUint32(headerBuf[0:4], uint32(len(encryptedChunk)))
			headerBuf[4] = isFinal

			if _, err := outFile.Write(headerBuf[:]); err != nil {
				pr.Close()
				return "", err
			}
			if _, err := outFile.Write(chunkNonce); err != nil {
				pr.Close()
				return "", err
			}
			if _, err := outFile.Write(encryptedChunk); err != nil {
				pr.Close()
				return "", err
			}

			chunkIndex++
		}

		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
		if readErr != nil {
			pr.Close()
			return "", readErr
		}
	}

	if err := outFile.Close(); err != nil {
		return "", err
	}

	if err := os.Rename(tmpTarget, targetPath); err != nil {
		return "", err
	}

	retentionDays := 14
	if envRet := os.Getenv("HERASAT_BACKUP_RETENTION_DAYS"); envRet != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(envRet)); err == nil && parsed > 0 {
			retentionDays = parsed
		}
	}
	rotateOldBackups(backupDir, retentionDays)

	_ = logAuditWithDetail("SYSTEM", "BackupEngine", "BACKUP_COMPLETED", nil, "127.0.0.1", filename)
	return targetPath, nil
}

func RestoreBackup(backupFilePath string, db *sql.DB) error {
	f, err := os.Open(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer f.Close()

	magicBuf := make([]byte, 5)
	if _, err := io.ReadFull(f, magicBuf); err != nil {
		return fmt.Errorf("corrupted backup header: %w", err)
	}

	if string(magicBuf) == backupMagicV1 {
		return restoreLegacyV1Backup(f, db)
	}

	if string(magicBuf) != backupMagicV2 {
		return errors.New("unrecognized backup format or corrupted header")
	}

	key := GetBackupKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	var mLenBuf [4]byte
	if _, err := io.ReadFull(f, mLenBuf[:]); err != nil {
		return fmt.Errorf("failed reading manifest length: %w", err)
	}
	mLen := binary.BigEndian.Uint32(mLenBuf[:])
	if mLen > 10<<20 {
		return errors.New("manifest payload exceeds safety threshold")
	}

	mNonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(f, mNonce); err != nil {
		return fmt.Errorf("failed reading manifest nonce: %w", err)
	}

	mCipher := make([]byte, mLen)
	if _, err := io.ReadFull(f, mCipher); err != nil {
		return fmt.Errorf("failed reading manifest ciphertext: %w", err)
	}

	mPlain, err := aesgcm.Open(nil, mNonce, mCipher, []byte("MANIFEST"))
	if err != nil {
		return fmt.Errorf("manifest integrity check failed: %w", err)
	}

	var manifest BackupManifest
	if err := json.Unmarshal(mPlain, &manifest); err != nil {
		return fmt.Errorf("corrupted manifest structure: %w", err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		chunkIndex := 0

		for {
			var hBuf [5]byte
			if _, err := io.ReadFull(f, hBuf[:]); err != nil {
				if err == io.EOF {
					break
				}
				pw.CloseWithError(err)
				return
			}

			payloadLen := binary.BigEndian.Uint32(hBuf[0:4])
			isFinal := hBuf[4]
			if payloadLen > 10<<20 {
				pw.CloseWithError(errors.New("chunk payload exceeds safety limit"))
				return
			}

			cNonce := make([]byte, aesgcm.NonceSize())
			if _, err := io.ReadFull(f, cNonce); err != nil {
				pw.CloseWithError(err)
				return
			}

			cPayload := make([]byte, payloadLen)
			if _, err := io.ReadFull(f, cPayload); err != nil {
				pw.CloseWithError(err)
				return
			}

			aad := []byte(fmt.Sprintf("CHUNK:%d", chunkIndex))
			chunkPlain, err := aesgcm.Open(nil, cNonce, cPayload, aad)
			if err != nil {
				pw.CloseWithError(fmt.Errorf("chunk %d authentication failed: %w", chunkIndex, ErrAuthenticationFailed))
				return
			}

			if _, err := pw.Write(chunkPlain); err != nil {
				return
			}

			chunkIndex++
			if isFinal == 0x01 {
				break
			}
		}
	}()

	tr := tar.NewReader(pr)
	absUploadDir, _ := filepath.Abs("./secure_uploads")
	_ = os.MkdirAll(absUploadDir, 0700)

	var extractedSQLPath string
	tmpDir, err := os.MkdirTemp("", "herasat_restore_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("archive stream read error: %w", err)
		}

		if hdr.Name == "database.sql.gz" {
			if hdr.Typeflag != tar.TypeReg {
				continue
			}
			extractedSQLPath = filepath.Join(tmpDir, "database.sql.gz")
			sf, err := os.OpenFile(extractedSQLPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
			if err != nil {
				return err
			}
			if _, err := io.Copy(sf, io.LimitReader(tr, 500<<20)); err != nil {
				sf.Close()
				return err
			}
			sf.Close()
		} else if strings.HasPrefix(hdr.Name, "uploads/") {
			if hdr.Typeflag != tar.TypeReg {
				continue
			}
			relName := strings.TrimPrefix(hdr.Name, "uploads/")
			targetPath := filepath.Join(absUploadDir, filepath.Clean(relName))
			if isWithinDirectory(absUploadDir, targetPath) {
				if fi, err := os.Lstat(targetPath); err == nil && (fi.Mode()&os.ModeSymlink != 0) {
					_ = os.Remove(targetPath)
				}
				_ = os.MkdirAll(filepath.Dir(targetPath), 0700)
				f, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
				if err == nil {
					_, _ = io.Copy(f, io.LimitReader(tr, 50<<20))
					f.Close()
				}
			}
		}
	}

	if extractedSQLPath != "" {
		if err := executeRestoredSQLGz(extractedSQLPath, db); err != nil {
			return err
		}
	}

	return nil
}

func executeRestoredSQLGz(sqlGzPath string, db *sql.DB) error {
	f, err := os.Open(sqlGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	psqlPath := getPsqlPath()
	dockerContainer := strings.TrimSpace(os.Getenv("HERASAT_DOCKER_CONTAINER"))
	connStr := os.Getenv("HERASAT_DB_CONN")

	if (psqlPath != "" || dockerContainer != "") && connStr != "" {
		safeConn, password := parseDBConn(connStr)
		var cmd *exec.Cmd
		if psqlPath != "" {
			cmd = exec.Command(psqlPath, safeConn)
			if password != "" {
				cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
			}
		} else {
			cmd = exec.Command("docker", "exec", "-i", "-e", "PGPASSWORD", dockerContainer, "psql", safeConn)
			if password != "" {
				cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
			}
		}
		cmd.Stdin = gzr
		var stderr strings.Builder
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("psql restore execution failed: %v - %s", err, stderr.String())
		}
		return nil
	}

	if db != nil {
		const maxMemSQLBytes = 50 << 20
		lr := io.LimitReader(gzr, maxMemSQLBytes+1)
		b, err := io.ReadAll(lr)
		if err != nil {
			return err
		}
		if int64(len(b)) > maxMemSQLBytes {
			return errors.New("restored SQL script exceeds maximum memory buffer limit (50MB); configure psql or docker utilities")
		}
		if len(b) > 0 {
			if _, err := db.Exec(string(b)); err != nil {
				return fmt.Errorf("failed executing restored SQL: %w", err)
			}
		}
	}

	return nil
}

func restoreLegacyV1Backup(f *os.File, db *sql.DB) error {
	const maxLegacyBackupSize = 100 << 20
	lr := io.LimitReader(f, maxLegacyBackupSize+1)
	data, err := io.ReadAll(lr)
	if err != nil {
		return err
	}
	if int64(len(data)) > maxLegacyBackupSize {
		return errors.New("legacy backup exceeds maximum supported payload limit (100MB)")
	}
	if len(data) < 12+16 {
		return errors.New("corrupted legacy backup: insufficient length")
	}

	nonce := data[:12]
	ciphertext := data[12:]

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
		return fmt.Errorf("backup decryption failed: authentication tag mismatch: %w", err)
	}

	gzr, err := gzip.NewReader(bytes.NewReader(decrypted))
	if err != nil {
		return fmt.Errorf("failed to initialize gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	absUploadDir, _ := filepath.Abs("./secure_uploads")

	var sqlContent string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if hdr.Name == "database.sql" {
			if b, err := io.ReadAll(tr); err == nil {
				sqlContent = string(b)
			}
		} else if strings.HasPrefix(hdr.Name, "uploads/") {
			if hdr.Typeflag != tar.TypeReg {
				continue
			}
			relName := strings.TrimPrefix(hdr.Name, "uploads/")
			targetPath := filepath.Join(absUploadDir, filepath.Clean(relName))
			if isWithinDirectory(absUploadDir, targetPath) {
				if fi, err := os.Lstat(targetPath); err == nil && (fi.Mode()&os.ModeSymlink != 0) {
					_ = os.Remove(targetPath)
				}
				_ = os.MkdirAll(filepath.Dir(targetPath), 0700)
				if outF, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); err == nil {
					_, _ = io.Copy(outF, io.LimitReader(tr, 50<<20))
					outF.Close()
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
