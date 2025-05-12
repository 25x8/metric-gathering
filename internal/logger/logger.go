package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

	// Создаем упрощенную конфигурацию с меньшим потреблением памяти
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      zapcore.OmitKey, // Отключаем логирование вызывающего кода
		FunctionKey:    zapcore.OmitKey, // Отключаем логирование имени функции
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey, // Отключаем стектрейсы
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Создаем базовую конфигурацию без сэмплирования
	cfg := zap.Config{
		Level:            lvl,
		Development:      false,
		Sampling:         nil, // Отключаем сэмплирование для экономии памяти
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// создаём логер на основе конфигурации
	zl, err := cfg.Build(
		zap.WithCaller(false), // Отключаем информацию о вызывающем коде
	)
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
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Обертка для записи ответа
		ww := &responseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(ww, r)

		// Логируем только важную информацию
		Log.Info("Request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", ww.statusCode),
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
