package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none';")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		if isCookieSecure() {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		origin := r.Header.Get("Origin")
		allowedOriginsEnv := os.Getenv("HERASAT_ALLOWED_ORIGINS")
		if origin != "" && allowedOriginsEnv != "" {
			for _, allowed := range strings.Split(allowedOriginsEnv, ",") {
				if strings.TrimSpace(allowed) == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}
		}

		if r.Method == http.MethodOptions {
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

func main() {
	InitSyslogShipper()
	defer GlobalSyslogShipper.Close()
	defer GlobalSSEBroker.Close()

	certFile := os.Getenv("HERASAT_TLS_CERT")
	keyFile := os.Getenv("HERASAT_TLS_KEY")
	tlsConfigured := certFile != "" && keyFile != ""
	cookieSecureEnv := os.Getenv("HERASAT_COOKIE_SECURE")
	enableTLSEnv := os.Getenv("HERASAT_ENABLE_TLS")
	allowInsecureHTTP := enableTLSEnv == "false" || os.Getenv("HERASAT_INSECURE_HTTP") == "true" || os.Getenv("HERASAT_DEV_MODE") == "1" || cookieSecureEnv == "false"

	if !tlsConfigured && cookieSecureEnv != "true" && !allowInsecureHTTP {
		log.Fatalf("FATAL: refusing to start without TLS and without HERASAT_COOKIE_SECURE=true. Set HERASAT_TLS_CERT/HERASAT_TLS_KEY or explicitly set HERASAT_COOKIE_SECURE=true behind a trusted TLS-terminating reverse proxy.")
	}

	if allowInsecureHTTP {
		log.Println("WARNING: Running in HTTP mode. Session cookies will not have Secure flag.")
		InitCookieSecurity(false)
	} else if cookieSecureEnv == "true" {
		if !tlsConfigured {
			log.Println("WARNING: HERASAT_COOKIE_SECURE=true set without TLS certificates; assuming trusted TLS-terminating reverse proxy.")
		}
		InitCookieSecurity(true)
	} else {
		InitCookieSecurity(tlsConfigured)
	}

	aesMasterKey := os.Getenv("HERASAT_AES_KEY")
	if len(aesMasterKey) != 32 {
		log.Fatalf("FATAL SECURITY ERROR: Environment variable HERASAT_AES_KEY must be set and exactly 32 bytes long.")
	}
	InitCrypto(aesMasterKey)

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

	loginLimiter := NewIPRateLimiter(0.08, 5)
	submitLimiter := NewIPRateLimiter(0.033, 3)
	trackLimiter := NewIPRateLimiter(0.5, 10)
	captchaLimiter := NewIPRateLimiter(1.0, 20)

	mux := http.NewServeMux()

	const maxJSONBodyBytes = 256 << 10 // 256 KB

	mux.HandleFunc("GET /health", healthCheck)
	mux.HandleFunc("GET /api/captcha", RateLimitMiddleware(captchaLimiter, GetCaptchaHandler))

	mux.HandleFunc("POST /api/report", RateLimitMiddleware(submitLimiter, SubmitReportHandler))
	mux.HandleFunc("POST /api/report/track", RateLimitMiddleware(trackLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(TrackReportHandler))))
	mux.HandleFunc("POST /api/report/follow-up", RateLimitMiddleware(submitLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(SubmitFollowUpHandler))))

	mux.HandleFunc("POST /api/admin/login", RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(LoginHandler))))
	mux.HandleFunc("POST /api/admin/login/verify-mfa", RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(VerifyMFAHandler))))
	mux.HandleFunc("GET /api/admin/events", AdminAuthMiddleware(AdminEventsHandler))
	mux.HandleFunc("POST /api/admin/2fa/setup", AdminAuthMiddleware(CSRFMiddleware(Setup2FAHandler)))
	mux.HandleFunc("POST /api/admin/2fa/enable", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(Enable2FAHandler)))))
	mux.HandleFunc("POST /api/admin/2fa/disable", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(Disable2FAHandler)))))
	mux.HandleFunc("POST /api/admin/logout", LogoutHandler)
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

	mux.HandleFunc("GET /api/profanity-rules", GetProfanityRulesHandler)

	mux.HandleFunc("POST /api/admin/change-my-password", AdminAuthMiddleware(CSRFMiddleware(RateLimitMiddleware(loginLimiter, LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ChangeMyPasswordHandler))))))

	mux.HandleFunc("POST /api/admin/report/delete", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(DeleteReportHandler)))))
	mux.HandleFunc("GET /api/admin/users", SuperuserAuthMiddleware(GetUsersHandler))
	mux.HandleFunc("POST /api/admin/users/create", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(CreateUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/update", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(UpdateUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/toggle-active", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ToggleUserActiveHandler)))))
	mux.HandleFunc("POST /api/admin/users/delete", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(DeleteUserHandler)))))
	mux.HandleFunc("POST /api/admin/users/reset-password", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(ResetUserPasswordHandler)))))
	mux.HandleFunc("POST /api/admin/users/lift-lockout", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(LiftUserLockoutHandler)))))
	mux.HandleFunc("GET /api/admin/sessions", SuperuserAuthMiddleware(GetActiveSessionsHandler))
	mux.HandleFunc("POST /api/admin/sessions/terminate", SuperuserAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(TerminateSessionHandler)))))
	mux.HandleFunc("GET /api/admin/audit-logs", SuperuserAuthMiddleware(GetAuditLogsHandler))
	mux.HandleFunc("GET /api/admin/system-status", SuperuserAuthMiddleware(GetSystemStatusHandler))
	mux.HandleFunc("GET /api/admin/analytics", AdminAuthMiddleware(GetAnalyticsHandler))

	mux.HandleFunc("GET /api/admin/notifications", AdminAuthMiddleware(GetNotificationsHandler))
	mux.HandleFunc("POST /api/admin/notifications/mark-read", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, RequireJSONMiddleware(MarkNotificationReadHandler)))))
	mux.HandleFunc("POST /api/admin/notifications/mark-all-read", AdminAuthMiddleware(CSRFMiddleware(LimitBodyMiddleware(maxJSONBodyBytes, MarkAllNotificationsReadHandler))))

	mux.Handle("/", http.FileServer(http.Dir("./frontend/public")))
	mux.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("./frontend/admin"))))

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
		log.Printf("Herasat secure backend initialized with HTTPS on %s (AFTA Compliant)...", bindAddr)
		log.Fatal(server.ListenAndServeTLS(certFile, keyFile))
	} else {
		log.Printf("Herasat secure backend initialized with HTTP on %s (AFTA Compliant)...", bindAddr)
		log.Fatal(server.ListenAndServe())
	}
}
