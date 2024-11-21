package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// MemStorage - структура для хранения метрик в памяти
type MemStorage struct {
	sync.Mutex
	gauges   map[string]float64
	counters map[string]int64
}

// MemStorageData - структура для сериализации метрик
type MemStorageData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

// NewMemStorage - конструктор для MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// SaveGaugeMetric - сохраняет метрику типа gauge
func (s *MemStorage) SaveGaugeMetric(name string, value float64) {
	s.Lock()
	s.gauges[name] = value
	s.Unlock()
}

// SaveCounterMetric - сохраняет метрику типа counter
func (s *MemStorage) SaveCounterMetric(name string, value int64) {
	s.Lock()
	s.counters[name] += value
	s.Unlock()
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
func SaveToFile(s *MemStorage, filePath string) error {
	s.Lock()
	defer s.Unlock()

	if filePath == "" {
		// Если путь к файлу не задан, пропускаем сохранение
		return nil
	}

	data := MemStorageData{
		Gauges:   s.gauges,
		Counters: s.counters,
	}

	file, err := os.Create(filePath)
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
func LoadFromFile(s *MemStorage, filePath string) error {
	s.Lock()
	defer s.Unlock()

	if filePath == "" {
		// Если путь к файлу не задан, пропускаем загрузку
		return nil
	}

	file, err := os.Open(filePath)
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
func RunPeriodicSave(s *MemStorage, filePath string, storeInterval time.Duration) {
	ticker := time.NewTicker(storeInterval)
	defer ticker.Stop()
	for range ticker.C {
		if err := SaveToFile(s, filePath); err != nil {
			log.Printf("Error saving metrics to file: %v", err)
		}
	}
}
