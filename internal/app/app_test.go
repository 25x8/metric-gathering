package app

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddlewareWithHash(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		header     string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "Valid SHA256 Key and Header",
			key:        "a3f7b5c8d9e4f1234567890abcdefabcdefabcdefabcdefabcdefabcdefabcdef",
			header:     "some-hash-value",
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name:       "Invalid Key (not SHA256)",
			key:        "invalidkey",
			header:     "some-hash-value",
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name:       "Missing HashSHA256 Header",
			key:        "a3f7b5c8d9e4f1234567890abcdefabcdefabcdefabcdefabcdefabcdefabcdef",
			header:     "",
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Моковый обработчик
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Создаем мидлваре
			middleware := MiddlewareWithHash(tt.key)

			// Создаем сервер
			handler := middleware(mockHandler)
			server := httptest.NewServer(handler)
			defer server.Close()

			// Создаем запрос
			req, err := http.NewRequest(http.MethodGet, server.URL, nil)
			assert.NoError(t, err)

			// Устанавливаем заголовок, если он задан
			if tt.header != "" {
				req.Header.Set("HashSHA256", tt.header)
			}

			// Выполняем запрос
			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Проверяем статус ответа
			assert.Equal(t, tt.wantStatus, resp.StatusCode, "Unexpected status code")

			// Проверяем тело ответа
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), tt.wantBody, "Response body does not match expected output")
		})
	}
}
