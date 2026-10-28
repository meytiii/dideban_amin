package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestCaptchaReplayPrevention(t *testing.T) {
	InitCrypto("01234567890123456789012345678901")

	answer := 42
	token := GenerateCaptchaToken(answer, 2*time.Minute)

	if !VerifyCaptcha(token, "42") {
		t.Fatalf("Expected first CAPTCHA verification to succeed")
	}

	if VerifyCaptcha(token, "42") {
		t.Fatalf("SECURITY VIOLATION: CAPTCHA replay succeeded on second attempt!")
	}

	token2 := GenerateCaptchaToken(99, 2*time.Minute)
	if VerifyCaptcha(token2, "100") {
		t.Fatalf("Expected incorrect answer to fail")
	}

	if !VerifyCaptcha(token2, "99") {
		t.Fatalf("Expected correct answer to succeed after a failed attempt on same token")
	}

	if VerifyCaptcha(token2, "99") {
		t.Fatalf("SECURITY VIOLATION: Replay attack succeeded on token2!")
	}
}

func TestIPRateLimiterBoundedCapacity(t *testing.T) {
	limiter := NewIPRateLimiter(1.0, 5)
	limiter.maxVisitors = 100

	for i := 0; i < 1000; i++ {
		ip := fmt.Sprintf("192.168.%d.%d", i/256, i%256)
		limiter.Allow(ip)
	}

	limiter.mu.Lock()
	count := len(limiter.visitors)
	limiter.mu.Unlock()

	if count > limiter.maxVisitors+50 {
		t.Fatalf("Rate limiter exceeded bounded capacity: %d (max: %d)", count, limiter.maxVisitors)
	}
}

func TestInspectZipArchive(t *testing.T) {

	safeBuf := new(bytes.Buffer)
	zw := zip.NewWriter(safeBuf)
	f, err := zw.Create("document.txt")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("Hello safe document"))
	zw.Close()

	if err := inspectZipArchive(safeBuf.Bytes()); err != nil {
		t.Fatalf("Expected safe zip to pass, got error: %v", err)
	}

	traversalBuf := new(bytes.Buffer)
	zw = zip.NewWriter(traversalBuf)
	f, err = zw.Create("../../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("malicious content"))
	zw.Close()

	if err := inspectZipArchive(traversalBuf.Bytes()); err == nil {
		t.Fatalf("SECURITY VIOLATION: Directory traversal zip was allowed!")
	}

	exeBuf := new(bytes.Buffer)
	zw = zip.NewWriter(exeBuf)
	f, err = zw.Create("subfolder/malware.exe")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("MZ..."))
	zw.Close()

	if err := inspectZipArchive(exeBuf.Bytes()); err == nil {
		t.Fatalf("SECURITY VIOLATION: Nested .exe in zip was allowed!")
	}

	batBuf := new(bytes.Buffer)
	zw = zip.NewWriter(batBuf)
	f, err = zw.Create("script.bat")
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("@echo off"))
	zw.Close()

	if err := inspectZipArchive(batBuf.Bytes()); err == nil {
		t.Fatalf("SECURITY VIOLATION: Nested .bat in zip was allowed!")
	}
}
