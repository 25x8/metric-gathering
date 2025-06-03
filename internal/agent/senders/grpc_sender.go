package senders

import (
	"github.com/25x8/metric-gathering/internal/grpc/client"
)

// GRPCSender - структура для отправки метрик через gRPC
type GRPCSender struct {
	client *client.MetricsClient
}

// NewGRPCSender - конструктор для GRPCSender
func NewGRPCSender(serverAddr string) (*GRPCSender, error) {
	grpcClient, err := client.NewMetricsClient(serverAddr)
	if err != nil {
		return nil, err
	}

	return &GRPCSender{
		client: grpcClient,
	}, nil
}

// SendBatch отправляет метрики пакетно через gRPC
func (s *GRPCSender) SendBatch(metrics map[string]interface{}) error {
	return s.client.SendMetrics(metrics)
}

// Send отправляет метрики по одной через gRPC (совместимость с интерфейсом HTTPSender)
func (s *GRPCSender) Send(metrics map[string]interface{}, key string) error {
	// В gRPC ключи для подписи не нужны, так как gRPC сам обеспечивает целостность
	return s.client.SendMetrics(metrics)
}

// Close закрывает gRPC соединение
func (s *GRPCSender) Close() error {
	return s.client.Close()
}
