package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// CompressWriter добавляет поддержку gzip для ответа
type CompressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (cw *CompressWriter) Write(data []byte) (int, error) {
	return cw.Writer.Write(data)
}

// GzipMiddleware обрабатывает запросы с gzip и добавляет сжатие ответов
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
