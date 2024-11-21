package senders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHTTPSender_Send - тестирование отправки метрик с помощью мока HTTP-сервера
func TestHTTPSender_Send(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/gauge/Alloc/12345.67", r.URL.EscapedPath())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPSender с URL тестового сервера
	sender := NewHTTPSender(server.URL)

	// Пример метрик
	metrics := map[string]interface{}{
		"Alloc": 12345.67,
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}

func TestHTTPSender_SendCounter(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/counter/PollCount/1", r.URL.EscapedPath())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPSender с URL тестового сервера
	sender := NewHTTPSender(server.URL)

	// Пример метрик
	metrics := map[string]interface{}{
		"PollCount": int64(1),
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}

// Helper functions to create pointers to float64 and int64
func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
