package main

import (
	"bytes"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSyslogAsyncWorkerDelivery(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on udp: %v", err)
	}
	defer pc.Close()

	addr := pc.LocalAddr().String()

	_ = os.Setenv("HERASAT_SYSLOG_ENABLED", "true")
	_ = os.Setenv("HERASAT_SYSLOG_NETWORK", "udp")
	_ = os.Setenv("HERASAT_SYSLOG_ADDR", addr)
	defer func() {
		_ = os.Unsetenv("HERASAT_SYSLOG_ENABLED")
		_ = os.Unsetenv("HERASAT_SYSLOG_NETWORK")
		_ = os.Unsetenv("HERASAT_SYSLOG_ADDR")
	}()

	shipper := NewSyslogShipper()
	defer shipper.Close()

	event := SyslogAuditEvent{
		EventID:      "TEST_AUDIT_101",
		Action:       "TEST_ACTION",
		Actor:        "test_admin",
		IP:           "127.0.0.1",
		Severity:     SyslogSevInfo,
		Timestamp:    time.Now().UTC(),
		Detail:       "Safe metadata detail",
		SecurityNote: "No secrets leaked",
	}

	ok := shipper.EnqueueAudit(event)
	if !ok {
		t.Fatalf("EnqueueAudit returned false on empty queue")
	}

	buf := make([]byte, 2048)
	_ = pc.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("failed to receive syslog packet: %v", err)
	}

	msg := string(buf[:n])
	if !strings.HasPrefix(msg, "<") {
		t.Fatalf("RFC 5424 message must start with <PRI>: %s", msg)
	}
	if !strings.Contains(msg, "herasat-portal") {
		t.Fatalf("expected APP-NAME in message: %s", msg)
	}
	if !strings.Contains(msg, "TEST_ACTION") {
		t.Fatalf("expected MSGID or action in message: %s", msg)
	}
}

func TestSyslogFormatRFC5424SanitizesSecrets(t *testing.T) {
	shipper := &SyslogShipper{
		appName:  "herasat-portal",
		hostname: "herasat-node01",
	}

	event := SyslogAuditEvent{
		EventID:   "AUDIT_SEC_01",
		Action:    "LOGIN_FAILURE",
		Actor:     "malicious_actor",
		IP:        "192.168.1.100",
		Severity:  SyslogSevWarning,
		Timestamp: time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC),
		Detail:    "Invalid password attempt with Argon2 hash: $argon2id$v=19$m=65536,t=3,p=4$secret_hash_not_allowed",
	}

	line := shipper.formatRFC5424(event)
	if strings.Contains(line, "$argon2id$") {
		t.Fatalf("RFC5424 formatter failed to strip password hash: %s", line)
	}
	if strings.Contains(line, "secret_hash_not_allowed") {
		t.Fatalf("RFC5424 formatter leaked secret content: %s", line)
	}
}

func TestSyslogQueueCapacityNonBlocking(t *testing.T) {
	shipper := &SyslogShipper{
		enabled: false,
		queue:   make(chan SyslogAuditEvent, 5),
	}

	for i := 0; i < 5; i++ {
		ok := shipper.EnqueueAudit(SyslogAuditEvent{Action: "FLOOD"})
		if !ok {
			t.Fatalf("enqueue should succeed while under capacity")
		}
	}

	ok := shipper.EnqueueAudit(SyslogAuditEvent{Action: "OVERFLOW"})
	if ok {
		t.Fatalf("enqueue should return false when queue is saturated")
	}
}
