package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func NewDBStorage(dsn string) (*DBStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	storage := &DBStorage{db: db}

	if err := storage.initTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *DBStorage) initTables() error {
	query := `
        CREATE TABLE IF NOT EXISTS gauges (
            name TEXT PRIMARY KEY,
            value DOUBLE PRECISION NOT NULL
        );
        CREATE TABLE IF NOT EXISTS counters (
            name TEXT PRIMARY KEY,
            value BIGINT NOT NULL
        );
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *DBStorage) SaveGaugeMetric(name string, value float64) error {
	query := `INSERT INTO gauges (name, value) VALUES ($1, $2)
              ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value;`
	_, err := s.db.Exec(query, name, value)
	return err
}

func (s *DBStorage) SaveCounterMetric(name string, delta int64) error {
	query := `INSERT INTO counters (name, value) VALUES ($1, $2)
              ON CONFLICT (name) DO UPDATE SET value = counters.value + EXCLUDED.value;`
	_, err := s.db.Exec(query, name, delta)
	return err
}

func (s *DBStorage) GetGaugeMetric(name string) (float64, error) {
	var value float64
	query := `SELECT value FROM gauges WHERE name = $1`
	err := s.db.QueryRow(query, name).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("metric not found")
	}
	return value, err
}

func (s *DBStorage) GetCounterMetric(name string) (int64, error) {
	var value int64
	query := `SELECT value FROM counters WHERE name = $1`
	err := s.db.QueryRow(query, name).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("metric not found")
	}
	return value, err
}

func (s *DBStorage) GetAllMetrics() map[string]interface{} {
	ctx := context.Background()
	allMetrics := make(map[string]interface{})

	// Получаем все gauge метрики
	gaugesQuery := `SELECT name, value FROM gauges`
	rows, err := s.db.QueryContext(ctx, gaugesQuery)
	if err != nil {
		log.Printf("Error fetching gauges: %v", err)
		return allMetrics
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			log.Printf("Error scanning gauge: %v", err)
			continue
		}
		allMetrics[name] = value
	}

	// Проверяем ошибки после итерации
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating gauge rows: %v", err)
	}

	// Получаем все counter метрики
	countersQuery := `SELECT name, value FROM counters`
	rows, err = s.db.QueryContext(ctx, countersQuery)
	if err != nil {
		log.Printf("Error fetching counters: %v", err)
		return allMetrics
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			log.Printf("Error scanning counter: %v", err)
			continue
		}
		allMetrics[name] = value
	}

	// Проверяем ошибки после итерации
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating counter rows: %v", err)
	}

	return allMetrics
}

func (s *DBStorage) Flush() error {
	// PostgreSQL уже сохраняет данные
	return nil
}

func (s *DBStorage) Load() error {
	return nil
}

func (s *DBStorage) UpdateMetricsBatch(metrics []Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
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
}
