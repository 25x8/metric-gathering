package collectors

import (
	"testing"
)

func BenchmarkMetricsCollector_Collect(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Collect()
	}
}

func BenchmarkMetricsCollector_CollectSystemMetrics(b *testing.B) {
	collector := NewMetricsCollector()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.CollectSystemMetrics()
	}
}

func BenchmarkMetricsCollector_GetMetrics(b *testing.B) {
	collector := NewMetricsCollector()
	// Соберем некоторые метрики, чтобы было что получать
	collector.Collect()
	collector.CollectSystemMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = collector.GetMetrics()
	}
}
