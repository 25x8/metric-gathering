package senders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHttpSender_Send - тестирование отправки метрик с помощью мока HTTP-сервера
func TestHttpSender_Send(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/gauge/Alloc/12345.67", r.URL.EscapedPath())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HttpSender с URL тестового сервера
	sender := NewHttpSender(server.URL)

	// Пример метрик
	metrics := map[string]interface{}{
		"Alloc": 12345.67,
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}

func TestHttpSender_SendCounter(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/counter/PollCount/1", r.URL.EscapedPath())
		assert.Equal(t, "text/plain", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HttpSender с URL тестового сервера
	sender := NewHttpSender(server.URL)

	// Пример метрик
	metrics := map[string]interface{}{
		"PollCount": int64(1),
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}
