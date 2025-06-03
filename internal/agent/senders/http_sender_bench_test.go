package senders

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkHTTPSender_Send(b *testing.B) {
	// Создаем тестовый сервер для бенчмарка
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewHTTPSender(ts.URL)
	testMetrics := map[string]interface{}{
		"gauge_metric1":   1.23,
		"counter_metric1": int64(42),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sender.Send(testMetrics, "", nil)
		if err != nil {
			b.Fatalf("Error sending metrics: %v", err)
		}
	}
}

func BenchmarkHTTPSender_SendWithHash(b *testing.B) {
	// Создаем тестовый сервер для бенчмарка
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	sender := NewHTTPSender(ts.URL)
	testMetrics := map[string]interface{}{
		"gauge_metric1":   1.23,
		"counter_metric1": int64(42),
	}

	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64-символьный ключ для SHA-256

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sender.Send(testMetrics, key, nil)
		if err != nil {
			b.Fatalf("Error sending metrics with hash: %v", err)
		}
	}
}
