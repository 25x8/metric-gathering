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
