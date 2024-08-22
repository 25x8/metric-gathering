package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveAndRetrieveGaugeMetric(t *testing.T) {
	store := NewMemStorage()

	// Сохраняем метрику типа gauge
	err := store.SaveGaugeMetric("Alloc", 12345.67)
	assert.NoError(t, err)

	// Извлекаем и проверяем значение
	value, err := store.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Equal(t, 12345.67, value)
}

func TestSaveAndRetrieveCounterMetric(t *testing.T) {
	store := NewMemStorage()

	// Сохраняем метрику типа counter
	err := store.SaveCounterMetric("PollCount", 1)
	assert.NoError(t, err)
	err = store.SaveCounterMetric("PollCount", 2)
	assert.NoError(t, err)

	// Извлекаем и проверяем значение
	value, err := store.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), value)
}

func TestGetNonExistentMetric(t *testing.T) {
	store := NewMemStorage()

	// Попытка получить несуществующую метрику
	_, err := store.GetGaugeMetric("NonExistent")
	assert.Error(t, err)
	assert.Equal(t, "metric not found", err.Error())
}

func TestGetAllMetrics(t *testing.T) {
	store := NewMemStorage()

	// Сохраняем несколько метрик
	store.SaveGaugeMetric("Alloc", 12345.67)
	store.SaveCounterMetric("PollCount", 3)

	// Извлекаем все метрики
	allMetrics := store.GetAllMetrics()

	// Проверяем значения
	assert.Equal(t, 12345.67, allMetrics["Alloc"])
	assert.Equal(t, int64(3), allMetrics["PollCount"])
}
