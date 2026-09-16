package main

import (
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetIPHardening(t *testing.T) {
	// (a) No trust config -> uses RemoteAddr
	os.Unsetenv("HERASAT_TRUSTED_PROXIES")
	reqA := httptest.NewRequest("GET", "/", nil)
	reqA.RemoteAddr = "10.0.0.5:12345"
	reqA.Header.Set("X-Forwarded-For", "203.0.113.195, 10.0.0.5")
	reqA.Header.Set("X-Real-IP", "203.0.113.195")

	ipA := getIP(reqA)
	if ipA != "10.0.0.5" {
		t.Errorf("Scenario (a): Expected RemoteAddr 10.0.0.5 when HERASAT_TRUSTED_PROXIES unset, got %s", ipA)
	}

	// (b) Trusted proxy + XFF -> uses right-to-left first non-trusted address
	os.Setenv("HERASAT_TRUSTED_PROXIES", "10.0.0.0/8,192.168.1.0/24")
	defer os.Unsetenv("HERASAT_TRUSTED_PROXIES")

	reqB := httptest.NewRequest("GET", "/", nil)
	reqB.RemoteAddr = "10.0.0.1:45678" // trusted proxy
	reqB.Header.Set("X-Forwarded-For", "198.51.100.22, 192.168.1.10") // client, trusted proxy hop

	ipB := getIP(reqB)
	if ipB != "198.51.100.22" {
		t.Errorf("Scenario (b): Expected client IP 198.51.100.22, got %s", ipB)
	}

	// Prevent prepended spoofed IP:
	// Attacker sends: "1.1.1.1, 198.51.100.22" through proxy 10.0.0.1
	reqB2 := httptest.NewRequest("GET", "/", nil)
	reqB2.RemoteAddr = "10.0.0.1:45678"
	reqB2.Header.Set("X-Forwarded-For", "1.1.1.1, 198.51.100.22") // attacker attempted prepending 1.1.1.1
	ipB2 := getIP(reqB2)
	if ipB2 != "198.51.100.22" {
		t.Errorf("Scenario (b2): Expected 198.51.100.22 (discarding prepended 1.1.1.1), got %s", ipB2)
	}

	// (c) Untrusted RemoteAddr + XFF -> uses RemoteAddr
	reqC := httptest.NewRequest("GET", "/", nil)
	reqC.RemoteAddr = "203.0.113.50:54321" // untrusted public IP
	reqC.Header.Set("X-Forwarded-For", "198.51.100.77")
	reqC.Header.Set("X-Real-IP", "198.51.100.77")

	ipC := getIP(reqC)
	if ipC != "203.0.113.50" {
		t.Errorf("Scenario (c): Expected RemoteAddr 203.0.113.50 when RemoteAddr not in trusted CIDRs, got %s", ipC)
	}
}
