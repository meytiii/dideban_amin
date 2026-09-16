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
		connStr = "host=127.0.0.1 port=5432 user=herasat_user password=herasat1405@ dbname=herasat_db sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration test: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("Skipping PostgreSQL integration test (ping failed): %v", err)
	}

	// Ensure tables exist
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
		username VARCHAR(64) PRIMARY KEY,
		consecutive_failures INT NOT NULL DEFAULT 0,
		locked_until TIMESTAMPTZ,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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

	// Simulate restart by creating a completely new SessionManager instance with empty in-memory cache
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

	// Clean up
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

	// Force insert two sessions directly to simulate concurrent sessions before revoke
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

	// Verify both tokens return kicked reason
	reason1, kicked1 := sm.GetKickedReason(token1)
	if !kicked1 || reason1 != "SECURITY_INCIDENT" {
		t.Errorf("Expected token1 kicked with SECURITY_INCIDENT, got reason=%q, kicked=%v", reason1, kicked1)
	}

	// Second read must be false because single-use
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

	// 1. Active count
	count := sm.ActiveCount()
	if count < 1 {
		t.Errorf("Expected ActiveCount >= 1, got %d", count)
	}

	// 2. GetAllActiveSessions
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

	// 3. UpdateActivity
	sm.UpdateActivity(token)

	// 4. RevokeSessionByToken
	revokedUser, ok := sm.RevokeSessionByToken(token, "ADMIN_TERMINATION")
	if !ok || revokedUser != testUser {
		t.Errorf("Expected RevokeSessionByToken to succeed for %s, got user=%q, ok=%v", testUser, revokedUser, ok)
	}

	// Verify kicked reason
	reason, kicked := sm.GetKickedReason(token)
	if !kicked || reason != "ADMIN_TERMINATION" {
		t.Errorf("Expected kicked reason ADMIN_TERMINATION, got reason=%q, kicked=%v", reason, kicked)
	}

	// 5. Delete session
	sm.Delete(token)

	// 6. RateLimitMiddleware
	limiter := NewIPRateLimiter(2.0, 2)
	dummyHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
	wrapped := RateLimitMiddleware(limiter, dummyHandler)

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.195:12345"

	// Request 1: allowed
	rr1 := httptest.NewRecorder()
	wrapped(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected 200 on first request, got %d", rr1.Code)
	}

	// Request 2: allowed
	rr2 := httptest.NewRecorder()
	wrapped(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Errorf("Expected 200 on second request, got %d", rr2.Code)
	}

	// Request 3: rate limited
	rr3 := httptest.NewRecorder()
	wrapped(rr3, req)
	if rr3.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests on burst limit exceeded, got %d", rr3.Code)
	}
}

