package senders

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тестирование отправки метрик с помощью мокового HTTP-сервера
func TestHTTPSender_Send(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что URL соответствует ожидаемому пути
		assert.Equal(t, "/update/metrics", r.URL.EscapedPath())

		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		// Читаем и распаковываем тело запроса
		var body []byte
		var err error
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			assert.NoError(t, err)
			defer gr.Close()
			body, err = ioutil.ReadAll(gr)
			assert.NoError(t, err)
		} else {
			body, err = ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
		}

		// Декодируем JSON
		var metrics map[string]interface{}
		err = json.Unmarshal(body, &metrics)
		assert.NoError(t, err)

		// Проверяем, что метрики содержат ожидаемые данные
		assert.Equal(t, 12345.67, metrics["Alloc"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPSender с URL тестового сервера
	sender := NewHTTPSender(server.URL + "/update/metrics")

	// Пример метрик
	metrics := map[string]interface{}{
		"Alloc": 12345.67,
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}

func TestHTTPSender_SendCounter(t *testing.T) {
	// Создаем тестовый HTTP-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что URL соответствует ожидаемому пути
		assert.Equal(t, "/update/metrics", r.URL.EscapedPath())

		// Проверяем заголовки
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

		// Читаем и распаковываем тело запроса
		var body []byte
		var err error
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			assert.NoError(t, err)
			defer gr.Close()
			body, err = ioutil.ReadAll(gr)
			assert.NoError(t, err)
		} else {
			body, err = ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
		}

		// Декодируем JSON
		var metrics map[string]interface{}
		err = json.Unmarshal(body, &metrics)
		assert.NoError(t, err)

		// Проверяем, что метрики содержат ожидаемые данные
		assert.Equal(t, float64(1), metrics["PollCount"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем HTTPSender с URL тестового сервера
	sender := NewHTTPSender(server.URL + "/update/metrics")

	// Пример метрик
	metrics := map[string]interface{}{
		"PollCount": int64(1),
	}

	// Отправляем метрики
	err := sender.Send(metrics)
	assert.NoError(t, err)
}
