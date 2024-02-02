package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"go.uber.org/zap"
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

func commonSend(body []byte, url string, hashkey string) {
	var compressBody io.ReadWriter = &bytes.Buffer{}
	var err error

	gzipWriter := gzip.NewWriter(compressBody)
	_, err = gzipWriter.Write(body)
	if err != nil {
		log.Logger.Info("Error convert to gzip.Writer:", zap.Error(err))
		return
	}
	err = gzipWriter.Close()
	if err != nil {
		log.Logger.Info("Error closing compressed:", zap.Error(err))
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, compressBody)
	if err != nil {
		log.Logger.Info("Error creating request:", zap.Error(err))
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	if hashkey != "" {
		compressedData := compressBody.(*bytes.Buffer).Bytes()
		req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(compressedData, hashkey)))
		//fmt.Println(string(helpers.CalculateHash(compressedData, hashkey)))
	}

	var resp *http.Response
	for i := 0; i < config.MaxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			err = resp.Body.Close()
			if err != nil {
				log.Logger.Info("Error closing body:", zap.Error(err))
				return
			}
			break
		} else {
			time.Sleep(retryDelays[i])
			continue
		}
	}
	if err != nil {
		log.Logger.Info("Max retries", zap.Error(err))
		return
	}
}
func sendMetric(m storage.Metrics, addr string, hashkey string) {
	url := fmt.Sprintf("http://%s/update/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	commonSend(body, url, hashkey)
}
func sendBatchMetrics(m []storage.Metrics, addr string, hashkey string) {
	url := fmt.Sprintf("http://%s/updates/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	commonSend(body, url, hashkey)
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
	var hash string
	var err error

	flag.IntVar(&pi, "p", 2, "pollInterval")
	flag.IntVar(&ri, "r", 10, "reportInterval")
	flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	flag.StringVar(&hash, "k", "123", "input hash")
	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		ri, err = strconv.Atoi(envReportInterval)
		if err != nil {
			log.Logger.Info("Invalid parameter REPORT_INTERVAL:", zap.Error(err))
			return
		}
	}
	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		pi, err = strconv.Atoi(envPoolInterval)
		if err != nil {
			log.Logger.Info("Invalid parameter POLL_INTERVAL:", zap.Error(err))
			return
		}
	}
	if envHashKey := os.Getenv("KEY"); envHashKey != "" {
		hash = envHashKey
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
			sendMetric(metrics[index], addr, hash)
		}
		m := storage.Metrics{
			ID:    "PollCount",
			MType: config.Counter,
			Delta: &pollCount,
		}
		sendMetric(m, addr, hash)
		sendBatchMetrics(metrics, addr, hash)
		mu.Unlock()
	}
}
