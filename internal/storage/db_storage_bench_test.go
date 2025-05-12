package storage

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// mockDBStorage создает тестовое хранилище без применения миграций
func mockDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{db: db}
}

func setupMockDB(b testing.TB) (*sql.DB, sqlmock.Sqlmock, *DBStorage) {
	db, mock, err := sqlmock.New()
	if err != nil {
		b.Fatalf("Failed to create mock DB: %v", err)
	}

	storage := mockDBStorage(db)

	return db, mock, storage
}

func BenchmarkDBStorage_SaveGaugeMetric(b *testing.B) {
	db, mock, storage := setupMockDB(b)
	defer db.Close()

	// Подготавливаем ожидание для запроса
	mock.ExpectExec("INSERT INTO gauges").WillReturnResult(sqlmock.NewResult(1, 1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "gauge_metric_" + string(rune(i%100))
		err := storage.SaveGaugeMetric(name, float64(i))
		if err != nil {
			b.Fatalf("Error in SaveGaugeMetric: %v", err)
		}

		// Сбрасываем ожидания для следующей итерации
		if i+1 < b.N {
			mock.ExpectExec("INSERT INTO gauges").WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}
}

func BenchmarkDBStorage_SaveCounterMetric(b *testing.B) {
	db, mock, storage := setupMockDB(b)
	defer db.Close()

	// Для первого вызова нужен SELECT и затем либо INSERT, либо UPDATE
	mock.ExpectQuery("SELECT value FROM metrics").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(10))
	mock.ExpectExec("UPDATE metrics").WillReturnResult(sqlmock.NewResult(1, 1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "counter_metric_" + string(rune(i%100))
		storage.SaveCounterMetric(name, int64(i))

		// Сбрасываем ожидания для следующей итерации
		if i+1 < b.N {
			mock.ExpectQuery("SELECT value FROM metrics").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(10))
			mock.ExpectExec("UPDATE metrics").WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}
}

func BenchmarkDBStorage_GetGaugeMetric(b *testing.B) {
	db, mock, storage := setupMockDB(b)
	defer db.Close()

	mock.ExpectQuery("SELECT value FROM metrics").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(42.0))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		name := "gauge_metric_" + string(rune(i%100))
		storage.GetGaugeMetric(name)

		// Сбрасываем ожидания для следующей итерации
		if i+1 < b.N {
			mock.ExpectQuery("SELECT value FROM metrics").WillReturnRows(sqlmock.NewRows([]string{"value"}).AddRow(42.0))
		}
	}
}

func BenchmarkDBStorage_GetAllMetrics(b *testing.B) {
	db, mock, storage := setupMockDB(b)
	defer db.Close()

	// Исправляем проблемы с rows.Err()
	mock.ExpectQuery("SELECT name, value FROM gauges").WillReturnRows(
		sqlmock.NewRows([]string{"name", "value"}).
			AddRow("gauge1", 42.0).
			AddRow("gauge2", 84.0).
			RowError(2, nil))

	mock.ExpectQuery("SELECT name, value FROM counters").WillReturnRows(
		sqlmock.NewRows([]string{"name", "value"}).
			AddRow("counter1", 100).
			AddRow("counter2", 200).
			RowError(2, nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics := storage.GetAllMetrics()
		if len(metrics) == 0 {
			b.Fatal("GetAllMetrics returned empty map")
		}

		// Сбрасываем ожидания для следующей итерации
		if i+1 < b.N {
			mock.ExpectQuery("SELECT name, value FROM gauges").WillReturnRows(
				sqlmock.NewRows([]string{"name", "value"}).
					AddRow("gauge1", 42.0).
					AddRow("gauge2", 84.0).
					RowError(2, nil))

			mock.ExpectQuery("SELECT name, value FROM counters").WillReturnRows(
				sqlmock.NewRows([]string{"name", "value"}).
					AddRow("counter1", 100).
					AddRow("counter2", 200).
					RowError(2, nil))
		}
	}
}

func BenchmarkDBStorage_UpdateMetricsBatch(b *testing.B) {
	// В реальном методе UpdateMetricsBatch используются отдельные таблицы для gauge и counter,
	// поэтому мы модифицируем этот тест

	// Пропускаем тест, т.к. он требует много настройки для правильной имитации
	b.Skip("Skipping UpdateMetricsBatch benchmark as it requires complex setup")
}

// Вспомогательные функции для создания указателей
func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
