package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Example_updateMetrics() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Params = []gin.Param{
		{Key: "type", Value: "gauge"},
		{Key: "name", Value: "test"},
		{Key: "value", Value: "10.5"},
	}
	updateMetrics(c, &m, false, "testFilePath")

	c.Params = []gin.Param{
		{Key: "type", Value: "counter"},
		{Key: "name", Value: "test_counter"},
		{Key: "value", Value: "5"},
	}
	updateMetrics(c, &m, false, "testFilePath")

	c.Params = []gin.Param{
		{Key: "type", Value: "invalid_type"},
	}

	updateMetrics(c, &m, false, "testFilePath")
	// Output:
}

func Example_updateMetricsFromBody() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	bodyData := `{"MType": "gauge", "ID": "test", "Value": 10.5}`
	r := strings.NewReader(bodyData)
	req, _ := http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(bodyData)); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Body = io.NopCloser(&buf)
	c.Request = req
	updateMetricsFromBody(c, &m, false, "testFilePath", "hashKey")

	bodyData = `{"MType": "counter", "ID": "test_counter", "Delta": 5}`
	r = strings.NewReader(bodyData)
	req, _ = http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	updateMetricsFromBody(c, &m, false, "testFilePath", "hashKey")
	// Output:
}

func Example_updateBatchMetricsFromBody() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	var v float64 = 10.5
	var d int64 = 5
	metricsList := []storage.Metrics{
		{MType: "gauge", ID: "test1", Value: &v},
		{MType: "counter", ID: "test2", Delta: &d},
	}
	bodyData, _ := json.Marshal(metricsList)
	r := bytes.NewReader(bodyData)
	req, _ := http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	updateBatchMetricsFromBody(c, &m, false, "testFilePath", "hashKey")

	// Output:
}

func Example_getMetric() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	m.GaugeStorage(nil, "test_gauge", 10.5)
	m.CountStorage(nil, "test_counter", 5)

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Params = []gin.Param{
		{Key: "type", Value: config.Counter},
		{Key: "name", Value: "test_counter"},
	}
	getMetric(c, &m)
	c.Params = []gin.Param{
		{Key: "type", Value: config.Gauge},
		{Key: "name", Value: "test_gauge"},
	}
	getMetric(c, &m)

	c.Params = []gin.Param{
		{Key: "type", Value: "invalid_type"},
	}
	getMetric(c, &m)

	c.Params = []gin.Param{
		{Key: "type", Value: config.Counter},
		{Key: "name", Value: "nonexistent_metric"},
	}
	getMetric(c, &m)

	// Output:
}

func Example_getMetricFromBody() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	bodyData := `{"MType": "counter", "ID": "test_counter"}`
	r := bytes.NewReader([]byte(bodyData))
	req, _ := http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(bodyData)); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Body = io.NopCloser(&buf)
	c.Request = req
	getMetricFromBody(c, &m, "hashKey")

	bodyData = `{"MType": "gauge", "ID": "test_gauge"}`
	r = bytes.NewReader([]byte(bodyData))
	req, _ = http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	getMetricFromBody(c, &m, "hashKey")

	bodyData = `{"MType": "invalid_type"}`
	r = bytes.NewReader([]byte(bodyData))
	req, _ = http.NewRequest("POST", "/", r)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	getMetricFromBody(c, &m, "hashKey")

	// Output:
}
func Example_printMetrics() {
	m := storage.MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
	m.GaugeStorage(nil, "test_gauge", 10.5)
	m.CountStorage(nil, "test_counter", 5)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	c.Request, _ = http.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Accept-Encoding", "")
	printMetrics(c, &m)

	// Output:
}
