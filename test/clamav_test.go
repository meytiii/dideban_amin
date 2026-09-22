package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"testing"
)

func startMockClamAVServer(t *testing.T, eicarSig string) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock clamav listener: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				cmdBuf := make([]byte, 10)
				n, _ := c.Read(cmdBuf)
				if n < 9 || !bytes.HasPrefix(cmdBuf, []byte("zINSTREAM\x00")) {
					return
				}

				var received bytes.Buffer
				for {
					var chunkSize uint32
					err := binary.Read(c, binary.BigEndian, &chunkSize)
					if err != nil || chunkSize == 0 {
						break
					}
					chunk := make([]byte, chunkSize)
					_, err = io.ReadFull(c, chunk)
					if err != nil {
						break
					}
					received.Write(chunk)
				}

				if strings.Contains(received.String(), eicarSig) {
					_, _ = c.Write([]byte("stream: Win.Test.EICAR_Hdb-1 FOUND\x00"))
				} else {
					_, _ = c.Write([]byte("stream: OK\x00"))
				}
			}(conn)
		}
	}()

	addr := listener.Addr().String()
	cleanup := func() {
		_ = listener.Close()
	}
	return addr, cleanup
}

func TestClamAVCleanFile(t *testing.T) {
	addr, cleanup := startMockClamAVServer(t, "EICAR-STANDARD-ANTIVIRUS-TEST-FILE")
	defer cleanup()

	_ = os.Setenv("HERASAT_CLAMAV_ENABLED", "true")
	_ = os.Setenv("HERASAT_CLAMAV_ADDR", addr)
	_ = os.Setenv("HERASAT_CLAMAV_TIMEOUT", "2s")
	defer func() {
		_ = os.Unsetenv("HERASAT_CLAMAV_ENABLED")
		_ = os.Unsetenv("HERASAT_CLAMAV_ADDR")
		_ = os.Unsetenv("HERASAT_CLAMAV_TIMEOUT")
	}()

	scanner := NewClamAVScanner()
	cleanContent := []byte("Normal clean non-malicious text content for reporting")
	ok, _, err := scanner.ScanStream(bytes.NewReader(cleanContent))
	if err != nil || !ok {
		t.Fatalf("expected clean scan to succeed, got ok=%v, error: %v", ok, err)
	}
}

func TestClamAVMalwareFound(t *testing.T) {
	eicarSig := "EICAR-STANDARD-ANTIVIRUS-TEST-FILE"
	addr, cleanup := startMockClamAVServer(t, eicarSig)
	defer cleanup()

	_ = os.Setenv("HERASAT_CLAMAV_ENABLED", "true")
	_ = os.Setenv("HERASAT_CLAMAV_ADDR", addr)
	_ = os.Setenv("HERASAT_CLAMAV_TIMEOUT", "2s")
	defer func() {
		_ = os.Unsetenv("HERASAT_CLAMAV_ENABLED")
		_ = os.Unsetenv("HERASAT_CLAMAV_ADDR")
		_ = os.Unsetenv("HERASAT_CLAMAV_TIMEOUT")
	}()

	scanner := NewClamAVScanner()
	eicarContent := []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$" + eicarSig + "!$H+H*")
	ok, virusName, err := scanner.ScanStream(bytes.NewReader(eicarContent))
	if err == nil || ok {
		t.Fatalf("expected malware detection error, got ok=%v, err=nil", ok)
	}
	if !strings.Contains(virusName, "EICAR") && !strings.Contains(err.Error(), "malware") {
		t.Fatalf("expected malware detected indicator, got: %s, err: %v", virusName, err)
	}
}

func TestClamAVFailClosedWhenUnreachable(t *testing.T) {
	_ = os.Setenv("HERASAT_CLAMAV_ENABLED", "true")
	_ = os.Setenv("HERASAT_CLAMAV_ADDR", "127.0.0.1:54321")
	_ = os.Setenv("HERASAT_CLAMAV_TIMEOUT", "500ms")
	defer func() {
		_ = os.Unsetenv("HERASAT_CLAMAV_ENABLED")
		_ = os.Unsetenv("HERASAT_CLAMAV_ADDR")
		_ = os.Unsetenv("HERASAT_CLAMAV_TIMEOUT")
	}()

	scanner := NewClamAVScanner()
	ok, _, err := scanner.ScanStream(bytes.NewReader([]byte("Some payload")))
	if err == nil || ok {
		t.Fatalf("expected fail closed error when clamav daemon is down, got ok=%v, err=nil", ok)
	}
	if !errors.Is(err, ErrClamAVUnavailable) {
		t.Fatalf("expected ErrClamAVUnavailable, got: %v", err)
	}
}

func TestClamAVDisabledBypassesNetwork(t *testing.T) {
	_ = os.Setenv("HERASAT_CLAMAV_ENABLED", "false")
	defer os.Unsetenv("HERASAT_CLAMAV_ENABLED")

	scanner := NewClamAVScanner()
	ok, _, err := scanner.ScanStream(bytes.NewReader([]byte("Any bytes")))
	if err != nil || !ok {
		t.Fatalf("when disabled, ScanStream should return ok=true and err=nil, got ok=%v, err: %v", ok, err)
	}
}
