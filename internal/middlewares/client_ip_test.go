package middlewares

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		remoteAddr     string
		headers        map[string]string
		expectedIP     string
		expectedPort   string
		expectedRemote string
	}{
		{
			name:           "direct connection with port",
			remoteAddr:     "203.0.113.1:54321",
			headers:        map[string]string{},
			expectedIP:     "203.0.113.1",
			expectedPort:   "54321",
			expectedRemote: "203.0.113.1:54321",
		},
		{
			name:           "direct connection without port",
			remoteAddr:     "203.0.113.1",
			headers:        map[string]string{},
			expectedIP:     "203.0.113.1",
			expectedPort:   "0",
			expectedRemote: "203.0.113.1:0",
		},
		{
			name:       "true-client-ip header",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "198.51.100.1",
			},
			expectedIP:     "198.51.100.1",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.1:12345",
		},
		{
			name:       "x-real-ip header",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.2",
			},
			expectedIP:     "198.51.100.2",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.2:12345",
		},
		{
			name:       "x-forwarded-for single IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.3",
			},
			expectedIP:     "198.51.100.3",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.3:12345",
		},
		{
			name:       "x-forwarded-for multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "198.51.100.4, 10.0.0.2, 10.0.0.3",
			},
			expectedIP:     "198.51.100.4",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.4:12345",
		},
		{
			name:       "x-forwarded-for with spaces",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "  198.51.100.5  , 10.0.0.2",
			},
			expectedIP:     "198.51.100.5",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.5:12345",
		},
		{
			name:       "priority: true-client-ip over x-real-ip",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "198.51.100.6",
				"X-Real-IP":      "198.51.100.7",
			},
			expectedIP:     "198.51.100.6",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.6:12345",
		},
		{
			name:       "priority: true-client-ip over x-forwarded-for",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP":  "198.51.100.8",
				"X-Forwarded-For": "198.51.100.9",
			},
			expectedIP:     "198.51.100.8",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.8:12345",
		},
		{
			name:       "priority: x-real-ip over x-forwarded-for",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP":       "198.51.100.10",
				"X-Forwarded-For": "198.51.100.11",
			},
			expectedIP:     "198.51.100.10",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.10:12345",
		},
		{
			name:       "invalid true-client-ip falls back to x-real-ip",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "not-an-ip",
				"X-Real-IP":      "198.51.100.12",
			},
			expectedIP:     "198.51.100.12",
			expectedPort:   "12345",
			expectedRemote: "198.51.100.12:12345",
		},
		{
			name:       "invalid x-forwarded-for falls back to remote addr",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "invalid-ip",
			},
			expectedIP:     "10.0.0.1",
			expectedPort:   "12345",
			expectedRemote: "10.0.0.1:12345",
		},
		{
			name:           "ipv6 address",
			remoteAddr:     "[2001:db8::1]:12345",
			headers:        map[string]string{},
			expectedIP:     "2001:db8::1",
			expectedPort:   "12345",
			expectedRemote: "[2001:db8::1]:12345",
		},
		{
			name:       "ipv6 in x-forwarded-for",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "2001:db8::2",
			},
			expectedIP:     "2001:db8::2",
			expectedPort:   "12345",
			expectedRemote: "[2001:db8::2]:12345",
		},
		{
			name:           "empty headers with valid remote",
			remoteAddr:     "192.168.1.1:8080",
			headers:        map[string]string{},
			expectedIP:     "192.168.1.1",
			expectedPort:   "8080",
			expectedRemote: "192.168.1.1:8080",
		},
		{
			name:       "preserve port 0 from original",
			remoteAddr: "10.0.0.1:0",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.13",
			},
			expectedIP:     "198.51.100.13",
			expectedPort:   "0",
			expectedRemote: "198.51.100.13:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that captures the modified request
			var capturedRemoteAddr string
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRemoteAddr = r.RemoteAddr
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with middleware
			handler := ClientIPMiddleware(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			// Set headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Execute request
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Verify RemoteAddr was set correctly
			if capturedRemoteAddr != tt.expectedRemote {
				t.Errorf("expected RemoteAddr %q, got %q", tt.expectedRemote, capturedRemoteAddr)
			}

			// Verify we can split it correctly
			host, port, err := net.SplitHostPort(capturedRemoteAddr)
			if err != nil {
				t.Fatalf("failed to split host port: %v", err)
			}

			if host != tt.expectedIP {
				t.Errorf("expected host %q, got %q", tt.expectedIP, host)
			}

			if port != tt.expectedPort {
				t.Errorf("expected port %q, got %q", tt.expectedPort, port)
			}
		})
	}
}

func TestExtractClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expectedIP string
	}{
		{
			name:       "no headers, valid remote addr with port",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "no headers, valid remote addr without port",
			remoteAddr: "192.168.1.1",
			headers:    map[string]string{},
			expectedIP: "192.168.1.1",
		},
		{
			name:       "true-client-ip present",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "203.0.113.1",
			},
			expectedIP: "203.0.113.1",
		},
		{
			name:       "x-real-ip present",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.2",
			},
			expectedIP: "203.0.113.2",
		},
		{
			name:       "x-forwarded-for present",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.3",
			},
			expectedIP: "203.0.113.3",
		},
		{
			name:       "all headers present - priority order",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP":  "203.0.113.4",
				"X-Real-IP":       "203.0.113.5",
				"X-Forwarded-For": "203.0.113.6",
			},
			expectedIP: "203.0.113.4",
		},
		{
			name:       "invalid true-client-ip, valid x-real-ip",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "not.valid.ip",
				"X-Real-IP":      "203.0.113.7",
			},
			expectedIP: "203.0.113.7",
		},
		{
			name:       "empty header values",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"True-Client-IP": "",
				"X-Real-IP":      "",
			},
			expectedIP: "10.0.0.1",
		},
		{
			name:       "whitespace in header value",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "  203.0.113.8  ",
			},
			expectedIP: "203.0.113.8",
		},
		{
			name:       "invalid remote addr",
			remoteAddr: "invalid",
			headers:    map[string]string{},
			expectedIP: "",
		},
		{
			name:       "ipv6 remote addr",
			remoteAddr: "[2001:db8::1]:12345",
			headers:    map[string]string{},
			expectedIP: "2001:db8::1",
		},
		{
			name:       "ipv6 in header",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Real-IP": "2001:db8::2",
			},
			expectedIP: "2001:db8::2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			result := extractClientIP(req)

			if result != tt.expectedIP {
				t.Errorf("expected IP %q, got %q", tt.expectedIP, result)
			}
		})
	}
}

func BenchmarkClientIPMiddleware(b *testing.B) {
	handler := ClientIPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkExtractClientIP(b *testing.B) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = extractClientIP(req)
	}
}
