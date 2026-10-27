package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidTOTPCode  = errors.New("invalid totp code")
	ErrTOTPReplayed     = errors.New("totp code already used")
	ErrInvalidMFAToken  = errors.New("invalid or expired mfa token")
	ErrMFAScopeMismatch = errors.New("invalid mfa token scope")
)

type MFAPendingPayload struct {
	Username  string `json:"username"`
	Scope     string `json:"scope"`
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
	Nonce     string `json:"nonce"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
}

type TOTPReplayStore struct {
	mu    sync.Mutex
	used  map[string]time.Time
	reapT *time.Ticker
}

var totpReplays = &TOTPReplayStore{
	used: make(map[string]time.Time),
}

var (
	mfaNoncesMu   sync.Mutex
	mfaUsedNonces = make(map[string]time.Time)
)

func init() {
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			now := time.Now()
			totpReplays.mu.Lock()
			for k, exp := range totpReplays.used {
				if now.After(exp) {
					delete(totpReplays.used, k)
				}
			}
			totpReplays.mu.Unlock()

			mfaNoncesMu.Lock()
			for k, exp := range mfaUsedNonces {
				if now.After(exp) {
					delete(mfaUsedNonces, k)
				}
			}
			mfaNoncesMu.Unlock()
		}
	}()
}

func GenerateBase32Secret() (string, error) {
	bytesArr := make([]byte, 20)
	if _, err := rand.Read(bytesArr); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytesArr), nil
}

func normalizeTOTPCode(code string) string {
	var sb strings.Builder
	for _, r := range code {
		if r >= '0' && r <= '9' {
			sb.WriteRune(r)
		} else if r >= '۰' && r <= '۹' {
			sb.WriteRune(r - '۰' + '0')
		} else if r >= '٠' && r <= '٩' {
			sb.WriteRune(r - '٠' + '0')
		}
	}
	return sb.String()
}

func GenerateTOTPCode(secretBase32 string, timestep int64) (string, error) {
	cleanSecret := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secretBase32), " ", ""))
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleanSecret)
	if err != nil {
		secretBytes, err = base32.StdEncoding.DecodeString(cleanSecret)
		if err != nil {
			return "", err
		}
	}

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(timestep))

	mac := hmac.New(sha1.New, secretBytes)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	binCode := (int(h[offset])&0x7f)<<24 |
		(int(h[offset+1])&0xff)<<16 |
		(int(h[offset+2])&0xff)<<8 |
		(int(h[offset+3]) & 0xff)

	otp := binCode % 1000000
	return fmt.Sprintf("%06d", otp), nil
}

func VerifyTOTPCode(username, secretBase32, code string) (bool, error) {
	code = normalizeTOTPCode(code)
	if len(code) != 6 {
		return false, ErrInvalidTOTPCode
	}

	cleanSecret := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(secretBase32), " ", ""))

	now := time.Now().Unix()
	currentStep := now / 30

	totpReplays.mu.Lock()
	defer totpReplays.mu.Unlock()

	for stepOffset := int64(-1); stepOffset <= 1; stepOffset++ {
		step := currentStep + stepOffset
		expectedCode, err := GenerateTOTPCode(cleanSecret, step)
		if err != nil {
			continue
		}

		if subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1 {
			replayKey := fmt.Sprintf("totp:%s:%d", username, step)
			if exp, exists := totpReplays.used[replayKey]; exists && time.Now().Before(exp) {
				return false, ErrTOTPReplayed
			}
			if DB != nil {
				var exists bool
				_ = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM used_mfa_nonces WHERE nonce = $1 AND expires_at > NOW())", replayKey).Scan(&exists)
				if exists {
					return false, ErrTOTPReplayed
				}
				_, _ = DB.Exec("INSERT INTO used_mfa_nonces (nonce, expires_at) VALUES ($1, NOW() + INTERVAL '120 seconds') ON CONFLICT (nonce) DO UPDATE SET expires_at = EXCLUDED.expires_at", replayKey)
			}

			totpReplays.used[replayKey] = time.Now().Add(120 * time.Second)
			return true, nil
		}
	}

	return false, ErrInvalidTOTPCode
}

func GenerateMFAPendingToken(username string, ipAndUA ...string) (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}

	ip := ""
	ua := ""
	if len(ipAndUA) > 0 {
		ip = ipAndUA[0]
	}
	if len(ipAndUA) > 1 {
		ua = ipAndUA[1]
	}

	now := time.Now()
	payload := MFAPendingPayload{
		Username:  username,
		Scope:     "mfa_pending",
		Iat:       now.Unix(),
		Exp:       now.Add(3 * time.Minute).Unix(),
		Nonce:     base64.RawURLEncoding.EncodeToString(nonceBytes),
		IP:        ip,
		UserAgent: ua,
	}

	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	b64Payload := base64.RawURLEncoding.EncodeToString(rawJSON)
	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(b64Payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return b64Payload + "." + sig, nil
}

func ValidateMFAPendingToken(token string, ipAndUA ...string) (*MFAPendingPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidMFAToken
	}

	clientIP := ""
	clientUA := ""
	if len(ipAndUA) > 0 {
		clientIP = ipAndUA[0]
	}
	if len(ipAndUA) > 1 {
		clientUA = ipAndUA[1]
	}

	b64Payload, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, GetHMACKey())
	mac.Write([]byte(b64Payload))
	expectedSig := mac.Sum(nil)

	actualSig, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return nil, ErrInvalidMFAToken
	}

	if subtle.ConstantTimeCompare(expectedSig, actualSig) != 1 {
		return nil, ErrInvalidMFAToken
	}

	rawJSON, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		return nil, ErrInvalidMFAToken
	}

	var payload MFAPendingPayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		return nil, ErrInvalidMFAToken
	}

	if payload.Scope != "mfa_pending" {
		return nil, ErrMFAScopeMismatch
	}

	if time.Now().Unix() > payload.Exp {
		return nil, ErrInvalidMFAToken
	}

	if payload.IP != "" && clientIP != "" && payload.IP != clientIP {
		return nil, ErrInvalidMFAToken
	}

	if payload.UserAgent != "" && clientUA != "" && payload.UserAgent != clientUA {
		return nil, ErrInvalidMFAToken
	}

	mfaNoncesMu.Lock()
	if exp, ok := mfaUsedNonces[payload.Nonce]; ok && time.Now().Before(exp) {
		mfaNoncesMu.Unlock()
		return nil, errors.New("mfa token already used")
	}
	mfaUsedNonces[payload.Nonce] = time.Unix(payload.Exp, 0)
	mfaNoncesMu.Unlock()

	if DB != nil {
		var exists bool
		_ = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM used_mfa_nonces WHERE nonce = $1)", payload.Nonce).Scan(&exists)
		if exists {
			return nil, errors.New("mfa token already used")
		}
		_, _ = DB.Exec("INSERT INTO used_mfa_nonces (nonce, expires_at) VALUES ($1, $2)", payload.Nonce, time.Unix(payload.Exp, 0))
	}

	return &payload, nil
}

func VerifyMFAHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MFAToken string `json:"mfa_token"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	clientIP := getIP(r)
	payload, err := ValidateMFAPendingToken(req.MFAToken, clientIP, r.UserAgent())
	if err != nil {
		http.Error(w, "توکن احراز هویت دومرحله‌ای منقضی شده یا نامعتبر است.", http.StatusUnauthorized)
		return
	}

	if isLocked, remaining := Lockouts.IsLocked(payload.Username, clientIP); isLocked {
		logAudit(payload.Username, "UNKNOWN", "LOCKED_MFA_ATTEMPT", nil, clientIP)
		minutes := int(remaining.Minutes()) + 1
		http.Error(w, fmt.Sprintf("حساب کاربری شما به دلیل تلاش‌های ناموفق متعدد به مدت %d دقیقه موقتاً مسدود شده است.", minutes), http.StatusTooManyRequests)
		return
	}

	var userID int64
	var storedHash, fullName, roleTitle, role, encSecret string
	var isSuperuser, totpEnabled bool
	var passwordChangedAt, dbNow time.Time
	var mustChangePassword bool

	err = DB.QueryRow(`
		SELECT id, password_hash, COALESCE(full_name, ''), role_title, COALESCE(role, 'edare_kol'), is_superuser,
		       COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false),
		       COALESCE(totp_secret, ''), COALESCE(totp_enabled, false), NOW()
		FROM admin_users
		WHERE username=$1 AND is_active=true`, payload.Username).Scan(
		&userID, &storedHash, &fullName, &roleTitle, &role, &isSuperuser, &passwordChangedAt, &mustChangePassword,
		&encSecret, &totpEnabled, &dbNow)
	if err != nil {
		http.Error(w, "کاربر یافت نشد یا حساب غیرفعال است.", http.StatusUnauthorized)
		return
	}

	if !totpEnabled || encSecret == "" {
		http.Error(w, "احراز هویت دومرحله‌ای برای این حساب فعال نیست.", http.StatusBadRequest)
		return
	}

	plainSecret, err := DecryptBoundField("admin_users", "totp_secret", userID, encSecret)
	if err != nil {
		http.Error(w, "خطای رمزگشایی کلید امنیتی دومرحله‌ای", http.StatusInternalServerError)
		return
	}

	ok, err := VerifyTOTPCode(payload.Username, plainSecret, req.TOTPCode)
	if !ok || err != nil {
		if err == ErrTOTPReplayed {
			http.Error(w, "این کد امنیتی قبلاً استفاده شده است. لطفاً منتظر تولید کد بعدی بمانید.", http.StatusConflict)
			return
		}
		failures, justLocked := Lockouts.RecordFailure(payload.Username, clientIP)
		ProgressiveDelay(failures)
		if justLocked {
			logAudit(payload.Username, roleTitle, "ACCOUNT_LOCKED", nil, clientIP)
		}
		logAudit(payload.Username, roleTitle, "FAILED_MFA_VERIFICATION", nil, clientIP)
		http.Error(w, "کد ۶ رقمی امنیتی وارد شده اشتباه است یا منقضی شده است.", http.StatusUnauthorized)
		return
	}

	isPasswordExpired := time.Since(passwordChangedAt) > 30*24*time.Hour
	requiresPasswordChange := mustChangePassword || isPasswordExpired

	Lockouts.Reset(payload.Username, clientIP)
	_, _ = DB.Exec("UPDATE admin_users SET last_login_at = NOW() WHERE username = $1", payload.Username)

	sessionToken := generateSessionToken()
	now := time.Now()
	sessionExpiry := now.Add(15 * time.Minute)
	Sessions.Set(sessionToken, AdminSession{
		Token:                  sessionToken,
		Username:               payload.Username,
		FullName:               fullName,
		Title:                  roleTitle,
		Role:                   role,
		IsSuperuser:            isSuperuser,
		IP:                     clientIP,
		UserAgent:              r.UserAgent(),
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

	logAudit(payload.Username, roleTitle, "LOGIN_MFA_SUCCESS", nil, clientIP)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Time", dbNow.Format(time.RFC3339Nano))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "login_successful",
		"username":             payload.Username,
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

func Setup2FAHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := adminFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	secret, err := GenerateBase32Secret()
	if err != nil {
		http.Error(w, "خطا در تولید کلید امنیتی", http.StatusInternalServerError)
		return
	}

	otpAuthURL := fmt.Sprintf("otpauth://totp/Herasat-HMD:%s?secret=%s&issuer=Herasat", admin.Username, secret)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"secret":      secret,
		"otpauth_url": otpAuthURL,
		"server_time": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func Enable2FAHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := adminFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Password string `json:"password"`
		Secret   string `json:"secret"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.Password = strings.TrimSpace(req.Password)
	req.Secret = strings.TrimSpace(req.Secret)
	req.TOTPCode = strings.TrimSpace(req.TOTPCode)

	if req.Password == "" || req.Secret == "" || req.TOTPCode == "" {
		http.Error(w, "کلمه عبور جاری، کلید امنیتی و کد تأیید الزامی است.", http.StatusBadRequest)
		return
	}

	clientIP := getIP(r)
	var userID int64
	var storedHash string
	err := DB.QueryRow(`SELECT id, password_hash FROM admin_users WHERE username=$1`, admin.Username).Scan(&userID, &storedHash)
	if err != nil {
		http.Error(w, "کاربر یافت نشد.", http.StatusUnauthorized)
		return
	}

	if !VerifyPassword(req.Password, storedHash) {
		failures, justLocked := Lockouts.RecordFailure(admin.Username, clientIP)
		ProgressiveDelay(failures)
		if justLocked {
			logAudit(admin.Username, admin.Title, "ACCOUNT_LOCKED", nil, clientIP)
		}
		http.Error(w, "کلمه عبور جاری اشتباه است.", http.StatusUnauthorized)
		return
	}

	valid, err := VerifyTOTPCode(admin.Username, req.Secret, req.TOTPCode)
	if !valid || err != nil {
		http.Error(w, "کد امنیتی تأیید نشد. لطفاً مجدداً امتحان کنید.", http.StatusBadRequest)
		return
	}

	encSecret, err := EncryptBoundField("admin_users", "totp_secret", userID, req.Secret)
	if err != nil {
		http.Error(w, "خطا در ذخیره‌سازی امن کلید", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`UPDATE admin_users SET totp_secret = $1, totp_enabled = true WHERE username = $2`, encSecret, admin.Username)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	if admin.Requires2FASetup {
		admin.Requires2FASetup = false
		Sessions.Set(admin.Token, admin)
	}

	logAudit(admin.Username, admin.Title, "ENABLE_2FA", nil, clientIP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "احراز هویت دومرحله‌ای با موفقیت فعال شد.",
	})
}

func Disable2FAHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := adminFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	clientIP := getIP(r)
	var userID int64
	var storedHash, encSecret string
	var totpEnabled, isSuperuser, totpEnforced bool
	err := DB.QueryRow(`
		SELECT id, password_hash, COALESCE(totp_secret, ''), COALESCE(totp_enabled, false), is_superuser, COALESCE(totp_enforced, false)
		FROM admin_users WHERE username=$1`, admin.Username).Scan(&userID, &storedHash, &encSecret, &totpEnabled, &isSuperuser, &totpEnforced)
	if err != nil || !totpEnabled || encSecret == "" {
		http.Error(w, "تنظیمات احراز هویت دومرحله‌ای نامعتبر است.", http.StatusBadRequest)
		return
	}

	if isSuperuser || totpEnforced || os.Getenv("HERASAT_REQUIRE_2FA") == "1" || os.Getenv("HERASAT_REQUIRE_2FA") == "true" {
		http.Error(w, "امکان غیرفعال‌سازی احراز هویت دومرحله‌ای طبق خط‌مشی امنیتی برای این حساب کاربری وجود ندارد.", http.StatusForbidden)
		return
	}

	if !VerifyPassword(req.Password, storedHash) {
		failures, justLocked := Lockouts.RecordFailure(admin.Username, clientIP)
		ProgressiveDelay(failures)
		if justLocked {
			logAudit(admin.Username, admin.Title, "ACCOUNT_LOCKED", nil, clientIP)
		}
		http.Error(w, "رمز عبور جاری اشتباه است.", http.StatusUnauthorized)
		return
	}

	plainSecret, err := DecryptBoundField("admin_users", "totp_secret", userID, encSecret)
	if err != nil {
		http.Error(w, "خطای رمزگشایی کلید امنیتی", http.StatusInternalServerError)
		return
	}

	valid, err := VerifyTOTPCode(admin.Username, plainSecret, req.TOTPCode)
	if !valid || err != nil {
		http.Error(w, "کد امنیتی اشتباه است.", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec(`UPDATE admin_users SET totp_secret = NULL, totp_enabled = false WHERE username = $1`, admin.Username)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "DISABLE_2FA", nil, clientIP)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "احراز هویت دومرحله‌ای با موفقیت غیرفعال شد.",
	})
}
