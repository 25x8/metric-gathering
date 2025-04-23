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

// BenchmarkGzipMiddleware_Compressed тестирует производительность middleware с сжатым запросом
func BenchmarkGzipMiddleware_Compressed(b *testing.B) {
	// Создаем сжатое тело запроса
	var compressed bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressed)
	_, err := gzipWriter.Write([]byte("This is test data for compression middleware benchmark"))
	if err != nil {
		b.Fatal(err)
	}
	err = gzipWriter.Close()
	if err != nil {
		b.Fatal(err)
	}

	// Создаем обработчик, который просто считывает все данные из тела запроса
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Оборачиваем обработчик в middleware
	wrapped := GzipMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Создаем запрос с сжатыми данными
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(compressed.Bytes()))
		req.Header.Set("Content-Encoding", "gzip")

		// Создаем ResponseRecorder для записи ответа
		w := httptest.NewRecorder()

		// Обрабатываем запрос
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkGzipMiddleware_Plain тестирует производительность middleware с обычным запросом
func BenchmarkGzipMiddleware_Plain(b *testing.B) {
	// Данные для запроса
	requestBody := strings.Repeat("This is test data for plain request middleware benchmark", 10)

	// Создаем обработчик, который просто считывает все данные из тела запроса
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

		// Записываем большой ответ, чтобы проверить сжатие
		w.Write([]byte(strings.Repeat("Response data that should be compressed", 50)))
	})

	// Оборачиваем обработчик в middleware
	wrapped := GzipMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Создаем запрос с обычными данными
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(requestBody))
		// Указываем, что клиент принимает сжатые данные
		req.Header.Set("Accept-Encoding", "gzip")

		// Создаем ResponseRecorder для записи ответа
		w := httptest.NewRecorder()

		// Обрабатываем запрос
		wrapped.ServeHTTP(w, req)
	}
}

// BenchmarkGzipMiddleware_LargeResponse тестирует производительность middleware с большим ответом
func BenchmarkGzipMiddleware_LargeResponse(b *testing.B) {
	// Создаем обработчик, который возвращает большой ответ
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Записываем очень большой ответ для проверки эффективности сжатия
		largeResponse := strings.Repeat("This is a large response that should benefit from compression. ", 1000)
		w.Write([]byte(largeResponse))
	})

	// Оборачиваем обработчик в middleware
	wrapped := GzipMiddleware(handler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Создаем простой запрос
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Указываем, что клиент принимает сжатые данные
		req.Header.Set("Accept-Encoding", "gzip")

		// Создаем ResponseRecorder для записи ответа
		w := httptest.NewRecorder()

		// Обрабатываем запрос
		wrapped.ServeHTTP(w, req)
	}
}
