package senders

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

// HTTPSender - структура для отправки метрик на сервер
type HTTPSender struct {
	ServerURL string
}

// NewHTTPSender - конструктор для HTTPSender
func NewHTTPSender(serverURL string) *HTTPSender {
	return &HTTPSender{
		ServerURL: serverURL,
	}
}

// Send - метод для отправки метрик на сервер
func (s *HTTPSender) Send(metrics map[string]interface{}) error {
	for key, value := range metrics {
		var metricType string
		switch value.(type) {
		case int64:
			metricType = "counter"
		case float64:
			metricType = "gauge"
		default:
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%v", s.ServerURL, metricType, key, value)

		// Сжатие тела запроса (в данном случае тело пустое, но это пригодится для реальных данных)
		var compressedBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressedBody)
		_, err := gzipWriter.Write([]byte{}) // Пустое тело для текущего примера
		if err != nil {
			return err
		}
		gzipWriter.Close()

		req, err := http.NewRequest(http.MethodPost, url, &compressedBody)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Обработка сжатого ответа, если сервер вернул gzip
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer gzipReader.Close()
			_, err = io.ReadAll(gzipReader) // Читаем тело ответа
			if err != nil {
				return err
			}
		}

		// Игнорируем тело, так как оно не используется
	}
	return nil
}
