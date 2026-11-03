package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDeactivatedUserLoginMessage(t *testing.T) {
	fullName := "رضا محمدی"
	safeUsername := "rmohammadi"
	greetingName := strings.TrimSpace(fullName)
	if greetingName == "" {
		greetingName = safeUsername
	}
	expectedMsg := fmt.Sprintf("همکار گرامی %s؛\nحساب شما غیرفعال گردیده است. جهت فعالسازی حساب کاربری با پشتیبان سامانه تماس حاصل فرمایید.", greetingName)

	if !strings.Contains(expectedMsg, "رضا محمدی") {
		t.Fatalf("expected message to contain full name")
	}
	if !strings.Contains(expectedMsg, "حساب شما غیرفعال گردیده است") {
		t.Fatalf("expected message to state account deactivated")
	}

	w := httptest.NewRecorder()
	writeAuthJSONError(w, http.StatusForbidden, "account_deactivated", expectedMsg)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["error"] != "account_deactivated" {
		t.Errorf("expected error=account_deactivated, got %v", body["error"])
	}
	if body["message"] != expectedMsg {
		t.Errorf("expected message=%q, got %q", expectedMsg, body["message"])
	}
}

func TestDeactivatedUserSessionTermination(t *testing.T) {
	sm := NewSessionManager(nil)
	token := "dummy_token_12345"
	sm.Set(token, AdminSession{
		Username:   "test_user",
		IP:         "127.0.0.1",
		UserAgent:  "agent",
		Role:       "edare_kol",
		Title:      "Admin",
		Expiry:     time.Now().Add(1 * time.Hour),
		LastActive: time.Now(),
	})

	sm.RevokeAllUserSessions("test_user", "DEACTIVATED")

	_, reason, exists := sm.GetWithReason(token)
	if exists {
		t.Fatalf("expected session to be terminated")
	}
	if reason != "DEACTIVATED" {
		t.Fatalf("expected reason=DEACTIVATED, got %s", reason)
	}
}
