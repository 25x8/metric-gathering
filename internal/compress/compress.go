package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// bufferedResponseWriter is a custom http.ResponseWriter that buffers the response body
type bufferedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		ResponseWriter: w,
		headers:        make(http.Header),
		body:           &bytes.Buffer{},
	}
}

func (w *bufferedResponseWriter) Header() http.Header {
	return w.headers
}

func (w *bufferedResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *bufferedResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

// GzipCompressMiddleware compresses the response if the client supports gzip encoding
func GzipCompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Client does not accept gzip
			next.ServeHTTP(w, r)
			return
		}

		// Buffer the response
		brw := newBufferedResponseWriter(w)
		next.ServeHTTP(brw, r)

		// Check Content-Type
		contentType := brw.headers.Get("Content-Type")
		if contentType == "application/json" || contentType == "text/html" {
			// Compress the response
			w.Header().Set("Content-Encoding", "gzip")
			for k, v := range brw.headers {
				w.Header()[k] = v
			}
			if brw.statusCode != 0 {
				w.WriteHeader(brw.statusCode)
			}
			gz := gzip.NewWriter(w)
			defer gz.Close()
			gz.Write(brw.body.Bytes())
		} else {
			// Do not compress
			for k, v := range brw.headers {
				w.Header()[k] = v
			}
			if brw.statusCode != 0 {
				w.WriteHeader(brw.statusCode)
			}
			w.Write(brw.body.Bytes())
		}
	})
}

// GzipDecompressMiddleware decompresses the request body if it's gzipped
func GzipDecompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decompress request body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = io.NopCloser(gr)
		}
		next.ServeHTTP(w, r)
	})
}
