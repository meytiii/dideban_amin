package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedHostsEnv := strings.TrimSpace(os.Getenv("HERASAT_ALLOWED_HOSTS"))
		if allowedHostsEnv != "" {
			host := r.Host
			if h, _, err := net.SplitHostPort(host); err == nil {
				host = h
			}
			allowed := false
			for _, ah := range strings.Split(allowedHostsEnv, ",") {
				if strings.EqualFold(strings.TrimSpace(ah), host) {
					allowed = true
					break
				}
			}
			if !allowed {
				http.Error(w, "Invalid Host header", http.StatusBadRequest)
				return
			}
		}

		allowedOriginsEnv := os.Getenv("HERASAT_ALLOWED_ORIGINS")
		var allowedOrigins []string
		if allowedOriginsEnv != "" {
			for _, o := range strings.Split(allowedOriginsEnv, ",") {
				o = strings.TrimSpace(o)
				if o != "" {
					allowedOrigins = append(allowedOrigins, o)
				}
			}
		}

		cspConnect := "connect-src 'self'"
		cspForm := "form-action 'self'"
		if len(allowedOrigins) > 0 {
			extra := " " + strings.Join(allowedOrigins, " ")
			cspConnect += extra
			cspForm += extra
		}

		csp := fmt.Sprintf("default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; object-src 'none'; base-uri 'self'; %s; %s; frame-ancestors 'none'; report-uri /api/security/csp-report;", cspForm, cspConnect)
		w.Header().Set("Content-Security-Policy", csp)
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		if isCookieSecure() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		origin := r.Header.Get("Origin")
		if len(allowedOrigins) > 0 {
			w.Header().Add("Vary", "Origin")
		}
		originAllowed := false
		if origin != "" && len(allowedOrigins) > 0 {
			for _, allowed := range allowedOrigins {
				if allowed == origin {
					originAllowed = true
					break
				}
			}
		}

		if originAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if r.Method == http.MethodOptions {
			if origin != "" && !originAllowed {
				http.Error(w, "CORS origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "secure_system_active"}`))
}

func CSPReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(r.Body, 64<<10))
	w.WriteHeader(http.StatusNoContent)
}

func loadDotEnv() {
	if envBytes, err := os.ReadFile(".env"); err == nil {
		lines := strings.Split(string(envBytes), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				k := strings.TrimSpace(parts[0])
				v := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
		}
	}
}

