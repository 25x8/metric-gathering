package senders

import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/25x8/metric-gathering/internal/utils"
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

func (s *HTTPSender) SendBatch(metrics map[string]interface{}, publicKey *rsa.PublicKey) error {
	if len(metrics) == 0 {
		return nil
	}

	var metricsSlice []Metric

	for key, value := range metrics {
		var metric Metric
		metric.ID = key
		switch v := value.(type) {
		case int64:
			metric.MType = "counter"
			metric.Delta = &v 
		case float64:
			metric.MType = "gauge"
			metric.Value = &v
		default:
			continue
		}
		metricsSlice = append(metricsSlice, metric)
	}

	if len(metricsSlice) == 0 {
		return nil
	}

	jsonData, err := json.Marshal(metricsSlice)
	if err != nil {
		return err
	}

	var compressedBody bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBody)
	_, err = gzipWriter.Write(jsonData)
	if err != nil {
		return err
	}
	gzipWriter.Close()

	compressedData := compressedBody.Bytes()

	if publicKey != nil && len(compressedData) > 100 {
		return fmt.Errorf("data too large for RSA encryption, use individual sends")
	}

	var requestBody []byte
	var isEncrypted bool

	if publicKey != nil {
		encryptedData, err := utils.EncryptWithPublicKey(compressedData, publicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		requestBody = encryptedData
		isEncrypted = true
	} else {
		requestBody = compressedData
	}

	url := fmt.Sprintf("%s/updates/", s.ServerURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
	if err != nil {
		return err
	}

	if isEncrypted {
		req.Header.Set("Content-Encrypted", "true")
	} else {
		req.Header.Set("Content-Encoding", "gzip")
	}

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

func (s *HTTPSender) Send(metrics map[string]interface{}, key string, publicKey *rsa.PublicKey) error {
	for keyName, value := range metrics {
		var metricType string
		switch value.(type) {
		case int64:
			metricType = "counter"
		case float64:
			metricType = "gauge"
		default:
			continue
		}

		url := fmt.Sprintf("%s/update/%s/%s/%v", s.ServerURL, metricType, keyName, value)

		var requestBody []byte
		var isEncrypted bool

		var compressedBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressedBody)
		_, err := gzipWriter.Write([]byte{})
		if err != nil {
			return err
		}
		gzipWriter.Close()

		compressedData := compressedBody.Bytes()

		if publicKey != nil {
			encryptedData, err := utils.EncryptWithPublicKey(compressedData, publicKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt data: %w", err)
			}
			requestBody = encryptedData
			isEncrypted = true
		} else {
			requestBody = compressedData
		}

		var hash string
		if key != "" {
			hash = utils.CalculateHash(requestBody, key)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBody))
		if err != nil {
			return err
		}

		if isEncrypted {
			req.Header.Set("Content-Encrypted", "true")
		} else {
			req.Header.Set("Content-Encoding", "gzip")
		}

		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "text/plain")

		if hash != "" {
			req.Header.Set("HashSHA256", hash)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.Header.Get("Content-Encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer gzipReader.Close()
			_, err = io.ReadAll(gzipReader)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
