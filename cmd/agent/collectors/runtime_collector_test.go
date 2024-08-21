package collectors

import (
	"github.com/25x8/metric-gathering/cmd/agent/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsCollector_CollectAndStore(t *testing.T) {
	mockStorage := mocks.NewMockStorage()
	collector := NewMetricsCollector()

	err := collector.CollectAndStore(mockStorage)
	assert.NoError(t, err)

	// Пример проверки сохраненных метрик
	value, err := mockStorage.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Greater(t, value, 0.0)

	counterValue, err := mockStorage.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), counterValue)
}
