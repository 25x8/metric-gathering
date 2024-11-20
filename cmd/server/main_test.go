package main

import (
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSaveAndRetrieveGaugeMetric(t *testing.T) {
	store := storage.NewMemStorage(0, "") // Используем синхронное сохранение и пустой путь к файлу

	// Сохраняем метрику типа gauge
	err := store.SaveGaugeMetric("Alloc", 12345.67)
	assert.NoError(t, err)

	// Извлекаем и проверяем значение
	value, err := store.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Equal(t, 12345.67, value)
}

func TestSaveAndRetrieveCounterMetric(t *testing.T) {
	store := storage.NewMemStorage(0, "") // Используем синхронное сохранение и пустой путь к файлу

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
	store := storage.NewMemStorage(0, "") // Используем синхронное сохранение и пустой путь к файлу

	// Попытка получить несуществующую метрику
	_, err := store.GetGaugeMetric("NonExistent")
	assert.Error(t, err)
	assert.Equal(t, "metric not found", err.Error())
}

func TestGetAllMetrics(t *testing.T) {
	store := storage.NewMemStorage(0, "") // Используем синхронное сохранение и пустой путь к файлу

	// Сохраняем несколько метрик
	store.SaveGaugeMetric("Alloc", 12345.67)
	store.SaveCounterMetric("PollCount", 3)

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
	defer os.Remove(tmpFile.Name()) // Удаляем файл после теста

	store := storage.NewMemStorage(0, tmpFile.Name())

	// Сохраняем метрики
	err = store.SaveGaugeMetric("Alloc", 12345.67)
	assert.NoError(t, err)
	err = store.SaveCounterMetric("PollCount", 3)
	assert.NoError(t, err)

	// Явно сохраняем метрики в файл
	err = store.SaveToFile()
	assert.NoError(t, err)

	// Создаём новое хранилище и загружаем метрики из файла
	newStore := storage.NewMemStorage(0, tmpFile.Name())
	err = newStore.LoadFromFile()
	assert.NoError(t, err)

	// Проверяем, что метрики загрузились корректно
	gaugeValue, err := newStore.GetGaugeMetric("Alloc")
	assert.NoError(t, err)
	assert.Equal(t, 12345.67, gaugeValue)

	counterValue, err := newStore.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), counterValue)
}
