package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	trustedProxiesLock     sync.RWMutex
	cachedTrustedProxies   []*net.IPNet
	trustedProxiesExplicit bool

	adminAllowedCIDRsLock   sync.RWMutex
	cachedAdminAllowedCIDRs []*net.IPNet
	adminAllowedExplicit    bool
	adminAllowAny           bool
)

func parseCIDRs(raw string) ([]*net.IPNet, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var nets []*net.IPNet
	for _, item := range strings.Split(raw, ",") {
		cidr := strings.TrimSpace(item)
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
		if err != nil || ipNet == nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", item, err)
		}
		nets = append(nets, ipNet)
	}
	return nets, nil
}

func ipInCIDRs(ip net.IP, cidrs []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, cidr := range cidrs {
		if cidr != nil && cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func parseTrustedProxies(envVal string) []*net.IPNet {
	nets, _ := parseCIDRs(envVal)
	return nets
}

func SetTrustedProxies(nets []*net.IPNet) {
	trustedProxiesLock.Lock()
	defer trustedProxiesLock.Unlock()
	cachedTrustedProxies = nets
	trustedProxiesExplicit = true
}

func GetTrustedProxies() []*net.IPNet {
	trustedProxiesLock.RLock()
	defer trustedProxiesLock.RUnlock()
	if !trustedProxiesExplicit {
		env := os.Getenv("HERASAT_TRUSTED_PROXIES")
		nets, _ := parseCIDRs(env)
		return nets
	}
	return cachedTrustedProxies
}

func SetAdminAllowedCIDRs(nets []*net.IPNet) {
	adminAllowedCIDRsLock.Lock()
	defer adminAllowedCIDRsLock.Unlock()
	cachedAdminAllowedCIDRs = nets
	adminAllowedExplicit = true
}

func GetAdminAllowedCIDRs() []*net.IPNet {
	adminAllowedCIDRsLock.RLock()
	defer adminAllowedCIDRsLock.RUnlock()
	if !adminAllowedExplicit {
		env := os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS")
		nets, _ := parseCIDRs(env)
		return nets
	}
	return cachedAdminAllowedCIDRs
}

func SetAdminAllowAny(allow bool) {
	adminAllowedCIDRsLock.Lock()
	defer adminAllowedCIDRsLock.Unlock()
	adminAllowAny = allow
	adminAllowedExplicit = true
}

func IsAdminAllowAny() bool {
	adminAllowedCIDRsLock.RLock()
	defer adminAllowedCIDRsLock.RUnlock()
	if !adminAllowedExplicit {
		return os.Getenv("HERASAT_ADMIN_ALLOW_ANY") == "true" || os.Getenv("HERASAT_DEV_MODE") == "1"
	}
	return adminAllowAny
}

func ResetNetworkSecurityForTest() {
	trustedProxiesLock.Lock()
	cachedTrustedProxies = nil
	trustedProxiesExplicit = false
	trustedProxiesLock.Unlock()

	adminAllowedCIDRsLock.Lock()
	cachedAdminAllowedCIDRs = nil
	adminAllowedExplicit = false
	adminAllowAny = false
	adminAllowedCIDRsLock.Unlock()
}

func InitNetworkSecurity() error {
	proxyEnv := strings.TrimSpace(os.Getenv("HERASAT_TRUSTED_PROXIES"))
	if proxyEnv != "" {
		nets, err := parseCIDRs(proxyEnv)
		if err != nil {
			return fmt.Errorf("invalid HERASAT_TRUSTED_PROXIES: %w", err)
		}
		SetTrustedProxies(nets)
	} else {
		SetTrustedProxies(nil)
	}

	certFile := os.Getenv("HERASAT_TLS_CERT")
	keyFile := os.Getenv("HERASAT_TLS_KEY")
	tlsConfigured := certFile != "" && keyFile != ""
	cookieSecure := isCookieSecure()
	devMode := os.Getenv("HERASAT_DEV_MODE") == "1"

	if cookieSecure && !tlsConfigured {
		if len(GetTrustedProxies()) == 0 {
			if !devMode {
				return errors.New("FATAL: HERASAT_COOKIE_SECURE=true is set without TLS (reverse proxy deployment), but HERASAT_TRUSTED_PROXIES is empty. You must configure HERASAT_TRUSTED_PROXIES to trust upstream reverse proxies, otherwise client IP detection, rate limits, lockouts, and IP whitelists will not function securely.")
			}
			log.Println("WARNING: HERASAT_COOKIE_SECURE=true without TLS and empty HERASAT_TRUSTED_PROXIES allowed due to HERASAT_DEV_MODE=1.")
		}
	}

	adminCIDREnv := strings.TrimSpace(os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS"))
	allowAny := os.Getenv("HERASAT_ADMIN_ALLOW_ANY") == "true"
	if adminCIDREnv == "" {
		if !allowAny && !devMode {
			return errors.New("FATAL: HERASAT_ADMIN_ALLOWED_CIDRS is required for security. Set allowed admin CIDRs (e.g. 127.0.0.1/32, 10.0.0.0/8), or explicitly set HERASAT_ADMIN_ALLOW_ANY=true or HERASAT_DEV_MODE=1 to bypass.")
		}
		SetAdminAllowAny(true)
		if !devMode {
			log.Println("WARNING: HERASAT_ADMIN_ALLOW_ANY=true is set; admin interface is accessible from any IP address.")
		} else {
			log.Println("WARNING: HERASAT_ADMIN_ALLOWED_CIDRS is empty; running in development mode (HERASAT_DEV_MODE=1).")
		}
	} else {
		nets, err := parseCIDRs(adminCIDREnv)
		if err != nil {
			return fmt.Errorf("invalid HERASAT_ADMIN_ALLOWED_CIDRS: %w", err)
		}
		SetAdminAllowedCIDRs(nets)
		SetAdminAllowAny(false)
	}

	return nil
}

func isAdminAccessAllowed(r *http.Request) bool {
	clientIPStr := getIP(r)
	clientIP := net.ParseIP(clientIPStr)

	adminAllowedCIDRsLock.RLock()
	allowedNets := cachedAdminAllowedCIDRs
	allowAny := adminAllowAny
	isExplicit := adminAllowedExplicit
	adminAllowedCIDRsLock.RUnlock()

	if !isExplicit {
		env := strings.TrimSpace(os.Getenv("HERASAT_ADMIN_ALLOWED_CIDRS"))
		if env == "" {
			return true
		}
		nets, err := parseCIDRs(env)
		if err != nil || len(nets) == 0 {
			return false
		}
		if clientIP == nil {
			return false
		}
		return ipInCIDRs(clientIP, nets)
	}

	if allowAny {
		return true
	}
	if len(allowedNets) == 0 {
		return false
	}
	if clientIP == nil {
		return false
	}
	return ipInCIDRs(clientIP, allowedNets)
}
