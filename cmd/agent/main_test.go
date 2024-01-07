package main

import (
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCollectMetrics(t *testing.T) {
	var metrics []storage.Metrics
	metrics = collectMetrics()
	if len(metrics) != 28 {
		t.Errorf("not correct metrics len")
	}
}

func TestSendMetric(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/update/"
		if r.URL.EscapedPath() != expectedPath {
			t.Errorf("Expected request path '%s', got '%s'", expectedPath, r.URL.EscapedPath())
		}
	}))
	randomValue := rand.Float64()
	sendMetric(storage.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
		Value: &randomValue,
	}, server.Listener.Addr().String())
	defer server.Close()

}
