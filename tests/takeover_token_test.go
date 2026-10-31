package main

import (
	"testing"
)

func TestTakeoverTokenLifecycle(t *testing.T) {
	username := "admin_takeover_user"
	ip := "127.0.0.1"
	ua := "TestAgent/1.0"

	token, err := GenerateTakeoverToken(username, ip, ua)
	if err != nil {
		t.Fatalf("GenerateTakeoverToken failed: %v", err)
	}

	payload, err := ValidateTakeoverToken(token, ip, ua)
	if err != nil {
		t.Fatalf("ValidateTakeoverToken failed on valid token: %v", err)
	}
	if payload.Username != username {
		t.Fatalf("expected username %s, got %s", username, payload.Username)
	}

	_, err = ValidateTakeoverToken(token, ip, ua)
	if err == nil {
		t.Fatalf("expected error on replaying token, got nil")
	}

	token2, _ := GenerateTakeoverToken(username, ip, ua)
	_, err = ValidateTakeoverToken(token2, "192.168.1.1", ua)
	if err == nil {
		t.Fatalf("expected error on IP mismatch, got nil")
	}
}
