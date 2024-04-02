package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountStorage(t *testing.T) {
	storage := &MemStorage{
		Counter: make(map[string]int64),
	}
	storage.CountStorage(context.Background(), "test_key", 5)
	assert.Equal(t, int64(5), storage.Counter["test_key"], "Expected value to be 5")
}

func TestMemStorage_GaugeStorage(t *testing.T) {
	storage := &MemStorage{
		Gauge: make(map[string]float64),
	}
	v := 5.5
	storage.GaugeStorage(context.Background(), "test_key", v)
	assert.Equal(t, v, storage.Gauge["test_key"], "Expected value to be 5")
}

func TestGetGauge(t *testing.T) {
	storage := &MemStorage{
		Gauge: map[string]float64{
			"test_key": 10.5,
		},
	}
	value, exists := storage.GetGauge(context.Background(), "test_key")
	assert.True(t, exists, "Expected key to exist")
	assert.Equal(t, 10.5, value, "Expected value to be 10.5")
}
func TestGetCount(t *testing.T) {
	storage := &MemStorage{
		Counter: map[string]int64{
			"test_key": 10,
		},
	}
	value, exists := storage.GetCount(context.Background(), "test_key")
	assert.True(t, exists, "Expected key to exist")
	assert.Equal(t, int64(10), value, "Expected value to be 10.5")
}

func TestGetStorage(t *testing.T) {
	storage := &MemStorage{
		Gauge: map[string]float64{
			"test_key": 10.5,
		},
		Counter: map[string]int64{
			"test_counter": 20,
		},
	}
	result := storage.GetStorage(context.Background())
	memStorage, ok := result.(MemStorage)
	assert.True(t, ok, "Expected result to be of type MemStorage")
	assert.Equal(t, storage.Gauge, memStorage.Gauge, "Expected Gauge maps to be equal")
	assert.Equal(t, storage.Counter, memStorage.Counter, "Expected Counter maps to be equal")
}

func TestUpdateBatch(t *testing.T) {
	storage := &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}

	var v float64 = 10.5
	var d int64 = 5

	testMetrics := []Metrics{
		{ID: "metric1", MType: "gauge", Value: &v},
		{ID: "metric2", MType: "counter", Delta: &d},
	}

	err := storage.UpdateBatch(context.Background(), testMetrics)
	assert.NoError(t, err, "Expected no error")

	gaugeValue, exists := storage.Gauge["metric1"]
	assert.True(t, exists, "Expected gauge metric1 to exist")
	assert.Equal(t, 10.5, gaugeValue, "Expected gauge metric1 value to be 10.5")

	counterValue, exists := storage.Counter["metric2"]
	assert.True(t, exists, "Expected counter metric2 to exist")
	assert.Equal(t, int64(5), counterValue, "Expected counter metric2 value to be 5")
}
