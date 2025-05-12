package storage

// Metrics представляет структуру данных для передачи метрик между сервисами.
// Используется как для входящих запросов, так и для ответов API.
// Поддерживает два типа метрик: gauge (плавающая точка) и counter (целочисленный счетчик).
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение для counter
	Value *float64 `json:"value,omitempty"` // значение для gauge
}
