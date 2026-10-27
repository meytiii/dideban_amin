package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func getTestDB(t *testing.T) *sql.DB {
	connStr := os.Getenv("HERASAT_DB_CONN")
	if connStr == "" {
		if os.Getenv("HERASAT_REQUIRE_DB_TESTS") == "1" {
			t.Fatalf("HERASAT_REQUIRE_DB_TESTS=1 is set but HERASAT_DB_CONN is unset")
		}
		t.Skip("Skipping PostgreSQL integration test: HERASAT_DB_CONN unset")
		return nil
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		if os.Getenv("HERASAT_REQUIRE_DB_TESTS") == "1" {
			t.Fatalf("Required PostgreSQL integration test failed to open connection: %v", err)
		}
		t.Skipf("Skipping PostgreSQL integration test: %v", err)
		return nil
	}
	if err := db.Ping(); err != nil {
		if os.Getenv("HERASAT_REQUIRE_DB_TESTS") == "1" {
			t.Fatalf("Required PostgreSQL integration test failed to ping database: %v", err)
		}
		t.Skipf("Skipping PostgreSQL integration test (ping failed): %v", err)
	}

	migrationSQL := `
	CREATE TABLE IF NOT EXISTS admin_sessions (
		token_hash CHAR(64) PRIMARY KEY,
		username VARCHAR(64) NOT NULL,
		full_name VARCHAR(128) NOT NULL,
		role_title VARCHAR(128) NOT NULL,
		role VARCHAR(32) NOT NULL,
		is_superuser BOOLEAN NOT NULL DEFAULT FALSE,
		ip VARCHAR(45) NOT NULL,
		user_agent TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		last_active TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expiry TIMESTAMPTZ NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_username ON admin_sessions(username);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_expiry ON admin_sessions(expiry);

	CREATE TABLE IF NOT EXISTS admin_session_kicks (
		token_hash CHAR(64) PRIMARY KEY,
		reason VARCHAR(64) NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS admin_login_attempts (
		username VARCHAR(64) NOT NULL,
		ip VARCHAR(45) NOT NULL DEFAULT '',
		consecutive_failures INT NOT NULL DEFAULT 0,
		locked_until TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (username, ip)
	);

	CREATE TABLE IF NOT EXISTS used_captcha_tokens (
		signature_b64 TEXT PRIMARY KEY,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	`
	if _, err := db.Exec(migrationSQL); err != nil {
		t.Fatalf("Failed to create test session tables: %v", err)
	}
	DB = db
	runMigrations()
	return db
}

func TestSessionPersistenceAcrossReinstantiation(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	sm1 := NewSessionManager(db)
	testToken := "test-token-persist-" + time.Now().Format("20060102150405.000000")
	testUser := "test_admin_persist"

	now := time.Now()
	expiry := now.Add(15 * time.Minute)
	sess := AdminSession{
		Username:    testUser,
		FullName:    "Test Persistence Admin",
		Title:       "Security Auditor",
		Role:        "superuser",
		IsSuperuser: true,
		IP:          "127.0.0.1",
		UserAgent:   "Go-Test-Agent",
		CreatedAt:   now,
		LastActive:  now,
		Expiry:      expiry,
	}

	sm1.Set(testToken, sess)

	sm2 := NewSessionManager(db)

	retrieved, ok := sm2.Get(testToken)
	if !ok {
		t.Fatalf("Expected session to be retrieved from PostgreSQL after SessionManager restart, but was not found")
	}

	if retrieved.Username != testUser {
		t.Errorf("Expected username %s, got %s", testUser, retrieved.Username)
	}
	if retrieved.Role != "superuser" {
		t.Errorf("Expected role superuser, got %s", retrieved.Role)
	}

	sm2.Delete(testToken)
	if _, ok := sm2.Get(testToken); ok {
		t.Errorf("Expected session to be deleted, but was still found")
	}
}

