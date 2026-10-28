package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGenerateTOTPSecret(t *testing.T) {
	secret1, err := GenerateBase32Secret()
	if err != nil {
		t.Fatalf("GenerateBase32Secret error: %v", err)
	}

	secret2, err := GenerateBase32Secret()
	if err != nil {
		t.Fatalf("GenerateBase32Secret error: %v", err)
	}

	if secret1 == secret2 {
		t.Fatalf("GenerateBase32Secret produced duplicate secret: %s", secret1)
	}

	raw, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret1)
	if err != nil {
		t.Fatalf("secret is not valid unpadded base32: %v", err)
	}
	if len(raw) != 20 {
		t.Fatalf("expected 20 raw bytes, got %d", len(raw))
	}
}

func TestTOTPGenerationAndVerification(t *testing.T) {
	secret, err := GenerateBase32Secret()
	if err != nil {
		t.Fatalf("GenerateBase32Secret error: %v", err)
	}

	username := "test_admin_user"
	step := time.Now().Unix() / 30

	code, err := GenerateTOTPCode(secret, step)
	if err != nil {
		t.Fatalf("GenerateTOTPCode error: %v", err)
	}

	if len(code) != 6 {
		t.Fatalf("TOTP code length must be 6 digits, got: %s", code)
	}

	ok, err := VerifyTOTPCode(username, secret, code)
	if err != nil || !ok {
		t.Fatalf("VerifyTOTPCode failed for valid code: ok=%v, err=%v", ok, err)
	}

	ok, err = VerifyTOTPCode(username, secret, code)
	if err != ErrTOTPReplayed {
		t.Fatalf("expected ErrTOTPReplayed on code reuse, got: ok=%v, err=%v", ok, err)
	}

	bogusCode := "000000"
	if bogusCode == code {
		bogusCode = "999999"
	}
	ok, err = VerifyTOTPCode(username, secret, bogusCode)
	if err != ErrInvalidTOTPCode {
		t.Fatalf("expected ErrInvalidTOTPCode for bogus code, got: ok=%v, err=%v", ok, err)
	}
}

func TestTOTPToleranceWindow(t *testing.T) {
	secret, _ := GenerateBase32Secret()
	username := "tolerance_user"

	currentStep := time.Now().Unix() / 30
	pastCode, err := GenerateTOTPCode(secret, currentStep-1)
	if err != nil {
		t.Fatalf("Generate past code failed: %v", err)
	}

	ok, err := VerifyTOTPCode(username, secret, pastCode)
	if err != nil || !ok {
		t.Fatalf("Verify past code (-1 timestep) failed: ok=%v, err=%v", ok, err)
	}

	futureCode, err := GenerateTOTPCode(secret, currentStep+1)
	if err != nil {
		t.Fatalf("Generate future code failed: %v", err)
	}

	ok, err = VerifyTOTPCode(username+"_2", secret, futureCode)
	if err != nil || !ok {
		t.Fatalf("Verify future code (+1 timestep) failed: ok=%v, err=%v", ok, err)
	}

	tooOldCode, err := GenerateTOTPCode(secret, currentStep-3)
	if err != nil {
		t.Fatalf("Generate too old code failed: %v", err)
	}
	ok, err = VerifyTOTPCode(username+"_3", secret, tooOldCode)
	if err != ErrInvalidTOTPCode {
		t.Fatalf("expected ErrInvalidTOTPCode for code outside tolerance, got: ok=%v, err=%v", ok, err)
	}
}

func TestConcurrentTOTPVerificationRace(t *testing.T) {
	secret, _ := GenerateBase32Secret()
	username := fmt.Sprintf("concurrent_user_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
	step := time.Now().Unix() / 30
	code, _ := GenerateTOTPCode(secret, step)

	const workers = 10
	var wg sync.WaitGroup
	results := make([]error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, results[idx] = VerifyTOTPCode(username, secret, code)
		}(i)
	}
	wg.Wait()

	successCount := 0
	replayCount := 0
	for _, res := range results {
		if res == nil {
			successCount++
		} else if res == ErrTOTPReplayed {
			replayCount++
		}
	}

	if successCount != 1 {
		t.Fatalf("expected exactly 1 success under concurrent verification, got: %d", successCount)
	}
	if replayCount != workers-1 {
		t.Fatalf("expected %d replays, got: %d", workers-1, replayCount)
	}
}

func TestMFAPendingTokenLifecycle(t *testing.T) {
	InitCrypto("12345678901234567890123456789012")

	token, err := GenerateMFAPendingToken("john_doe")
	if err != nil {
		t.Fatalf("GenerateMFAPendingToken failed: %v", err)
	}

	payload, err := ValidateMFAPendingToken(token)
	if err != nil {
		t.Fatalf("ValidateMFAPendingToken valid token failed: %v", err)
	}

	if payload.Username != "john_doe" || payload.Scope != "mfa_pending" {
		t.Fatalf("unexpected payload values: %+v", payload)
	}

	parts := strings.Split(token, ".")
	if len(parts) == 2 {
		tamperedToken := parts[0] + "xyz." + parts[1]
		_, err := ValidateMFAPendingToken(tamperedToken)
		if err != ErrInvalidMFAToken {
			t.Fatalf("expected ErrInvalidMFAToken for tampered payload, got: %v", err)
		}
	}

	expiredPayload := MFAPendingPayload{
		Username: "expired_user",
		Exp:      time.Now().Add(-10 * time.Second).Unix(),
		Iat:      time.Now().Add(-20 * time.Second).Unix(),
		Scope:    "mfa_pending",
		Nonce:    "expired_nonce_123",
	}
	rawExpired, _ := json.Marshal(expiredPayload)
	b64Exp := base64.RawURLEncoding.EncodeToString(rawExpired)
	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(b64Exp))
	sigExp := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	expiredToken := b64Exp + "." + sigExp

	_, err = ValidateMFAPendingToken(expiredToken)
	if err != ErrInvalidMFAToken {
		t.Fatalf("expected ErrInvalidMFAToken for expired token, got: %v", err)
	}

	wrongScopePayload := MFAPendingPayload{
		Username: "wrong_scope_user",
		Exp:      time.Now().Add(100 * time.Second).Unix(),
		Iat:      time.Now().Unix(),
		Scope:    "admin_session",
		Nonce:    "wrong_scope_nonce",
	}
	rawWrong, _ := json.Marshal(wrongScopePayload)
	b64Wrong := base64.RawURLEncoding.EncodeToString(rawWrong)
	macWrong := hmac.New(sha256.New, GetHMACKey())
	macWrong.Write([]byte(b64Wrong))
	sigWrong := base64.RawURLEncoding.EncodeToString(macWrong.Sum(nil))
	wrongScopeToken := b64Wrong + "." + sigWrong

	_, err = ValidateMFAPendingToken(wrongScopeToken)
	if err != ErrMFAScopeMismatch {
		t.Fatalf("expected ErrMFAScopeMismatch for non-mfa_pending scope, got: %v", err)
	}
}
