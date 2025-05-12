package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressWriter реализует интерфейс http.ResponseWriter с поддержкой gzip-сжатия.
// Оборачивает стандартный ResponseWriter и перенаправляет вывод через gzip-сжатие.
type CompressWriter struct {
	http.ResponseWriter           // встроенный ResponseWriter
	Writer              io.Writer // Writer с поддержкой gzip
}

// Write реализует интерфейс io.Writer для CompressWriter.
// Записывает данные через gzip-сжатие.
func (cw *CompressWriter) Write(data []byte) (int, error) {
	return cw.Writer.Write(data)
}

// GzipMiddleware обеспечивает обработку gzip-сжатых запросов и добавляет сжатие для ответов.
// Автоматически распаковывает тело запроса, если оно сжато с помощью gzip.
// Если клиент поддерживает gzip (указано в заголовке Accept-Encoding),
// ответ также будет сжат.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Разжимает тело запроса, если используется gzip
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		// Проверяет, поддерживает ли клиент gzip
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gw := gzip.NewWriter(w)
			defer gw.Close()
			w = &CompressWriter{ResponseWriter: w, Writer: gw}
		}

		next.ServeHTTP(w, r)
	})
}
