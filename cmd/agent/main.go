package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func commonSend(body []byte, url string) {
	var compressBody io.ReadWriter = &bytes.Buffer{}
	var err error
	gzipWriter := gzip.NewWriter(compressBody)
	_, err = gzipWriter.Write(body)
	if err != nil {
		fmt.Println("Error convert to gzip.Writer:", err)
		return
	}
	err = gzipWriter.Close()
	if err != nil {
		fmt.Println("Error closing compressed:", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, compressBody)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for i := 0; i < config.MaxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		} else {
			time.Sleep(retryDelays[i])
			continue
		}
	}
	if err != nil {
		fmt.Println("Max retries", err)
		return
	}

	err = resp.Body.Close()
	if err != nil {
		fmt.Println("Error closing body:", err)
		return
	}
}
func sendMetric(m storage.Metrics, addr string) {
	url := fmt.Sprintf("http://%s/update/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error convert to JSON:", err)
		return
	}
	commonSend(body, url)
}
func sendBatchMetrics(m []storage.Metrics, addr string) {
	url := fmt.Sprintf("http://%s/updates/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error convert to JSON:", err)
		return
	}
	commonSend(body, url)
}

func collectMetrics() []storage.Metrics {
	metrics := []storage.Metrics{}
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	val := reflect.ValueOf(memStats)
	selectedFields := []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys", "PauseTotalNs",
		"StackInuse", "StackSys", "Sys", "TotalAlloc"}
	for _, fieldName := range selectedFields {
		var field float64
		if val.FieldByName(fieldName).Kind() == reflect.Uint64 {
			field = float64(val.FieldByName(fieldName).Uint())
		} else if val.FieldByName(fieldName).Kind() == reflect.Float64 {
			field = val.FieldByName(fieldName).Float()
		}

		m := storage.Metrics{
			ID:    fieldName,
			MType: config.Gauge,
			Value: &field,
		}
		metrics = append(metrics, m)
	}
	randomValue := rand.Float64()
	m := storage.Metrics{
		ID:    "RandomValue",
		MType: config.Gauge,
		Value: &randomValue,
	}
	metrics = append(metrics, m)
	return metrics
}

func main() {
	var addr string
	var pi int
	var ri int
	var err error

	flag.IntVar(&pi, "p", 2, "pollInterval")
	flag.IntVar(&ri, "r", 10, "reportInterval")
	flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		ri, err = strconv.Atoi(envReportInterval)
		if err != nil {
			fmt.Println("Invalid parameter REPORT_INTERVAL:", err)
			return
		}
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		pi, err = strconv.Atoi(envPoolInterval)
		if err != nil {
			fmt.Println("Invalid parameter POLL_INTERVAL:", err)
			return
		}
	}

	pollInterval := time.Duration(pi) * time.Second
	reportInterval := time.Duration(ri) * time.Second

	var pollCount int64
	var metrics []storage.Metrics

	var mu sync.Mutex

	go func() {
		for {
			mu.Lock()
			metrics = collectMetrics()
			pollCount++
			mu.Unlock()
			time.Sleep(pollInterval)
		}

	}()

	for {
		time.Sleep(reportInterval)
		mu.Lock()
		for index := range metrics {
			sendMetric(metrics[index], addr)
		}
		m := storage.Metrics{
			ID:    "PollCount",
			MType: config.Counter,
			Delta: &pollCount,
		}
		sendMetric(m, addr)
		sendBatchMetrics(metrics, addr)
		mu.Unlock()
	}
}
