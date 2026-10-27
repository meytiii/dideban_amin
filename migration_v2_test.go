package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReencryptLegacyData(t *testing.T) {
	db := getTestDB(t)

	legacyMasterKey := "12345678901234567890123456789012"
	InitCrypto(legacyMasterKey)

	block, _ := aes.NewCipher(aesKey)
	aesgcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, aesgcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	plainDesc := "Legacy report description before key separation"
	ciphertext := aesgcm.Seal(nonce, nonce, []byte(plainDesc), nil)
	legacyV1Field := "v1:" + base64.StdEncoding.EncodeToString(ciphertext)

	trackingCode := fmt.Sprintf("TEST-MIG-V2-%d", time.Now().UnixNano())
	var reportID int64
	err := db.QueryRow(`
		INSERT INTO reports (tracking_code, category, report_type, priority, description, created_at, updated_at)
		VALUES ($1, 'administrative', 'complaint', 'normal', $2, NOW(), NOW())
		RETURNING id
	`, trackingCode, legacyV1Field).Scan(&reportID)
	if err != nil {
		t.Fatalf("Failed to insert test report: %v", err)
	}

	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "legacy_evidence.bin")
	rawFileBytes := []byte("Legacy evidence file content HENC01")
	header := []byte("HENC\x01")
	out := append(header, nonce...)
	legacyBlob := aesgcm.Seal(out, nonce, rawFileBytes, nil)
	if err := os.WriteFile(testFilePath, legacyBlob, 0600); err != nil {
		t.Fatalf("Failed to write test evidence file: %v", err)
	}

	legacyPathEnc := "v1:" + base64.StdEncoding.EncodeToString(aesgcm.Seal(nonce, nonce, []byte(testFilePath), nil))
	_, err = db.Exec(`
		INSERT INTO evidence_files (report_id, file_path, file_hash, mime_type, file_name, file_size)
		VALUES ($1, $2, 'dummyhash', 'application/octet-stream', 'test.bin', $3)
	`, reportID, legacyPathEnc, int64(len(rawFileBytes)))
	if err != nil {
		t.Fatalf("Failed to insert evidence_files row: %v", err)
	}

	if err := reencryptLegacyData(db); err != nil {
		t.Fatalf("reencryptLegacyData failed: %v", err)
	}

	var migratedDesc string
	err = db.QueryRow(`SELECT description FROM reports WHERE id = $1`, reportID).Scan(&migratedDesc)
	if err != nil {
		t.Fatalf("Failed to query migrated report: %v", err)
	}
	if !strings.HasPrefix(migratedDesc, "v2:f:") && !strings.HasPrefix(migratedDesc, "v3:") {
		t.Fatalf("Expected migrated description to start with v2:f: or v3:, got %s", migratedDesc)
	}
	decDesc, err := DecryptAES(migratedDesc)
	if err != nil || decDesc != plainDesc {
		t.Fatalf("Decryption failed on migrated description: %v (got %q)", err, decDesc)
	}

	migratedDiskBytes, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read migrated disk file: %v", err)
	}
	if !strings.HasPrefix(string(migratedDiskBytes), "HENC\x02") && !strings.HasPrefix(string(migratedDiskBytes), "HENC\x03") {
		t.Fatalf("Expected migrated file on disk to have HENC\\x02 or HENC\\x03 header")
	}
	decDiskBytes, err := DecryptFileBytes(migratedDiskBytes)
	if err != nil || string(decDiskBytes) != string(rawFileBytes) {
		t.Fatalf("Decrypted file blob mismatch: %v (got %q)", err, string(decDiskBytes))
	}
}
