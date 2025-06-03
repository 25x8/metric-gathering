package server

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/25x8/metric-gathering/internal/grpc/pb"
	"github.com/25x8/metric-gathering/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func bufDialer(ctx context.Context, _ string) (net.Conn, error) {
	return lis.Dial()
}

func setupTestServer() (*grpc.Server, pb.MetricsServiceClient) {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()

	memStorage := storage.NewMemStorage("/tmp/test-metrics.json")
	pb.RegisterMetricsServiceServer(s, NewMetricsServer(memStorage))

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// ↓ главное изменение — «passthrough://»
	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}

	return s, pb.NewMetricsServiceClient(conn)
}


func TestUpdateMetric_Gauge(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 42.5
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_gauge",
			Type:  "gauge",
			Value: &value,
		},
	}

	resp, err := client.UpdateMetric(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "test_gauge", resp.Metric.Id)
	assert.Equal(t, "gauge", resp.Metric.Type)
	assert.Equal(t, value, *resp.Metric.Value)
}

func TestUpdateMetric_Counter(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	delta := int64(10)
	req := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "test_counter",
			Type:  "counter",
			Delta: &delta,
		},
	}

	resp, err := client.UpdateMetric(ctx, req)
	require.NoError(t, err)

	assert.Equal(t, "test_counter", resp.Metric.Id)
	assert.Equal(t, "counter", resp.Metric.Type)
	assert.Equal(t, delta, *resp.Metric.Delta)
}

func TestUpdateMetrics_Batch(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	value := 123.45
	delta := int64(5)

	req := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "batch_gauge",
				Type:  "gauge",
				Value: &value,
			},
			{
				Id:    "batch_counter",
				Type:  "counter",
				Delta: &delta,
			},
		},
	}

	resp, err := client.UpdateMetrics(ctx, req)
	require.NoError(t, err)

	assert.True(t, resp.Success)
}

func TestGetMetric(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Сначала добавим метрику
	value := 99.9
	updateReq := &pb.UpdateMetricRequest{
		Metric: &pb.Metric{
			Id:    "get_test",
			Type:  "gauge",
			Value: &value,
		},
	}

	_, err := client.UpdateMetric(ctx, updateReq)
	require.NoError(t, err)

	// Теперь получим её
	getReq := &pb.GetMetricRequest{
		Id:   "get_test",
		Type: "gauge",
	}

	resp, err := client.GetMetric(ctx, getReq)
	require.NoError(t, err)

	assert.Equal(t, "get_test", resp.Metric.Id)
	assert.Equal(t, "gauge", resp.Metric.Type)
	assert.Equal(t, value, *resp.Metric.Value)
}

func TestGetAllMetrics(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Добавим несколько метрик
	value1 := 111.1
	value2 := 222.2
	delta1 := int64(100)

	updateReq := &pb.UpdateMetricsRequest{
		Metrics: []*pb.Metric{
			{
				Id:    "all_gauge1",
				Type:  "gauge",
				Value: &value1,
			},
			{
				Id:    "all_gauge2",
				Type:  "gauge",
				Value: &value2,
			},
			{
				Id:    "all_counter1",
				Type:  "counter",
				Delta: &delta1,
			},
		},
	}

	_, err := client.UpdateMetrics(ctx, updateReq)
	require.NoError(t, err)

	// Получим все метрики
	getAllReq := &pb.GetAllMetricsRequest{}
	resp, err := client.GetAllMetrics(ctx, getAllReq)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(resp.Metrics), 3)

	// Проверим, что наши метрики присутствуют
	metricNames := make(map[string]bool)
	for _, metric := range resp.Metrics {
		metricNames[metric.Id] = true
	}

	assert.True(t, metricNames["all_gauge1"])
	assert.True(t, metricNames["all_gauge2"])
	assert.True(t, metricNames["all_counter1"])
}

func TestPing(t *testing.T) {
	server, client := setupTestServer()
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.PingRequest{}
	resp, err := client.Ping(ctx, req)
	require.NoError(t, err)

	assert.True(t, resp.Healthy)
}
