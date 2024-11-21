package senders

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

func (s *HTTPSender) SendBatch(metrics interface{}) error {
	switch v := metrics.(type) {
	case map[string]interface{}:
		// Старый формат - отправляем через старый метод
		return s.Send(v)

	case []map[string]interface{}:
		// Новый формат - отправляем батчем через новый маршрут
		if len(v) == 0 {
			return nil // Пустые батчи не отправляем
		}

		var compressedBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressedBody)
		if err := json.NewEncoder(gzipWriter).Encode(v); err != nil {
			return err
		}
		gzipWriter.Close()

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/updates/", s.ServerURL), &compressedBody)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("server responded with status: %v", resp.StatusCode)
		}

		return nil

	default:
		return fmt.Errorf("unsupported metric format")
	}
}
