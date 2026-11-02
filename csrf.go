package main

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	csrfAuditMu       sync.Mutex
	csrfAuditThrottle = make(map[string]time.Time)
)

func logThrottledCSRFAudit(r *http.Request, alertType string) {
	clientIP := getIP(r)
	username := "UNKNOWN"
	sessionKey := clientIP
	if admin, ok := adminFromContext(r); ok && admin.Username != "" {
		username = admin.Username
		sessionKey = username + ":" + clientIP
	} else if c, err := r.Cookie("herasat_admin_session"); err == nil && c.Value != "" {
		if s, ok := Sessions.Get(c.Value); ok && s.Username != "" {
			username = s.Username
			sessionKey = username + ":" + clientIP
		} else {
			sessionKey = hashToken(c.Value) + ":" + clientIP
		}
	}

	safeAlert := truncateRunes(alertType, 160)
	key := safeAlert + ":" + sessionKey
	now := time.Now()

	csrfAuditMu.Lock()
	lastLog, exists := csrfAuditThrottle[key]
	if exists && now.Sub(lastLog) < time.Minute {
		csrfAuditMu.Unlock()
		return
	}
	csrfAuditThrottle[key] = now
	if len(csrfAuditThrottle) > 10000 {
		for k, t := range csrfAuditThrottle {
			if now.Sub(t) >= time.Minute {
				delete(csrfAuditThrottle, k)
			}
		}
	}
	csrfAuditMu.Unlock()

	logAudit(username, "SECURITY_ALERT", safeAlert, nil, clientIP)
}

func isOriginAllowed(origin string, r *http.Request) bool {
	allowed := strings.TrimSpace(os.Getenv("HERASAT_ALLOWED_ORIGINS"))
	if allowed != "" {
		origins := strings.Split(allowed, ",")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
		return false
	}
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	originHost := u.Host
	if h, _, err := net.SplitHostPort(originHost); err == nil {
		originHost = h
	}
	return strings.EqualFold(host, originHost)
}

func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			origin := r.Header.Get("Origin")
			if origin == "" {
				if ref := r.Header.Get("Referer"); ref != "" {
					if u, err := url.Parse(ref); err == nil {
						origin = u.Scheme + "://" + u.Host
					}
				}
			}
			if origin != "" && !isOriginAllowed(origin, r) {
				safeOrigin := truncateRunes(origin, 128)
				logThrottledCSRFAudit(r, "CSRF_INVALID_ORIGIN:"+safeOrigin)
				writeAuthJSONError(w, http.StatusForbidden, "origin_invalid", "درخواست ارسالی از مبدأ شبکه غیرمجاز رد گردید.")
				return
			}

			csrfHeader := r.Header.Get("X-CSRF-Token")
			cookie, err := r.Cookie("herasat_csrf")
			if err != nil || cookie.Value == "" || csrfHeader == "" {
				logThrottledCSRFAudit(r, "CSRF_MISSING")
				writeAuthJSONError(w, http.StatusForbidden, "csrf_missing", "شناسه امنیتی درخواست ارسال نشده است.")
				return
			}

			if subtle.ConstantTimeCompare([]byte(csrfHeader), []byte(cookie.Value)) != 1 {
				logThrottledCSRFAudit(r, "CSRF_MISMATCH")
				writeAuthJSONError(w, http.StatusForbidden, "csrf_invalid", "شناسه امنیتی درخواست معتبر نمی‌باشد.")
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}
