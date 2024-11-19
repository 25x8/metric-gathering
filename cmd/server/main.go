package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/25x8/metric-gathering/internal/logger"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

// MemStorage - структура для хранения метрик в памяти
type MemStorage struct {
	sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение для counter
	Value *float64 `json:"value,omitempty"` // значение для gauge
}

// NewMemStorage - конструктор для MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// SaveGaugeMetric - сохраняет метрику типа gauge
func (s *MemStorage) SaveGaugeMetric(name string, value float64) error {
	s.Lock()
	defer s.Unlock()
	s.gauges[name] = value
	return nil
}

// SaveCounterMetric - сохраняет метрику типа counter
func (s *MemStorage) SaveCounterMetric(name string, value int64) error {
	s.Lock()
	defer s.Unlock()
	s.counters[name] += value
	return nil
}

// GetGaugeMetric - получает значение метрики типа gauge
func (s *MemStorage) GetGaugeMetric(name string) (float64, error) {
	s.Lock()
	defer s.Unlock()
	value, exists := s.gauges[name]
	if !exists {
		return 0, fmt.Errorf("metric not found")
	}
	return value, nil
}

// GetCounterMetric - получает значение метрики типа counter
func (s *MemStorage) GetCounterMetric(name string) (int64, error) {
	s.Lock()
	defer s.Unlock()
	value, exists := s.counters[name]
	if !exists {
		return 0, fmt.Errorf("metric not found")
	}
	return value, nil
}

// GetAllMetrics - возвращает все метрики
func (s *MemStorage) GetAllMetrics() map[string]interface{} {
	s.Lock()
	defer s.Unlock()

	allMetrics := make(map[string]interface{})
	for name, value := range s.gauges {
		allMetrics[name] = value
	}
	for name, value := range s.counters {
		allMetrics[name] = value
	}
	return allMetrics
}

// handleGetValue - обработчик для получения значения метрики
func handleGetValue(s *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		metricType := vars["type"]
		metricName := vars["name"]

		switch metricType {
		case "gauge":
			value, err := s.GetGaugeMetric(metricName)
			if err != nil {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%v", value)

		case "counter":
			value, err := s.GetCounterMetric(metricName)
			if err != nil {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			fmt.Fprintf(w, "%v", value)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
		}
	}
}

func handleGetValueJSON(s *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
			return
		}

		var m Metrics
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&m)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if m.ID == "" || m.MType == "" {
			http.Error(w, "ID and MType are required", http.StatusBadRequest)
			return
		}

		switch m.MType {
		case "gauge":
			value, err := s.GetGaugeMetric(m.ID)
			if err != nil {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			m.Value = &value
		case "counter":
			delta, err := s.GetCounterMetric(m.ID)
			if err != nil {
				http.Error(w, "Metric not found", http.StatusNotFound)
				return
			}
			m.Delta = &delta
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

// handleGetAllMetrics - обработчик для получения всех метрик в виде HTML
func handleGetAllMetrics(s *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allMetrics := s.GetAllMetrics()

		tmpl := `
		<html>
		<head><title>Metrics</title></head>
		<body>
			<h1>All Metrics</h1>
			<table border="1">
				<tr>
					<th>Name</th>
					<th>Value</th>
				</tr>
				{{range $name, $value := .}}
				<tr>
					<td>{{$name}}</td>
					<td>{{$value}}</td>
				</tr>
				{{end}}
			</table>
		</body>
		</html>
		`

		t := template.Must(template.New("metrics").Parse(tmpl))
		t.Execute(w, allMetrics)
	}
}

// handleUpdateMetric - обработчик для обновления метрики
func handleUpdateMetric(s *MemStorage) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		metricType := vars["type"]
		metricName := vars["name"]
		metricValue := vars["value"]

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
			s.SaveGaugeMetric(metricName, value)

		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			s.SaveCounterMetric(metricName, value)

		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Metric %s updated", metricName)
	})
}

func handleUpdateMetricJSON(s *MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
			return
		}

		var m Metrics
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&m)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if m.ID == "" || m.MType == "" {
			http.Error(w, "ID and MType are required", http.StatusBadRequest)
			return
		}

		switch m.MType {
		case "gauge":
			if m.Value == nil {
				http.Error(w, "Value is required for gauge", http.StatusBadRequest)
				return
			}
			s.SaveGaugeMetric(m.ID, *m.Value)
			// Обновляем значение для ответа
			updatedValue, _ := s.GetGaugeMetric(m.ID)
			m.Value = &updatedValue
		case "counter":
			if m.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			s.SaveCounterMetric(m.ID, *m.Delta)
			// Обновляем значение для ответа
			updatedDelta, _ := s.GetCounterMetric(m.ID)
			m.Delta = &updatedDelta
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

func main() {
	// Определение флага для адреса сервера
	addr := flag.String("a", "localhost:8080", "HTTP server address")

	// Парсинг флагов
	flag.Parse()

	// Чтение переменной окружения
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		*addr = envAddr
	}

	storage := NewMemStorage()
	r := mux.NewRouter()

	// logger
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	defer logger.Sync()

	// Маршруты для обновления метрик и получения их значений
	r.Handle("/update/{type}/{name}/{value}", logger.RequestLogger(handleUpdateMetric(storage))).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", logger.RequestLogger(handleGetValue(storage))).Methods(http.MethodGet)
	r.Handle("/", logger.RequestLogger(handleGetAllMetrics(storage))).Methods(http.MethodGet)

	// Маршруты для работы с JSON
	r.Handle("/update/", logger.RequestLogger(handleUpdateMetricJSON(storage))).Methods(http.MethodPost)
	r.Handle("/value/", logger.RequestLogger(handleGetValueJSON(storage))).Methods(http.MethodPost)

	log.Printf("Server started at %s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
