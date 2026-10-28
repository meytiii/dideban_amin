package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSessionDeadlockAndPermutations(t *testing.T) {
	sessionManager := NewSessionManager(nil)
	Sessions = sessionManager

	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	wrapped := AdminAuthMiddleware(baseHandler)

	tests := []struct {
		name        string
		reqPwd      bool
		req2FA      bool
		path        string
		method      string
		expectCode  int
		expectError string
	}{
		{"both_allow_me", true, true, "/api/admin/me", "GET", http.StatusOK, ""},
		{"both_allow_logout", true, true, "/api/admin/logout", "POST", http.StatusOK, ""},
		{"both_allow_change_pwd", true, true, "/api/admin/change-my-password", "POST", http.StatusOK, ""},
		{"both_allow_2fa_setup", true, true, "/api/admin/2fa/setup", "POST", http.StatusOK, ""},
		{"both_allow_2fa_enable", true, true, "/api/admin/2fa/enable", "POST", http.StatusOK, ""},
		{"both_reject_reports", true, true, "/api/admin/reports", "GET", http.StatusForbidden, "password_change_required"},

		{"pwd_only_allow_me", true, false, "/api/admin/me", "GET", http.StatusOK, ""},
		{"pwd_only_allow_logout", true, false, "/api/admin/logout", "POST", http.StatusOK, ""},
		{"pwd_only_allow_change_pwd", true, false, "/api/admin/change-my-password", "POST", http.StatusOK, ""},
		{"pwd_only_reject_2fa_setup", true, false, "/api/admin/2fa/setup", "POST", http.StatusForbidden, "password_change_required"},
		{"pwd_only_reject_reports", true, false, "/api/admin/reports", "GET", http.StatusForbidden, "password_change_required"},

		{"totp_only_allow_me", false, true, "/api/admin/me", "GET", http.StatusOK, ""},
		{"totp_only_allow_logout", false, true, "/api/admin/logout", "POST", http.StatusOK, ""},
		{"totp_only_allow_2fa_setup", false, true, "/api/admin/2fa/setup", "POST", http.StatusOK, ""},
		{"totp_only_allow_2fa_enable", false, true, "/api/admin/2fa/enable", "POST", http.StatusOK, ""},
		{"totp_only_reject_change_pwd", false, true, "/api/admin/change-my-password", "POST", http.StatusForbidden, "2fa_setup_required"},
		{"totp_only_reject_reports", false, true, "/api/admin/reports", "GET", http.StatusForbidden, "2fa_setup_required"},

		{"clean_allow_reports", false, false, "/api/admin/reports", "GET", http.StatusOK, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := fmt.Sprintf("token-%s-%d", tc.name, time.Now().UnixNano())
			sessionManager.Set(token, AdminSession{
				Token:                  token,
				Username:               "testuser",
				Role:                   "edare_kol",
				RequiresPasswordChange: tc.reqPwd,
				Requires2FASetup:       tc.req2FA,
				CreatedAt:              time.Now(),
				LastActive:             time.Now(),
				Expiry:                 time.Now().Add(15 * time.Minute),
			})

			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: token})
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != tc.expectCode {
				t.Fatalf("path %s expected status %d, got %d (body: %s)", tc.path, tc.expectCode, rr.Code, rr.Body.String())
			}

			if tc.expectError != "" {
				var resp map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to decode response JSON: %v", err)
				}
				if resp["error"] != tc.expectError {
					t.Fatalf("expected error '%s', got '%s'", tc.expectError, resp["error"])
				}
			}
		})
	}
}

func TestChangeMyPasswordPreservesRequires2FA(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	DB = db

	username := fmt.Sprintf("pw_user_%d", time.Now().UnixNano())
	pwdPlain := "OldValidPass1234!@#"
	pwdNew := "NewValidPass5678!@#"

	hash, err := HashPassword(pwdPlain)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	var userID int64
	err = db.QueryRow(`
		INSERT INTO admin_users (username, password_hash, full_name, role_title, role, is_active, must_change_password, totp_enforced, created_at, password_changed_at)
		VALUES ($1, $2, 'Test User', 'Title', 'superadmin', true, true, true, NOW(), NOW() - INTERVAL '2 days')
		RETURNING id`, username, hash).Scan(&userID)
	if err != nil {
		t.Fatalf("insert admin user failed: %v", err)
	}
	defer func() {
		_, _ = db.Exec("DELETE FROM admin_password_history WHERE user_id = $1", userID)
		_, _ = db.Exec("DELETE FROM admin_users WHERE id = $1", userID)
	}()

	sessionToken := fmt.Sprintf("sess_pw_%d", time.Now().UnixNano())
	sessionManager := NewSessionManager(db)
	Sessions = sessionManager

	sessionManager.Set(sessionToken, AdminSession{
		Token:                  sessionToken,
		Username:               username,
		FullName:               "Test User",
		Title:                  "Title",
		Role:                   "superadmin",
		IsSuperuser:            true,
		RequiresPasswordChange: true,
		Requires2FASetup:       true,
		CreatedAt:              time.Now(),
		LastActive:             time.Now(),
		Expiry:                 time.Now().Add(15 * time.Minute),
	})

	bodyJSON := fmt.Sprintf(`{"current_password":"%s","new_password":"%s"}`, pwdPlain, pwdNew)
	req := httptest.NewRequest("POST", "/api/admin/change-my-password", bytes.NewBufferString(bodyJSON))
	req.AddCookie(&http.Cookie{Name: "herasat_admin_session", Value: sessionToken})
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	wrapped := AdminAuthMiddleware(http.HandlerFunc(ChangeMyPasswordHandler))
	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from ChangeMyPasswordHandler, got %d: %s", rr.Code, rr.Body.String())
	}

	cookies := rr.Result().Cookies()
	var newSessionToken string
	for _, c := range cookies {
		if c.Name == "herasat_admin_session" {
			newSessionToken = c.Value
		}
	}
	if newSessionToken == "" {
		t.Fatalf("expected new session cookie to be set")
	}

	newSession, exists := sessionManager.Get(newSessionToken)
	if !exists {
		t.Fatalf("new session not found in session store")
	}

	if newSession.RequiresPasswordChange {
		t.Fatalf("expected RequiresPasswordChange to be false after password change")
	}

	if !newSession.Requires2FASetup {
		t.Fatalf("expected Requires2FASetup to be preserved as true across password change")
	}
}

