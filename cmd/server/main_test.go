package main

import (
	"os"
	"testing"

	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndRetrieveGaugeMetric(t *testing.T) {
	store := storage.NewMemStorage("") // Создаем новое хранилище без файла

	// Сохраняем метрику типа gauge
	err := store.SaveGaugeMetric("Alloc", 12345.67)
	require.NoError(t, err)

	// Извлекаем и проверяем значение
	value, err := store.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Equal(t, 12345.67, value)
}

func TestSaveAndRetrieveCounterMetric(t *testing.T) {
	store := storage.NewMemStorage("")

	// Сохраняем метрику типа counter
	err := store.SaveCounterMetric("PollCount", 1)
	require.NoError(t, err)

	err = store.SaveCounterMetric("PollCount", 2)
	require.NoError(t, err)

	// Извлекаем и проверяем значение
	value, err := store.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), value)
}

func TestGetNonExistentMetric(t *testing.T) {
	store := storage.NewMemStorage("")

	// Попытка получить несуществующую метрику
	_, err := store.GetGaugeMetric("NonExistent")
	assert.Error(t, err)
	assert.Equal(t, "metric not found", err.Error())
}

func TestGetAllMetrics(t *testing.T) {
	store := storage.NewMemStorage("")

	// Сохраняем несколько метрик
	err := store.SaveGaugeMetric("Alloc", 12345.67)
	require.NoError(t, err)

	err = store.SaveCounterMetric("PollCount", 3)
	require.NoError(t, err)

	// Извлекаем все метрики
	allMetrics := store.GetAllMetrics()

	// Проверяем значения
	assert.Equal(t, 12345.67, allMetrics["Alloc"])
	assert.Equal(t, int64(3), allMetrics["PollCount"])
}

func TestSaveAndLoadMetrics(t *testing.T) {
	// Создаём временный файл
	tmpFile, err := os.CreateTemp("", "metrics_test_*.json")
	assert.NoError(t, err)

	tmpName := tmpFile.Name()
	err = tmpFile.Close()
	require.NoError(t, err)

	defer func() {
		err := os.Remove(tmpName) // Удаляем файл после теста
		if err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	store := storage.NewMemStorage(tmpName)

	// Сохраняем метрики
	err = store.SaveGaugeMetric("Alloc", 12345.67)
	require.NoError(t, err)

	err = store.SaveCounterMetric("PollCount", 3)
	require.NoError(t, err)

	// Явно сохраняем метрики в файл
	err = store.Flush()
	assert.NoError(t, err)

	// Создаём новое хранилище и загружаем метрики из файла
	newStore := storage.NewMemStorage(tmpName)
	err = newStore.Load()
	assert.NoError(t, err)

	// Проверяем, что метрики загрузились корректно
	gaugeValue, err := newStore.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Equal(t, 12345.67, gaugeValue)

	counterValue, err := newStore.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), counterValue)
}
