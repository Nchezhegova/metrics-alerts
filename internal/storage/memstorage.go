package storage

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

type MStorage interface {
	CountStorage(string, int64)
	GaugeStorage(string, float64)
	GetStorage() MemStorage
	GetCount(string) (int64, bool)
	GetGauge(string) (float64, bool)
}

func (s *MemStorage) CountStorage(k string, v int64) {
	s.Counter[k] += v
}

func (s *MemStorage) GaugeStorage(k string, v float64) {
	s.Gauge[k] = v
}

func (s *MemStorage) GetStorage() MemStorage {
	return *s
}

func (s *MemStorage) GetGauge(key string) (float64, bool) {
	v, exists := s.Gauge[key]
	return v, exists
}

func (s *MemStorage) GetCount(key string) (int64, bool) {
	v, exists := s.Counter[key]
	return v, exists
}
