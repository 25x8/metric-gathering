-- +goose Up

CREATE TABLE IF NOT EXISTS gauges (
                                      name TEXT PRIMARY KEY,
                                      value DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS counters (
                                        name TEXT PRIMARY KEY,
                                        value BIGINT NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS gauges;
DROP TABLE IF EXISTS counters;