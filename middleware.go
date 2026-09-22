package main

import (
	"net"
	"net/http"
	"os"
	"strings"
)

func LimitBodyMiddleware(maxBytes int64, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	}
}

func RequireJSONMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func NoStoreMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		next.ServeHTTP(w, r)
	}
}

func AdminIPWhitelistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedEnv := strings.TrimSpace(os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS"))
		if allowedEnv == "" {
			next.ServeHTTP(w, r)
			return
		}

		allowedNets := parseTrustedProxies(allowedEnv)
		if len(allowedNets) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		clientIPStr := getIP(r)
		clientIP := net.ParseIP(clientIPStr)
		if clientIP == nil || !ipInCIDRs(clientIP, allowedNets) {
			logAudit("UNKNOWN", "SECURITY_ALERT", "ADMIN_IP_WHITELIST_BLOCKED:"+clientIPStr, nil, clientIPStr)
			http.Error(w, "دسترسی به بخش مدیریت از این آدرس شبکه مجاز نمی‌باشد.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