func main() {
	loadDotEnv()
	InitSyslogShipper()
	defer GlobalSyslogShipper.Close()
	defer GlobalSSEBroker.Close()

	certFile := os.Getenv("HERASAT_TLS_CERT")
	keyFile := os.Getenv("HERASAT_TLS_KEY")
	tlsConfigured := certFile != "" && keyFile != ""
	cookieSecureEnv := os.Getenv("HERASAT_COOKIE_SECURE")
	enableTLSEnv := os.Getenv("HERASAT_ENABLE_TLS")
	allowInsecureHTTP := enableTLSEnv == "false" || os.Getenv("HERASAT_INSECURE_HTTP") == "true" || os.Getenv("HERASAT_DEV_MODE") == "1" || cookieSecureEnv == "false" || cookieSecureEnv == "0"

	if !tlsConfigured && cookieSecureEnv != "true" && cookieSecureEnv != "1" && !allowInsecureHTTP {
		log.Fatalf("FATAL: refusing to start without TLS and without HERASAT_COOKIE_SECURE=true. Set HERASAT_TLS_CERT/HERASAT_TLS_KEY or explicitly set HERASAT_COOKIE_SECURE=true behind a trusted TLS-terminating reverse proxy.")
	}

	if allowInsecureHTTP {
		log.Println("WARNING: Running in HTTP mode. Session cookies will not have Secure flag.")
		InitCookieSecurity(false)
	} else if cookieSecureEnv == "true" || cookieSecureEnv == "1" {
		if !tlsConfigured {
			log.Println("WARNING: HERASAT_COOKIE_SECURE=true set without TLS certificates; assuming trusted TLS-terminating reverse proxy.")
		}
		InitCookieSecurity(true)
	} else {
		InitCookieSecurity(tlsConfigured)
	}

	if err := InitNetworkSecurity(); err != nil {
		log.Fatalf("%v", err)
	}

	clamScanner := NewClamAVScanner()
	if err := CheckClamAVStartup(clamScanner); err != nil {
		log.Fatalf("%v", err)
	}

	aesMasterKey := os.Getenv("HERASAT_AES_KEY")
	if len(aesMasterKey) != 32 {
		log.Fatalf("FATAL SECURITY ERROR: Environment variable HERASAT_AES_KEY must be set and exactly 32 bytes long.")
	}
	if isKnownPlaceholderKey(aesMasterKey) {
		log.Fatalf("FATAL SECURITY ERROR: HERASAT_AES_KEY is using an insecure placeholder key. Generate a cryptographically secure 32-byte secret.")
	}
	InitCrypto(aesMasterKey)

	if err := RunCryptographicSelfTests(); err != nil {
		log.Fatalf("FATAL SECURITY ERROR: Cryptographic self-tests failed: %v", err)
	}

	dbConnStr := strings.TrimSpace(os.Getenv("HERASAT_DB_CONN"))
	if dbConnStr == "" && os.Getenv("HERASAT_DEV_MODE") == "1" {
		if envBytes, err := os.ReadFile(".env"); err == nil {
			lines := strings.Split(string(envBytes), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "HERASAT_DB_CONN=") {
					dbConnStr = strings.TrimPrefix(line, "HERASAT_DB_CONN=")
					dbConnStr = strings.Trim(dbConnStr, "\"'")
					break
				}
			}
		}
	}
	if dbConnStr == "" {
		log.Fatal("FATAL: HERASAT_DB_CONN environment variable is required.")
	}
	InitDB(dbConnStr)
	defer DB.Close()

	if err := VerifyAuditChainOnStartup(DB); err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	if UsedCaptchas.db == nil || Lockouts.db == nil {
		log.Fatal("FATAL: security stores not initialized")
	}

	if os.Getenv("HERASAT_REENCRYPT_ON_BOOT") == "true" {
		log.Println("[INFO] HERASAT_REENCRYPT_ON_BOOT is set; starting legacy data re-encryption...")
		if err := reencryptLegacyData(DB); err != nil {
			log.Fatalf("[FATAL] Re-encryption of legacy data failed: %v", err)
		}
		log.Println("[INFO] Re-encryption of legacy data completed successfully.")
	}

	Sessions.StartReaper(5 * time.Minute)
	UsedCaptchas.StartReaper(2 * time.Minute)
	StartAutoBackup(DB)

	loginLimiter := NewIPRateLimiter(0.08, 5)
	submitLimiter := NewIPRateLimiter(0.033, 3)
	trackLimiter := NewIPRateLimiter(0.5, 10)
	captchaLimiter := NewIPRateLimiter(1.0, 20)

	timeLimiter := NewIPRateLimiter(2.0, 30)

	mux := http.NewServeMux()

	const maxJSONBodyBytes = 256 << 10

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /api/server-time", AdminIPWhitelistMiddleware(RateLimitMiddleware(timeLimiter, GetServerTimeHandler)))
	mux.HandleFunc("GET /api/admin/server-time", AdminIPWhitelistMiddleware(RateLimitMiddleware(timeLimiter, GetServerTimeHandler)))
	mux.HandleFunc("GET /api/captcha", RateLimitMiddleware(captchaLimiter, GetCaptchaHandler))

	mux.HandleFunc("POST /api/report", RateLimitMiddleware(submitLimiter, SubmitReportHandler))
	mux.HandleFunc("POST /api/report/track", RateLimitMiddleware(trackLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(TrackReportHandler))))
	mux.HandleFunc("POST /api/report/follow-up", RateLimitMiddleware(submitLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(SubmitFollowUpHandler))))

	mux.HandleFunc("POST /api/security/csp-report", CSPReportHandler)

	mux.HandleFunc("POST /api/admin/login", AdminIPWhitelistMiddleware(RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(LoginHandler)))))
	mux.HandleFunc("POST /api/admin/login/verify-mfa", AdminIPWhitelistMiddleware(RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(VerifyMFAHandler)))))
	mux.HandleFunc("GET /api/admin/events", AdminAuthMiddleware(AdminEventsHandler))
	mux.HandleFunc("POST /api/admin/2fa/setup", AdminAuthMiddleware(CSRFMiddleware(Setup2FAHandler)))
	mux.HandleFunc("POST /api/admin/2fa/enable", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(Enable2FAHandler)))))
	mux.HandleFunc("POST /api/admin/2fa/disable", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(Disable2FAHandler)))))
	mux.HandleFunc("POST /api/admin/logout", AdminAuthMiddleware(CSRFMiddleware(LogoutHandler)))
	mux.HandleFunc("GET /api/admin/me", AdminAuthMiddleware(GetCurrentAdminHandler))
	mux.HandleFunc("POST /api/admin/toggle-read", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ToggleReadStatusHandler)))))
	mux.HandleFunc("POST /api/admin/respond", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(RespondToReportHandler)))))
	mux.HandleFunc("POST /api/admin/report/internal-notes", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(UpdateInternalNotesHandler)))))
	mux.HandleFunc("POST /api/admin/report/set-priority", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(SetReportPriorityHandler)))))
	mux.HandleFunc("POST /api/admin/report/classify", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ClassifyReportHandler)))))
	mux.HandleFunc("GET /api/admin/reports", AdminAuthMiddleware(GetReportsHandler))
	mux.HandleFunc("GET /api/admin/report", AdminAuthMiddleware(GetReportDetailsHandler))
	mux.HandleFunc("GET /api/admin/download", AdminAuthMiddleware(DownloadAttachmentHandler))
	mux.HandleFunc("GET /api/admin/report/export-docx", AdminAuthMiddleware(ExportReportDocxHandler))
	mux.HandleFunc("GET /api/admin/forwarding-recipients", AdminAuthMiddleware(GetForwardingRecipientsHandler))
	mux.HandleFunc("POST /api/admin/report/forward", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ForwardReportHandler)))))
	mux.HandleFunc("GET /api/admin/report/forward-history", AdminAuthMiddleware(GetForwardingHistoryHandler))
	mux.HandleFunc("POST /api/admin/report/regional-response", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(SubmitRegionalResponseHandler)))))
	mux.HandleFunc("POST /api/admin/report/remove-forward", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(RemoveForwardingHandler)))))

	mux.HandleFunc("GET /api/profanity-rules", SuperuserAuthMiddleware(GetProfanityRulesHandler))

	mux.HandleFunc("POST /api/admin/change-my-password", AdminAuthMiddleware(CSRFMiddleware(RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ChangeMyPasswordHandler))))))

	mux.HandleFunc("POST /api/admin/report/delete", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(DeleteReportHandler)))))
	mux.HandleFunc("GET /api/admin/users", SuperuserAuthMiddleware(GetUsersHandler))
	mux.HandleFunc("POST /api/admin/users/create", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(CreateUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/update", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(UpdateUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/toggle-active", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ToggleUserActiveHandler)))))
	mux.HandleFunc("POST /api/admin/users/delete", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(DeleteUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/reset-password", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ResetUserPasswordHandler)))))
	mux.HandleFunc("POST /api/admin/users/lift-lockout", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(LiftUserLockoutHandler)))))
	mux.HandleFunc("POST /api/admin/users/disable-2fa", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(DisableUser2FAHandler)))))
	mux.HandleFunc("GET /api/admin/sessions", SuperuserAuthMiddleware(GetActiveSessionsHandler))
	mux.HandleFunc("POST /api/admin/sessions/terminate", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(TerminateSessionHandler)))))
	mux.HandleFunc("GET /api/admin/audit-logs", SuperuserAuthMiddleware(GetAuditLogsHandler))
	mux.HandleFunc("GET /api/admin/audit-logs/verify", SuperuserAuthMiddleware(VerifyAuditLogsHandler))
	mux.HandleFunc("GET /api/admin/system-status", SuperuserAuthMiddleware(GetSystemStatusHandler))
	mux.HandleFunc("GET /api/admin/analytics", AdminAuthMiddleware(GetAnalyticsHandler))

	mux.HandleFunc("GET /api/admin/notifications", AdminAuthMiddleware(GetNotificationsHandler))
	mux.HandleFunc("POST /api/admin/notifications/mark-read", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(MarkNotificationReadHandler)))))
	mux.HandleFunc("POST /api/admin/notifications/mark-all-read", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, MarkAllNotificationsReadHandler))))

	mux.Handle("/", StaticCacheMiddleware(GzipMiddleware(Custom404FileServer("./frontend/public", "./frontend/public/404.html"))))
	mux.Handle("/admin/", AdminIPWhitelistMiddleware(http.StripPrefix("/admin/", StaticCacheMiddleware(GzipMiddleware(Custom404FileServer("./frontend/admin", "./frontend/admin/404.html")))).ServeHTTP))

	bindAddr := os.Getenv("HERASAT_BIND_ADDR")
	if bindAddr == "" {
		bindPort := os.Getenv("HERASAT_PORT")
		if bindPort == "" {
			bindPort = "8080"
		}

		bindAddr = "0.0.0.0:" + bindPort
	}

	server := &http.Server{
		Addr:              bindAddr,
		Handler:           secureHeadersMiddleware(mux),
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	if tlsConfigured {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			},
		}
		log.Printf("Herasat secure backend initialized with HTTPS on %s (AFTA Compliant)...", bindAddr)
		log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		log.Printf("Herasat secure backend initialized with HTTP on %s (AFTA Compliant)...", bindAddr)
		log.Fatal(server.ListenAndServe())
	}
}
