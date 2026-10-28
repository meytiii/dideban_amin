package main

import (
	"crypto/subtle"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

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
				http.Error(w, "Forbidden: Invalid Origin", http.StatusForbidden)
				return
			}

			csrfHeader := r.Header.Get("X-CSRF-Token")
			cookie, err := r.Cookie("herasat_csrf")
			if err != nil || cookie.Value == "" || csrfHeader == "" {
				http.Error(w, "Forbidden: Missing CSRF token", http.StatusForbidden)
				return
			}

			if subtle.ConstantTimeCompare([]byte(csrfHeader), []byte(cookie.Value)) != 1 {
				http.Error(w, "Forbidden: Invalid CSRF token", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	}
}
