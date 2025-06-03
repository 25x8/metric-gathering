package server

import (
	"context"
	"fmt"
	"net"

	pb "github.com/25x8/metric-gathering/internal/grpc/pb"
	"github.com/25x8/metric-gathering/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsServer реализует gRPC сервис для метрик
type MetricsServer struct {
	pb.UnimplementedMetricsServiceServer
	storage storage.Storage
}

// NewMetricsServer создает новый gRPC сервер для метрик
func NewMetricsServer(storage storage.Storage) *MetricsServer {
	return &MetricsServer{
		storage: storage,
	}
}

// UpdateMetric обновляет одну метрику
func (s *MetricsServer) UpdateMetric(ctx context.Context, req *pb.UpdateMetricRequest) (*pb.UpdateMetricResponse, error) {
	if req.Metric == nil {
		return nil, status.Error(codes.InvalidArgument, "metric is required")
	}

	metric := req.Metric
	if metric.Id == "" || metric.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "metric id and type are required")
	}

	switch metric.Type {
	case "gauge":
		if metric.Value == nil {
			return nil, status.Error(codes.InvalidArgument, "value is required for gauge metric")
		}
		err := s.storage.SaveGaugeMetric(metric.Id, *metric.Value)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save gauge metric: %v", err))
		}

		// Получаем обновленное значение
		updatedValue, err := s.storage.GetGaugeMetric(metric.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get updated gauge metric: %v", err))
		}
		metric.Value = &updatedValue

	case "counter":
		if metric.Delta == nil {
			return nil, status.Error(codes.InvalidArgument, "delta is required for counter metric")
		}
		err := s.storage.SaveCounterMetric(metric.Id, *metric.Delta)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save counter metric: %v", err))
		}

		// Получаем обновленное значение
		updatedDelta, err := s.storage.GetCounterMetric(metric.Id)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get updated counter metric: %v", err))
		}
		metric.Delta = &updatedDelta

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}

	return &pb.UpdateMetricResponse{
		Metric: metric,
	}, nil
}

// UpdateMetrics пакетное обновление метрик
func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if len(req.Metrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "metrics list cannot be empty")
	}

	// Конвертируем gRPC метрики в storage метрики
	storageMetrics := make([]storage.Metrics, 0, len(req.Metrics))
	for _, metric := range req.Metrics {
		if metric.Id == "" || metric.Type == "" {
			return nil, status.Error(codes.InvalidArgument, "all metrics must have id and type")
		}

		storageMetric := storage.Metrics{
			ID:    metric.Id,
			MType: metric.Type,
		}

		switch metric.Type {
		case "gauge":
			if metric.Value == nil {
				return nil, status.Error(codes.InvalidArgument, "value is required for gauge metric")
			}
			storageMetric.Value = metric.Value
		case "counter":
			if metric.Delta == nil {
				return nil, status.Error(codes.InvalidArgument, "delta is required for counter metric")
			}
			storageMetric.Delta = metric.Delta
		default:
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid metric type: %s", metric.Type))
		}

		storageMetrics = append(storageMetrics, storageMetric)
	}

	// Выполняем пакетное обновление
	err := s.storage.UpdateMetricsBatch(storageMetrics)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update metrics batch: %v", err))
	}

	return &pb.UpdateMetricsResponse{
		Success: true,
	}, nil
}

// GetMetric получает значение метрики
func (s *MetricsServer) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	if req.Id == "" || req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "metric id and type are required")
	}

	metric := &pb.Metric{
		Id:   req.Id,
		Type: req.Type,
	}

	switch req.Type {
	case "gauge":
		value, err := s.storage.GetGaugeMetric(req.Id)
		if err != nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("gauge metric not found: %v", err))
		}
		metric.Value = &value

	case "counter":
		delta, err := s.storage.GetCounterMetric(req.Id)
		if err != nil {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("counter metric not found: %v", err))
		}
		metric.Delta = &delta

	default:
		return nil, status.Error(codes.InvalidArgument, "invalid metric type")
	}

	return &pb.GetMetricResponse{
		Metric: metric,
	}, nil
}

// GetAllMetrics получает все метрики
func (s *MetricsServer) GetAllMetrics(ctx context.Context, req *pb.GetAllMetricsRequest) (*pb.GetAllMetricsResponse, error) {
	allMetrics := s.storage.GetAllMetrics()

	metrics := make([]*pb.Metric, 0, len(allMetrics))
	for name, value := range allMetrics {
		metric := &pb.Metric{
			Id: name,
		}

		switch v := value.(type) {
		case float64:
			metric.Type = "gauge"
			metric.Value = &v
		case int64:
			metric.Type = "counter"
			metric.Delta = &v
		default:
			continue // Пропускаем неизвестные типы
		}

		metrics = append(metrics, metric)
	}

	return &pb.GetAllMetricsResponse{
		Metrics: metrics,
	}, nil
}

// Ping проверяет доступность сервера
func (s *MetricsServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	// В реальном приложении здесь можно добавить проверку соединения с БД
	return &pb.PingResponse{
		Healthy: true,
	}, nil
}

// GRPCServer обертка для управления gRPC сервером
type GRPCServer struct {
	server   *grpc.Server
	listener net.Listener
}

// NewGRPCServer создает новый gRPC сервер
func NewGRPCServer(addr string, storage storage.Storage) (*GRPCServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	grpcServer := grpc.NewServer()
	metricsServer := NewMetricsServer(storage)

	pb.RegisterMetricsServiceServer(grpcServer, metricsServer)

	return &GRPCServer{
		server:   grpcServer,
		listener: listener,
	}, nil
}

// Start запускает gRPC сервер
func (s *GRPCServer) Start() error {
	return s.server.Serve(s.listener)
}

// Stop останавливает gRPC сервер
func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}

// Addr возвращает адрес, на котором слушает сервер
func (s *GRPCServer) Addr() net.Addr {
	return s.listener.Addr()
}
