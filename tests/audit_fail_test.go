package main

import (
	"bytes"
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

	var reportID int64
	err := db.QueryRow(`
		INSERT INTO reports (tracking_code, category, report_type, priority, description, created_at, updated_at)
		VALUES ('TEST-AUDIT-FAIL-1', 'administrative', 'complaint', 'normal', 'test desc', NOW(), NOW())
		RETURNING id
	`).Scan(&reportID)
	if err != nil {
		t.Fatalf("Failed to seed test report: %v", err)
	}

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

	adminSession := AdminSession{
		Username:    "superuser_test",
		FullName:    "Super User",
		Title:       "Support",
		Role:        "superuser",
		IsSuperuser: true,
	}
	ctx := contextWithAdmin(req.Context(), adminSession)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	DeleteReportHandler(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("Expected HTTP 503 Service Unavailable when audit log fails, got %d: %s", rr.Code, rr.Body.String())
	}

	_, err = db.Exec("ALTER TABLE audit_logs_temp_disabled RENAME TO audit_logs")
	if err != nil {
		t.Fatalf("Failed to restore audit_logs table: %v", err)
	}

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM reports WHERE id = $1)", reportID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to query reports: %v", err)
	}
	if !exists {
		t.Fatalf("Report was deleted despite audit failure! Expected rollback.")
	}

	_, _ = db.Exec("DELETE FROM reports WHERE id = $1", reportID)
}
