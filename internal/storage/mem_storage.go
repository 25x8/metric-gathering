package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
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

	return retryFileOperation(func() error {
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
	})
}

func (s *MemStorage) Load() error {
	s.Lock()
	defer s.Unlock()

	if s.filePath == "" {
		return nil
	}

	return retryFileOperation(func() error {
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
				if value, ok := v.(float64); ok {
					s.gauges[k] = value
				}
			}
		}

		if counters, ok := data["counters"].(map[string]interface{}); ok {
			for k, v := range counters {
				if value, ok := v.(float64); ok {
					s.counters[k] = int64(value)
				}
			}
		}

		return nil
	})
}

// RunPeriodicSave - запускает периодическое сохранение метрик
func RunPeriodicSave(s *MemStorage, filePath string, storeInterval time.Duration) {
	// Используем ticker для точного периодического выполнения
	ticker := time.NewTicker(storeInterval)
	defer ticker.Stop()

	for {
		// Ждем следующего тика
		<-ticker.C

		// Сохраняем метрики
		if err := s.Flush(); err != nil {
			log.Printf("Error saving metrics to file: %v", err)
		}
	}
}

func (s *MemStorage) UpdateMetricsBatch(metrics []Metrics) error {
	s.Lock()
	defer s.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case "counter":
			if metric.Delta == nil {
				continue
			}
			s.counters[metric.ID] += *metric.Delta
		case "gauge":
			if metric.Value == nil {
				continue
			}
			s.gauges[metric.ID] = *metric.Value
		default:
			continue
		}
	}
	return nil
}

// retryFileOperation выполняет операцию с файлами с повторными попытками в случае временных ошибок
func retryFileOperation(operation func() error) error {
	maxRetries := 4 // Первоначальная попытка + 3 дополнительных
	var err error
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		if !isFileRetriableError(err) {
			return err
		}
		if i < len(delays) {
			time.Sleep(delays[i])
		}
	}
	return err
}

// isFileRetriableError проверяет, является ли ошибка файловой системы временной
func isFileRetriableError(err error) bool {
	if err == nil {
		return false
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		var errno syscall.Errno
		if errors.As(pathErr.Err, &errno) {
			if errors.Is(errno, syscall.EAGAIN) || errors.Is(errno, syscall.EACCES) {
				return true
			}
		}
	}
	return false
}
