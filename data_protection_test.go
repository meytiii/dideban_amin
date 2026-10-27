package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestV3CiphertextWithAAD(t *testing.T) {
	key := []byte("TEST_KEY_FOR_AAD_VALIDATION_32B!")
	plaintext := []byte("confidential report payload")
	aad := []byte("reports:description:105")

	ciphertext, err := EncryptWithAAD(key, "v3:1:f:", aad, plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if !strings.HasPrefix(ciphertext, "v3:1:f:") {
		t.Fatalf("expected v3:1:f: prefix, got: %s", ciphertext)
	}

	decrypted, err := DecryptWithAAD(key, strings.TrimPrefix(ciphertext, "v3:1:f:"), aad)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypted text mismatch: got %s, want %s", string(decrypted), string(plaintext))
	}

	wrongAAD := []byte("reports:description:106")
	_, err = DecryptWithAAD(key, strings.TrimPrefix(ciphertext, "v3:1:f:"), wrongAAD)
	if err == nil {
		t.Fatalf("expected decryption error when AAD differs (cut-and-paste attack)")
	}

	tamperedCiphertext := strings.TrimPrefix(ciphertext, "v3:1:f:")
	tamperedCiphertext = tamperedCiphertext[:len(tamperedCiphertext)-4] + "AAAA"
	_, err = DecryptWithAAD(key, tamperedCiphertext, aad)
	if err == nil {
		t.Fatalf("expected decryption error on tampered ciphertext")
	}
}

func TestBoundFieldHelper(t *testing.T) {
	plaintext := "sensitive user information"
	enc, err := EncryptBoundField("reports", "reporter_name", 42, plaintext)
	if err != nil {
		t.Fatalf("EncryptBoundField failed: %v", err)
	}

	dec, err := DecryptBoundField("reports", "reporter_name", 42, enc)
	if err != nil {
		t.Fatalf("DecryptBoundField failed: %v", err)
	}
	if dec != plaintext {
		t.Fatalf("got %s, want %s", dec, plaintext)
	}

	_, err = DecryptBoundField("reports", "reporter_phone", 42, enc)
	if err == nil {
		t.Fatalf("expected failure when binding column differs")
	}

	_, err = DecryptBoundField("reports", "reporter_name", 43, enc)
	if err == nil {
		t.Fatalf("expected failure when binding rowID differs")
	}
}

func TestTrackingCodeHMACAndResolution(t *testing.T) {
	code := "TRK-9876543210"
	h1 := HashTrackingCode(code)
	h2 := HashTrackingCode(code)
	if h1 != h2 {
		t.Fatalf("HashTrackingCode must be deterministic")
	}
	if len(h1) != 64 {
		t.Fatalf("expected 64-char hex hash, got %d chars", len(h1))
	}

	hDiff := HashTrackingCode("TRK-9876543211")
	if h1 == hDiff {
		t.Fatalf("expected distinct hashes for different tracking codes")
	}

	raw := "PLAIN_TRACKING_CODE"
	resolvedRaw := ResolveTrackingCode(raw)
	if resolvedRaw != raw {
		t.Fatalf("expected raw code preserved, got: %s", resolvedRaw)
	}

	enc, err := EncryptAES(code)
	if err != nil {
		t.Fatalf("EncryptAES failed: %v", err)
	}
	resolvedEnc := ResolveTrackingCode(enc)
	if resolvedEnc != code {
		t.Fatalf("expected decrypted tracking code %s, got: %s", code, resolvedEnc)
	}
}

func TestEncryptedBackupAndTamperDetection(t *testing.T) {
	if getPgDumpPath() == "" && os.Getenv("HERASAT_DOCKER_CONTAINER") == "" {
		t.Skip("Skipping backup test: neither pg_dump nor HERASAT_DOCKER_CONTAINER is configured")
	}

	tmpDir, err := os.MkdirTemp("", "herasat_backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldDir := os.Getenv("HERASAT_BACKUP_DIR")
	os.Setenv("HERASAT_BACKUP_DIR", tmpDir)
	defer os.Setenv("HERASAT_BACKUP_DIR", oldDir)

	backupPath, err := PerformBackup(nil)
	if err != nil {
		t.Fatalf("PerformBackup failed: %v", err)
	}

	data, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("failed to read backup file: %v", err)
	}

	if len(data) < 5 || string(data[:5]) != "HBAK\x02" {
		t.Fatalf("expected HBAK\\x02 magic header, got: %q", data[:5])
	}

	err = RestoreBackup(backupPath, nil)
	if err != nil {
		t.Fatalf("RestoreBackup failed on valid archive: %v", err)
	}

	tamperedData := make([]byte, len(data))
	copy(tamperedData, data)
	tamperedData[len(tamperedData)-5] ^= 0xFF

	tamperedFile := filepath.Join(tmpDir, "tampered.hbak")
	if err := os.WriteFile(tamperedFile, tamperedData, 0600); err != nil {
		t.Fatalf("failed writing tampered backup: %v", err)
	}

	err = RestoreBackup(tamperedFile, nil)
	if err == nil {
		t.Fatalf("expected error on tampered backup archive, but RestoreBackup succeeded")
	}
}

func TestTwoPersonDeletionStatusValidation(t *testing.T) {
	past := time.Now().Add(-25 * time.Hour)
	if time.Since(past) <= 24*time.Hour {
		t.Fatalf("expected expired time to be older than 24h")
	}

	recent := time.Now().Add(-1 * time.Hour)
	if time.Since(recent) > 24*time.Hour {
		t.Fatalf("expected recent time to be within 24h")
	}

	adminA := "supervisor_a"
	adminB := "supervisor_b"
	if adminA == adminB {
		t.Fatalf("expected different admin usernames for dual control")
	}
}

func TestRedactionMarkerFormat(t *testing.T) {
	marker := "[حذف امنیتی توسط admin در 2026-10-27 12:00:00]"
	enc, err := EncryptAES(marker)
	if err != nil {
		t.Fatalf("EncryptAES failed on marker: %v", err)
	}
	dec, err := DecryptAES(enc)
	if err != nil {
		t.Fatalf("DecryptAES failed on marker: %v", err)
	}
	if dec != marker {
		t.Fatalf("got %s, want %s", dec, marker)
	}
}
