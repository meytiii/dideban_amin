package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var auditFailures uint64

func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		CaptchaID     string `json:"captcha_id"`
		CaptchaAnswer string `json:"captcha_answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !VerifyCaptcha(creds.CaptchaID, creds.CaptchaAnswer) {
		http.Error(w, "کد امنیتی اشتباه است یا منقضی شده است.", http.StatusBadRequest)
		return
	}

	creds.Username = strings.TrimSpace(creds.Username)
	creds.Password = strings.TrimSpace(creds.Password)

	if isLocked, remaining := Lockouts.IsLocked(creds.Username); isLocked {
		log.Printf("-> LOGIN BLOCKED: Account '%s' is locked. Remaining: %v", creds.Username, remaining)
		logAudit(creds.Username, "UNKNOWN", "LOCKED_LOGIN_ATTEMPT", nil, getIP(r))
		minutes := int(remaining.Minutes()) + 1
		http.Error(w, fmt.Sprintf("حساب کاربری شما به دلیل تلاش‌های ناموفق متعدد به مدت %d دقیقه موقتاً مسدود شده است.", minutes), http.StatusTooManyRequests)
		return
	}

	log.Printf("-> LOGIN ATTEMPT: Username='%s'", creds.Username)

	var storedHash, fullName, roleTitle, role, totpSecret string
	var isSuperuser, totpEnabled bool
	var passwordChangedAt, dbNow time.Time
	var mustChangePassword bool
	var lastLoginAt sql.NullTime
	err := DB.QueryRow(`
		SELECT password_hash, COALESCE(full_name, ''), role_title, COALESCE(role, 'edare_kol'), is_superuser,
		       COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false),
		       COALESCE(totp_secret, ''), COALESCE(totp_enabled, false), last_login_at, NOW()
		FROM admin_users
		WHERE username=$1 AND is_active=true`, creds.Username).Scan(
		&storedHash, &fullName, &roleTitle, &role, &isSuperuser, &passwordChangedAt, &mustChangePassword,
		&totpSecret, &totpEnabled, &lastLoginAt, &dbNow)
	if err != nil {
		log.Printf("-> LOGIN FAILED: Database error or user not found: %v", err)
		if _, justLocked := Lockouts.RecordFailure(creds.Username); justLocked {
			logAudit(creds.Username, "UNKNOWN", "ACCOUNT_LOCKED", nil, getIP(r))
		}
		logAudit(creds.Username, "UNKNOWN", "FAILED_LOGIN", nil, getIP(r))
		http.Error(w, "نام کاربری یا رمز عبور اشتباه است.", http.StatusUnauthorized)
		return
	}

	ok, needsRehash := VerifyPasswordWithRehash(creds.Password, storedHash)
	if !ok {
		log.Println("-> LOGIN FAILED: Password hash mismatch")
		if _, justLocked := Lockouts.RecordFailure(creds.Username); justLocked {
			logAudit(creds.Username, "UNKNOWN", "ACCOUNT_LOCKED", nil, getIP(r))
		}
		logAudit(creds.Username, "UNKNOWN", "FAILED_LOGIN", nil, getIP(r))
		http.Error(w, "نام کاربری یا رمز عبور اشتباه است.", http.StatusUnauthorized)
		return
	}

	if lastLoginAt.Valid && time.Since(lastLoginAt.Time) > 90*24*time.Hour {
		_, _ = DB.Exec(`UPDATE admin_users SET is_active = false WHERE username = $1`, creds.Username)
		logAudit(creds.Username, roleTitle, "ACCOUNT_DORMANT_AUTO_LOCKED", nil, getIP(r))
		http.Error(w, "حساب کاربری شما به دلیل عدم فعالیت بیش از ۹۰ روز به صورت خودکار غیرفعال شده است. لطفاً با مدیر ارشد تماس بگیرید.", http.StatusForbidden)
		return
	}

	if needsRehash {
		if newHash, err := HashPassword(creds.Password); err == nil {
			_, _ = DB.Exec(`UPDATE admin_users SET password_hash = $1 WHERE username = $2`, newHash, creds.Username)
			logAudit(creds.Username, roleTitle, "PASSWORD_REHASHED", nil, getIP(r))
		}
	}

	if totpEnabled && totpSecret != "" {
		mfaToken, err := GenerateMFAPendingToken(creds.Username)
		if err != nil {
			http.Error(w, "خطای تولید شناسه ورود دومرحله‌ای", http.StatusInternalServerError)
			return
		}

		logAudit(creds.Username, roleTitle, "MFA_REQUIRED", nil, getIP(r))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "mfa_required",
			"mfa_token": mfaToken,
		})
		return
	}

	isPasswordExpired := time.Since(passwordChangedAt) > 30*24*time.Hour
	requiresPasswordChange := mustChangePassword || isPasswordExpired

	Lockouts.Reset(creds.Username)
	_, _ = DB.Exec(`UPDATE admin_users SET last_login_at = NOW() WHERE username = $1`, creds.Username)

	sessionToken := generateSessionToken()
	now := time.Now()
	sessionExpiry := now.Add(15 * time.Minute)
	Sessions.Set(sessionToken, AdminSession{
		Token:       sessionToken,
		Username:    creds.Username,
		FullName:    fullName,
		Title:       roleTitle,
		Role:        role,
		IsSuperuser: isSuperuser,
		IP:          getIP(r),
		UserAgent:   r.UserAgent(),
		CreatedAt:   now,
		LastActive:  now,
		Expiry:      sessionExpiry,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "herasat_admin_session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  sessionExpiry,
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteLaxMode,
	})

	csrfBytes := make([]byte, 32)
	rand.Read(csrfBytes)
	csrfToken := base64.RawURLEncoding.EncodeToString(csrfBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     "herasat_csrf",
		Value:    csrfToken,
		Path:     "/",
		Expires:  sessionExpiry,
		HttpOnly: false,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteStrictMode,
	})

	log.Println("-> LOGIN SUCCESS: Session cookie issued")
	logAudit(creds.Username, roleTitle, "LOGIN", nil, getIP(r))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Time", dbNow.Format(time.RFC3339Nano))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "login_successful",
		"username":             creds.Username,
		"full_name":            fullName,
		"role_title":           roleTitle,
		"role":                 role,
		"is_superuser":         isSuperuser,
		"must_change_password": requiresPasswordChange,
		"is_password_expired":  isPasswordExpired,
		"password_changed_at":  passwordChangedAt.Format(time.RFC3339),
		"csrf_token":           csrfToken,
		"server_time":          dbNow.Format(time.RFC3339Nano),
		"server_time_unix_ms":  dbNow.UnixMilli(),
	})
}

func GetCurrentAdminHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var passwordChangedAt, dbNow time.Time
	var mustChangePassword bool
	var dbRole string
	var totpEnabled bool
	var err error = fmt.Errorf("db unavailable")
	if DB != nil {
		err = DB.QueryRow(`
		SELECT COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false), COALESCE(role, 'edare_kol'), COALESCE(totp_enabled, false), NOW()
		FROM admin_users
		WHERE username = $1`, admin.Username).Scan(&passwordChangedAt, &mustChangePassword, &dbRole, &totpEnabled, &dbNow)
	}
	if err != nil {
		dbNow = time.Now()
	}

	if dbRole != "" {
		admin.Role = dbRole
		admin.IsSuperuser = (dbRole == "superuser")
	}

	isPasswordExpired := time.Since(passwordChangedAt) > 30*24*time.Hour
	requiresPasswordChange := mustChangePassword || isPasswordExpired

	csrfToken := ""
	if csrfCookie, err := r.Cookie("herasat_csrf"); err == nil && csrfCookie.Value != "" {
		csrfToken = csrfCookie.Value
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Time", dbNow.Format(time.RFC3339Nano))
	if csrfToken != "" {
		w.Header().Set("X-CSRF-Token", csrfToken)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":             admin.Username,
		"full_name":            admin.FullName,
		"role_title":           admin.Title,
		"role":                 admin.Role,
		"is_superuser":         admin.IsSuperuser,
		"must_change_password": requiresPasswordChange,
		"is_password_expired":  isPasswordExpired,
		"password_changed_at":  passwordChangedAt.Format(time.RFC3339),
		"totp_enabled":         totpEnabled,
		"server_time":          dbNow.Format(time.RFC3339Nano),
		"server_time_unix_ms":  dbNow.UnixMilli(),
		"csrf_token":           csrfToken,
	})
}

func parseTrustedProxies(envVal string) []*net.IPNet {
	if strings.TrimSpace(envVal) == "" {
		return nil
	}
	var nets []*net.IPNet
	for _, cidr := range strings.Split(envVal, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet != nil {
			nets = append(nets, ipNet)
		}
	}
	return nets
}

func ipInCIDRs(ip net.IP, cidrs []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func getIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	remoteIP := net.ParseIP(strings.TrimSpace(remoteHost))

	trustedNets := parseTrustedProxies(os.Getenv("HERASAT_TRUSTED_PROXIES"))
	if len(trustedNets) == 0 || !ipInCIDRs(remoteIP, trustedNets) {
		return remoteHost
	}

	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		hops := strings.Split(xff, ",")
		for i := len(hops) - 1; i >= 0; i-- {
			hop := strings.TrimSpace(hops[i])
			hopIP := net.ParseIP(hop)
			if hopIP != nil && !ipInCIDRs(hopIP, trustedNets) {
				return hop
			}
		}
	}

	if xReal := strings.TrimSpace(r.Header.Get("X-Real-IP")); xReal != "" {
		xRealIP := net.ParseIP(xReal)
		if xRealIP != nil && !ipInCIDRs(xRealIP, trustedNets) {
			return xReal
		}
	}

	return remoteHost
}

func computeAuditHash(prevHash, username, title, action string, targetID *int64, ip string, t time.Time) string {
	targetStr := "none"
	if targetID != nil {
		targetStr = strconv.FormatInt(*targetID, 10)
	}
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s", prevHash, username, title, action, targetStr, ip, t.Format(time.RFC3339Nano))
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func getLatestAuditHash(db *sql.DB) string {
	var h string
	err := db.QueryRow("SELECT COALESCE(entry_hash, '') FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&h)
	if err != nil || h == "" {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return h
}

func getLatestAuditHashTx(tx *sql.Tx) string {
	var h string
	err := tx.QueryRow("SELECT COALESCE(entry_hash, '') FROM audit_logs ORDER BY id DESC LIMIT 1").Scan(&h)
	if err != nil || h == "" {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return h
}

func logAudit(username, title, action string, targetID *int64, ip string) error {
	ShipSyslogAudit(username, action, targetID, ip)
	if DB == nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG FAILURE: DB is nil (user=%s, action=%s)", username, action)
		return fmt.Errorf("DB is nil")
	}
	now := time.Now().UTC()
	prevHash := getLatestAuditHash(DB)
	entryHash := computeAuditHash(prevHash, username, title, action, targetID, ip, now)
	_, err := DB.Exec(`INSERT INTO audit_logs (admin_username, admin_title, action, target_report_id, ip_address, prev_hash, entry_hash, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		username, title, action, targetID, ip, prevHash, entryHash, now)
	if err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG FAILURE: %v (user=%s, action=%s)", err, username, action)
		return err
	}
	return nil
}

