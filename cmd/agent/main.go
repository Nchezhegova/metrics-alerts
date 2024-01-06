package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Функция для отправки метрики на сервер
//func sendMetric(metricType, name string, value interface{}, addr string) {
//	url := fmt.Sprintf("http://%s/update/%s/%s/%v", addr, metricType, name, value)
//	resp, err := http.Post(url, "text/plain", nil)
//	if err != nil {
//		fmt.Println("Error sending metric:", err)
//		return
//	}
//	err = resp.Body.Close()
//	if err != nil {
//		fmt.Println("Error closing body:", err)
//		return
//	}
//}

func sendMetric(m storage.Metrics, addr string) {
	url := fmt.Sprintf("http://%s/update/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error convert to JSON:", err)
		return
	}
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error sending metric:", err)
		return
	}
	err = resp.Body.Close()
	if err != nil {
		fmt.Println("Error closing body:", err)
		return
	}
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
			MType: "gauge",
			Value: &field,
		}
		metrics = append(metrics, m)
	}
	randomValue := rand.Float64()
	m := storage.Metrics{
		ID:    "RandomValue",
		MType: "gauge",
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
		for index, _ := range metrics {
			//sendMetric("gauge", name, value, addr)
			sendMetric(metrics[index], addr)
		}
		m := storage.Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &pollCount,
		}

		sendMetric(m, addr)
		mu.Unlock()
	}
}
