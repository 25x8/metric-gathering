package collectors

import (
	"github.com/25x8/metric-gathering/cmd/agent/storage"
	"math/rand"
	"runtime"
	"sync"
)

// MetricsCollector - структура для сбора метрик
type MetricsCollector struct {
	PollCount int64
	mu        sync.Mutex
}

// NewMetricsCollector - конструктор для MetricsCollector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// Collect - метод для сбора метрик
func (c *MetricsCollector) Collect() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.PollCount++

	metrics := map[string]interface{}{
		"Alloc":         float64(memStats.Alloc),
		"BuckHashSys":   float64(memStats.BuckHashSys),
		"Frees":         float64(memStats.Frees),
		"GCCPUFraction": float64(memStats.GCCPUFraction),
		"GCSys":         float64(memStats.GCSys),
		"HeapAlloc":     float64(memStats.HeapAlloc),
		"HeapIdle":      float64(memStats.HeapIdle),
		"HeapInuse":     float64(memStats.HeapInuse),
		"HeapObjects":   float64(memStats.HeapObjects),
		"HeapReleased":  float64(memStats.HeapReleased),
		"HeapSys":       float64(memStats.HeapSys),
		"LastGC":        float64(memStats.LastGC),
		"Lookups":       float64(memStats.Lookups),
		"MCacheInuse":   float64(memStats.MCacheInuse),
		"MCacheSys":     float64(memStats.MCacheSys),
		"MSpanInuse":    float64(memStats.MSpanInuse),
		"MSpanSys":      float64(memStats.MSpanSys),
		"Mallocs":       float64(memStats.Mallocs),
		"NextGC":        float64(memStats.NextGC),
		"NumForcedGC":   float64(memStats.NumForcedGC),
		"NumGC":         float64(memStats.NumGC),
		"OtherSys":      float64(memStats.OtherSys),
		"PauseTotalNs":  float64(memStats.PauseTotalNs),
		"StackInuse":    float64(memStats.StackInuse),
		"StackSys":      float64(memStats.StackSys),
		"Sys":           float64(memStats.Sys),
		"TotalAlloc":    float64(memStats.TotalAlloc),
		"PollCount":     int64(c.PollCount), // Явное указание типа int64
		"RandomValue":   rand.Float64(),     // Random gauge value
	}

	return metrics
}

// CollectAndStore - метод для сбора метрик и их сохранения в хранилище
func (c *MetricsCollector) CollectAndStore(store *storage.MemStorage) error {
	metrics := c.Collect()

	for name, value := range metrics {
		switch v := value.(type) {
		case float64:
			if err := store.SaveGaugeMetric(name, v); err != nil {
				return err
			}
		case int64:
			if err := store.SaveCounterMetric(name, v); err != nil {
				return err
			}
		default:
			// Игнорировать неподдерживаемые типы
		}
	}

	return nil
}

// CollectBatch - метод для сбора метрик в формате для отправки батчами
func (c *MetricsCollector) CollectBatch() []map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.PollCount++

	// Формируем массив метрик
	metrics := []map[string]interface{}{
		{"name": "Alloc", "type": "gauge", "value": float64(memStats.Alloc)},
		{"name": "BuckHashSys", "type": "gauge", "value": float64(memStats.BuckHashSys)},
		{"name": "Frees", "type": "gauge", "value": float64(memStats.Frees)},
		{"name": "GCCPUFraction", "type": "gauge", "value": float64(memStats.GCCPUFraction)},
		{"name": "GCSys", "type": "gauge", "value": float64(memStats.GCSys)},
		{"name": "HeapAlloc", "type": "gauge", "value": float64(memStats.HeapAlloc)},
		{"name": "HeapIdle", "type": "gauge", "value": float64(memStats.HeapIdle)},
		{"name": "HeapInuse", "type": "gauge", "value": float64(memStats.HeapInuse)},
		{"name": "HeapObjects", "type": "gauge", "value": float64(memStats.HeapObjects)},
		{"name": "HeapReleased", "type": "gauge", "value": float64(memStats.HeapReleased)},
		{"name": "HeapSys", "type": "gauge", "value": float64(memStats.HeapSys)},
		{"name": "LastGC", "type": "gauge", "value": float64(memStats.LastGC)},
		{"name": "Lookups", "type": "gauge", "value": float64(memStats.Lookups)},
		{"name": "MCacheInuse", "type": "gauge", "value": float64(memStats.MCacheInuse)},
		{"name": "MCacheSys", "type": "gauge", "value": float64(memStats.MCacheSys)},
		{"name": "MSpanInuse", "type": "gauge", "value": float64(memStats.MSpanInuse)},
		{"name": "MSpanSys", "type": "gauge", "value": float64(memStats.MSpanSys)},
		{"name": "Mallocs", "type": "gauge", "value": float64(memStats.Mallocs)},
		{"name": "NextGC", "type": "gauge", "value": float64(memStats.NextGC)},
		{"name": "NumForcedGC", "type": "gauge", "value": float64(memStats.NumForcedGC)},
		{"name": "NumGC", "type": "gauge", "value": float64(memStats.NumGC)},
		{"name": "OtherSys", "type": "gauge", "value": float64(memStats.OtherSys)},
		{"name": "PauseTotalNs", "type": "gauge", "value": float64(memStats.PauseTotalNs)},
		{"name": "StackInuse", "type": "gauge", "value": float64(memStats.StackInuse)},
		{"name": "StackSys", "type": "gauge", "value": float64(memStats.StackSys)},
		{"name": "Sys", "type": "gauge", "value": float64(memStats.Sys)},
		{"name": "TotalAlloc", "type": "gauge", "value": float64(memStats.TotalAlloc)},
		{"name": "PollCount", "type": "counter", "value": int64(c.PollCount)},
		{"name": "RandomValue", "type": "gauge", "value": rand.Float64()},
	}

	return metrics
}
