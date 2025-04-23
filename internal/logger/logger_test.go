package logger

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestInitialize(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectError bool
	}{
		{"Info level", "info", false},
		{"Debug level", "debug", false},
		{"Error level", "error", false},
		{"Invalid level", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Initialize(tt.level)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, Log)
			}
		})
	}
}

func TestRequestLogger(t *testing.T) {
	// Создаем буфер для перехвата вывода логгера
	var buf bytes.Buffer
	testEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		TimeKey:    "ts",
	})
	testCore := zapcore.NewCore(testEncoder, zapcore.AddSync(&buf), zapcore.InfoLevel)
	Log = zap.New(testCore)

	// Создаем тестовый HTTP сервер
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	})

	// Оборачиваем наш обработчик в RequestLogger
	loggedHandler := RequestLogger(handler)

	// Создаем тестовый HTTP запрос и ответ
	req := httptest.NewRequest("GET", "/test", nil)
	resp := httptest.NewRecorder()

	// Вызываем обработчик
	loggedHandler.ServeHTTP(resp, req)

	// Проверяем результат
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Hello, world!")
	assert.Contains(t, buf.String(), "Request")
	assert.Contains(t, buf.String(), "/test")
	assert.Contains(t, buf.String(), "GET")
}

func TestResponseWriterWrapper(t *testing.T) {
	// Создаем тестовый ResponseWriter
	recorder := httptest.NewRecorder()
	wrapper := &responseWriterWrapper{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
	}

	// Устанавливаем заголовок ответа
	wrapper.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, wrapper.statusCode)

	// Записываем данные в ответ
	data := []byte("Test data")
	n, err := wrapper.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), wrapper.responseSize)
	assert.Equal(t, "Test data", recorder.Body.String())
}
