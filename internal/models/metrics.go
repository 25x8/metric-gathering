package models

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение для counter
	Value *float64 `json:"value,omitempty"` // значение для gauge
}

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type Metric struct {
	Type  string      `json:"type"`  // Тип метрики (gauge или counter)
	Name  string      `json:"name"`  // Имя метрики
	Value interface{} `json:"value"` // Значение метрики (float64 для gauge, int64 для counter)
}
