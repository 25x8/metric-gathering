package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGzipMiddleware_CompressedRequest(t *testing.T) {
	// Создаем тестовые данные
	testData := "This is test data that should be decompressed by middleware"

	// Сжимаем данные
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, err := gzipWriter.Write([]byte(testData))
	if err != nil {
		t.Fatal(err)
	}
	err = gzipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Создаем обработчик, который проверяет, что данные правильно распакованы
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}

		if string(data) != testData {
			t.Errorf("Expected body %q, got %q", testData, string(data))
		}

		w.WriteHeader(http.StatusOK)
	})

	// Оборачиваем обработчик в middleware
	wrappedHandler := GzipMiddleware(handler)

	// Создаем запрос с сжатыми данными
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressed.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Обрабатываем запрос
	wrappedHandler.ServeHTTP(w, req)

	// Проверяем код ответа
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGzipMiddleware_CompressedResponse(t *testing.T) {
	// Большие данные, которые должны быть сжаты
	largeResponse := strings.Repeat("This is a large response that should be compressed. ", 100)

	// Создаем обработчик, который отправляет большой ответ
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(largeResponse))
	})

	// Оборачиваем обработчик в middleware
	wrappedHandler := GzipMiddleware(handler)

	// Создаем запрос, указывающий, что клиент принимает сжатые данные
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Обрабатываем запрос
	wrappedHandler.ServeHTTP(w, req)

	// Проверяем, что заголовок указывает на сжатие
	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected Content-Encoding: gzip header not found")
	}

	// Декодируем ответ и проверяем содержимое
	reader, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	if string(decompressed) != largeResponse {
		t.Error("Decompressed response does not match expected content")
	}
}
