package main

// Import section with a brief description.
import (
	"bytes"
	"compress/gzip"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/cmd/grpcprotocol"
	"github.com/Nchezhegova/metrics-alerts/cmd/grpcprotocol/proto"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// RetryDelays holds the retry delays.
var RetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

// link flags
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)
var key *rsa.PublicKey
var Client = &http.Client{}

// printBuildInfo prints the build information.
func printBuildInfo() {
	// Command to get the commit value
	cmd := exec.Command("git", "log", "--pretty=format:'%h'", "--abbrev-commit", "-1")
	output, err := cmd.Output()
	if err != nil {
		log.Logger.Info("Ошибка при получении значения коммита:", zap.Error(err))
	} else {
		buildCommit = strings.Trim(string(output), "'\n")
	}

	// Command to get the date value
	cmd = exec.Command("git", "log", "--pretty=format:%cd", "--date=short", "-1")
	output, err = cmd.Output()
	if err != nil {
		log.Logger.Info("Ошибка при получении значения даты:", zap.Error(err))
	} else {
		buildDate = strings.TrimSpace(string(output))
	}
	log.Logger.Info("Build version:", zap.String("version", buildVersion))
	log.Logger.Info("Build date:", zap.String("date", buildDate))
	log.Logger.Info("Build commit:", zap.String("commit", buildCommit))
}

// commonSend sends data with metrics independent of the body
func commonSend(body []byte, url string, hashkey string, dsc *proto.DataServiceClient) {
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

	var encryptCompressBody []byte
	if key != nil {
		encryptCompressBody, err = helpers.EncryptData(compressBody.(*bytes.Buffer).Bytes(), key)
		if err != nil {
			log.Logger.Info("Error closing compressed:", zap.Error(err))
			return
		}
	} else {
		encryptCompressBody = compressBody.(*bytes.Buffer).Bytes()
	}

	if dsc != nil {
		grpcprotocol.TestSend(*dsc, encryptCompressBody)
	} else {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(encryptCompressBody))
		if err != nil {
			log.Logger.Info("Error creating request:", zap.Error(err))
			return
		}
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Real-IP", config.IP)

		if hashkey != "" {
			compressedData := compressBody.(*bytes.Buffer).Bytes()
			req.Header.Set("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(compressedData, hashkey)))
		}
		var resp *http.Response
		for i := 0; i < config.MaxRetries; i++ {
			resp, err = Client.Do(req)
			if err == nil {
				err = resp.Body.Close()
				if err != nil {
					log.Logger.Info("Error closing body:", zap.Error(err))
					return
				}
				break
			} else {
				time.Sleep(RetryDelays[i])
				continue
			}
		}
		if err != nil {
			log.Logger.Info("Max retries", zap.Error(err))
			return
		}
	}
}

// sendMetric specifies the url and prepares the body with the one metric
func sendMetric(m storage.Metrics, addr string, hashkey string, dsc *proto.DataServiceClient) {
	url := fmt.Sprintf("http://%s/update/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	commonSend(body, url, hashkey, dsc)
}

// sendBatchMetrics specifies the URL and prepares the body with a bunch of metrics
func sendBatchMetrics(m []storage.Metrics, addr string, hashkey string) {
	url := fmt.Sprintf("http://%s/updates/", addr)

	body, err := json.Marshal(m)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	commonSend(body, url, hashkey, nil)
}

// collectMetrics collects metrics MemStats and RandomValue
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

// collectgopsutilMetrics  collects metrics VirtualMemory
func collectgopsutilMetrics() []storage.Metrics {
	metrics := []storage.Metrics{}
	memoryStats, err := mem.VirtualMemory()
	if err != nil {
		log.Logger.Info("Error getting memory stats:", zap.Error(err))
	} else {
		total := float64(memoryStats.Total)
		m := storage.Metrics{
			ID:    "TotalMemory",
			MType: config.Gauge,
			Value: &total,
		}
		metrics = append(metrics, m)

		free := float64(memoryStats.Free)
		m = storage.Metrics{
			ID:    "FreeMemory",
			MType: config.Gauge,
			Value: &free,
		}
		metrics = append(metrics, m)
	}
	numCPU := float64(runtime.NumCPU())
	m := storage.Metrics{
		ID:    "CPUutilization1",
		MType: config.Gauge,
		Value: &numCPU,
	}
	metrics = append(metrics, m)
	return metrics
}

func workers(jobs <-chan storage.Metrics, addr string, hashkey string, wg *sync.WaitGroup, dsc *proto.DataServiceClient) {
	for {
		job, ok := <-jobs
		if !ok {
			return
		}
		sendMetric(job, addr, hashkey, dsc)
		wg.Done()
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	wgs := sync.WaitGroup{}

	printBuildInfo()
	conf := config.NewConfig()
	conf.SetConfigFromFlags()
	conf.SetConfigFromEnv()
	err := conf.SetConfigFromJSON()
	if err != nil {
		log.Logger.Info("Error loading configuration:", zap.Error(err))
		return
	}

	if conf.KeyPath != "" {
		key, err = helpers.ConvertPublicKey(conf.KeyPath)
		if err != nil {
			log.Logger.Info("Error reading public key:", zap.Error(err))
			return
		}
	}
	pollInterval := time.Duration(conf.PollInterval) * time.Second
	reportInterval := time.Duration(conf.ReportInterval) * time.Second

	var pollCount int64
	var metrics []storage.Metrics
	var psMetrics []storage.Metrics
	var mu sync.Mutex

	dsc := grpcprotocol.StartGRPCClient(conf.GrpcAddr)

	jobs := make(chan storage.Metrics, conf.RateLimit)
	defer close(jobs)

	go func() {
		for {
			mu.Lock()
			metrics = collectMetrics()
			pollCount++
			mu.Unlock()
			time.Sleep(pollInterval)
		}

	}()
	// gopsutil
	go func() {
		for {
			mu.Lock()
			psMetrics = collectgopsutilMetrics()
			mu.Unlock()
			time.Sleep(pollInterval)
		}

	}()

	for w := 1; w <= conf.RateLimit; w++ {
		go workers(jobs, conf.Addr, conf.Hash, &wgs, dsc)
	}

	for {
		time.Sleep(reportInterval)
		mu.Lock()
		for index := range metrics {
			wgs.Add(1)
			jobs <- metrics[index]
		}
		for index := range psMetrics {
			wgs.Add(1)
			jobs <- psMetrics[index]
		}
		m := storage.Metrics{
			ID:    "PollCount",
			MType: config.Counter,
			Delta: &pollCount,
		}
		sendMetric(m, conf.Addr, conf.Hash, dsc)
		mu.Unlock()

		select {
		case <-sigs:
			log.Logger.Info("Shutting down agent")
			wgs.Wait()
			os.Exit(0)
		default:
		}
	}
}
