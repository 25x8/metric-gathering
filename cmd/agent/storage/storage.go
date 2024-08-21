package storage

import (
	"errors"
	"sync"
)

// MetricRepository - интерфейс для работы с хранилищем метрик
type MetricRepository interface {
	SaveGaugeMetric(name string, value float64) error
	SaveCounterMetric(name string, value int64) error
	GetGaugeMetric(name string) (float64, error)
	GetCounterMetric(name string) (int64, error)
}

// MemStorage - структура для хранения метрик в памяти
type MemStorage struct {
	sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
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
		return 0, errors.New("metric not found")
	}
	return value, nil
}

// GetCounterMetric - получает значение метрики типа counter
func (s *MemStorage) GetCounterMetric(name string) (int64, error) {
	s.Lock()
	defer s.Unlock()
	value, exists := s.counters[name]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}