func TestTokenAwarePDFScanner(t *testing.T) {
	cleanPDF, err := os.ReadFile("testdata/clean_photo.pdf")
	if err != nil {
		t.Fatalf("failed to read testdata/clean_photo.pdf: %v", err)
	}
	if err := ScanPDFForActiveContent(cleanPDF); err != nil {
		t.Fatalf("expected clean 5MB photo-heavy PDF to pass, got error: %v", err)
	}

	flatePDF, err := os.ReadFile("testdata/flate_js.pdf")
	if err != nil {
		t.Fatalf("failed to read testdata/flate_js.pdf: %v", err)
	}
	if err := ScanPDFForActiveContent(flatePDF); err == nil {
		t.Fatalf("expected PDF with /JavaScript in Flate stream to fail")
	}

	hexPDF, err := os.ReadFile("testdata/hex_js.pdf")
	if err != nil {
		t.Fatalf("failed to read testdata/hex_js.pdf: %v", err)
	}
	if err := ScanPDFForActiveContent(hexPDF); err == nil {
		t.Fatalf("expected PDF with hex-escaped /J#61vaScript to fail")
	}

	encPDF, err := os.ReadFile("testdata/encrypted.pdf")
	if err != nil {
		t.Fatalf("failed to read testdata/encrypted.pdf: %v", err)
	}
	if err := ScanPDFForActiveContent(encPDF); err == nil {
		t.Fatalf("expected encrypted PDF to fail")
	}
}

func TestInspectOOXMLFixtures(t *testing.T) {
	cleanXLSX, err := os.ReadFile("testdata/clean_printer_settings.xlsx")
	if err != nil {
		t.Fatalf("failed to read testdata/clean_printer_settings.xlsx: %v", err)
	}
	if err := InspectOOXML(cleanXLSX); err != nil {
		t.Fatalf("expected clean XLSX with printerSettings1.bin to pass, got: %v", err)
	}

	macroXLSX, err := os.ReadFile("testdata/macro_vba.xlsx")
	if err != nil {
		t.Fatalf("failed to read testdata/macro_vba.xlsx: %v", err)
	}
	if err := InspectOOXML(macroXLSX); err == nil {
		t.Fatalf("expected XLSX with vbaProject.bin to fail")
	}
}

func TestAuditChainConcurrentAndTamperDetection(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	DB = db

	_, _ = db.Exec("ALTER TABLE audit_logs DISABLE TRIGGER ALL")
	_, _ = db.Exec("TRUNCATE TABLE audit_logs")
	_, _ = db.Exec("ALTER TABLE audit_logs ENABLE TRIGGER ALL")

	concurrency := 8
	iterationsPerWorker := 5
	var wg sync.WaitGroup

	errCh := make(chan error, concurrency*iterationsPerWorker)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterationsPerWorker; j++ {
				action := fmt.Sprintf("ACTION_W%d_%d", workerID, j)
				detail := fmt.Sprintf("detail payload %d-%d", workerID, j)
				if err := logAuditWithDetail(fmt.Sprintf("user_%d", workerID), "Auditor", action, nil, "127.0.0.1", detail); err != nil {
					errCh <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent audit logging failed: %v", err)
		}
	}

	res, err := VerifyAuditChainDetailed(db)
	if err != nil {
		t.Fatalf("VerifyAuditChainDetailed failed: %v", err)
	}
	expectedCount := concurrency * iterationsPerWorker
	if res.VerifiedCount != expectedCount {
		t.Fatalf("expected %d verified entries, got %d", expectedCount, res.VerifiedCount)
	}

	var targetID int64
	err = db.QueryRow("SELECT id FROM audit_logs ORDER BY id ASC OFFSET 10 LIMIT 1").Scan(&targetID)
	if err != nil {
		t.Fatalf("failed to query row to mutate: %v", err)
	}

	_, _ = db.Exec("ALTER TABLE audit_logs DISABLE TRIGGER ALL")
	_, err = db.Exec("UPDATE audit_logs SET detail = 'tampered content' WHERE id = $1", targetID)
	_, _ = db.Exec("ALTER TABLE audit_logs ENABLE TRIGGER ALL")
	if err != nil {
		t.Fatalf("failed to mutate row for tamper test: %v", err)
	}

	tamperRes, tamperErr := VerifyAuditChainDetailed(db)
	if tamperErr == nil {
		t.Fatalf("expected VerifyAuditChainDetailed to fail after row tampering")
	}
	if tamperRes.BrokenRowID != targetID {
		t.Fatalf("expected broken row id to be %d, got %d (err: %v)", targetID, tamperRes.BrokenRowID, tamperErr)
	}
}
