package middlewares

import (
	"net"
	"net/http"
	"strings"
)

// ClientIPMiddleware extracts the real client IP from proxy headers and sets
// RemoteAddr to "IP:port" format for consistency throughout the application.
func ClientIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := extractClientIP(r)

		if clientIP != "" {
			_, port, err := net.SplitHostPort(r.RemoteAddr)
			if err == nil && port != "" {
				r.RemoteAddr = net.JoinHostPort(clientIP, port)
			} else {
				r.RemoteAddr = net.JoinHostPort(clientIP, "0")
			}
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientIP(r *http.Request) string {
	if ip := r.Header.Get("True-Client-IP"); ip != "" {
		if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
			return parsed.String()
		}
	}

	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		if parsed := net.ParseIP(strings.TrimSpace(ip)); parsed != nil {
			return parsed.String()
		}
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsed := net.ParseIP(ip); parsed != nil {
				return parsed.String()
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		if parsed := net.ParseIP(r.RemoteAddr); parsed != nil {
			return parsed.String()
		}
		return ""
	}

	if parsed := net.ParseIP(host); parsed != nil {
		return parsed.String()
	}

	return ""
}
