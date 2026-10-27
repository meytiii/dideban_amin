package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func isWithinDirectory(parentDir, targetPath string) bool {
	absParent, err := filepath.Abs(filepath.Clean(parentDir))
	if err != nil {
		return false
	}
	absTarget, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return false
	}
	if absParent == absTarget {
		return true
	}
	prefix := absParent + string(os.PathSeparator)
	return strings.HasPrefix(absTarget, prefix)
}

type adminContextKeyType struct{}

var adminContextKey = adminContextKeyType{}

func contextWithAdmin(ctx context.Context, session AdminSession) context.Context {
	ctx = context.WithValue(ctx, adminContextKey, session)
	return context.WithValue(ctx, "admin", session)
}

func adminFromContext(r *http.Request) (AdminSession, bool) {
	if r == nil {
		return AdminSession{}, false
	}
	val := r.Context().Value(adminContextKey)
	if val == nil {
		val = r.Context().Value("admin")
	}
	if val == nil {
		return AdminSession{}, false
	}
	s, ok := val.(AdminSession)
	return s, ok
}

type AdminSession struct {
	Token                  string    `json:"token"`
	Username               string    `json:"username"`
	FullName               string    `json:"full_name"`
	Title                  string    `json:"role_title"`
	Role                   string    `json:"role"`
	IsSuperuser            bool      `json:"is_superuser"`
	IP                     string    `json:"ip"`
	UserAgent              string    `json:"user_agent"`
	Requires2FASetup       bool      `json:"requires_2fa_setup"`
	RequiresPasswordChange bool      `json:"requires_password_change"`
	CreatedAt              time.Time `json:"created_at"`
	LastActive             time.Time `json:"last_active"`
	Expiry                 time.Time `json:"expiry"`
}

type ActiveSessionItem struct {
	SessionID              string `json:"session_id"`
	Token                  string `json:"token"`
	Username               string `json:"username"`
	FullName               string `json:"full_name"`
	RoleTitle              string `json:"role_title"`
	Role                   string `json:"role"`
	IsSuperuser            bool   `json:"is_superuser"`
	IP                     string `json:"ip"`
	UserAgent              string `json:"user_agent"`
	Requires2FASetup       bool   `json:"requires_2fa_setup"`
	RequiresPasswordChange bool   `json:"requires_password_change"`
	CreatedAt              string `json:"created_at"`
	LastActive             string `json:"last_active"`
	Expiry                 string `json:"expiry"`
	TTLSeconds             int64  `json:"ttl_seconds"`
}

type cachedSession struct {
	session     AdminSession
	cachedUntil time.Time
}

type SessionManager struct {
	mu           sync.RWMutex
	db           *sql.DB
	cache        map[string]cachedSession
	sessions     map[string]AdminSession
	kickedTokens map[string]string
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", sum[:])
}

func NewSessionManager(db *sql.DB) *SessionManager {
	return &SessionManager{
		db:           db,
		cache:        make(map[string]cachedSession),
		sessions:     make(map[string]AdminSession),
		kickedTokens: make(map[string]string),
	}
}

var Sessions = NewSessionManager(nil)

func (sm *SessionManager) init(db *sql.DB) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.db = db
}

func (sm *SessionManager) getDB() *sql.DB {
	if sm.db != nil {
		return sm.db
	}
	return DB
}

