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

func TestMetricsCollector_CollectBatch(t *testing.T) {
	collector := NewMetricsCollector()

	metrics := collector.CollectBatch()
	if len(metrics) == 0 {
		t.Errorf("Expected metrics batch, got empty")
	}

	found := false
	for _, metric := range metrics {
		if metric["name"] == "Alloc" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected metric 'Alloc' in batch, but not found")
	}
}
