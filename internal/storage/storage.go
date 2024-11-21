package storage

type Storage interface {
	SaveGaugeMetric(name string, value float64) error
	SaveCounterMetric(name string, delta int64) error
	GetGaugeMetric(name string) (float64, error)
	GetCounterMetric(name string) (int64, error)
	GetAllMetrics() map[string]interface{} // Возвращает все метрики
	Flush() error
	Load() error
}
