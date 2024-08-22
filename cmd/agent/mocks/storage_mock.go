package mocks

import (
	"errors"
	"sync"
)

type MockStorage struct {
	mu       sync.Mutex
	Gauges   map[string]float64
	Counters map[string]int64
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}
}

func (m *MockStorage) SaveGaugeMetric(name string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Gauges[name] = value
	return nil
}

func (m *MockStorage) SaveCounterMetric(name string, value int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Counters[name] += value
	return nil
}

func (m *MockStorage) GetGaugeMetric(name string) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, exists := m.Gauges[name]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}

func (m *MockStorage) GetCounterMetric(name string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	value, exists := m.Counters[name]
	if !exists {
		return 0, errors.New("metric not found")
	}
	return value, nil
}
