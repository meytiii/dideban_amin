package main

import (
	"sync"
	"testing"
	"time"
)

func TestAccountLockoutConcurrentFailures(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	testUser := "lockout_victim_test_" + time.Now().Format("20060102150405.000000")
	alm := NewAccountLockoutManager(db)

	// Clean any previous state
	alm.Reset(testUser)

	var wg sync.WaitGroup
	workers := 5
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			alm.RecordFailure(testUser)
		}()
	}
	wg.Wait()

	// Create new manager instance to verify DB persistence without cache
	alm2 := NewAccountLockoutManager(db)
	isLocked, remaining := alm2.IsLocked(testUser)
	if !isLocked {
		t.Fatalf("Expected account %s to be locked after 5 concurrent failures, but was not locked", testUser)
	}

	if remaining <= 0 || remaining > 16*time.Minute {
		t.Errorf("Expected lockout duration ~15 minutes, got %v", remaining)
	}

	// Reset should delete row
	alm2.Reset(testUser)
	isLockedAfterReset, _ := alm2.IsLocked(testUser)
	if isLockedAfterReset {
		t.Errorf("Expected account %s to be unlocked after Reset, but was still locked", testUser)
	}
}
