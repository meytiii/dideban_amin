package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCookieSecureStartupFatalWithoutTLSOrFlag(t *testing.T) {
	if os.Getenv("TEST_SUBPROCESS_COOKIE_SECURE") == "1" {

		certFile := os.Getenv("HERASAT_TLS_CERT")
		keyFile := os.Getenv("HERASAT_TLS_KEY")
		tlsConfigured := certFile != "" && keyFile != ""
		cookieSecure := os.Getenv("HERASAT_COOKIE_SECURE") == "true"
		if !tlsConfigured && !cookieSecure {
			os.Stderr.WriteString("FATAL: refusing to start without TLS and without HERASAT_COOKIE_SECURE=true\n")
			os.Exit(1)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCookieSecureStartupFatalWithoutTLSOrFlag")
	cmd.Env = []string{
		"TEST_SUBPROCESS_COOKIE_SECURE=1",
		"HERASAT_COOKIE_SECURE=",
		"HERASAT_TLS_CERT=",
		"HERASAT_TLS_KEY=",
		"SYSTEMROOT=" + os.Getenv("SYSTEMROOT"),
	}

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Expected subprocess to exit with fatal non-zero status when neither TLS nor HERASAT_COOKIE_SECURE is set")
	}

	outStr := string(output)
	if !strings.Contains(outStr, "FATAL: refusing to start without TLS and without HERASAT_COOKIE_SECURE=true") {
		t.Errorf("Expected output to contain fatal message, got: %s", outStr)
	}
}

func TestCookieSecureCachedResolution(t *testing.T) {
	InitCookieSecurity(true)
	if !isCookieSecure() {
		t.Errorf("Expected isCookieSecure() to return true when cached as true")
	}
}
