package main

import (
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/25x8/metric-gathering/internal/logger"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// MemStorage - структура для хранения метрик в памяти
type MemStorage struct {
	sync.Mutex
	gauges        map[string]float64
	counters      map[string]int64
	storeInterval time.Duration
	filePath      string
}

// MemStorageData - структура для сериализации метрик
type MemStorageData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение для counter
	Value *float64 `json:"value,omitempty"` // значение для gauge
}

// compressWriter добавляет поддержку gzip для ответа
type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (cw *compressWriter) Write(data []byte) (int, error) {
	return cw.Writer.Write(data)
}

// gzipMiddleware обрабатывает запросы с gzip и добавляет сжатие ответов
func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разжимает тело запроса, если используется gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		// Проверяет, поддерживает ли клиент gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gw := gzip.NewWriter(w)
			defer gw.Close()
			w = &compressWriter{ResponseWriter: w, Writer: gw}
		}

		next.ServeHTTP(w, r)
	})
}

// NewMemStorage - конструктор для MemStorage
func NewMemStorage(storeInterval time.Duration, filePath string) *MemStorage {
	return &MemStorage{
		gauges:        make(map[string]float64),
		counters:      make(map[string]int64),
		storeInterval: storeInterval,
		filePath:      filePath,
	}
}

// SaveGaugeMetric - сохраняет метрику типа gauge
func (s *MemStorage) SaveGaugeMetric(name string, value float64) error {
	s.Lock()
	s.gauges[name] = value
	s.Unlock()

	if s.storeInterval == 0 {
		if err := s.SaveToFile(); err != nil {
			log.Printf("Error saving metrics: %v", err)
			return err
		}
	}
	return nil
}

// SaveCounterMetric - сохраняет метрику типа counter
func (s *MemStorage) SaveCounterMetric(name string, value int64) error {
	s.Lock()
	s.counters[name] += value
	s.Unlock()

	if s.storeInterval == 0 {
		if err := s.SaveToFile(); err != nil {
			log.Printf("Error saving metrics: %v", err)
			return err
		}
	}
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

// SaveToFile - сохраняет метрики в файл
func (s *MemStorage) SaveToFile() error {
	s.Lock()
	defer s.Unlock()

	if s.filePath == "" {
		// Если путь к файлу не задан, пропускаем сохранение
		return nil
	}

	data := MemStorageData{
		Gauges:   s.gauges,
		Counters: s.counters,
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(data)
	if err != nil {
		return err
	}

	return nil
}

// LoadFromFile - загружает метрики из файла
func (s *MemStorage) LoadFromFile() error {
	s.Lock()
	defer s.Unlock()

	if s.filePath == "" {
		// Если путь к файлу не задан, пропускаем загрузку
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	data := MemStorageData{}
	err = decoder.Decode(&data)
	if err != nil {
		return err
	}

	s.gauges = data.Gauges
	s.counters = data.Counters

	return nil
}

// RunPeriodicSave - запускает периодическое сохранение метрик
func (s *MemStorage) RunPeriodicSave() {
	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := s.SaveToFile(); err != nil {
			log.Printf("Error saving metrics to file: %v", err)
		}
	}
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

		w.Header().Set("Content-Type", "text/html")

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
			updatedValue, _ := s.GetGaugeMetric(m.ID)
			m.Value = &updatedValue
		case "counter":
			if m.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			s.SaveCounterMetric(m.ID, *m.Delta)
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
	// Определение флагов
	addrFlag := flag.String("a", "localhost:8080", "HTTP server address")
	storeIntervalFlag := flag.Int("i", 300, "Store interval in seconds (0 for synchronous saving)")
	fileStoragePathFlag := flag.String("f", "/tmp/metrics-db.json", "File storage path")
	restoreFlag := flag.Bool("r", true, "Restore metrics from file at startup")

	// Парсинг флагов
	flag.Parse()

	// Чтение переменных окружения с приоритетом
	addr := *addrFlag
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	var storeInterval time.Duration
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		intervalSec, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Fatalf("Invalid STORE_INTERVAL: %v", err)
		}
		storeInterval = time.Duration(intervalSec) * time.Second
	} else if storeIntervalFlag != nil {
		storeInterval = time.Duration(*storeIntervalFlag) * time.Second
	} else {
		storeInterval = 300 * time.Second
	}

	fileStoragePath := *fileStoragePathFlag
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		fileStoragePath = envFileStoragePath
	}

	restore := *restoreFlag
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		var err error
		restore, err = strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatalf("Invalid RESTORE value: %v", err)
		}
	}

	storage := NewMemStorage(storeInterval, fileStoragePath)

	// Восстановление метрик из файла при старте
	if restore {
		if err := storage.LoadFromFile(); err != nil {
			log.Printf("Error loading metrics from file: %v", err)
		}
	}

	// Запуск периодического сохранения метрик
	if storeInterval > 0 {
		go storage.RunPeriodicSave()
	}

	// Обработка сигнала завершения для сохранения метрик
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := storage.SaveToFile(); err != nil {
			log.Printf("Error saving metrics on shutdown: %v", err)
		}
		os.Exit(0)
	}()

	r := mux.NewRouter()

	// logger
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	defer logger.Sync()

	// Маршруты для обновления метрик и получения их значений
	r.Handle("/update/{type}/{name}/{value}", gzipMiddleware(logger.RequestLogger(handleUpdateMetric(storage)))).Methods(http.MethodPost)
	r.Handle("/value/{type}/{name}", gzipMiddleware(logger.RequestLogger(handleGetValue(storage)))).Methods(http.MethodGet)
	r.Handle("/", gzipMiddleware(logger.RequestLogger(handleGetAllMetrics(storage)))).Methods(http.MethodGet)

	// Маршруты для работы с JSON
	r.Handle("/update/", gzipMiddleware(logger.RequestLogger(handleUpdateMetricJSON(storage)))).Methods(http.MethodPost)
	r.Handle("/value/", gzipMiddleware(logger.RequestLogger(handleGetValueJSON(storage)))).Methods(http.MethodPost)

	log.Printf("Server started at %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
