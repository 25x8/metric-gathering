package senders

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Metric Структура метрики
type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

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

func (s *HTTPSender) SendBatch(metrics map[string]interface{}) error {
	if len(metrics) == 0 {
		return nil // Don't send empty batches
	}

	var metricsSlice []Metric

	for key, value := range metrics {
		var metric Metric
		metric.ID = key
		switch v := value.(type) {
		case int64:
			metric.MType = "counter"
			metric.Delta = &v // Send as delta
		case float64:
			metric.MType = "gauge"
			metric.Value = &v
		default:
			continue
		}
		metricsSlice = append(metricsSlice, metric)
	}

	if len(metricsSlice) == 0 {
		return nil // No metrics to send
	}

	// Serialize metrics to JSON
	jsonData, err := json.Marshal(metricsSlice)
	if err != nil {
		return err
	}

	// Compress data
	var compressedBody bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBody)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		return err
	}
	gzipWriter.Close()

	// Send request
	url := fmt.Sprintf("%s/updates/", s.ServerURL)
	req, err := http.NewRequest(http.MethodPost, url, &compressedBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %s", resp.Status)
	}

	return nil
}

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
