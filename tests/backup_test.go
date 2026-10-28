package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBackupRotation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "herasat_backup_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldFile := filepath.Join(tempDir, "backup_herasat_20200101_120000.sql.gz")
	if err := os.WriteFile(oldFile, []byte("old backup data"), 0600); err != nil {
		t.Fatalf("failed to write old file: %v", err)
	}
	oldTime := time.Now().AddDate(0, 0, -20)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	recentFile := filepath.Join(tempDir, "backup_herasat_20261026_120000.sql.gz")
	if err := os.WriteFile(recentFile, []byte("recent backup data"), 0600); err != nil {
		t.Fatalf("failed to write recent file: %v", err)
	}

	otherFile := filepath.Join(tempDir, "unrelated_file.txt")
	if err := os.WriteFile(otherFile, []byte("preserve me"), 0600); err != nil {
		t.Fatalf("failed to write other file: %v", err)
	}
	if err := os.Chtimes(otherFile, oldTime, oldTime); err != nil {
		t.Fatalf("failed to set mtime: %v", err)
	}

	rotateOldBackups(tempDir, 14)

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Errorf("expected old backup file to be removed, but it exists")
	}

	if _, err := os.Stat(recentFile); err != nil {
		t.Errorf("expected recent backup file to remain, but got error: %v", err)
	}

	if _, err := os.Stat(otherFile); err != nil {
		t.Errorf("expected non-backup file to remain, but got error: %v", err)
	}
}
