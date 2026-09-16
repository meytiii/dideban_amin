package main

import (
	"testing"
	"time"
)

func TestCaptchaPersistenceTryRedeem(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	store1 := NewUsedCaptchaStore(db)

	testSig := "sig_test_" + time.Now().Format("20060102150405.000000")
	futureExpiry := time.Now().Add(2 * time.Minute).Unix()

	// First redeem should succeed
	if !store1.TryRedeem(testSig, futureExpiry) {
		t.Fatalf("Expected first TryRedeem to succeed")
	}

	// Second redeem with a new store instance (simulate restart) must fail via DB persistence
	store2 := NewUsedCaptchaStore(db)
	if store2.TryRedeem(testSig, futureExpiry) {
		t.Fatalf("SECURITY VIOLATION: Second TryRedeem on same signature succeeded across store restart!")
	}

	// Signature with expiresAt in the past must fail immediately
	pastSig := "sig_past_" + time.Now().Format("20060102150405.000000")
	pastExpiry := time.Now().Add(-10 * time.Second).Unix()
	if store2.TryRedeem(pastSig, pastExpiry) {
		t.Fatalf("SECURITY VIOLATION: TryRedeem with past expiresAt succeeded!")
	}
}
