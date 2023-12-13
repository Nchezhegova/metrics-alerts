package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestCollectMetrics(t *testing.T) {
	var mu sync.Mutex
	metrics := make(map[string]interface{})

	collectMetrics(&metrics, &mu)

	testMetrics := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction",
		"GCSys", "HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC",
		"Lookups", "MCacheInuse", "MCacheSys", "MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
		"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc"}

	for _, metric := range testMetrics {
		if _, exists := metrics[metric]; !exists {
			t.Errorf("Metric %s was not collected", metric)
		}
	}

	if _, exists := metrics["RandomValue"]; !exists {
		t.Errorf("Custom metric RandomValue was not collected")
	}
}

func TestSendMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/update/gauge/SomeMetric/123.45"
		if r.URL.EscapedPath() != expectedPath {
			t.Errorf("Expected request path '%s', got '%s'", expectedPath, r.URL.EscapedPath())
		}
	}))
	sendMetric("gauge", "SomeMetric", 123.45, server.Listener.Addr().String())
	defer server.Close()

}
