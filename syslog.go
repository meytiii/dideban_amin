package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SyslogSeverity int

const (
	SeverityAlert   SyslogSeverity = 1
	SeverityWarning SyslogSeverity = 4
	SeverityNotice  SyslogSeverity = 5
	SeverityInfo    SyslogSeverity = 6
)

type SyslogMessage struct {
	Severity SyslogSeverity
	Action   string
	Username string
	IP       string
	TargetID *int64
	Detail   string
	Time     time.Time
}

type SyslogShipper struct {
	enabled  bool
	network  string
	addr     string
	queue    chan SyslogMessage
	hostname string
	pid      string
	conn     net.Conn
	mu       sync.Mutex
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

var GlobalSyslogShipper *SyslogShipper

func InitSyslogShipper() *SyslogShipper {
	enabled := strings.ToLower(os.Getenv("HERASAT_SYSLOG_ENABLED")) == "true"
	network := strings.ToLower(os.Getenv("HERASAT_SYSLOG_NETWORK"))
	if network == "" {
		network = "udp"
	}
	addr := os.Getenv("HERASAT_SYSLOG_ADDR")
	if addr == "" {
		addr = "192.168.10.50:514"
	}

	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "herasat-node"
	}

	shipper := &SyslogShipper{
		enabled:  enabled,
		network:  network,
		addr:     addr,
		queue:    make(chan SyslogMessage, 2048),
		hostname: hostname,
		pid:      strconv.Itoa(os.Getpid()),
		stopCh:   make(chan struct{}),
	}

	GlobalSyslogShipper = shipper

	if enabled {
		shipper.wg.Add(1)
		go shipper.worker()
	}

	return shipper
}

func (s *SyslogShipper) Enqueue(msg SyslogMessage) {
	if !s.enabled {
		return
	}
	if msg.Time.IsZero() {
		msg.Time = time.Now()
	}

	select {
	case s.queue <- msg:
	default:
	}
}

func (s *SyslogShipper) Close() {
	if !s.enabled {
		return
	}
	close(s.stopCh)
	s.wg.Wait()
	s.mu.Lock()
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.mu.Unlock()
}

func (s *SyslogShipper) connect() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil {
		return nil
	}

	dialer := net.Dialer{Timeout: 3 * time.Second}

	switch s.network {
	case "tcp+tls", "tls":
		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
		conn, err := tls.DialWithDialer(&dialer, "tcp", s.addr, tlsConfig)
		if err != nil {
			return err
		}
		s.conn = conn
		return nil

	case "tcp":
		conn, err := dialer.Dial("tcp", s.addr)
		if err != nil {
			return err
		}
		s.conn = conn
		return nil

	default:
		conn, err := dialer.Dial("udp", s.addr)
		if err != nil {
			return err
		}
		s.conn = conn
		return nil
	}
}

func (s *SyslogShipper) worker() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			s.flushRemaining()
			return

		case msg := <-s.queue:
			s.sendMsg(msg)
		}
	}
}

func (s *SyslogShipper) flushRemaining() {
	for {
		select {
		case msg := <-s.queue:
			s.sendMsg(msg)
		default:
			return
		}
	}
}

func (s *SyslogShipper) sendMsg(msg SyslogMessage) {
	if err := s.connect(); err != nil {
		return
	}

	formatted := s.formatRFC5424(msg)
	payload := []byte(formatted + "\n")

	s.mu.Lock()
	if s.conn == nil {
		s.mu.Unlock()
		return
	}

	s.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := s.conn.Write(payload)
	if err != nil {
		s.conn.Close()
		s.conn = nil
	}
	s.mu.Unlock()
}

func (s *SyslogShipper) formatRFC5424(msg SyslogMessage) string {
	priVal := 4*8 + int(msg.Severity)
	ts := msg.Time.Format(time.RFC3339)

	var targetStr string
	if msg.TargetID != nil {
		targetStr = strconv.FormatInt(*msg.TargetID, 10)
	} else {
		targetStr = "none"
	}

	sd := fmt.Sprintf(`[herasat@meta user="%s" ip="%s" target_id="%s"]`,
		sanitizeSyslogParam(msg.Username),
		sanitizeSyslogParam(msg.IP),
		targetStr)

	cleanDetail := sanitizeSyslogParam(msg.Detail)
	if cleanDetail == "" {
		cleanDetail = msg.Action
	}

	return fmt.Sprintf("<%d>1 %s %s herasat %s %s %s %s",
		priVal, ts, s.hostname, s.pid, msg.Action, sd, cleanDetail)
}

func sanitizeSyslogParam(input string) string {
	clean := strings.ReplaceAll(input, "\"", "")
	clean = strings.ReplaceAll(clean, "\n", " ")
	clean = strings.ReplaceAll(clean, "\r", "")
	clean = strings.ReplaceAll(clean, "\x00", "")
	return strings.TrimSpace(clean)
}

func MapActionToSyslogSeverity(action string) SyslogSeverity {
	switch action {
	case "ACCOUNT_LOCKED", "LOCKED_LOGIN_ATTEMPT", "CLAMAV_SERVICE_UNAVAILABLE",
		"ILLEGAL_FILE_ACCESS_ATTEMPT", "WORM_TAMPER_ATTEMPT", "CRYPTO_AUTH_FAILURE":
		return SeverityAlert

	case "FAILED_LOGIN", "FAILED_MFA_VERIFICATION", "REJECTED_VULGAR_CONTENT",
		"REJECTED_MALICIOUS_FILE", "REJECTED_UNSAFE_ZIP_PAYLOAD":
		return SeverityWarning

	default:
		return SeverityInfo
	}
}

func ShipSyslogAudit(username, action string, targetID *int64, ip string) {
	if GlobalSyslogShipper == nil || !GlobalSyslogShipper.enabled {
		return
	}

	sev := MapActionToSyslogSeverity(action)
	GlobalSyslogShipper.Enqueue(SyslogMessage{
		Severity: sev,
		Action:   action,
		Username: username,
		IP:       ip,
		TargetID: targetID,
		Detail:   fmt.Sprintf("Action %s performed by %s", action, username),
		Time:     time.Now(),
	})
}
