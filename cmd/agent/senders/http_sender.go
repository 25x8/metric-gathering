package senders

import (
	"bytes"
	"fmt"
	"net/http"
)

// HttpSender - структура для отправки метрик на сервер
type HttpSender struct {
	ServerURL string
}

// NewHttpSender - конструктор для HttpSender
func NewHttpSender(serverURL string) *HttpSender {
	return &HttpSender{
		ServerURL: serverURL,
	}
}

// Send - метод для отправки метрик на сервер
func (s *HttpSender) Send(metrics map[string]interface{}) error {
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
