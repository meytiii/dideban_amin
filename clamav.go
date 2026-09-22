package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

var (
	ErrClamAVUnavailable = errors.New("clamav service unavailable")
	ErrMalwareDetected   = errors.New("malware detected in file stream")
)

type ClamAVScanner struct {
	enabled bool
	address string
	timeout time.Duration
}

func NewClamAVScanner() *ClamAVScanner {
	enabled := strings.ToLower(os.Getenv("HERASAT_CLAMAV_ENABLED")) == "true"
	addr := os.Getenv("HERASAT_CLAMAV_ADDR")
	if addr == "" {
		addr = "127.0.0.1:3310"
	}
	timeoutStr := os.Getenv("HERASAT_CLAMAV_TIMEOUT")
	timeout := 10 * time.Second
	if timeoutStr != "" {
		if parsed, err := time.ParseDuration(timeoutStr); err == nil && parsed > 0 {
			timeout = parsed
		}
	}
	return &ClamAVScanner{
		enabled: enabled,
		address: addr,
		timeout: timeout,
	}
}

func (s *ClamAVScanner) IsEnabled() bool {
	return s.enabled
}

func (s *ClamAVScanner) ScanStream(r io.Reader) (bool, string, error) {
	if !s.enabled {
		return true, "", nil
	}

	conn, err := net.DialTimeout("tcp", s.address, s.timeout)
	if err != nil {
		return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(s.timeout)); err != nil {
		return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	if _, err := conn.Write([]byte("zINSTREAM\x00")); err != nil {
		return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	buf := make([]byte, 4096)
	lenBuf := make([]byte, 4)

	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			binary.BigEndian.PutUint32(lenBuf, uint32(n))
			if _, err := conn.Write(lenBuf); err != nil {
				return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
			}
			if _, err := conn.Write(buf[:n]); err != nil {
				return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return false, "", readErr
		}
	}

	binary.BigEndian.PutUint32(lenBuf, 0)
	if _, err := conn.Write(lenBuf); err != nil {
		return false, "", fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	var respBuf bytes.Buffer
	readRespBuf := make([]byte, 1024)
	for {
		n, err := conn.Read(readRespBuf)
		if n > 0 {
			respBuf.Write(readRespBuf[:n])
			if bytes.Contains(readRespBuf[:n], []byte{0}) || bytes.Contains(readRespBuf[:n], []byte{'\n'}) {
				break
			}
		}
		if err != nil {
			break
		}
	}

	respStr := strings.TrimSpace(strings.ReplaceAll(respBuf.String(), "\x00", ""))

	if strings.HasSuffix(respStr, "OK") {
		return true, "", nil
	}

	if strings.Contains(respStr, "FOUND") {
		virusName := "Unknown"
		parts := strings.Split(respStr, ":")
		if len(parts) >= 2 {
			sub := strings.TrimSpace(parts[1])
			virusName = strings.TrimSuffix(sub, "FOUND")
			virusName = strings.TrimSpace(virusName)
		}
		return false, virusName, ErrMalwareDetected
	}

	return false, "", fmt.Errorf("%w: unexpected clamav response: %s", ErrClamAVUnavailable, respStr)
}

func (s *ClamAVScanner) ScanBytes(data []byte) (bool, string, error) {
	return s.ScanStream(bytes.NewReader(data))
}
