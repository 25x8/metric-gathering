package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		realIP         string
		remoteAddr     string
		expectedStatus int
	}{
		{
			name:           "Empty trusted subnet - should allow",
			trustedSubnet:  "",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP in trusted subnet - should allow",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet - should deny",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Localhost IP in localhost subnet - should allow",
			trustedSubnet:  "127.0.0.0/8",
			realIP:         "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid CIDR - should return 500",
			trustedSubnet:  "invalid-cidr",
			realIP:         "192.168.1.100",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Invalid IP - should return 400",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "invalid-ip",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "No X-Real-IP header, use RemoteAddr",
			trustedSubnet:  "192.168.1.0/24",
			realIP:         "",
			remoteAddr:     "192.168.1.50:12345",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем простой обработчик, который возвращает 200
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Оборачиваем в middleware
			middleware := TrustedSubnetMiddleware(tt.trustedSubnet)
			wrappedHandler := middleware(handler)

			// Создаем запрос
			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			// Создаем ResponseRecorder
			w := httptest.NewRecorder()

			// Выполняем запрос
			wrappedHandler.ServeHTTP(w, req)

			// Проверяем результат
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name          string
		xForwardedFor string
		xRealIP       string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "X-Forwarded-For with single IP",
			xForwardedFor: "192.168.1.100",
			expectedIP:    "192.168.1.100",
		},
		{
			name:          "X-Forwarded-For with multiple IPs",
			xForwardedFor: "192.168.1.100, 10.0.0.1, 172.16.0.1",
			expectedIP:    "192.168.1.100",
		},
		{
			name:       "X-Real-IP when no X-Forwarded-For",
			xRealIP:    "192.168.1.200",
			expectedIP: "192.168.1.200",
		},
		{
			name:       "RemoteAddr when no headers",
			remoteAddr: "192.168.1.50:12345",
			expectedIP: "192.168.1.50",
		},
		{
			name:       "RemoteAddr without port",
			remoteAddr: "192.168.1.60",
			expectedIP: "192.168.1.60",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			if tt.remoteAddr != "" {
				req.RemoteAddr = tt.remoteAddr
			}

			result := getClientIP(req)
			assert.Equal(t, tt.expectedIP, result)
		})
	}
}
