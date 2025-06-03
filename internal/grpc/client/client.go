package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/25x8/metric-gathering/internal/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// MetricsClient gRPC клиент для отправки метрик
type MetricsClient struct {
	conn   *grpc.ClientConn
	client pb.MetricsServiceClient
}

// NewMetricsClient создает новый gRPC клиент
func NewMetricsClient(serverAddr string) (*MetricsClient, error) {
	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewMetricsServiceClient(conn)

	return &MetricsClient{
		conn:   conn,
		client: client,
	}, nil
}

// SendMetrics отправляет метрики пакетно через gRPC
func (c *MetricsClient) SendMetrics(metrics map[string]interface{}) error {
	if len(metrics) == 0 {
		return nil
	}

	// Конвертируем map в protobuf метрики
	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for name, value := range metrics {
		metric := &pb.Metric{
			Id: name,
		}

		switch v := value.(type) {
		case int64:
			metric.Type = "counter"
			metric.Delta = &v
		case float64:
			metric.Type = "gauge"
			metric.Value = &v
		default:
			continue // Пропускаем неизвестные типы
		}

		pbMetrics = append(pbMetrics, metric)
	}

	if len(pbMetrics) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := c.client.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	})

	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	return nil
}

// SendMetric отправляет одну метрику через gRPC
func (c *MetricsClient) SendMetric(name string, value interface{}) error {
	metric := &pb.Metric{
		Id: name,
	}

	switch v := value.(type) {
	case int64:
		metric.Type = "counter"
		metric.Delta = &v
	case float64:
		metric.Type = "gauge"
		metric.Value = &v
	default:
		return fmt.Errorf("unsupported metric type: %T", value)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.UpdateMetric(ctx, &pb.UpdateMetricRequest{
		Metric: metric,
	})

	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}

	return nil
}

// GetMetric получает значение метрики через gRPC
func (c *MetricsClient) GetMetric(name, metricType string) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetMetric(ctx, &pb.GetMetricRequest{
		Id:   name,
		Type: metricType,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get metric: %w", err)
	}

	switch resp.Metric.Type {
	case "gauge":
		if resp.Metric.Value != nil {
			return *resp.Metric.Value, nil
		}
	case "counter":
		if resp.Metric.Delta != nil {
			return *resp.Metric.Delta, nil
		}
	}

	return nil, fmt.Errorf("invalid metric response")
}

// GetAllMetrics получает все метрики через gRPC
func (c *MetricsClient) GetAllMetrics() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.GetAllMetrics(ctx, &pb.GetAllMetricsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get all metrics: %w", err)
	}

	metrics := make(map[string]interface{})
	for _, metric := range resp.Metrics {
		switch metric.Type {
		case "gauge":
			if metric.Value != nil {
				metrics[metric.Id] = *metric.Value
			}
		case "counter":
			if metric.Delta != nil {
				metrics[metric.Id] = *metric.Delta
			}
		}
	}

	return metrics, nil
}

// Ping проверяет доступность сервера через gRPC
func (c *MetricsClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}

	if !resp.Healthy {
		return fmt.Errorf("server is not healthy")
	}

	return nil
}

// Close закрывает соединение
func (c *MetricsClient) Close() error {
	return c.conn.Close()
}
