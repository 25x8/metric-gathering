package senders

import (
	"bytes"
	"fmt"
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
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(nil))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}
	return nil
}
