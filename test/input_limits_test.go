package main

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestInputSanitizationAndLimits(t *testing.T) {
	db := getTestDB(t)
	DB = db

	// Helper for generating CAPTCHA
	token := GenerateCaptchaToken(15, 2*time.Minute)
	answer := "15"

	// 1. SubmitReportHandler: description > 5000 runes returns 400
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("captcha_id", token)
		_ = writer.WriteField("captcha_answer", answer)
		_ = writer.WriteField("category", "financial")
		_ = writer.WriteField("description", strings.Repeat("گزارش تستی ", 10000)) // ~110,000 runes
		_ = writer.Close()

		req := httptest.NewRequest("POST", "/api/report", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		SubmitReportHandler(w, req)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for 100k runes description, got %d", w.Result().StatusCode)
		}
	}

	// 2. SubmitReportHandler: name > 100 runes returns 400
	{
		newToken := GenerateCaptchaToken(22, 2*time.Minute)
		newAns := "22"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("captcha_id", newToken)
		_ = writer.WriteField("captcha_answer", newAns)
		_ = writer.WriteField("category", "financial")
		_ = writer.WriteField("description", strings.Repeat("شرح گزارش معتبر برای تست سیستم حراست استان همدان. ", 2))
		_ = writer.WriteField("reporter_name", strings.Repeat("نام کاربری بسیار طولانی ", 10)) // > 100 runes
		_ = writer.Close()

		req := httptest.NewRequest("POST", "/api/report", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		SubmitReportHandler(w, req)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for name > 100 runes, got %d", w.Result().StatusCode)
		}
	}

	// 3. UpdateInternalNotesHandler: notes > 4000 runes returns 400
	{
		adminSession := AdminSession{
			Username: "admin_test_notes",
			Role:     "super_admin",
		}
		jsonBody := `{"report_id": 1, "internal_notes": "` + strings.Repeat("یادداشت محرمانه ", 500) + `"}`
		req := httptest.NewRequest("POST", "/api/admin/report/internal-notes", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "admin", adminSession))
		w := httptest.NewRecorder()

		UpdateInternalNotesHandler(w, req)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for notes > 4000 runes, got %d", w.Result().StatusCode)
		}
	}

	// 4. RespondToReportHandler: response > 3000 runes returns 400
	{
		adminSession := AdminSession{
			Username: "admin_test_resp",
			Role:     "super_admin",
		}
		jsonBody := `{"report_id": 1, "response": "` + strings.Repeat("پاسخ اداری رسمی ", 400) + `"}`
		req := httptest.NewRequest("POST", "/api/admin/respond", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "admin", adminSession))
		w := httptest.NewRecorder()

		RespondToReportHandler(w, req)
		if w.Result().StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for response > 3000 runes, got %d", w.Result().StatusCode)
		}
	}

	// 5. Unit test helper functions: truncateRunes, stripNul, sanitizePhone, isValidTrackingCode
	{
		if stripNul("hello\x00world\x00") != "helloworld" {
			t.Errorf("stripNul failed")
		}
		if sanitizePhone("+98 (918) 123-4567abc") != "+98 (918) 123-4567" {
			t.Errorf("sanitizePhone failed")
		}
		if !isValidTrackingCode("ABC123XYZ") || isValidTrackingCode("abc123") || isValidTrackingCode("ABC 123") || isValidTrackingCode("") {
			t.Errorf("isValidTrackingCode failed")
		}
		if truncateRunes("سلام دوستان", 4) != "سلام" {
			t.Errorf("truncateRunes failed: got %q, want %q", truncateRunes("سلام دوستان", 4), "سلام")
		}
	}
}
