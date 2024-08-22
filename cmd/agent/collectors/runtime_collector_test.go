package collectors

import (
	"github.com/25x8/metric-gathering/cmd/agent/storage"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector_CollectAndStore(t *testing.T) {
	store := storage.NewMemStorage()
	collector := NewMetricsCollector()

	err := collector.CollectAndStore(store)
	assert.NoError(t, err)

	// Проверка метрики типа gauge
	value, err := store.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Greater(t, value, 0.0)

	// Проверка счетчика PollCount
	counterValue, err := store.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), counterValue)
}
