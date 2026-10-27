package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetReportsPaginationAndFiltering(t *testing.T) {
	db := getTestDB(t)
	DB = db

	defer func() {
		db.Exec("DELETE FROM reports WHERE tracking_code LIKE 'PAG_TEST_%'")
		db.Exec("DELETE FROM admin_report_views WHERE admin_username = 'test_pag_admin'")
		db.Exec("DELETE FROM admin_users WHERE username = 'test_pag_admin'")
	}()

	db.Exec("DELETE FROM admin_users WHERE username = 'test_pag_admin'")
	_, err := db.Exec(`
		INSERT INTO admin_users (username, password_hash, full_name, role_title, role, is_active, created_at)
		VALUES ('test_pag_admin', 'dummy_hash', 'Pag Admin', 'مدیر ارشد', 'super_admin', true, NOW())
	`)
	if err != nil {
		t.Fatalf("failed to insert test admin user: %v", err)
	}

	categories := []string{"financial", "administrative", "security"}
	priorities := []string{"low", "normal", "high"}
	classifications := []string{"confidential", "secret", "top_secret"}

	for i := 1; i <= 15; i++ {
		trackingCode := fmt.Sprintf("PAG_TEST_%03d", i)
		cat := categories[i%3]
		prio := priorities[i%3]
		class := classifications[i%3]
		encDesc, _ := EncryptAES(fmt.Sprintf("Test description for report %d", i))

		var id int64
		err := db.QueryRow(`
			INSERT INTO reports (tracking_code, category, report_type, priority, description, classification, created_at, updated_at)
			VALUES ($1, $2, 'complaint', $3, $4, $5, NOW(), NOW())
			RETURNING id
		`, trackingCode, cat, prio, encDesc, class).Scan(&id)
		if err != nil {
			t.Fatalf("failed to insert test report %d: %v", i, err)
		}

		if i <= 5 {
			db.Exec(`
				INSERT INTO admin_report_views (report_id, admin_username, viewed_at)
				VALUES ($1, 'test_pag_admin', NOW())
			`, id)
		}
	}

	adminSession := AdminSession{
		Username: "test_pag_admin",
		Role:     "super_admin",
	}

	{
		req := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_", nil)
		req = req.WithContext(contextWithAdmin(req.Context(), adminSession))
		w := httptest.NewRecorder()

		GetReportsHandler(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		var result PaginatedReportsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result.Total != 15 {
			t.Errorf("expected total 15, got %d", result.Total)
		}
		if len(result.Reports) != 15 {
			t.Errorf("expected 15 reports, got %d", len(result.Reports))
		}
		if result.Page != 1 || result.PageSize != 25 {
			t.Errorf("expected page=1, pageSize=25; got page=%d, pageSize=%d", result.Page, result.PageSize)
		}
	}

	{
		req := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_&page=1&page_size=5", nil)
		req = req.WithContext(contextWithAdmin(req.Context(), adminSession))
		w := httptest.NewRecorder()

		GetReportsHandler(w, req)
		var result PaginatedReportsResponse
		json.NewDecoder(w.Result().Body).Decode(&result)

		if result.Total != 15 {
			t.Errorf("expected total 15, got %d", result.Total)
		}
		if len(result.Reports) != 5 {
			t.Errorf("expected 5 reports on page 1, got %d", len(result.Reports))
		}
		if result.Page != 1 || result.PageSize != 5 {
			t.Errorf("expected page=1, pageSize=5; got page=%d, pageSize=%d", result.Page, result.PageSize)
		}

		req2 := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_&page=2&page_size=5", nil)
		req2 = req2.WithContext(contextWithAdmin(req2.Context(), adminSession))
		w2 := httptest.NewRecorder()

		GetReportsHandler(w2, req2)
		var result2 PaginatedReportsResponse
		json.NewDecoder(w2.Result().Body).Decode(&result2)

		if len(result2.Reports) != 5 {
			t.Errorf("expected 5 reports on page 2, got %d", len(result2.Reports))
		}
		if result2.Page != 2 {
			t.Errorf("expected page=2; got page=%d", result2.Page)
		}

		if result.Reports[0].ID == result2.Reports[0].ID {
			t.Errorf("expected disjoint pages, but report IDs match: %d", result.Reports[0].ID)
		}
	}

	{
		req := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_&category=financial", nil)
		req = req.WithContext(contextWithAdmin(req.Context(), adminSession))
		w := httptest.NewRecorder()

		GetReportsHandler(w, req)
		var result PaginatedReportsResponse
		json.NewDecoder(w.Result().Body).Decode(&result)

		if result.Total != 5 {
			t.Errorf("expected 5 financial reports, got %d", result.Total)
		}
		for _, rep := range result.Reports {
			if rep.Category != "financial" {
				t.Errorf("expected financial category, got %s", rep.Category)
			}
		}
	}

	{
		reqUnread := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_&read_status=unread", nil)
		reqUnread = reqUnread.WithContext(contextWithAdmin(reqUnread.Context(), adminSession))
		wUnread := httptest.NewRecorder()

		GetReportsHandler(wUnread, reqUnread)
		var resUnread PaginatedReportsResponse
		json.NewDecoder(wUnread.Result().Body).Decode(&resUnread)

		if resUnread.Total != 10 {
			t.Errorf("expected 10 unread reports, got %d", resUnread.Total)
		}
		for _, rep := range resUnread.Reports {
			if rep.IsRead {
				t.Errorf("expected unread report, but is_read was true for ID %d", rep.ID)
			}
		}

		reqRead := httptest.NewRequest("GET", "/api/admin/reports?q=PAG_TEST_&read_status=read", nil)
		reqRead = reqRead.WithContext(contextWithAdmin(reqRead.Context(), adminSession))
		wRead := httptest.NewRecorder()

		GetReportsHandler(wRead, reqRead)
		var resRead PaginatedReportsResponse
		json.NewDecoder(wRead.Result().Body).Decode(&resRead)

		if resRead.Total != 5 {
			t.Errorf("expected 5 read reports, got %d", resRead.Total)
		}
		for _, rep := range resRead.Reports {
			if !rep.IsRead {
				t.Errorf("expected read report, but is_read was false for ID %d", rep.ID)
			}
		}
	}
}
