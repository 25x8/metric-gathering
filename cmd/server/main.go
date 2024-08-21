package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// MemStorage - структура для хранения метрик
type MemStorage struct {
	sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

// NewMemStorage - инициализация нового хранилища
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

// UpdateGauge - метод для обновления gauge метрики
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.Lock()
	defer s.Unlock()
	s.Gauges[name] = value
}

// UpdateCounter - метод для обновления counter метрики
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.Lock()
	defer s.Unlock()
	s.Counters[name] += value
}

// HTTPHandler - основной обработчик запросов
func (s *MemStorage) HTTPHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка метода запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is supported", http.StatusMethodNotAllowed)
		return
	}

	// Разбор URL
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/update/"), "/")
	if len(parts) != 3 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := parts[0]
	metricName := parts[1]
	metricValue := parts[2]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		s.UpdateGauge(metricName, value)

	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		s.UpdateCounter(metricName, value)

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Metric %s updated", metricName)
}

func main() {
	storage := NewMemStorage()
	http.HandleFunc("/update/", storage.HTTPHandler)

	log.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