func logAuditTx(tx *sql.Tx, username, title, action string, targetID *int64, ip string) error {
	ShipSyslogAudit(username, action, targetID, ip)
	if tx == nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG TX FAILURE: tx is nil (user=%s, action=%s)", username, action)
		return fmt.Errorf("tx is nil")
	}
	now := time.Now().UTC()
	prevHash := getLatestAuditHashTx(tx)
	entryHash := computeAuditHash(prevHash, username, title, action, targetID, ip, now)
	_, err := tx.Exec(`INSERT INTO audit_logs (admin_username, admin_title, action, target_report_id, ip_address, prev_hash, entry_hash, created_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		username, title, action, targetID, ip, prevHash, entryHash, now)
	if err != nil {
		atomic.AddUint64(&auditFailures, 1)
		log.Printf("CRITICAL AUDIT LOG TX FAILURE: %v (user=%s, action=%s)", err, username, action)
		return err
	}
	return nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("herasat_admin_session")
	if err == nil && cookie.Value != "" {
		if session, exists := Sessions.Get(cookie.Value); exists {
			logAudit(session.Username, session.Title, "LOGOUT", nil, getIP(r))
		}
		Sessions.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "herasat_admin_session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "herasat_csrf",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   isCookieSecure(),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "logged_out"}`))
}
