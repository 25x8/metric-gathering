package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidSHA256(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "Valid SHA256",
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
			expected: true,
		},
		{
			name:     "Invalid length",
			key:      "0123456789abcdef",
			expected: false,
		},
		{
			name:     "Invalid characters",
			key:      "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdez",
			expected: false,
		},
		{
			name:     "Empty string",
			key:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidSHA256(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMiddlewareWithHash(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		requestHash    string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Empty key passes through",
			key:            "",
			requestHash:    "anyhash",
			requestBody:    "test body",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid key format passes through",
			key:            "invalidkey",
			requestHash:    "anyhash",
			requestBody:    "test body",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый обработчик, который всегда возвращает OK
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Создаем middleware с заданным ключом
			middleware := MiddlewareWithHash(tt.key)

			// Создаем тестовый запрос с телом
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.requestBody))
			if tt.requestHash != "" {
				req.Header.Set("HashSHA256", tt.requestHash)
			}

			// Создаем recorder для записи ответа
			rr := httptest.NewRecorder()

			// Применяем middleware
			middleware(nextHandler).ServeHTTP(rr, req)

			// Проверяем статус ответа
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
