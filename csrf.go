package main

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

func isOriginAllowed(origin string) bool {
	allowed := os.Getenv("HERASAT_ALLOWED_ORIGINS")
	if strings.TrimSpace(allowed) == "" {
		return true
	}
	origins := strings.Split(allowed, ",")
	for _, o := range origins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		method := strings.ToUpper(r.Method)
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete {
			if origin := r.Header.Get("Origin"); origin != "" {
				if !isOriginAllowed(origin) {
					http.Error(w, "Forbidden: Invalid Origin", http.StatusForbidden)
					return
				}
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
