package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuditFailClosed(t *testing.T) {
	db := getTestDB(t)
	oldDB := DB
	DB = db
	defer func() { DB = oldDB }()

	// Seed test report
	var reportID int64
	err := db.QueryRow(`
		INSERT INTO reports (tracking_code, category, report_type, priority, description, created_at, updated_at)
		VALUES ('TEST-AUDIT-FAIL-1', 'administrative', 'complaint', 'normal', 'test desc', NOW(), NOW())
		RETURNING id
	`).Scan(&reportID)
	if err != nil {
		t.Fatalf("Failed to seed test report: %v", err)
	}

	// Rename audit_logs table to simulate failure
	_, err = db.Exec("ALTER TABLE audit_logs RENAME TO audit_logs_temp_disabled")
	if err != nil {
		t.Fatalf("Failed to rename audit_logs: %v", err)
	}
	defer func() {
		_, _ = db.Exec("ALTER TABLE audit_logs_temp_disabled RENAME TO audit_logs")
	}()

	body, _ := json.Marshal(map[string]interface{}{
		"report_id": reportID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/report/delete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Superuser session context
	adminSession := AdminSession{
		Username:    "superuser_test",
		FullName:    "Super User",
		Title:       "Support",
		Role:        "superuser",
		IsSuperuser: true,
	}
	ctx := context.WithValue(req.Context(), "admin", adminSession)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	DeleteReportHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected HTTP 503 Service Unavailable when audit log fails, got %d: %s", rr.Code, rr.Body.String())
	}

	// Restore audit_logs table before verifying report wasn't deleted
	_, err = db.Exec("ALTER TABLE audit_logs_temp_disabled RENAME TO audit_logs")
	if err != nil {
		t.Fatalf("Failed to restore audit_logs table: %v", err)
	}

	// Assert report was NOT deleted due to transaction rollback
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM reports WHERE id = $1)", reportID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to query reports: %v", err)
	}
	if !exists {
		t.Fatalf("Report was deleted despite audit failure! Expected rollback.")
	}

	// Cleanup
	_, _ = db.Exec("DELETE FROM reports WHERE id = $1", reportID)
}
