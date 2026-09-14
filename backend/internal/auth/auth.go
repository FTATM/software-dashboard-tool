package auth

import (
	"net"
	"net/http"
	"strings"
)

type ContextKey string

const (
	CookieName               = "authToken"
	AuthUserIdKey ContextKey = "userId"
)

type ClientInfo struct {
	IP        string `json:"ip" db:"ip"`
	UserAgent string `json:"userAgent" db:"user_agent"`
}

func GetClientInfo(r *http.Request) *ClientInfo {
	var ip string

	// 1. Check Cloudflare header
	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		ip = cfIP
	}

	// 2. Check X-Forwarded-For (comma-separated list; first entry is the client)
	if ip == "" {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			ip = strings.TrimSpace(parts[0])
		}
	}

	// 3. Check X-Real-IP (standard Nginx header)
	if ip == "" {
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			ip = strings.TrimSpace(xri)
		}
	}

	// 4. Fallback to direct connection socket (r.RemoteAddr is formatted as "IP:port")
	if ip == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err == nil {
			ip = host
		} else {
			ip = r.RemoteAddr
		}
	}

	return &ClientInfo{
		IP:        ip,
		UserAgent: r.UserAgent(), // Contains Browser, OS, and Architecture
	}
}
