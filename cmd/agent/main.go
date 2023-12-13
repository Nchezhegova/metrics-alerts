package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

// Функция для отправки метрики на сервер
func sendMetric(metricType, name string, value interface{}, addr string) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", addr, metricType, name, value)
	_, err := http.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Println("Error sending metric:", err)
	}
}

func collectMetrics(metrics *map[string]interface{}, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	(*metrics)["Alloc"] = memStats.Alloc
	(*metrics)["BuckHashSys"] = memStats.BuckHashSys
	(*metrics)["Frees"] = memStats.Frees
	(*metrics)["GCCPUFraction"] = memStats.GCCPUFraction
	(*metrics)["GCSys"] = memStats.GCSys
	(*metrics)["HeapAlloc"] = memStats.HeapAlloc
	(*metrics)["HeapIdle"] = memStats.HeapIdle
	(*metrics)["HeapInuse"] = memStats.HeapInuse
	(*metrics)["HeapObjects"] = memStats.HeapObjects
	(*metrics)["HeapReleased"] = memStats.HeapReleased
	(*metrics)["HeapSys"] = memStats.HeapSys
	(*metrics)["LastGC"] = memStats.LastGC
	(*metrics)["Lookups"] = memStats.Lookups
	(*metrics)["MCacheInuse"] = memStats.MCacheInuse
	(*metrics)["MCacheSys"] = memStats.MCacheSys
	(*metrics)["MSpanInuse"] = memStats.MSpanInuse
	(*metrics)["MSpanSys"] = memStats.MSpanSys
	(*metrics)["Mallocs"] = memStats.Mallocs
	(*metrics)["NextGC"] = memStats.NextGC
	(*metrics)["NumForcedGC"] = memStats.NumForcedGC
	(*metrics)["NumGC"] = memStats.NumGC
	(*metrics)["OtherSys"] = memStats.OtherSys
	(*metrics)["PauseTotalNs"] = memStats.PauseTotalNs
	(*metrics)["StackInuse"] = memStats.StackInuse
	(*metrics)["StackSys"] = memStats.StackSys
	(*metrics)["Sys"] = memStats.Sys
	(*metrics)["TotalAlloc"] = memStats.TotalAlloc

	(*metrics)["RandomValue"] = rand.Float64()

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
			return
		}
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		pi, err = strconv.Atoi(envPoolInterval)
		if err != nil {
			return
		}

	}

	pollInterval := time.Duration(pi) * time.Second
	reportInterval := time.Duration(ri) * time.Second

	var pollCount int64
	metrics := make(map[string]interface{})
	var mu sync.Mutex

	go func() {
		for {
			collectMetrics(&metrics, &mu)
			pollCount++
			time.Sleep(pollInterval)
		}
	}()

	for {
		<-time.After(reportInterval)
		mu.Lock()
		for name, value := range metrics {
			go sendMetric("gauge", name, value, addr)
		}
		sendMetric("counter", "PollCount", pollCount, addr)
		mu.Unlock()
	}
}
