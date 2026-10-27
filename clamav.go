package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
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
	val := strings.ToLower(strings.TrimSpace(os.Getenv("HERASAT_CLAMAV_ENABLED")))
	enabled := false
	if val == "true" || val == "1" {
		enabled = true
	} else if val == "false" || val == "0" {
		enabled = false
	} else {
		enabled = os.Getenv("HERASAT_DEV_MODE") != "1"
	}

	addr := os.Getenv("HERASAT_CLAMAV_ADDR")
	if addr == "" {
		host := os.Getenv("CLAMAV_URL")
		port := os.Getenv("CLAMAV_PORT")
		if host != "" {
			if port != "" {
				addr = net.JoinHostPort(host, port)
			} else {
				addr = net.JoinHostPort(host, "3310")
			}
		} else {
			addr = "127.0.0.1:3310"
		}
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

func (s *ClamAVScanner) Ping() error {
	if !s.enabled {
		return nil
	}
	conn, err := net.DialTimeout("tcp", s.address, s.timeout)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(s.timeout)); err != nil {
		return fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	if _, err := conn.Write([]byte("zPING\x00")); err != nil {
		return fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	buf := make([]byte, 16)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("%w: %v", ErrClamAVUnavailable, err)
	}

	resp := strings.TrimSpace(strings.ReplaceAll(string(buf[:n]), "\x00", ""))
	if !strings.Contains(resp, "PONG") {
		return fmt.Errorf("%w: unexpected clamav ping response: %s", ErrClamAVUnavailable, resp)
	}
	return nil
}

func CheckClamAVStartup(scanner *ClamAVScanner) error {
	if !scanner.IsEnabled() {
		log.Println("WARNING: ClamAV antivirus scanner is DISABLED. Uploaded files will not be scanned for malware! Do not run with ClamAV disabled in production.")
		return nil
	}

	err := scanner.Ping()
	if err != nil {
		required := os.Getenv("HERASAT_CLAMAV_REQUIRED") == "1" || os.Getenv("HERASAT_CLAMAV_REQUIRED") == "true"
		if required {
			return fmt.Errorf("FATAL: ClamAV is required (HERASAT_CLAMAV_REQUIRED=1) but daemon at %s is unreachable: %w", scanner.address, err)
		}
		log.Printf("WARNING: ClamAV daemon at %s is currently unreachable: %v. Scans will fail closed on file uploads until service is available.", scanner.address, err)
	} else {
		log.Printf("ClamAV antivirus daemon reachable at %s (PING/PONG verified)", scanner.address)
	}
	return nil
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
