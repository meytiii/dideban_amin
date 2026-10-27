package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	dummyArgon2Hash string
)

func init() {
	dummyArgon2Hash, _ = HashPassword("DummyTimingDefensePassword#2026!")
}

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

	creds.Username = strings.ToLower(strings.TrimSpace(creds.Username))
	creds.Password = strings.TrimSpace(creds.Password)
	clientIP := getIP(r)
	safeUsername := truncateRunes(strings.ReplaceAll(strings.ReplaceAll(creds.Username, "\n", ""), "\r", ""), 32)

	if isLocked, remaining := Lockouts.IsLocked(creds.Username, clientIP); isLocked {
		logAudit(safeUsername, "UNKNOWN", "LOCKED_LOGIN_ATTEMPT", nil, clientIP)
		minutes := int(remaining.Minutes()) + 1
		http.Error(w, fmt.Sprintf("حساب کاربری شما به دلیل تلاش‌های ناموفق متعدد به مدت %d دقیقه موقتاً مسدود شده است.", minutes), http.StatusTooManyRequests)
		return
	}

	var storedHash, fullName, roleTitle, role, totpSecret string
	var isSuperuser, totpEnabled, totpEnforced bool
	var passwordChangedAt, dbNow time.Time
	var mustChangePassword bool
	var lastLoginAt sql.NullTime
	err := DB.QueryRow(`
		SELECT password_hash, COALESCE(full_name, ''), role_title, COALESCE(role, 'edare_kol'), is_superuser,
		       COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false),
		       COALESCE(totp_secret, ''), COALESCE(totp_enabled, false), COALESCE(totp_enforced, false),
		       last_login_at, NOW()
		FROM admin_users
		WHERE username=$1 AND is_active=true`, creds.Username).Scan(
		&storedHash, &fullName, &roleTitle, &role, &isSuperuser, &passwordChangedAt, &mustChangePassword,
		&totpSecret, &totpEnabled, &totpEnforced, &lastLoginAt, &dbNow)
	if err != nil {
		_ = VerifyPassword(creds.Password, dummyArgon2Hash)
		ProgressiveDelay(1)
		logAudit(safeUsername, "UNKNOWN", "FAILED_LOGIN", nil, clientIP)
		http.Error(w, "نام کاربری یا رمز عبور اشتباه است.", http.StatusUnauthorized)
		return
	}

	ok, needsRehash := VerifyPasswordWithRehash(creds.Password, storedHash)
	if !ok {
		failures, justLocked := Lockouts.RecordFailure(creds.Username, clientIP)
		ProgressiveDelay(failures)
		if justLocked {
			logAudit(safeUsername, "UNKNOWN", "ACCOUNT_LOCKED", nil, clientIP)
		}
		logAudit(safeUsername, "UNKNOWN", "FAILED_LOGIN", nil, clientIP)
		http.Error(w, "نام کاربری یا رمز عبور اشتباه است.", http.StatusUnauthorized)
		return
	}

	if lastLoginAt.Valid && time.Since(lastLoginAt.Time) > 90*24*time.Hour {
		_, _ = DB.Exec(`UPDATE admin_users SET is_active = false WHERE username = $1`, creds.Username)
		logAudit(safeUsername, roleTitle, "ACCOUNT_DORMANT_AUTO_LOCKED", nil, clientIP)
		http.Error(w, "حساب کاربری شما به دلیل عدم فعالیت بیش از ۹۰ روز به صورت خودکار غیرفعال شده است. لطفاً با مدیر ارشد تماس بگیرید.", http.StatusForbidden)
		return
	}

	if needsRehash {
		if newHash, err := HashPassword(creds.Password); err == nil {
			_, _ = DB.Exec(`UPDATE admin_users SET password_hash = $1 WHERE username = $2`, newHash, creds.Username)
			logAudit(safeUsername, roleTitle, "PASSWORD_REHASHED", nil, clientIP)
		}
	}

	if totpEnabled && totpSecret != "" {
		mfaToken, err := GenerateMFAPendingToken(creds.Username, clientIP, r.UserAgent())
		if err != nil {
			http.Error(w, "خطای تولید شناسه ورود دومرحله‌ای", http.StatusInternalServerError)
			return
		}

		logAudit(safeUsername, roleTitle, "MFA_REQUIRED", nil, clientIP)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "mfa_required",
			"mfa_token": mfaToken,
		})
		return
	}

	require2FA := isSuperuser || totpEnforced || os.Getenv("HERASAT_REQUIRE_2FA") == "1" || os.Getenv("HERASAT_REQUIRE_2FA") == "true"
	isPasswordExpired := time.Since(passwordChangedAt) > 30*24*time.Hour
	requiresPasswordChange := mustChangePassword || isPasswordExpired

	Lockouts.Reset(creds.Username, clientIP)
	_, _ = DB.Exec(`UPDATE admin_users SET last_login_at = NOW() WHERE username = $1`, creds.Username)

	sessionToken := generateSessionToken()
	now := time.Now()
	sessionExpiry := now.Add(15 * time.Minute)
	Sessions.Set(sessionToken, AdminSession{
		Token:                  sessionToken,
		Username:               creds.Username,
		FullName:               fullName,
		Title:                  roleTitle,
		Role:                   role,
		IsSuperuser:            isSuperuser,
		IP:                     clientIP,
		UserAgent:              r.UserAgent(),
		Requires2FASetup:       require2FA,
		RequiresPasswordChange: requiresPasswordChange,
		CreatedAt:              now,
		LastActive:             now,
		Expiry:                 sessionExpiry,
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

	logAudit(safeUsername, roleTitle, "LOGIN", nil, clientIP)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Time", dbNow.Format(time.RFC3339Nano))
	w.WriteHeader(http.StatusOK)

	statusStr := "login_successful"
	if require2FA {
		statusStr = "2fa_setup_required"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               statusStr,
		"username":             creds.Username,
		"full_name":            fullName,
		"role_title":           roleTitle,
		"role":                 role,
		"is_superuser":         isSuperuser,
		"requires_2fa_setup":   require2FA,
		"must_change_password": requiresPasswordChange,
		"is_password_expired":  isPasswordExpired,
		"password_changed_at":  passwordChangedAt.Format(time.RFC3339),
		"csrf_token":           csrfToken,
		"server_time":          dbNow.Format(time.RFC3339Nano),
		"server_time_unix_ms":  dbNow.UnixMilli(),
	})
}

func GetCurrentAdminHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := adminFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var passwordChangedAt, dbNow time.Time
	var mustChangePassword bool
	var dbRole string
	var totpEnabled, totpEnforced bool
	var err error = fmt.Errorf("db unavailable")
	if DB != nil {
		err = DB.QueryRow(`
		SELECT COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false), COALESCE(role, 'edare_kol'), COALESCE(totp_enabled, false), COALESCE(totp_enforced, false), NOW()
		FROM admin_users
		WHERE username = $1`, admin.Username).Scan(&passwordChangedAt, &mustChangePassword, &dbRole, &totpEnabled, &totpEnforced, &dbNow)
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
	require2FA := admin.IsSuperuser || totpEnforced || os.Getenv("HERASAT_REQUIRE_2FA") == "1" || os.Getenv("HERASAT_REQUIRE_2FA") == "true"
	requires2FASetup := admin.Requires2FASetup || (require2FA && !totpEnabled)

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
		"requires_2fa_setup":   requires2FASetup,
		"must_change_password": requiresPasswordChange,
		"is_password_expired":  isPasswordExpired,
		"password_changed_at":  passwordChangedAt.Format(time.RFC3339),
		"totp_enabled":         totpEnabled,
		"server_time":          dbNow.Format(time.RFC3339Nano),
		"server_time_unix_ms":  dbNow.UnixMilli(),
		"csrf_token":           csrfToken,
	})
}

func getIP(r *http.Request) string {
	remoteHost, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		remoteHost = r.RemoteAddr
	}
	remoteIP := net.ParseIP(strings.TrimSpace(remoteHost))

	trustedNets := GetTrustedProxies()
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
