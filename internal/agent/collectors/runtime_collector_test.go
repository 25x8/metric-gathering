package collectors

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/25x8/metric-gathering/internal/agent/senders"
	"github.com/25x8/metric-gathering/internal/agent/storage"
	"github.com/stretchr/testify/assert"
)

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func TestMetricsCollector_Collect(t *testing.T) {
	collector := NewMetricsCollector()
	collector.Collect()

	collector.mu.Lock()
	defer collector.mu.Unlock()

	// Проверяем, что метрики не пустые
	assert.NotNil(t, collector.metrics)

	// Ожидаемые ключи метрик
	expectedKeys := []string{
		"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc",
		"HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs",
		"NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs", "StackInuse",
		"StackSys", "Sys", "TotalAlloc", "PollCount", "RandomValue",
	}

	for _, key := range expectedKeys {
		_, ok := collector.metrics[key]
		assert.True(t, ok, "Метрика %s не найдена", key)
	}

	// Проверяем типы некоторых метрик
	assert.IsType(t, float64(0), collector.metrics["Alloc"])
	assert.IsType(t, int64(0), collector.metrics["PollCount"])
	assert.IsType(t, float64(0), collector.metrics["RandomValue"])
}

func TestMetricsCollector_PollCount(t *testing.T) {
	collector := NewMetricsCollector()

	collector.Collect()
	collector.mu.Lock()
	pollCount1 := collector.metrics["PollCount"].(int64)
	collector.mu.Unlock()

	collector.Collect()
	collector.mu.Lock()
	pollCount2 := collector.metrics["PollCount"].(int64)
	collector.mu.Unlock()

	assert.Equal(t, pollCount1+1, pollCount2, "PollCount должен увеличиваться на 1")
}

func TestMetricsCollector_CollectAndStore(t *testing.T) {
	collector := NewMetricsCollector()
	store := storage.NewMemStorage()

	err := collector.CollectAndStore(store)
	assert.NoError(t, err)

	// Проверяем, что метрики сохранены в хранилище
	gaugeMetrics := []string{"Alloc", "HeapAlloc", "RandomValue"}
	counterMetrics := []string{"PollCount"}

	for _, name := range gaugeMetrics {
		value, err := store.GetGaugeMetric(name)
		assert.NoError(t, err)
		assert.IsType(t, float64(0), value)
	}

	for _, name := range counterMetrics {
		value, err := store.GetCounterMetric(name)
		assert.NoError(t, err)
		assert.IsType(t, int64(0), value)
	}
}

func TestMetricsCollector_CollectAndStoreMultipleTimes(t *testing.T) {
	collector := NewMetricsCollector()
	store := storage.NewMemStorage()

	// Первый сбор и сохранение метрик
	err := collector.CollectAndStore(store)
	assert.NoError(t, err)

	// Второй сбор и сохранение метрик
	err = collector.CollectAndStore(store)
	assert.NoError(t, err)

	err = collector.CollectAndStore(store)
	assert.NoError(t, err)

	err = collector.CollectAndStore(store)
	assert.NoError(t, err)

	err = collector.CollectAndStore(store)
	assert.NoError(t, err)

	value, err := store.GetCounterMetric("PollCount")
	assert.NoError(t, err)
	assert.Equal(t, int64(5), value, "PollCount должен быть равен 5 после двух сборов")
}

func TestHTTPSender_Send(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "Должен использовать метод POST")
		assert.Equal(t, "/updates/", r.URL.Path, "URL-адрес должен быть /updates/")
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"), "Должен использовать gzip сжатие")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "Тип содержимого должен быть application/json")

		// Распаковываем gzip
		gzipReader, err := gzip.NewReader(r.Body)
		assert.NoError(t, err, "Должен быть корректный gzip-формат")
		defer gzipReader.Close()

		body, err := io.ReadAll(gzipReader)
		assert.NoError(t, err, "Тело запроса должно быть прочитано без ошибок")

		var metrics []Metric
		err = json.Unmarshal(body, &metrics)
		assert.NoError(t, err, "Тело запроса должно быть корректным JSON")

		assert.Greater(t, len(metrics), 0, "Должен быть хотя бы один элемент в метриках")
		assert.Equal(t, "gauge", metrics[0].MType, "Тип метрики должен быть 'counter'")

		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	sender := senders.NewHTTPSender(testServer.URL)

	t.Run("Send empty metrics", func(t *testing.T) {
		err := sender.Send(map[string]interface{}{}, "", nil)
		assert.NoError(t, err, "Отправка пустых метрик не должна вызывать ошибок")
	})

	t.Run("Send valid metrics", func(t *testing.T) {
		metrics := map[string]interface{}{
			"Alloc":       float64(123),
			"BuckHashSys": 56.78,
		}
		err := sender.SendBatch(metrics, nil)
		assert.NoError(t, err, "Отправка валидных метрик не должна вызывать ошибок")

	})

	t.Run("Send invalid metrics", func(t *testing.T) {
		metrics := map[string]interface{}{
			"unsupported_metric": "string_value",
		}
		err := sender.Send(metrics, "", nil)
		assert.NoError(t, err, "Метрики с неподдерживаемыми типами просто игнорируются")
	})

}
