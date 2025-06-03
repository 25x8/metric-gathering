package app

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidSHA256(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{
			name: "Valid SHA256",
			key:  "a1b2c3d4e5f67890123456789012345678901234567890123456789012345678",
			want: true,
		},
		{
			name: "Invalid length",
			key:  "short",
			want: false,
		},
		{
			name: "Invalid characters",
			key:  "g1b2c3d4e5f6789012345678901234567890123456789012345678901234567",
			want: false,
		},
		{
			name: "Empty string",
			key:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSHA256(tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMiddlewareWithHash(t *testing.T) {
	keyBytes := make([]byte, 32)
	_, err := rand.Read(keyBytes)
	require.NoError(t, err)
	validKey := hex.EncodeToString(keyBytes)

	tests := []struct {
		name           string
		key            string
		expectBypassed bool
	}{
		{
			name:           "Empty key passes through",
			key:            "",
			expectBypassed: true,
		},
		{
			name:           "Invalid key format passes through",
			key:            "invalid-key",
			expectBypassed: true,
		},
		{
			name:           "Valid key requires hash check",
			key:            validKey,
			expectBypassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Handler reached"))
			})

			middleware := MiddlewareWithHash(tt.key)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			if tt.expectBypassed {
				// Если ключ пустой или невалидный, запрос должен пройти
				assert.Equal(t, http.StatusOK, w.Code)
				assert.Equal(t, "Handler reached", w.Body.String())
			} else {
				// Если ключ валидный, middleware должен требовать хэш
				assert.NotEqual(t, http.StatusOK, w.Code)
				assert.NotEqual(t, "Handler reached", w.Body.String())
			}
		})
	}
}

func TestMiddlewareWithTrustedSubnet(t *testing.T) {
	tests := []struct {
		name           string
		trustedSubnet  string
		clientIP       string
		expectedStatus int
	}{
		{
			name:           "Valid IP in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "192.168.1.100",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP not in trusted subnet",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Missing X-Real-IP header",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid IP address",
			trustedSubnet:  "192.168.1.0/24",
			clientIP:       "invalid-ip",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Local subnet allows localhost",
			trustedSubnet:  "127.0.0.0/8",
			clientIP:       "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Handler reached"))
			})

			middleware := MiddlewareWithTrustedSubnet(tt.trustedSubnet)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "/test", nil)
			if tt.clientIP != "" {
				req.Header.Set("X-Real-IP", tt.clientIP)
			}
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "Handler reached", w.Body.String())
			}
		})
	}
}
