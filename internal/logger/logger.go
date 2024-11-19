package logger

import (
	"github.com/25x8/metric-gathering/internal/response"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции Initialize, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

// Sync закрывает логгер и завершает все фоновые процессы
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// RequestLogger — middleware-логер для входящих HTTP-запросов.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Проверяем, реализует ли ResponseWriter интерфейс ResponseInfo
		var ri response.ResponseInfo
		if existingRI, ok := w.(response.ResponseInfo); ok {
			// Если уже реализует, используем его
			ri = existingRI
		} else {
			// Иначе создаем свою обертку
			ww := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}
			w = ww
			ri = ww
		}

		next.ServeHTTP(w, r)

		// Используем интерфейс ResponseInfo для получения информации
		Log.Info("Request processed",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", ri.Status()),
			zap.Int("response_size", ri.Size()),
		)
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.responseSize += size
	return size, err
}

// Реализуем интерфейс response.ResponseInfo
func (w *responseWriterWrapper) Status() int {
	return w.statusCode
}

func (w *responseWriterWrapper) Size() int {
	return w.responseSize
}
