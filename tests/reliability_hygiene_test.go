package main

import (
	"strings"
	"testing"
	"time"
)

func TestStripXMLControlChars(t *testing.T) {
	input := "گزارش تخلف\x00\x01\x08 اداری\x0b\x0c همراه با تب\t و خط جدید\n و بازگشت\r و پایان\x1f."
	clean := stripXMLControlChars(input)

	if strings.ContainsAny(clean, "\x00\x01\x08\x0b\x0c\x1f") {
		t.Fatalf("stripXMLControlChars did not remove forbidden control characters: %q", clean)
	}

	if !strings.Contains(clean, "\t") || !strings.Contains(clean, "\n") || !strings.Contains(clean, "\r") {
		t.Fatalf("stripXMLControlChars must preserve valid whitespace characters: %q", clean)
	}

	expectedPrefix := "گزارش تخلف اداری"
	if !strings.HasPrefix(clean, expectedPrefix) {
		t.Fatalf("expected prefix %q, got %q", expectedPrefix, clean)
	}
}

func TestResolvePythonBinary(t *testing.T) {
	bin, err := resolvePythonBinary()
	if err != nil {
		t.Logf("python not detected in environment: %v", err)
	} else if bin == "" {
		t.Fatalf("expected non-empty path when error is nil")
	}
}

func TestNormalizeRateLimitKey(t *testing.T) {
	ipv4 := "192.168.1.55"
	k4 := normalizeRateLimitKey(ipv4)
	if k4 != "192.168.1.55" {
		t.Fatalf("expected IPv4 to remain unchanged, got: %s", k4)
	}

	ipv6A := "2001:db8:abcd:1234:0000:0000:0000:0001"
	ipv6B := "2001:db8:abcd:1234:ffff:eeee:dddd:cccc"
	kA := normalizeRateLimitKey(ipv6A)
	kB := normalizeRateLimitKey(ipv6B)

	if kA != kB {
		t.Fatalf("expected same /64 subnet key for %s and %s, got %s vs %s", ipv6A, ipv6B, kA, kB)
	}
	if !strings.HasSuffix(kA, "/64") {
		t.Fatalf("expected /64 suffix on IPv6 key, got: %s", kA)
	}
}

func TestDeterministicLRUEviction(t *testing.T) {
	limiter := &IPRateLimiter{
		visitors:    make(map[string]*visitor),
		rate:        1.0,
		capacity:    5.0,
		maxVisitors: 3,
	}

	baseTime := time.Now().Add(-10 * time.Minute)
	limiter.visitors["10.0.0.1"] = &visitor{lastSeen: baseTime}
	limiter.visitors["10.0.0.2"] = &visitor{lastSeen: baseTime.Add(2 * time.Minute)}
	limiter.visitors["10.0.0.3"] = &visitor{lastSeen: baseTime.Add(4 * time.Minute)}

	now := time.Now()
	limiter.evictStaleLocked(now)

	if len(limiter.visitors) >= 3 {
		t.Fatalf("expected eviction when visitors count reaches maxVisitors")
	}
}
