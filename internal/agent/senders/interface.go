package senders

import "crypto/rsa"

// MetricsSender общий интерфейс для отправки метрик
type MetricsSender interface {
	SendBatch(metrics map[string]interface{}) error
	Send(metrics map[string]interface{}, key string) error
}

// HTTPMetricsSender интерфейс для HTTP отправки с поддержкой шифрования
type HTTPMetricsSender interface {
	SendBatch(metrics map[string]interface{}, publicKey *rsa.PublicKey) error
	Send(metrics map[string]interface{}, key string, publicKey *rsa.PublicKey) error
}
