package main

import (
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

	shipper := InitSyslogShipper()
	defer shipper.Close()

	msg := SyslogMessage{
		Severity: SeverityInfo,
		Action:   "TEST_ACTION",
		Username: "test_admin",
		IP:       "127.0.0.1",
		Detail:   "Safe metadata detail",
		Time:     time.Now().UTC(),
	}

	shipper.Enqueue(msg)

	buf := make([]byte, 2048)
	_ = pc.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := pc.ReadFrom(buf)
	if err != nil {
		t.Fatalf("failed to receive syslog packet: %v", err)
	}

	packet := string(buf[:n])
	if !strings.HasPrefix(packet, "<") {
		t.Fatalf("RFC 5424 message must start with <PRI>: %s", packet)
	}
	if !strings.Contains(packet, "herasat") {
		t.Fatalf("expected APP-NAME in message: %s", packet)
	}
	if !strings.Contains(packet, "TEST_ACTION") {
		t.Fatalf("expected action in message: %s", packet)
	}
}

func TestSyslogFormatRFC5424SanitizesParams(t *testing.T) {
	shipper := &SyslogShipper{
		hostname: "herasat-node01",
		pid:      "1234",
	}

	msg := SyslogMessage{
		Severity: SeverityWarning,
		Action:   "LOGIN_FAILURE",
		Username: "admin\"\nmalicious",
		IP:       "192.168.1.100\r",
		Detail:   "Invalid password \"attempt\"",
		Time:     time.Date(2026, 9, 22, 10, 0, 0, 0, time.UTC),
	}

	line := shipper.formatRFC5424(msg)
	if strings.Contains(line, "admin\"") {
		t.Fatalf("RFC5424 formatter failed to sanitize quotes: %s", line)
	}
	if strings.Contains(line, "\n") {
		t.Fatalf("RFC5424 formatter leaked newline: %s", line)
	}
	if !strings.Contains(line, "<36>1") {
		t.Fatalf("expected PRI <36>1 for Warning severity: %s", line)
	}
}

func TestSyslogQueueCapacityNonBlocking(t *testing.T) {
	shipper := &SyslogShipper{
		enabled: true,
		queue:   make(chan SyslogMessage, 5),
	}

	for i := 0; i < 5; i++ {
		shipper.Enqueue(SyslogMessage{Action: "FLOOD"})
	}

	done := make(chan struct{})
	go func() {
		shipper.Enqueue(SyslogMessage{Action: "OVERFLOW"})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Enqueue blocked when queue was full")
	}
}
