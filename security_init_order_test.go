package main

import (
	"testing"
	"time"
)

func TestSecurityStoresInitOrder(t *testing.T) {

	uninitStore := NewUsedCaptchaStore(nil)
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		uninitStore.TryRedeem("some_sig", time.Now().Add(time.Minute).Unix())
	}()

	if !panicked {
		t.Errorf("expected TryRedeem on uninitialized store to panic")
	}

	uninitLockouts := NewAccountLockoutManager(nil)
	lockoutPanicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				lockoutPanicked = true
			}
		}()
		uninitLockouts.IsLocked("test_user")
	}()

	if !lockoutPanicked {
		t.Errorf("expected IsLocked on uninitialized lockout manager to panic")
	}

	db := getTestDB(t)
	uninitStore.init(db)
	uninitLockouts.init(db)

	if uninitStore.db == nil {
		t.Errorf("expected store db to be initialized")
	}
	if uninitLockouts.db == nil {
		t.Errorf("expected lockouts db to be initialized")
	}
}
