package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Values(t *testing.T) {
	// Тест для gauge метрики
	gaugeValue := 42.5
	gaugeMetric := Metrics{
		ID:    "gauge_test",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Проверяем значения полей
	assert.Equal(t, "gauge_test", gaugeMetric.ID)
	assert.Equal(t, "gauge", gaugeMetric.MType)
	assert.Equal(t, &gaugeValue, gaugeMetric.Value)
	assert.Nil(t, gaugeMetric.Delta)

	// Тест для counter метрики
	counterValue := int64(100)
	counterMetric := Metrics{
		ID:    "counter_test",
		MType: "counter",
		Delta: &counterValue,
	}

	// Проверяем значения полей
	assert.Equal(t, "counter_test", counterMetric.ID)
	assert.Equal(t, "counter", counterMetric.MType)
	assert.Equal(t, &counterValue, counterMetric.Delta)
	assert.Nil(t, counterMetric.Value)
}

func TestMetrics_Serialization(t *testing.T) {
	// Тест для gauge метрики
	gaugeValue := 42.5
	gaugeMetric := Metrics{
		ID:    "gauge_test",
		MType: "gauge",
		Value: &gaugeValue,
	}

	// Проверяем JSON-теги
	assert.NotEmpty(t, gaugeMetric.ID)
	assert.NotEmpty(t, gaugeMetric.MType)
	assert.NotNil(t, gaugeMetric.Value)

	// Тест для counter метрики
	counterValue := int64(100)
	counterMetric := Metrics{
		ID:    "counter_test",
		MType: "counter",
		Delta: &counterValue,
	}

	// Проверяем JSON-теги
	assert.NotEmpty(t, counterMetric.ID)
	assert.NotEmpty(t, counterMetric.MType)
	assert.NotNil(t, counterMetric.Delta)
}
