package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/pressly/goose/v3"
	"log"
	"net"
	"time"
)

const (
	Gauge   = "gauge"
	Counter = "counter"
)

type DBStorage struct {
	db *sql.DB
}

func (s *DBStorage) DB() *sql.DB {
	return s.db
}

func NewDBStorage(db *sql.DB) (*DBStorage, error) {
	ctx := context.Background()

	// Используем retryOperation для проверки соединения
	err := retryOperation(ctx, func() error {
		return db.PingContext(ctx)
	})

	if err != nil {
		return nil, fmt.Errorf("database connection check failed: %w", err)
	}

	storage := &DBStorage{db: db}

	// Настраиваем goose
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %w", err)
	}
	goose.SetTableName("goose_db_version")

	// Применяем миграции с использованием retryOperation
	log.Println("Applying database migrations...")
	err = retryOperation(ctx, func() error {
		return goose.Up(db, "migrations")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	log.Println("Database migrations applied successfully.")

	return storage, nil
}

func (s *DBStorage) SaveGaugeMetric(name string, value float64) error {
	query := `INSERT INTO gauges (name, value) VALUES ($1, $2)
              ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;`

	ctx := context.Background()

	return retryOperation(ctx, func() error {
		_, err := s.db.Exec(query, name, value)
		return err
	})
}

func (s *DBStorage) SaveCounterMetric(name string, delta int64) error {
	query := `INSERT INTO counters (name, value) VALUES ($1, $2)
              ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value;`

	ctx := context.Background()

	return retryOperation(ctx, func() error {
		_, err := s.db.Exec(query, name, delta)
		return err
	})
}

func (s *DBStorage) GetGaugeMetric(name string) (float64, error) {
	var value float64
	query := `SELECT value FROM gauges WHERE name = $1`

	ctx := context.Background()

	err := retryOperation(ctx, func() error {
		return s.db.QueryRow(query, name).Scan(&value)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("metric not found")
	}
	return value, err
}

func (s *DBStorage) GetCounterMetric(name string) (int64, error) {
	var value int64
	query := `SELECT value FROM counters WHERE name = $1`

	ctx := context.Background()

	err := retryOperation(ctx, func() error {
		return s.db.QueryRow(query, name).Scan(&value)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("metric not found")
	}
	return value, err
}

func (s *DBStorage) GetAllMetrics() map[string]interface{} {
	ctx := context.Background()
	allMetrics := make(map[string]interface{})

	// Получаем все gauge метрики
	gaugesQuery := `SELECT name, value FROM gauges`
	var gaugeRows *sql.Rows

	err := retryOperation(ctx, func() error {
		var err error
		gaugeRows, err = s.db.QueryContext(ctx, gaugesQuery)
		if err := gaugeRows.Err(); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		log.Printf("Error fetching gauges: %v", err)
		return allMetrics
	}
	defer func() {
		if err := gaugeRows.Close(); err != nil {
			log.Printf("Error closing gauge rows: %v", err)
		}
	}()

	for gaugeRows.Next() {
		var name string
		var value float64
		if err := gaugeRows.Scan(&name, &value); err != nil {
			log.Printf("Error scanning gauge: %v", err)
			continue
		}
		allMetrics[name] = value
	}

	// Проверка ошибок после итерации
	if err := gaugeRows.Err(); err != nil {
		log.Printf("Error iterating gauge rows: %v", err)
	}

	// Получаем все counter метрики
	countersQuery := `SELECT name, value FROM counters`
	var counterRows *sql.Rows
	err = retryOperation(ctx, func() error {
		var err error
		counterRows, err = s.db.QueryContext(ctx, countersQuery)
		if err := counterRows.Err(); err != nil {
			return err
		}
		return err
	})
	if err != nil {
		log.Printf("Error fetching counters: %v", err)
		return allMetrics
	}
	defer func() {
		if err := counterRows.Close(); err != nil {
			log.Printf("Error closing counter rows: %v", err)
		}
	}()

	for counterRows.Next() {
		var name string
		var value int64
		if err := counterRows.Scan(&name, &value); err != nil {
			log.Printf("Error scanning counter: %v", err)
			continue
		}
		allMetrics[name] = value
	}

	// Проверка ошибок после итерации
	if err := counterRows.Err(); err != nil {
		log.Printf("Error iterating counter rows: %v", err)
	}

	return allMetrics
}

func (s *DBStorage) UpdateMetricsBatch(metrics []Metrics) error {
	ctx := context.Background()

	return retryOperation(ctx, func() error {
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		defer func() {
			if err != nil {
				_ = tx.Rollback()
			} else {
				err = tx.Commit()
			}
		}()

		for _, metric := range metrics {
			switch metric.MType {
			case Counter:
				if metric.Delta == nil {
					continue
				}
				query := `INSERT INTO counters (name, value) VALUES ($1, $2)
                          ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value;`
				_, err = tx.Exec(query, metric.ID, *metric.Delta)
				if err != nil {
					return err
				}
			case Gauge:
				if metric.Value == nil {
					continue
				}
				query := `INSERT INTO gauges (name, value) VALUES ($1, $2)
                          ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;`
				_, err = tx.Exec(query, metric.ID, *metric.Value)
				if err != nil {
					return err
				}
			default:
				continue
			}
		}
		return nil
	})
}

// retryOperation повторяет запрос при retriable ошибках
func retryOperation(ctx context.Context, operation func() error) error {
	maxRetries := 4
	var err error
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}
		if !isRetriableError(err) {
			return err
		}
		if i < len(delays) {
			select {
			case <-time.After(delays[i]):
				// Переход к следующей итерации
			case <-ctx.Done():
				// Если контекст отменён, возвращаем его ошибку
				return ctx.Err()
			}
		}
	}
	return err
}

// isRetriableError проверяет, является ли ошибка временной и стоит ли повторить попытку
func isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	// Проверка ошибок PostgreSQL
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.SerializationFailure,
			pgerrcode.DeadlockDetected,
			pgerrcode.UniqueViolation,
			pgerrcode.ConnectionException,
			pgerrcode.ConnectionDoesNotExist,
			pgerrcode.ConnectionFailure,
			pgerrcode.CrashShutdown,
			pgerrcode.CannotConnectNow,
			pgerrcode.IOError:
			return true
		default:
			return false
		}
	}

	// Проверка ошибок сети
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// Ошибка sql драйвера
	if errors.Is(err, sql.ErrConnDone) {
		return true
	}

	return false
}