func (sm *SessionManager) Set(token string, session AdminSession) {
	tokenHash := hashToken(token)
	session.Token = token

	if db := sm.getDB(); db != nil {

		rows, err := db.Query(`SELECT token_hash FROM admin_sessions WHERE username = $1 AND token_hash != $2`, session.Username, tokenHash)
		if err == nil {
			var oldHashes []string
			for rows.Next() {
				var oh string
				if err := rows.Scan(&oh); err == nil {
					oldHashes = append(oldHashes, oh)
				}
			}
			rows.Close()

			for _, oh := range oldHashes {
				_, _ = db.Exec(`INSERT INTO admin_session_kicks (token_hash, reason, created_at)
					VALUES ($1, 'CONCURRENT_LOGIN', NOW())
					ON CONFLICT (token_hash) DO UPDATE SET reason = 'CONCURRENT_LOGIN'`, oh)
			}
			if len(oldHashes) > 0 {
				_, _ = db.Exec(`DELETE FROM admin_sessions WHERE username = $1 AND token_hash != $2`, session.Username, tokenHash)
			}
		}

		_, _ = db.Exec(`INSERT INTO admin_sessions (token_hash, username, full_name, role_title, role, is_superuser, ip, user_agent, requires_2fa_setup, requires_password_change, created_at, last_active, expiry)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			ON CONFLICT (token_hash) DO UPDATE SET
				full_name = EXCLUDED.full_name,
				role_title = EXCLUDED.role_title,
				role = EXCLUDED.role,
				is_superuser = EXCLUDED.is_superuser,
				ip = EXCLUDED.ip,
				user_agent = EXCLUDED.user_agent,
				requires_2fa_setup = EXCLUDED.requires_2fa_setup,
				requires_password_change = EXCLUDED.requires_password_change,
				last_active = EXCLUDED.last_active,
				expiry = EXCLUDED.expiry`,
			tokenHash, session.Username, session.FullName, session.Title, session.Role, session.IsSuperuser,
			session.IP, session.UserAgent, session.Requires2FASetup, session.RequiresPasswordChange,
			session.CreatedAt, session.LastActive, session.Expiry)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for th, cs := range sm.cache {
		if cs.session.Username == session.Username && th != tokenHash {
			delete(sm.cache, th)
		}
	}
	for oldToken, oldSession := range sm.sessions {
		if oldSession.Username == session.Username && oldToken != token {
			sm.kickedTokens[oldToken] = "CONCURRENT_LOGIN"
			delete(sm.sessions, oldToken)
			GlobalSSEBroker.BroadcastSessionRevoked(oldSession.Username, hashToken(oldToken))
		}
	}

	sm.cache[tokenHash] = cachedSession{
		session:     session,
		cachedUntil: time.Now().Add(60 * time.Second),
	}
	sm.sessions[token] = session
}

func (sm *SessionManager) Get(token string) (AdminSession, bool) {
	tokenHash := hashToken(token)
	now := time.Now()

	sm.mu.RLock()
	if c, ok := sm.cache[tokenHash]; ok {
		if now.Before(c.cachedUntil) && now.Before(c.session.Expiry) && now.Sub(c.session.LastActive) <= 15*time.Minute && now.Sub(c.session.CreatedAt) <= 12*time.Hour {
			sess := c.session
			sm.mu.RUnlock()
			return sess, true
		}
	}
	sm.mu.RUnlock()

	if db := sm.getDB(); db != nil {
		var s AdminSession
		err := db.QueryRow(`
			SELECT username, full_name, role_title, role, is_superuser, ip, user_agent, requires_2fa_setup, requires_password_change, created_at, last_active, expiry
			FROM admin_sessions
			WHERE token_hash = $1`, tokenHash).Scan(
			&s.Username, &s.FullName, &s.Title, &s.Role, &s.IsSuperuser,
			&s.IP, &s.UserAgent, &s.Requires2FASetup, &s.RequiresPasswordChange,
			&s.CreatedAt, &s.LastActive, &s.Expiry)
		if err == nil {
			if now.After(s.Expiry) || now.Sub(s.LastActive) > 15*time.Minute || now.Sub(s.CreatedAt) > 12*time.Hour {
				sm.Delete(token)
				return AdminSession{}, false
			}
			s.Token = token

			sm.mu.Lock()
			sm.cache[tokenHash] = cachedSession{
				session:     s,
				cachedUntil: now.Add(60 * time.Second),
			}
			sm.sessions[token] = s
			sm.mu.Unlock()

			return s, true
		}
		if err == sql.ErrNoRows {
			sm.mu.Lock()
			delete(sm.cache, tokenHash)
			delete(sm.sessions, token)
			sm.mu.Unlock()
			return AdminSession{}, false
		}
	}

	sm.mu.RLock()
	session, exists := sm.sessions[token]
	sm.mu.RUnlock()

	if !exists {
		return AdminSession{}, false
	}

	if now.After(session.Expiry) || now.Sub(session.LastActive) > 15*time.Minute || now.Sub(session.CreatedAt) > 12*time.Hour {
		sm.Delete(token)
		return AdminSession{}, false
	}

	return session, true
}

func (sm *SessionManager) GetKickedReason(token string) (string, bool) {
	tokenHash := hashToken(token)

	if db := sm.getDB(); db != nil {
		var reason string
		err := db.QueryRow(`SELECT reason FROM admin_session_kicks WHERE token_hash = $1`, tokenHash).Scan(&reason)
		if err == nil && reason != "" {
			_, _ = db.Exec(`DELETE FROM admin_session_kicks WHERE token_hash = $1`, tokenHash)
			return reason, true
		}
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	reason, exists := sm.kickedTokens[token]
	if exists {
		delete(sm.kickedTokens, token)
		return reason, true
	}
	return "", false
}

func (sm *SessionManager) RevokeAllUserSessions(username, reason string) int {
	count := 0
	tokenHashList := make([]string, 0)

	if db := sm.getDB(); db != nil {
		rows, err := db.Query(`SELECT token_hash FROM admin_sessions WHERE username = $1`, username)
		if err == nil {
			for rows.Next() {
				var th string
				if err := rows.Scan(&th); err == nil {
					tokenHashList = append(tokenHashList, th)
				}
			}
			rows.Close()

			for _, th := range tokenHashList {
				_, _ = db.Exec(`INSERT INTO admin_session_kicks (token_hash, reason, created_at)
					VALUES ($1, $2, NOW())
					ON CONFLICT (token_hash) DO UPDATE SET reason = EXCLUDED.reason`, th, reason)
			}
			res, err := db.Exec(`DELETE FROM admin_sessions WHERE username = $1`, username)
			if err == nil {
				if ra, err := res.RowsAffected(); err == nil {
					count = int(ra)
				}
			}
		}
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for th, cs := range sm.cache {
		if cs.session.Username == username {
			delete(sm.cache, th)
		}
	}

	memCount := 0
	for token, s := range sm.sessions {
		if s.Username == username {
			sm.kickedTokens[token] = reason
			delete(sm.sessions, token)
			memCount++
		}
	}

	if count == 0 {
		count = memCount
	}
	return count
}

func (sm *SessionManager) RevokeSessionByToken(token, reason string) (string, bool) {
	tokenHash := hashToken(token)
	var username string

	if db := sm.getDB(); db != nil {
		err := db.QueryRow(`SELECT username FROM admin_sessions WHERE token_hash = $1`, tokenHash).Scan(&username)
		if err == nil {
			_, _ = db.Exec(`INSERT INTO admin_session_kicks (token_hash, reason, created_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT (token_hash) DO UPDATE SET reason = EXCLUDED.reason`, tokenHash, reason)
			_, _ = db.Exec(`DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
		}
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.cache, tokenHash)

	if s, ok := sm.sessions[token]; ok {
		sm.kickedTokens[token] = reason
		delete(sm.sessions, token)
		if username == "" {
			username = s.Username
		}
		return username, true
	}

	if username != "" {
		return username, true
	}
	return "", false
}

func (sm *SessionManager) GetAllActiveSessions() []ActiveSessionItem {
	now := time.Now()

	if db := sm.getDB(); db != nil {
		rows, err := db.Query(`
			SELECT token_hash, username, full_name, role_title, role, is_superuser, ip, user_agent, requires_2fa_setup, requires_password_change, created_at, last_active, expiry
			FROM admin_sessions
			WHERE expiry > NOW()
			ORDER BY last_active DESC`)
		if err == nil {
			defer rows.Close()
			var list []ActiveSessionItem
			for rows.Next() {
				var tokenHash, username, fullName, roleTitle, role, ip, userAgent string
				var isSuperuser, req2FA, reqPwd bool
				var createdAt, lastActive, expiry time.Time
				if err := rows.Scan(&tokenHash, &username, &fullName, &roleTitle, &role, &isSuperuser, &ip, &userAgent, &req2FA, &reqPwd, &createdAt, &lastActive, &expiry); err == nil {
					ttl := int64(expiry.Sub(now).Seconds())
					if ttl < 0 {
						ttl = 0
					}
					cleanHash := strings.TrimSpace(tokenHash)
					shortHash := cleanHash
					if len(shortHash) > 8 {
						shortHash = shortHash[:8]
					}
					sessID := "SES-" + strings.ToUpper(shortHash)
					list = append(list, ActiveSessionItem{
						SessionID:              sessID,
						Token:                  cleanHash,
						Username:               username,
						FullName:               fullName,
						RoleTitle:              roleTitle,
						Role:                   role,
						IsSuperuser:            isSuperuser,
						IP:                     ip,
						UserAgent:              userAgent,
						Requires2FASetup:       req2FA,
						RequiresPasswordChange: reqPwd,
						CreatedAt:              createdAt.Format(time.RFC3339),
						LastActive:             lastActive.Format(time.RFC3339),
						Expiry:                 expiry.Format(time.RFC3339),
						TTLSeconds:             ttl,
					})
				}
			}
			if list == nil {
				list = []ActiveSessionItem{}
			}
			return list
		}
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var list []ActiveSessionItem
	for token, s := range sm.sessions {
		if now.Before(s.Expiry) {
			ttl := int64(s.Expiry.Sub(now).Seconds())
			if ttl < 0 {
				ttl = 0
			}
			sessID := "SES-" + strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(token)))[:8])
			list = append(list, ActiveSessionItem{
				SessionID:              sessID,
				Token:                  token,
				Username:               s.Username,
				FullName:               s.FullName,
				RoleTitle:              s.Title,
				Role:                   s.Role,
				IsSuperuser:            s.IsSuperuser,
				IP:                     s.IP,
				UserAgent:              s.UserAgent,
				Requires2FASetup:       s.Requires2FASetup,
				RequiresPasswordChange: s.RequiresPasswordChange,
				CreatedAt:              s.CreatedAt.Format(time.RFC3339),
				LastActive:             s.LastActive.Format(time.RFC3339),
				Expiry:                 s.Expiry.Format(time.RFC3339),
				TTLSeconds:             ttl,
			})
		}
	}
	if list == nil {
		list = []ActiveSessionItem{}
	}
	return list
}

func (sm *SessionManager) UpdateActivity(token string) {
	tokenHash := hashToken(token)
	now := time.Now()

	sm.mu.RLock()
	var createdAt time.Time
	if c, ok := sm.cache[tokenHash]; ok {
		createdAt = c.session.CreatedAt
	} else if s, ok := sm.sessions[token]; ok {
		createdAt = s.CreatedAt
	}
	sm.mu.RUnlock()

	if createdAt.IsZero() {
		if db := sm.getDB(); db != nil {
			_ = db.QueryRow(`SELECT created_at FROM admin_sessions WHERE token_hash = $1`, tokenHash).Scan(&createdAt)
		}
	}

	if createdAt.IsZero() {
		createdAt = now
	}

	maxExpiry := createdAt.Add(12 * time.Hour)
	if now.After(maxExpiry) {
		sm.Delete(token)
		return
	}

	newExpiry := now.Add(15 * time.Minute)
	if newExpiry.After(maxExpiry) {
		newExpiry = maxExpiry
	}

	if db := sm.getDB(); db != nil {
		_, _ = db.Exec(`UPDATE admin_sessions SET last_active = $1, expiry = $2 WHERE token_hash = $3`, now, newExpiry, tokenHash)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if c, ok := sm.cache[tokenHash]; ok {
		c.session.LastActive = now
		c.session.Expiry = newExpiry
		c.cachedUntil = now.Add(60 * time.Second)
		sm.cache[tokenHash] = c
	}
	if s, ok := sm.sessions[token]; ok {
		s.LastActive = now
		s.Expiry = newExpiry
		sm.sessions[token] = s
	}
}

func (sm *SessionManager) Delete(token string) {
	tokenHash := hashToken(token)

	if db := sm.getDB(); db != nil {
		_, _ = db.Exec(`DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
		_, _ = db.Exec(`DELETE FROM admin_session_kicks WHERE token_hash = $1`, tokenHash)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.cache, tokenHash)
	delete(sm.sessions, token)
	delete(sm.kickedTokens, token)
}

func (sm *SessionManager) ActiveCount() int {
	if db := sm.getDB(); db != nil {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM admin_sessions WHERE expiry > NOW()`).Scan(&count)
		if err == nil {
			return count
		}
	}

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, s := range sm.sessions {
		if now.Before(s.Expiry) {
			count++
		}
	}
	return count
}

func (sm *SessionManager) StartReaper(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			now := time.Now()

			if db := sm.getDB(); db != nil {
				_, _ = db.Exec(`DELETE FROM admin_sessions WHERE expiry <= NOW()`)
				_, _ = db.Exec(`DELETE FROM admin_session_kicks WHERE created_at < NOW() - INTERVAL '1 day'`)
			}

			sm.mu.Lock()
			for th, c := range sm.cache {
				if now.After(c.cachedUntil) || now.After(c.session.Expiry) {
					delete(sm.cache, th)
				}
			}
			for token, session := range sm.sessions {
				if now.After(session.Expiry) {
					delete(sm.sessions, token)
				}
			}
			sm.mu.Unlock()
		}
	}()
}

type visitor struct {
	tokens     float64
	lastSeen   time.Time
	lastRefill time.Time
}

type IPRateLimiter struct {
	mu          sync.Mutex
	visitors    map[string]*visitor
	rate        float64
	capacity    float64
	maxVisitors int
}

func NewIPRateLimiter(rate float64, capacity float64) *IPRateLimiter {
	limiter := &IPRateLimiter{
		visitors:    make(map[string]*visitor),
		rate:        rate,
		capacity:    capacity,
		maxVisitors: 10000,
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			limiter.mu.Lock()
			now := time.Now()
			for ip, v := range limiter.visitors {
				if now.Sub(v.lastSeen) > 10*time.Minute {
					delete(limiter.visitors, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()

	return limiter
}

func (limiter *IPRateLimiter) evictStaleLocked(now time.Time) {

	for ip, v := range limiter.visitors {
		if now.Sub(v.lastSeen) > 3*time.Minute {
			delete(limiter.visitors, ip)
		}
	}

	if len(limiter.visitors) >= limiter.maxVisitors {
		type ipSeen struct {
			ip       string
			lastSeen time.Time
		}
		entries := make([]ipSeen, 0, len(limiter.visitors))
		for ip, v := range limiter.visitors {
			entries = append(entries, ipSeen{ip: ip, lastSeen: v.lastSeen})
		}

		toEvict := len(entries) / 5
		if toEvict < 1 {
			toEvict = 1
		}
		for i := 0; i < toEvict && i < len(entries); i++ {
			delete(limiter.visitors, entries[i].ip)
		}
	}
}

func (limiter *IPRateLimiter) Allow(ip string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	v, exists := limiter.visitors[ip]
	if !exists {

		if len(limiter.visitors) >= limiter.maxVisitors {
			limiter.evictStaleLocked(now)
		}

		limiter.visitors[ip] = &visitor{
			tokens:     limiter.capacity - 1.0,
			lastSeen:   now,
			lastRefill: now,
		}
		return true
	}

	v.lastSeen = now
	elapsed := now.Sub(v.lastRefill).Seconds()
	v.lastRefill = now
	v.tokens += elapsed * limiter.rate
	if v.tokens > limiter.capacity {
		v.tokens = limiter.capacity
	}

	if v.tokens >= 1.0 {
		v.tokens -= 1.0
		return true
	}

	return false
}

func RateLimitMiddleware(limiter *IPRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		if !limiter.Allow(ip) {
			http.Error(w, "تعداد درخواست‌های شما بیش از حد مجاز است. لطفاً کمی صبر کنید.", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

type cachedAccountFailure struct {
	consecutiveFailures int
	lockedUntil         time.Time
	cachedAt            time.Time
}

type AccountLockoutManager struct {
	mu       sync.Mutex
	db       *sql.DB
	initOnce sync.Once
	accounts map[string]*cachedAccountFailure
}

func NewAccountLockoutManager(db *sql.DB) *AccountLockoutManager {
	return &AccountLockoutManager{
		db:       db,
		accounts: make(map[string]*cachedAccountFailure),
	}
}

var Lockouts = NewAccountLockoutManager(nil)

func (alm *AccountLockoutManager) init(db *sql.DB) {
	alm.initOnce.Do(func() {
		alm.mu.Lock()
		alm.db = db
		alm.mu.Unlock()
	})
}

func (alm *AccountLockoutManager) getDB() *sql.DB {
	alm.mu.Lock()
	db := alm.db
	alm.mu.Unlock()
	if db != nil {
		return db
	}
	if alm != Lockouts {
		panic("FATAL: AccountLockoutManager called before database initialization")
	}
	if DB != nil {
		alm.init(DB)
		return DB
	}
	if os.Getenv("HERASAT_DEV_MODE") == "1" || flag.Lookup("test.v") != nil {
		return nil
	}
	panic("FATAL: AccountLockoutManager called before database initialization")
}

func (alm *AccountLockoutManager) IsLocked(username string, ip ...string) (bool, time.Duration) {
	now := time.Now()
	ipStr := ""
	if len(ip) > 0 {
		ipStr = strings.TrimSpace(ip[0])
	}
	key := username
	if ipStr != "" {
		key = username + "@" + ipStr
	}

	alm.mu.Lock()
	cached, exists := alm.accounts[key]
	if exists && now.Sub(cached.cachedAt) < 1*time.Second {
		if now.Before(cached.lockedUntil) {
			dur := time.Until(cached.lockedUntil)
			alm.mu.Unlock()
			return true, dur
		}
		if !cached.lockedUntil.IsZero() {
			delete(alm.accounts, key)
		}
		alm.mu.Unlock()
		return false, 0
	}
	alm.mu.Unlock()

	if db := alm.getDB(); db != nil {
		var failures int
		var lockedUntil sql.NullTime
		var err error
		if ipStr != "" {
			err = db.QueryRow(`SELECT consecutive_failures, locked_until FROM admin_login_attempts WHERE username = $1 AND ip = $2`, username, ipStr).
				Scan(&failures, &lockedUntil)
		} else {
			err = db.QueryRow(`SELECT consecutive_failures, locked_until FROM admin_login_attempts WHERE username = $1 ORDER BY consecutive_failures DESC LIMIT 1`, username).
				Scan(&failures, &lockedUntil)
		}
		if err == nil {
			var lTime time.Time
			if lockedUntil.Valid {
				lTime = lockedUntil.Time
			}

			alm.mu.Lock()
			alm.accounts[key] = &cachedAccountFailure{
				consecutiveFailures: failures,
				lockedUntil:         lTime,
				cachedAt:            now,
			}
			alm.mu.Unlock()

			if lockedUntil.Valid && now.Before(lTime) {
				return true, time.Until(lTime)
			}
			return false, 0
		}
		if err == sql.ErrNoRows {
			alm.mu.Lock()
			delete(alm.accounts, key)
			alm.mu.Unlock()
			return false, 0
		}
	}

	alm.mu.Lock()
	defer alm.mu.Unlock()

	if info, ok := alm.accounts[key]; ok {
		if now.Before(info.lockedUntil) {
			return true, time.Until(info.lockedUntil)
		}
		if !info.lockedUntil.IsZero() {
			delete(alm.accounts, key)
		}
	}
	return false, 0
}

func (alm *AccountLockoutManager) RecordFailure(username string, ip ...string) (consecutiveFailures int, justLocked bool) {
	now := time.Now()
	ipStr := ""
	if len(ip) > 0 {
		ipStr = strings.TrimSpace(ip[0])
	}
	key := username
	if ipStr != "" {
		key = username + "@" + ipStr
	}

	if db := alm.getDB(); db != nil {
		var failures int
		var lockedUntil sql.NullTime

		err := db.QueryRow(`
			INSERT INTO admin_login_attempts (username, ip, consecutive_failures, locked_until, updated_at)
			VALUES ($1, $2, 1, NULL, NOW())
			ON CONFLICT (username, ip) DO UPDATE SET
				consecutive_failures = admin_login_attempts.consecutive_failures + 1,
				locked_until = CASE
					WHEN admin_login_attempts.consecutive_failures + 1 >= 5 THEN NOW() + INTERVAL '15 minutes'
					ELSE admin_login_attempts.locked_until
				END,
				updated_at = NOW()
			RETURNING consecutive_failures, locked_until`, username, ipStr).Scan(&failures, &lockedUntil)

		if err == nil {
			var lTime time.Time
			if lockedUntil.Valid {
				lTime = lockedUntil.Time
			}
			justLocked = (failures == 5)

			alm.mu.Lock()
			alm.accounts[key] = &cachedAccountFailure{
				consecutiveFailures: failures,
				lockedUntil:         lTime,
				cachedAt:            now,
			}
			alm.mu.Unlock()

			return failures, justLocked
		}
	}

	alm.mu.Lock()
	defer alm.mu.Unlock()

	info, exists := alm.accounts[key]
	if !exists {
		info = &cachedAccountFailure{cachedAt: now}
		alm.accounts[key] = info
	}

	info.consecutiveFailures++
	info.cachedAt = now
	if info.consecutiveFailures >= 5 {
		info.lockedUntil = now.Add(15 * time.Minute)
		return info.consecutiveFailures, true
	}

	return info.consecutiveFailures, false
}

func (alm *AccountLockoutManager) Reset(username string, ip ...string) {
	ipStr := ""
	if len(ip) > 0 {
		ipStr = strings.TrimSpace(ip[0])
	}

	if db := alm.getDB(); db != nil {
		if ipStr != "" {
			_, _ = db.Exec(`DELETE FROM admin_login_attempts WHERE username = $1 AND ip = $2`, username, ipStr)
		} else {
			_, _ = db.Exec(`DELETE FROM admin_login_attempts WHERE username = $1`, username)
		}
	}

	alm.mu.Lock()
	defer alm.mu.Unlock()

	if ipStr != "" {
		delete(alm.accounts, username+"@"+ipStr)
	} else {
		for k := range alm.accounts {
			if k == username || strings.HasPrefix(k, username+"@") {
				delete(alm.accounts, k)
			}
		}
	}
}

func ProgressiveDelay(failures int) {
	switch {
	case failures <= 1:
		time.Sleep(250 * time.Millisecond)
	case failures == 2:
		time.Sleep(500 * time.Millisecond)
	case failures == 3:
		time.Sleep(1 * time.Second)
	default:
		time.Sleep(2 * time.Second)
	}
}

var (
	cachedCookieSecure   bool
	cookieSecureExplicit bool
	cookieSecureOnce     sync.Once
)

func InitCookieSecurity(secure bool) {
	cachedCookieSecure = secure
	cookieSecureExplicit = true
}

func ResetCookieSecurityForTest() {
	cachedCookieSecure = false
	cookieSecureExplicit = false
	cookieSecureOnce = sync.Once{}
}

func isCookieSecure() bool {
	if !cookieSecureExplicit {
		cookieSecureOnce.Do(func() {
			val := strings.ToLower(strings.TrimSpace(os.Getenv("HERASAT_COOKIE_SECURE")))
			cachedCookieSecure = val == "true" || val == "1"
		})
	}
	return cachedCookieSecure
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func stripNul(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

func sanitizePhone(s string) string {
	var sb strings.Builder
	for _, ch := range s {
		if (ch >= '0' && ch <= '9') || ch == '+' || ch == '-' || ch == '(' || ch == ')' || ch == ' ' {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

func isValidTrackingCode(code string) bool {
	if len(code) == 0 || len(code) > 32 {
		return false
	}
	for _, ch := range code {
		if !((ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}
