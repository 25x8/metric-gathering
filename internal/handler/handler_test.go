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

// setupHandler creates a handler with in-memory storage for testing
func setupHandler() *Handler {
	memStorage := storage.NewMemStorage("") // Empty string for file path in tests

	// Setup some test data
	memStorage.SaveGaugeMetric("gauge_test", 42.0)
	memStorage.SaveCounterMetric("counter_test", int64(100))

	return &Handler{
		Storage: memStorage,
	}
}

// TestHandleGetValue проверяет получение метрики через URL
func TestHandleGetValue(t *testing.T) {
	h := setupHandler()
	router := mux.NewRouter()
	router.HandleFunc("/value/{type}/{name}", h.HandleGetValue).Methods(http.MethodGet)

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "Valid gauge metric",
			url:        "/value/gauge/gauge_test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Valid counter metric",
			url:        "/value/counter/counter_test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid metric type",
			url:        "/value/invalid/test",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetValue() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestHandleGetValueJSON проверяет получение метрики через JSON
func TestHandleGetValueJSON(t *testing.T) {
	h := setupHandler()
	router := mux.NewRouter()
	router.HandleFunc("/value/", h.HandleGetValueJSON).Methods(http.MethodPost)

	tests := []struct {
		name       string
		metric     storage.Metrics
		wantStatus int
	}{
		{
			name: "Valid gauge metric",
			metric: storage.Metrics{
				ID:    "gauge_test",
				MType: "gauge",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Valid counter metric",
			metric: storage.Metrics{
				ID:    "counter_test",
				MType: "counter",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Invalid metric type",
			metric: storage.Metrics{
				ID:    "test",
				MType: "invalid",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.metric)
			req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleGetValueJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusOK {
				var response storage.Metrics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				if response.ID != tt.metric.ID || response.MType != tt.metric.MType {
					t.Errorf("Response does not match request: got %+v, want ID=%s, MType=%s",
						response, tt.metric.ID, tt.metric.MType)
				}

				// Проверяем, что значение было установлено
				if tt.metric.MType == "gauge" && response.Value == nil {
					t.Error("Expected non-nil Value for gauge metric")
				} else if tt.metric.MType == "counter" && response.Delta == nil {
					t.Error("Expected non-nil Delta for counter metric")
				}
			}
		})
	}
}

// TestHandleUpdateMetric проверяет обновление метрики через URL
func TestHandleUpdateMetric(t *testing.T) {
	h := setupHandler()
	router := mux.NewRouter()
	router.HandleFunc("/update/{type}/{name}/{value}", h.HandleUpdateMetric).Methods(http.MethodPost)

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "Update gauge metric",
			url:        "/update/gauge/test_gauge/42.0",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Update counter metric",
			url:        "/update/counter/test_counter/100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid metric type",
			url:        "/update/invalid/test/42",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid gauge value",
			url:        "/update/gauge/test/invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid counter value",
			url:        "/update/counter/test/invalid",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleUpdateMetric() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
