package storage

import (
	"context"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

//go:generate  mockgen -build_flags=--mod=mod -destination=mocks/mock_store.go -package=mocks . MStorage
type MStorage interface {
	CountStorage(context.Context, string, int64)
	GaugeStorage(context.Context, string, float64)
	GetStorage(context.Context) interface{}
	GetCount(context.Context, string) (int64, bool)
	GetGauge(context.Context, string) (float64, bool)
	SetStartData(MemStorage)
	UpdateBatch(context.Context, []Metrics) error
}

func (s *MemStorage) CountStorage(c context.Context, k string, v int64) {
	s.Counter[k] += v
}

func (s *MemStorage) GaugeStorage(c context.Context, k string, v float64) {
	s.Gauge[k] = v
}

func (s *MemStorage) GetStorage(c context.Context) interface{} {
	return *s
}

func (s *MemStorage) SetStartData(storage MemStorage) {
	s.Gauge = storage.Gauge
	s.Counter = storage.Counter
}

func (s *MemStorage) GetGauge(c context.Context, key string) (float64, bool) {
	v, exists := s.Gauge[key]
	return v, exists
}

func (s *MemStorage) GetCount(c context.Context, key string) (int64, bool) {
	v, exists := s.Counter[key]
	return v, exists
}

func (s *MemStorage) UpdateBatch(c context.Context, list []Metrics) error {
	for _, metric := range list {
		switch metric.MType {
		case config.Gauge:
			k := metric.ID
			v := metric.Value
			s.GaugeStorage(c, k, *v)

		case config.Counter:
			k := metric.ID
			v := metric.Delta
			s.CountStorage(c, k, *v)
			vNew, _ := s.GetCount(c, metric.ID)
			metric.Delta = &vNew
		default:
			err := fmt.Errorf("unknowning metric type")
			return err
		}
	}
	return nil
}
