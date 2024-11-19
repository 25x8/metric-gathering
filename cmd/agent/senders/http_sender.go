package senders

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
)

type HTTPSender struct {
	ServerURL string
}

func NewHTTPSender(serverURL string) *HTTPSender {
	return &HTTPSender{
		ServerURL: serverURL,
	}
}

func (s *HTTPSender) Send(metrics map[string]interface{}) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	// Compress the data using gzip
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}

	// Create a new HTTP request with the compressed data
	req, err := http.NewRequest("POST", s.ServerURL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle the response if necessary
	return nil
}
