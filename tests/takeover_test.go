package main

import (
	"testing"
	"time"
)

func TestTakeoverCooldown(t *testing.T) {
	sm := NewSessionManager(nil)
	username := "test_takeover_user"

	isCooldown, rem := sm.CheckTakeoverCooldown(username)
	if isCooldown || rem != 0 {
		t.Fatalf("expected no cooldown initially, got: %v, %d", isCooldown, rem)
	}

	sm.RecordTakeover(username)

	isCooldown, rem = sm.CheckTakeoverCooldown(username)
	if !isCooldown || rem <= 0 || rem > 60 {
		t.Fatalf("expected cooldown between 1 and 60, got: %v, %d", isCooldown, rem)
	}

	token1 := "token_dev_1"
	sess1 := AdminSession{
		Token:      token1,
		Username:   username,
		Role:       "edare_kol",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Expiry:     time.Now().Add(15 * time.Minute),
	}
	sm.Set(token1, sess1)

	activeSess, exists := sm.GetActiveSession(username)
	if !exists || activeSess == nil {
		t.Fatalf("expected active session for user, got none")
	}

	token2 := "token_dev_2"
	sess2 := AdminSession{
		Token:      token2,
		Username:   username,
		Role:       "edare_kol",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Expiry:     time.Now().Add(15 * time.Minute),
	}
	sm.Set(token2, sess2)

	reason, wasKicked := sm.GetKickedReason(token1)
	if !wasKicked || reason != "CONCURRENT_LOGIN" {
		t.Fatalf("expected token1 to be kicked with CONCURRENT_LOGIN, got reason: %s, kicked: %v", reason, wasKicked)
	}

	_, exists = sm.Get(token1)
	if exists {
		t.Fatalf("expected token1 to be deleted from active sessions")
	}

	_, exists = sm.Get(token2)
	if !exists {
		t.Fatalf("expected token2 to remain active")
	}
}
