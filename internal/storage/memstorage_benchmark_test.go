package storage

import (
	"context"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"testing"
)

func BenchmarkCountStorage(b *testing.B) {
	memStorage := &MemStorage{
		Counter: make(map[string]int64),
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memStorage.CountStorage(ctx, "test", 1)
	}
}

func BenchmarkGaugeStorage(b *testing.B) {
	memStorage := &MemStorage{
		Gauge: make(map[string]float64),
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memStorage.GaugeStorage(ctx, "test", 1.0)
	}
}

func BenchmarkGetGauge(b *testing.B) {
	memStorage := &MemStorage{
		Gauge: map[string]float64{"test": 1.0},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memStorage.GetGauge(ctx, "test")
	}
}

func BenchmarkGetCount(b *testing.B) {
	memStorage := &MemStorage{
		Counter: map[string]int64{"test": 1},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memStorage.GetCount(ctx, "test")
	}
}

func BenchmarkUpdateBatch(b *testing.B) {
	memStorage := &MemStorage{
		Counter: make(map[string]int64),
		Gauge:   make(map[string]float64),
	}
	v := 1.0
	var d int64 = 1
	metricsList := []Metrics{
		{MType: config.Gauge, ID: "test_gauge", Value: &v},
		{MType: config.Counter, ID: "test_counter", Delta: &d},
	}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memStorage.UpdateBatch(ctx, metricsList)
	}
}
