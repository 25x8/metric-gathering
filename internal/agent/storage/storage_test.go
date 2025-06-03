package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()
	assert.NotNil(t, storage)
	assert.NotNil(t, storage.gauges)
	assert.NotNil(t, storage.counters)
	assert.Empty(t, storage.gauges)
	assert.Empty(t, storage.counters)
}

func TestMemStorage_SaveAndGetGaugeMetric(t *testing.T) {
	storage := NewMemStorage()

	// Сохраняем метрику
	err := storage.SaveGaugeMetric("test_gauge", 123.456)
	assert.NoError(t, err)

	// Получаем сохраненную метрику
	value, err := storage.GetGaugeMetric("test_gauge")
	assert.NoError(t, err)
	assert.Equal(t, 123.456, value)

	// Пробуем получить несуществующую метрику
	_, err = storage.GetGaugeMetric("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemStorage_SaveAndGetCounterMetric(t *testing.T) {
	storage := NewMemStorage()

	// Сохраняем метрику
	err := storage.SaveCounterMetric("test_counter", 42)
	assert.NoError(t, err)

	// Получаем сохраненную метрику
	value, err := storage.GetCounterMetric("test_counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), value)

	// Добавляем значение к существующей метрике
	err = storage.SaveCounterMetric("test_counter", 10)
	assert.NoError(t, err)

	// Проверяем, что значение увеличилось
	value, err = storage.GetCounterMetric("test_counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(52), value)

	// Пробуем получить несуществующую метрику
	_, err = storage.GetCounterMetric("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	storage := NewMemStorage()

	// Пустое хранилище
	metrics := storage.GetAllMetrics()
	assert.Empty(t, metrics)

	// Добавляем метрики
	storage.SaveGaugeMetric("gauge1", 1.1)
	storage.SaveGaugeMetric("gauge2", 2.2)
	storage.SaveCounterMetric("counter1", 10)
	storage.SaveCounterMetric("counter2", 20)

	// Получаем все метрики
	metrics = storage.GetAllMetrics()
	assert.Len(t, metrics, 4)
	assert.Equal(t, 1.1, metrics["gauge1"])
	assert.Equal(t, 2.2, metrics["gauge2"])
	assert.Equal(t, int64(10), metrics["counter1"])
	assert.Equal(t, int64(20), metrics["counter2"])
}
