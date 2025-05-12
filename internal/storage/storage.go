package storage

// Storage определяет интерфейс для хранения и управления метриками.
// Реализации этого интерфейса обеспечивают сохранение метрик в памяти или базе данных.
type Storage interface {
	// SaveGaugeMetric сохраняет метрику типа gauge с указанным именем и значением.
	// Возвращает ошибку, если операция завершилась неудачно.
	SaveGaugeMetric(name string, value float64) error

	// SaveCounterMetric увеличивает значение метрики типа counter на указанную дельту.
	// Если метрика с таким именем не существует, создается новая со значением delta.
	// Возвращает ошибку, если операция завершилась неудачно.
	SaveCounterMetric(name string, delta int64) error

	// GetGaugeMetric возвращает значение метрики типа gauge с указанным именем.
	// Если метрика не найдена, возвращается ошибка.
	GetGaugeMetric(name string) (float64, error)

	// GetCounterMetric возвращает значение метрики типа counter с указанным именем.
	// Если метрика не найдена, возвращается ошибка.
	GetCounterMetric(name string) (int64, error)

	// GetAllMetrics возвращает все сохраненные метрики в виде карты,
	// где ключ - имя метрики, а значение - ее текущее значение.
	GetAllMetrics() map[string]interface{}

	// UpdateMetricsBatch обновляет несколько метрик одновременно.
	// Принимает массив структур Metrics и обновляет соответствующие метрики в хранилище.
	// Возвращает ошибку, если операция завершилась неудачно.
	UpdateMetricsBatch(metrics []Metrics) error
}
