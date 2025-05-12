package storage

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewDBStorage(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию миграций для тестов
	originalGooseUp := gooseUp
	defer func() { gooseUp = originalGooseUp }()

	// Подменяем функцию gooseUp на нашу тестовую
	gooseUp = func(db *sql.DB, dir string, opts ...goose.OptionsFunc) error {
		// Эмулируем создание таблиц миграциями
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS gauges")
		if err != nil {
			return err
		}
		_, err = db.Exec("CREATE TABLE IF NOT EXISTS counters")
		return err
	}

	// Переопределяем функцию повторного выполнения операций для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Ожидаем запрос ping для проверки соединения
	mock.ExpectPing()

	// Ожидаем запросы для создания таблиц (они будут вызваны из gooseUp)
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS gauges")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS counters")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Создаем хранилище
	storage, err := NewDBStorage(db)
	assert.NoError(t, err)
	assert.NotNil(t, storage)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_SaveGaugeMetric(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Ожидаем запрос на сохранение метрики
	mock.ExpectExec("INSERT INTO gauges").
		WithArgs("test_gauge", 123.456).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Сохраняем метрику
	err = storage.SaveGaugeMetric("test_gauge", 123.456)
	assert.NoError(t, err)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_SaveCounterMetric(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Ожидаем запрос на сохранение метрики
	mock.ExpectExec("INSERT INTO counters").
		WithArgs("test_counter", int64(32)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Сохраняем метрику
	err = storage.SaveCounterMetric("test_counter", 32)
	assert.NoError(t, err)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_GetGaugeMetric(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Подготавливаем заглушку для запроса метрики
	rows := sqlmock.NewRows([]string{"value"}).AddRow(123.456)
	mock.ExpectQuery("SELECT value FROM gauges").
		WithArgs("test_gauge").
		WillReturnRows(rows)

	// Получаем метрику
	value, err := storage.GetGaugeMetric("test_gauge")
	assert.NoError(t, err)
	assert.Equal(t, 123.456, value)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_GetCounterMetric(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Подготавливаем заглушку для запроса метрики
	rows := sqlmock.NewRows([]string{"value"}).AddRow(42)
	mock.ExpectQuery("SELECT value FROM counters").
		WithArgs("test_counter").
		WillReturnRows(rows)

	// Получаем метрику
	value, err := storage.GetCounterMetric("test_counter")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), value)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_GetAllMetrics(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Подготавливаем заглушки для gauge метрик
	gaugeRows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("gauge1", 123.456).
		AddRow("gauge2", 789.012)
	mock.ExpectQuery("SELECT name, value FROM gauges").
		WillReturnRows(gaugeRows)

	// Подготавливаем заглушки для counter метрик
	counterRows := sqlmock.NewRows([]string{"name", "value"}).
		AddRow("counter1", 42).
		AddRow("counter2", 84)
	mock.ExpectQuery("SELECT name, value FROM counters").
		WillReturnRows(counterRows)

	// Получаем все метрики
	metrics := storage.GetAllMetrics()

	// Проверяем, что все метрики получены
	assert.Equal(t, float64(123.456), metrics["gauge1"])
	assert.Equal(t, float64(789.012), metrics["gauge2"])
	assert.Equal(t, int64(42), metrics["counter1"])
	assert.Equal(t, int64(84), metrics["counter2"])

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestDBStorage_UpdateMetricsBatch(t *testing.T) {
	// Создаем фейковое подключение к базе данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Переопределяем функцию retryOperation для тестов
	originalRetryOperation := retryOperation
	defer func() { retryOperation = originalRetryOperation }()

	retryOperation = func(ctx context.Context, operation func() error) error {
		return operation()
	}

	// Создаем хранилище
	storage := &DBStorage{db: db}

	// Подготавливаем метрики для сохранения
	delta := int64(42)
	value := float64(123.456)
	metrics := []Metrics{
		{
			ID:    "counter1",
			MType: Counter,
			Delta: &delta,
		},
		{
			ID:    "gauge1",
			MType: Gauge,
			Value: &value,
		},
	}

	// Ожидаем транзакцию
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO counters").
		WithArgs("counter1", delta).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO gauges").
		WithArgs("gauge1", value).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Обновляем метрики пакетом
	err = storage.UpdateMetricsBatch(metrics)
	assert.NoError(t, err)

	// Убеждаемся, что все ожидания выполнены
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
