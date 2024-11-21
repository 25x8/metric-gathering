package storage

import (
	"encoding/json"
	"fmt"
	"github.com/25x8/metric-gathering/internal/models"
	"log"
	"os"
	"sync"
	"time"
)

type MemStorage struct {
	sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
	filePath string
}

func NewMemStorage(filePath string) *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
		filePath: filePath,
	}
}

func (s *MemStorage) SaveGaugeMetric(name string, value float64) error {
	s.Lock()
	defer s.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) SaveCounterMetric(name string, delta int64) error {
	s.Lock()
	defer s.Unlock()
	s.counters[name] += delta
	return nil
}

func (s *MemStorage) GetGaugeMetric(name string) (float64, error) {
	s.Lock()
	defer s.Unlock()
	value, exists := s.gauges[name]
	if !exists {
		return 0, fmt.Errorf("metric not found")
	}
	return value, nil
}

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

func (s *MemStorage) Flush() error {
	s.Lock()
	defer s.Unlock()

	if s.filePath == "" {
		return nil
	}

	file, err := os.Create(s.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := map[string]interface{}{
		"gauges":   s.gauges,
		"counters": s.counters,
	}

	return json.NewEncoder(file).Encode(data)
}

func (s *MemStorage) Load() error {
	s.Lock()
	defer s.Unlock()

	if s.filePath == "" {
		return nil
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Если файла нет, ничего не делаем
		}
		return err
	}
	defer file.Close()

	data := map[string]interface{}{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	if gauges, ok := data["gauges"].(map[string]interface{}); ok {
		for k, v := range gauges {
			s.gauges[k] = v.(float64)
		}
	}

	if counters, ok := data["counters"].(map[string]interface{}); ok {
		for k, v := range counters {
			s.counters[k] = int64(v.(float64))
		}
	}

	return nil
}

// RunPeriodicSave - запускает периодическое сохранение метрик
func RunPeriodicSave(s *MemStorage, filePath string, storeInterval time.Duration) {
	ticker := time.NewTicker(storeInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.Flush(); err != nil {
			log.Printf("Error saving metrics to file: %v", err)
		}
	}
}

func (s *MemStorage) SaveMetric(metricType, name string, value interface{}) error {
	switch metricType {
	case models.Gauge:
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("invalid value type for gauge: %T", value)
		}
		return s.SaveGaugeMetric(name, v)
	case models.Counter:
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("invalid value type for counter: %T", value)
		}
		return s.SaveCounterMetric(name, v)
	default:
		return fmt.Errorf("unsupported metric type: %s", metricType)
	}
}
