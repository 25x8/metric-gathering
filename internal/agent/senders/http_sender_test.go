package senders

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Value *float64 `json:"value,omitempty"`
	Delta *int64   `json:"delta,omitempty"`
}

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
	err := sender.Send(metrics, "")
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
	err := sender.Send(metrics, "")
	assert.NoError(t, err)
}

// Helper functions to create pointers to float64 and int64
func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestPostMetrics(t *testing.T) {
	tests := []struct {
		name   string
		method string
		value  float64
		delta  int64
		update int
		ok     bool
		static bool
	}{
		{method: "counter", name: "PollCount"},
		{method: "gauge", name: "RandomValue"},
		{method: "gauge", name: "Alloc"},
		{method: "gauge", name: "BuckHashSys", static: true},
		{method: "gauge", name: "Frees"},
		{method: "gauge", name: "GCCPUFraction", static: true},
		{method: "gauge", name: "GCSys", static: true},
		{method: "gauge", name: "HeapAlloc"},
		{method: "gauge", name: "HeapIdle"},
		{method: "gauge", name: "HeapInuse"},
		{method: "gauge", name: "HeapObjects"},
		{method: "gauge", name: "HeapReleased", static: true},
		{method: "gauge", name: "HeapSys", static: true},
		{method: "gauge", name: "LastGC", static: true},
		{method: "gauge", name: "Lookups", static: true},
		{method: "gauge", name: "MCacheInuse", static: true},
		{method: "gauge", name: "MCacheSys", static: true},
		{method: "gauge", name: "MSpanInuse", static: true},
		{method: "gauge", name: "MSpanSys", static: true},
		{method: "gauge", name: "Mallocs"},
		{method: "gauge", name: "NextGC", static: true},
		{method: "gauge", name: "NumForcedGC", static: true},
		{method: "gauge", name: "NumGC", static: true},
		{method: "gauge", name: "OtherSys", static: true},
		{method: "gauge", name: "PauseTotalNs", static: true},
		{method: "gauge", name: "StackInuse", static: true},
		{method: "gauge", name: "StackSys", static: true},
		{method: "gauge", name: "Sys", static: true},
		{method: "gauge", name: "TotalAlloc"},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the Content-Type header
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare the request body
			metrics := Metrics{
				ID:    tt.name,
				MType: tt.method,
			}
			body, err := json.Marshal(metrics)
			assert.NoError(t, err)

			// Perform the POST request
			resp, err := http.Post(server.URL+"/value/", "application/json", bytes.NewBuffer(body))
			assert.NoError(t, err)
			defer resp.Body.Close()

			// Verify the response status
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