func TestSessionRevokeAllUserSessions(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	sm := NewSessionManager(db)
	targetUser := "user_to_revoke_test"

	token1 := "test-rev-token-1-" + time.Now().Format("20060102150405.000001")
	token2 := "test-rev-token-2-" + time.Now().Format("20060102150405.000002")

	now := time.Now()
	expiry := now.Add(15 * time.Minute)

	sess1 := AdminSession{
		Username:    targetUser,
		FullName:    "Revoke Target",
		Title:       "Officer",
		Role:        "edare_kol",
		IsSuperuser: false,
		IP:          "10.0.0.1",
		UserAgent:   "Agent-1",
		CreatedAt:   now,
		LastActive:  now,
		Expiry:      expiry,
	}
	sess2 := AdminSession{
		Username:    targetUser,
		FullName:    "Revoke Target",
		Title:       "Officer",
		Role:        "edare_kol",
		IsSuperuser: false,
		IP:          "10.0.0.2",
		UserAgent:   "Agent-2",
		CreatedAt:   now,
		LastActive:  now,
		Expiry:      expiry,
	}

	th1 := hashToken(token1)
	th2 := hashToken(token2)
	_, _ = db.Exec(`INSERT INTO admin_sessions (token_hash, username, full_name, role_title, role, is_superuser, ip, user_agent, created_at, last_active, expiry)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		th1, sess1.Username, sess1.FullName, sess1.Title, sess1.Role, sess1.IsSuperuser, sess1.IP, sess1.UserAgent, sess1.CreatedAt, sess1.LastActive, sess1.Expiry)
	_, _ = db.Exec(`INSERT INTO admin_sessions (token_hash, username, full_name, role_title, role, is_superuser, ip, user_agent, created_at, last_active, expiry)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		th2, sess2.Username, sess2.FullName, sess2.Title, sess2.Role, sess2.IsSuperuser, sess2.IP, sess2.UserAgent, sess2.CreatedAt, sess2.LastActive, sess2.Expiry)

	revokedCount := sm.RevokeAllUserSessions(targetUser, "SECURITY_INCIDENT")
	if revokedCount != 2 {
		t.Errorf("Expected RevokeAllUserSessions to revoke 2 sessions, got %d", revokedCount)
	}

	reason1, kicked1 := sm.GetKickedReason(token1)
	if !kicked1 || reason1 != "SECURITY_INCIDENT" {
		t.Errorf("Expected token1 kicked with SECURITY_INCIDENT, got reason=%q, kicked=%v", reason1, kicked1)
	}

	if _, kickedAgain := sm.GetKickedReason(token1); kickedAgain {
		t.Errorf("Expected GetKickedReason to be single-use, but second read was true")
	}
}

func TestSessionManagementAndRateLimiter(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	sm := NewSessionManager(db)
	testUser := "user_manage_test"
	token := "test-manage-token-" + time.Now().Format("20060102150405.000000")

	sess := AdminSession{
		Username:    testUser,
		FullName:    "Manager Admin",
		Title:       "Supervisor",
		Role:        "admin",
		IsSuperuser: false,
		IP:          "192.168.1.100",
		UserAgent:   "TestAgent/1.0",
		CreatedAt:   time.Now(),
		LastActive:  time.Now(),
		Expiry:      time.Now().Add(30 * time.Minute),
	}

	sm.Set(token, sess)

	count := sm.ActiveCount()
	if count < 1 {
		t.Errorf("Expected ActiveCount >= 1, got %d", count)
	}

	sessions := sm.GetAllActiveSessions()
	found := false
	for _, s := range sessions {
		if s.Username == testUser {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected user %s to be listed in GetAllActiveSessions", testUser)
	}

	sm.UpdateActivity(token)

	revokedUser, ok := sm.RevokeSessionByToken(token, "ADMIN_TERMINATION")
	if !ok || revokedUser != testUser {
		t.Errorf("Expected RevokeSessionByToken to succeed for %s, got user=%q, ok=%v", testUser, revokedUser, ok)
	}

	reason, kicked := sm.GetKickedReason(token)
	if !kicked || reason != "ADMIN_TERMINATION" {
		t.Errorf("Expected kicked reason ADMIN_TERMINATION, got reason=%q, kicked=%v", reason, kicked)
	}

	sm.Delete(token)

	limiter := NewIPRateLimiter(2.0, 2)
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	wrapped := RateLimitMiddleware(limiter, dummyHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.195:12345"

	rr1 := httptest.NewRecorder()
	wrapped(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected 200 on first request, got %d", rr1.Code)
	}

	rr2 := httptest.NewRecorder()
	wrapped(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected 200 on second request, got %d", rr2.Code)
	}

	rr3 := httptest.NewRecorder()
	wrapped(rr3, req)
	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests on burst limit exceeded, got %d", rr3.Code)
	}
}
