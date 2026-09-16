package main

import (
	"testing"
	"time"
)

func TestReportsUpdatedAtTrigger(t *testing.T) {
	db := getTestDB(t)
	DB = db

	// Ensure trigger migration has run
	runMigrations()

	var reportID int64
	initialTime := time.Now().Add(-1 * time.Hour)

	err := db.QueryRow(`
		INSERT INTO reports (tracking_code, category, report_type, priority, description, created_at, updated_at)
		VALUES ('TRG_TEST_001', 'security', 'complaint', 'normal', 'trigger test', $1, $1)
		RETURNING id
	`, initialTime).Scan(&reportID)
	if err != nil {
		t.Fatalf("failed to insert test report: %v", err)
	}
	defer func() {
		db.Exec("DELETE FROM reports WHERE id = $1", reportID)
	}()

	// Query updated_at right after insert
	var createdAt, updatedAtBefore time.Time
	err = db.QueryRow(`SELECT created_at, updated_at FROM reports WHERE id = $1`, reportID).Scan(&createdAt, &updatedAtBefore)
	if err != nil {
		t.Fatalf("failed to query initial timestamps: %v", err)
	}

	// Wait 50ms so time advances noticeably
	time.Sleep(50 * time.Millisecond)

	// Perform update without explicitly setting updated_at
	_, err = db.Exec(`UPDATE reports SET priority = 'urgent' WHERE id = $1`, reportID)
	if err != nil {
		t.Fatalf("failed to update report priority: %v", err)
	}

	// Query updated_at after update
	var updatedAtAfter time.Time
	err = db.QueryRow(`SELECT updated_at FROM reports WHERE id = $1`, reportID).Scan(&updatedAtAfter)
	if err != nil {
		t.Fatalf("failed to query updated_at after update: %v", err)
	}

	if !updatedAtAfter.After(updatedAtBefore) {
		t.Fatalf("expected updated_at (%v) to be strictly after previous updated_at (%v)", updatedAtAfter, updatedAtBefore)
	}
}
