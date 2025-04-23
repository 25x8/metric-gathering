package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/gorilla/mux"
)

// MockStorage - простая реализация интерфейса Storage для тестирования
type MockStorage struct{}

func (m *MockStorage) SaveGaugeMetric(name string, value float64) error {
	return nil
}

func (m *MockStorage) SaveCounterMetric(name string, delta int64) error {
	return nil
}

func (m *MockStorage) GetGaugeMetric(name string) (float64, error) {
	return 42.0, nil
}

func (m *MockStorage) GetCounterMetric(name string) (int64, error) {
	return 100, nil
}

func (m *MockStorage) GetAllMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	metrics["gauge_test"] = 42.0
	metrics["counter_test"] = int64(100)
	return metrics
}

func (m *MockStorage) UpdateMetricsBatch(metrics []storage.Metrics) error {
	return nil
}

// setupHandlerBench создает тестовый обработчик для бенчмарков
func setupHandlerBench() *Handler {
	mockStorage := &MockStorage{}
	return &Handler{
		Storage: mockStorage,
	}
}

// BenchmarkHandleGetValue проверяет производительность получения значения метрики
func BenchmarkHandleGetValue(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/value/{type}/{name}", h.HandleGetValue).Methods(http.MethodGet)

	// Создаем запрос
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test_metric", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHandleGetValueJSON проверяет производительность получения значения метрики через JSON
func BenchmarkHandleGetValueJSON(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/value/", h.HandleGetValueJSON).Methods(http.MethodPost)

	// Создаем JSON запрос
	metric := storage.Metrics{
		ID:    "test_metric",
		MType: "gauge",
	}

	body, _ := json.Marshal(metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHandleUpdateMetric проверяет производительность обновления метрики
func BenchmarkHandleUpdateMetric(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/update/{type}/{name}/{value}", h.HandleUpdateMetric).Methods(http.MethodPost)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/gauge/test_metric/42.0", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHandleUpdateMetricJSON проверяет производительность обновления метрики через JSON
func BenchmarkHandleUpdateMetricJSON(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/update/", h.HandleUpdateMetricJSON).Methods(http.MethodPost)

	// Создаем метрику для обновления
	value := 42.0
	metric := storage.Metrics{
		ID:    "test_metric",
		MType: "gauge",
		Value: &value,
	}

	body, _ := json.Marshal(metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHandleGetAllMetrics проверяет производительность получения всех метрик
func BenchmarkHandleGetAllMetrics(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/", h.HandleGetAllMetrics).Methods(http.MethodGet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkHandleUpdatesBatch проверяет производительность пакетного обновления метрик
func BenchmarkHandleUpdatesBatch(b *testing.B) {
	h := setupHandlerBench()
	router := mux.NewRouter()
	router.HandleFunc("/updates/", h.HandleUpdatesBatch).Methods(http.MethodPost)

	// Создаем несколько метрик для обновления
	value := 42.0
	delta := int64(100)
	metrics := []storage.Metrics{
		{ID: "gauge_metric", MType: "gauge", Value: &value},
		{ID: "counter_metric", MType: "counter", Delta: &delta},
	}

	body, _ := json.Marshal(metrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
