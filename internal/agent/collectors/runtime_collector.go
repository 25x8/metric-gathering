package collectors

import (
	"github.com/25x8/metric-gathering/internal/agent/storage"
	"math/rand"
	"runtime"
	"sync"
)

// MetricsCollector - структура для сбора метрик
type MetricsCollector struct {
	PollCount     int64
	lastPollCount int64
	mu            sync.Mutex
	metrics       map[string]interface{}
}

// NewMetricsCollector - конструктор для MetricsCollector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]interface{}),
	}
}

// Collect - метод для сбора метрик
func (c *MetricsCollector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.metrics["Alloc"] = float64(memStats.Alloc)
	c.metrics["BuckHashSys"] = float64(memStats.BuckHashSys)
	c.metrics["Frees"] = float64(memStats.Frees)
	c.metrics["GCCPUFraction"] = memStats.GCCPUFraction // Это уже float64
	c.metrics["GCSys"] = float64(memStats.GCSys)
	c.metrics["HeapAlloc"] = float64(memStats.HeapAlloc)
	c.metrics["HeapIdle"] = float64(memStats.HeapIdle)
	c.metrics["HeapInuse"] = float64(memStats.HeapInuse)
	c.metrics["HeapObjects"] = float64(memStats.HeapObjects)
	c.metrics["HeapReleased"] = float64(memStats.HeapReleased)
	c.metrics["HeapSys"] = float64(memStats.HeapSys)
	c.metrics["LastGC"] = float64(memStats.LastGC)
	c.metrics["Lookups"] = float64(memStats.Lookups)
	c.metrics["MCacheInuse"] = float64(memStats.MCacheInuse)
	c.metrics["MCacheSys"] = float64(memStats.MCacheSys)
	c.metrics["MSpanInuse"] = float64(memStats.MSpanInuse)
	c.metrics["MSpanSys"] = float64(memStats.MSpanSys)
	c.metrics["Mallocs"] = float64(memStats.Mallocs)
	c.metrics["NextGC"] = float64(memStats.NextGC)
	c.metrics["NumForcedGC"] = float64(memStats.NumForcedGC)
	c.metrics["NumGC"] = float64(memStats.NumGC)
	c.metrics["OtherSys"] = float64(memStats.OtherSys)
	c.metrics["PauseTotalNs"] = float64(memStats.PauseTotalNs)
	c.metrics["StackInuse"] = float64(memStats.StackInuse)
	c.metrics["StackSys"] = float64(memStats.StackSys)
	c.metrics["Sys"] = float64(memStats.Sys)
	c.metrics["TotalAlloc"] = float64(memStats.TotalAlloc)
	c.metrics["RandomValue"] = rand.Float64()

	c.PollCount++
	c.metrics["PollCount"] = c.PollCount

}

// CollectAndStore - метод для сбора метрик и их сохранения в хранилище
func (c *MetricsCollector) CollectAndStore(store *storage.MemStorage) error {
	c.Collect()
	metrics := c.GetMetrics()

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

	c.mu.Lock()
	c.lastPollCount = c.PollCount
	defer c.mu.Unlock()

	return nil
}

func (c *MetricsCollector) GetMetrics() map[string]interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()

	metricsCopy := make(map[string]interface{})
	for k, v := range c.metrics {
		if k == "PollCount" {
			// Compute delta for the counter
			delta := c.PollCount - c.lastPollCount
			metricsCopy[k] = delta
		} else {
			metricsCopy[k] = v
		}
	}

	return metricsCopy
}
