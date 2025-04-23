package storage

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_Basic(t *testing.T) {
	// Создаем временный тестовый файл
	tempFile, err := os.CreateTemp("", "test_metrics_*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Создаем хранилище
	storage := NewMemStorage(tempFile.Name())

	// Тестируем сохранение и получение метрик gauge
	err = storage.SaveGaugeMetric("test_gauge", 123.456)
	require.NoError(t, err)

	value, err := storage.GetGaugeMetric("test_gauge")
	require.NoError(t, err)
	assert.Equal(t, 123.456, value)

	// Тестируем получение несуществующей метрики
	_, err = storage.GetGaugeMetric("non_existent")
	assert.Error(t, err)

	// Тестируем сохранение и получение метрик counter
	err = storage.SaveCounterMetric("test_counter", 42)
	require.NoError(t, err)

	// Counter metrics are cumulative
	err = storage.SaveCounterMetric("test_counter", 10)
	require.NoError(t, err)

	value2, err := storage.GetCounterMetric("test_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(52), value2)

	// Тестируем получение несуществующей метрики counter
	_, err = storage.GetCounterMetric("non_existent")
	assert.Error(t, err)

	// Тестируем получение всех метрик
	allMetrics := storage.GetAllMetrics()
	assert.Len(t, allMetrics, 2)
	assert.Equal(t, 123.456, allMetrics["test_gauge"])
	assert.Equal(t, int64(52), allMetrics["test_counter"])
}

func TestMemStorage_Flush(t *testing.T) {
	// Создаем временный тестовый файл
	tempFile, err := os.CreateTemp("", "test_metrics_*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Создаем хранилище и добавляем метрики
	storage := NewMemStorage(tempFile.Name())
	storage.SaveGaugeMetric("test_gauge", 123.456)
	storage.SaveCounterMetric("test_counter", 42)

	// Сохраняем метрики в файл
	err = storage.Flush()
	require.NoError(t, err)

	// Создаем новое хранилище и загружаем метрики из файла
	storage2 := NewMemStorage(tempFile.Name())
	err = storage2.Load()
	require.NoError(t, err)

	// Проверяем, что метрики загружены правильно
	value, err := storage2.GetGaugeMetric("test_gauge")
	require.NoError(t, err)
	assert.Equal(t, 123.456, value)

	value2, err := storage2.GetCounterMetric("test_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(42), value2)
}

func TestMemStorage_UpdateMetricsBatch(t *testing.T) {
	storage := NewMemStorage("")

	// Подготовка тестовых данных
	gaugeValue := 123.456
	counterValue := int64(42)

	metrics := []Metrics{
		{
			ID:    "batch_gauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "batch_counter",
			MType: "counter",
			Delta: &counterValue,
		},
	}

	// Обновляем метрики пакетно
	err := storage.UpdateMetricsBatch(metrics)
	require.NoError(t, err)

	// Проверяем сохранение gauge метрики
	value, err := storage.GetGaugeMetric("batch_gauge")
	require.NoError(t, err)
	assert.Equal(t, 123.456, value)

	// Проверяем сохранение counter метрики
	value2, err := storage.GetCounterMetric("batch_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(42), value2)

	// Проверяем второе пакетное обновление
	gaugeValue = 654.321
	counterValue = int64(10)

	metrics = []Metrics{
		{
			ID:    "batch_gauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "batch_counter",
			MType: "counter",
			Delta: &counterValue,
		},
	}

	err = storage.UpdateMetricsBatch(metrics)
	require.NoError(t, err)

	// Проверяем обновление gauge метрики
	value, err = storage.GetGaugeMetric("batch_gauge")
	require.NoError(t, err)
	assert.Equal(t, 654.321, value)

	// Counter должен увеличиться
	value2, err = storage.GetCounterMetric("batch_counter")
	require.NoError(t, err)
	assert.Equal(t, int64(52), value2)
}

func TestMemStorage_InvalidMetricsBatch(t *testing.T) {
	storage := NewMemStorage("")

	// Тестирование с недостающими значениями
	nilDelta := []Metrics{
		{
			ID:    "invalid_counter",
			MType: "counter",
			Delta: nil,
		},
	}

	err := storage.UpdateMetricsBatch(nilDelta)
	require.NoError(t, err) // Ошибки нет, но метрика не должна быть сохранена

	_, err = storage.GetCounterMetric("invalid_counter")
	assert.Error(t, err)

	// Тестирование с неподдерживаемым типом
	invalidType := []Metrics{
		{
			ID:    "invalid_type",
			MType: "unsupported",
			Value: new(float64), // Не важно значение, тип не поддерживается
		},
	}

	err = storage.UpdateMetricsBatch(invalidType)
	require.NoError(t, err) // Ошибки нет, но метрика не должна быть сохранена

	_, err = storage.GetGaugeMetric("invalid_type")
	assert.Error(t, err)
}

func TestMemStorage_GetAllMetrics(t *testing.T) {
	storage := NewMemStorage("")

	// Проверка на пустом хранилище
	allMetrics := storage.GetAllMetrics()
	assert.Empty(t, allMetrics)

	// Добавляем метрики
	storage.SaveGaugeMetric("gauge1", 1.1)
	storage.SaveGaugeMetric("gauge2", 2.2)
	storage.SaveCounterMetric("counter1", 3)
	storage.SaveCounterMetric("counter2", 4)

	// Проверяем, что все метрики возвращаются
	allMetrics = storage.GetAllMetrics()
	assert.Len(t, allMetrics, 4)
	assert.Equal(t, 1.1, allMetrics["gauge1"])
	assert.Equal(t, 2.2, allMetrics["gauge2"])
	assert.Equal(t, int64(3), allMetrics["counter1"])
	assert.Equal(t, int64(4), allMetrics["counter2"])
}

func TestMemStorage_EmptyFilePath(t *testing.T) {
	storage := NewMemStorage("")

	// Добавляем метрики
	storage.SaveGaugeMetric("test_gauge", 123.456)
	storage.SaveCounterMetric("test_counter", 42)

	// Проверяем, что Flush не вернет ошибку при пустом пути к файлу
	err := storage.Flush()
	assert.NoError(t, err)

	// Проверяем, что Load не вернет ошибку при пустом пути к файлу
	err = storage.Load()
	assert.NoError(t, err)
}

func TestMemStorage_InvalidFile(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "test_invalid_file")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Создаем файл с некорректным JSON
	invalidFilePath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFilePath, []byte("{invalid json"), 0644)
	require.NoError(t, err)

	// Создаем хранилище с указанием на некорректный файл
	storage := NewMemStorage(invalidFilePath)

	// Проверяем, что Load вернет ошибку при чтении некорректного JSON
	err = storage.Load()
	assert.Error(t, err)
}

func TestMemStorage_RetryOperation(t *testing.T) {
	// Мокаем операцию, которая всегда возвращает ошибку
	operation := func() error {
		return &os.PathError{
			Op:   "open",
			Path: "/nonexistent",
			Err:  syscall.EACCES,
		}
	}

	// Проверяем, что после всех попыток ошибка все равно будет возвращена
	err := retryFileOperation(operation)
	assert.Error(t, err)

	// Проверяем, что retryFileOperation возвращает nil, если операция успешна
	err = retryFileOperation(func() error {
		return nil
	})
	assert.NoError(t, err)
}

func TestRunPeriodicSaveShort(t *testing.T) {
	// Пропускаем тест, т.к. нельзя переназначить RunPeriodicSave
	t.Skip("Cannot mock RunPeriodicSave function")
}

func TestMemStorage_FlushNonexistentDir(t *testing.T) {
	// Создаем хранилище с несуществующим каталогом
	nonExistentPath := "/tmp/nonexistent/dir/test_metrics.json"
	storage := NewMemStorage(nonExistentPath)

	// Добавляем метрики
	storage.SaveGaugeMetric("test_gauge", 123.456)
	storage.SaveCounterMetric("test_counter", 42)

	// Проверяем, что Flush вернет ошибку при несуществующем каталоге
	err := storage.Flush()
	assert.Error(t, err)
}

func TestFileRetriableError(t *testing.T) {
	// Тест для isFileRetriableError с nil
	assert.False(t, isFileRetriableError(nil))

	// Тест для isFileRetriableError с обычной ошибкой
	assert.False(t, isFileRetriableError(errors.New("normal error")))

	// Тест для isFileRetriableError с временной ошибкой
	pathErr := &os.PathError{
		Op:   "open",
		Path: "/test",
		Err:  syscall.EAGAIN,
	}
	assert.True(t, isFileRetriableError(pathErr))

	// Тест для isFileRetriableError с другой временной ошибкой
	pathErr = &os.PathError{
		Op:   "open",
		Path: "/test",
		Err:  syscall.EACCES,
	}
	assert.True(t, isFileRetriableError(pathErr))

	// Тест для isFileRetriableError с нетемпоральной ошибкой
	pathErr = &os.PathError{
		Op:   "open",
		Path: "/test",
		Err:  syscall.ENOENT,
	}
	assert.False(t, isFileRetriableError(pathErr))
}

func TestMemStorage_NonexistentFileLoad(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "test_nonexistent_file")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Создаем путь к несуществующему файлу
	nonexistentFilePath := filepath.Join(tempDir, "nonexistent.json")

	// Создаем хранилище с указанием на несуществующий файл
	storage := NewMemStorage(nonexistentFilePath)

	// Проверяем, что Load не вернет ошибку при несуществующем файле
	err = storage.Load()
	assert.NoError(t, err)

	// Проверяем, что метрики пустые
	_, err = storage.GetGaugeMetric("some_metric")
	assert.Error(t, err)
}

func TestMemStorage_RetryFileOperationSuccess(t *testing.T) {
	// Создаем временный тестовый файл
	tempFile, err := os.CreateTemp("", "test_retry_*.json")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Счетчик вызовов
	callCount := 0
	expectedCallCount := 1

	// Операция, которая завершается успешно с первого раза
	operation := func() error {
		callCount++
		return nil
	}

	// Выполняем операцию с повторами
	err = retryFileOperation(operation)
	assert.NoError(t, err)
	assert.Equal(t, expectedCallCount, callCount)
}

func TestMemStorage_RetryFileOperationNonRetriableError(t *testing.T) {
	// Счетчик вызовов
	callCount := 0
	expectedCallCount := 1

	// Операция, которая завершается с нетемпоральной ошибкой
	operation := func() error {
		callCount++
		return errors.New("non-retriable error")
	}

	// Выполняем операцию с повторами
	err := retryFileOperation(operation)
	assert.Error(t, err)
	assert.Equal(t, expectedCallCount, callCount)
}

func TestMemStorage_RetryFileOperationTemporaryError(t *testing.T) {
	// Счетчик вызовов
	callCount := 0
	maxRetries := 4 // Соответствует константе в retryFileOperation

	// Операция, которая всегда завершается с временной ошибкой
	operation := func() error {
		callCount++
		return &os.PathError{
			Op:   "open",
			Path: "/test",
			Err:  syscall.EAGAIN,
		}
	}

	// Выполняем операцию с повторами
	err := retryFileOperation(operation)
	assert.Error(t, err)
	assert.Equal(t, maxRetries, callCount)
}

func TestMemStorage_FlushFileError(t *testing.T) {
	// Создаем хранилище с путем к каталогу (вместо файла)
	// Это приведет к ошибке при создании файла
	dirPath, err := os.MkdirTemp("", "test_dir_error")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	storage := NewMemStorage(dirPath) // Путь к каталогу вместо файла

	// Добавляем метрики
	storage.SaveGaugeMetric("test_gauge", 123.456)

	// Проверяем, что Flush вернет ошибку
	err = storage.Flush()
	assert.Error(t, err)
}
