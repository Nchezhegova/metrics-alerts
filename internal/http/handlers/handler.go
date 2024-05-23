package handlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/http/middleware"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var mu sync.Mutex

// updateMetrics updates one metric from url params
func updateMetrics(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string) {
	mu.Lock()
	defer mu.Unlock()

	switch c.Param("type") {
	case config.Gauge:
		k := c.Param("name")
		v, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		m.GaugeStorage(c, k, v)

	case config.Counter:
		k := c.Param("name")
		v, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		m.CountStorage(c, k, v)
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}

// updateMetricsFromBody updates one metric from body
func updateMetricsFromBody(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string, hashKey string) {
	mu.Lock()
	defer mu.Unlock()

	var metrics storage.Metrics
	var b io.ReadCloser

	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer gz.Close()
		b = gz
		c.Header("Accept-Encoding", "gzip")

	} else {
		b = c.Request.Body
	}

	decoder := json.NewDecoder(b)
	err := decoder.Decode(&metrics)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch metrics.MType {
	case config.Gauge:
		k := metrics.ID
		v := metrics.Value
		m.GaugeStorage(c, k, *v)

	case config.Counter:
		k := metrics.ID
		v := metrics.Delta
		m.CountStorage(c, k, *v)
		vNew, _ := m.GetCount(c, metrics.ID)
		metrics.Delta = &vNew

	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	metricsByte, err := json.Marshal(metrics)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		compressBody := helpers.CompressResp(metricsByte)
		compressedData := compressBody.String()
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", "application/json")
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", metricsByte)
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}

// updateBatchMetricsFromBody updates a batch of metrics from the body
func updateBatchMetricsFromBody(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string, hashKey string) {
	mu.Lock()
	defer mu.Unlock()

	var metricsList []storage.Metrics
	var b io.ReadCloser

	if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer gz.Close()
		b = gz
		c.Header("Accept-Encoding", "gzip")

	} else {
		b = c.Request.Body
	}

	decoder := json.NewDecoder(b)
	err := decoder.Decode(&metricsList)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err = m.UpdateBatch(c, metricsList)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	metricsByte, err := json.Marshal(metricsList)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		compressBody := helpers.CompressResp(metricsByte)
		compressedData := compressBody.String()
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", "application/json")
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", metricsByte)
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}

// getMetric displays the value by the key that came in the url
func getMetric(c *gin.Context, m storage.MStorage) {
	switch c.Param("type") {
	case config.Counter:
		v, exists := m.GetCount(c, c.Param("name"))
		if exists {
			c.JSON(http.StatusOK, v)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case config.Gauge:
		v, exists := m.GetGauge(c, c.Param("name"))
		if exists {
			c.JSON(http.StatusOK, v)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}

// getMetric displays the value by the key that came in the body
func getMetricFromBody(c *gin.Context, m storage.MStorage, hashKey string) {

	var metrics storage.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	switch metrics.MType {
	case config.Counter:
		v, exists := m.GetCount(c, metrics.ID)
		if exists {
			metrics.Delta = &v
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case config.Gauge:
		v, exists := m.GetGauge(c, metrics.ID)
		if exists {
			metrics.Value = &v
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	metricsByte, err := json.Marshal(metrics)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		compressBody := helpers.CompressResp(metricsByte)
		compressedData := compressBody.String()
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", "application/json")
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		if hashKey != "" {
			c.Header("HashSHA256", base64.StdEncoding.EncodeToString(helpers.CalculateHash(metricsByte, hashKey)))
		}
		c.Data(http.StatusOK, "application/json", metricsByte)
	}
}

// printMetrics prints all metrics
func printMetrics(c *gin.Context, m storage.MStorage) {
	res := m.GetStorage(c)
	metricsByte, err := json.Marshal(res)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		_, err = gzipWriter.Write(metricsByte)
		if err != nil {
			log.Logger.Info("Error convert to gzip.Writer:", zap.Error(err))
			return
		}

		err = gzipWriter.Close()
		if err != nil {
			log.Logger.Info("Error closing compressed:", zap.Error(err))
			return
		}

		compressedData := compressBody.String()
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", "text/html")

		c.Data(http.StatusOK, "text/html", []byte(compressedData))
	}
	c.String(http.StatusOK, string(metricsByte))
}

// checkDB checks DB connection
func checkDB(c *gin.Context, db *sql.DB) {
	if db != nil {
		err := storage.CheckConnect(db)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// checkHash checks hash-key from request
func checkHash(c *gin.Context, hashKey string) bool {
	if hashKey == "" {
		return true
	}
	if c.GetHeader("HashSHA256") != "" {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, c.Request.Body)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return false
		}
		hashServe := base64.StdEncoding.EncodeToString(helpers.CalculateHash(buf.Bytes(), hashKey))
		hashAgent := (c.GetHeader("HashSHA256"))
		if hashServe == hashAgent {
			c.Request.Body = io.NopCloser(&buf)
			return true
		} else {
			c.AbortWithStatus(http.StatusBadRequest)
			return false
		}
	}
	return true
}

// StartServ starts the server and routes requests
func StartServ(m storage.MStorage, addr string, storeInterval int, filePath string, restore bool,
	hashKey string, keyPath string, trustedSubnet string) {
	r := gin.Default()
	r.ContextWithFallback = true

	r.Use(log.GinLogger(log.Logger), gin.Recovery())
	if trustedSubnet != "" {
		r.Use(middleware.CheckIP(trustedSubnet), gin.Recovery())
	}
	syncWrite := helpers.SetWriterFile(m, storeInterval, filePath, restore)

	var key *rsa.PrivateKey
	var err error
	if keyPath != "" {
		key, err = helpers.ConvertPrivateKey(keyPath)
		if err != nil {
			log.Logger.Info("Error convert to private key:", zap.Error(err))
			os.Exit(1)
		}
	}

	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		updateMetrics(c, m, syncWrite, filePath)
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		getMetric(c, m)
	})
	r.POST("/value/", func(c *gin.Context) {
		if checkHash(c, hashKey) {
			getMetricFromBody(c, m, hashKey)
		} else {
			log.Logger.Info("Problem with hashkey")
		}
	})
	r.GET("/", func(c *gin.Context) {
		printMetrics(c, m)
	})
	r.GET("/ping", func(c *gin.Context) {
		checkDB(c, storage.DB)
	})

	r.Use(middleware.DecryptBody(key))
	{
		r.POST("/updates/", func(c *gin.Context) {
			if checkHash(c, hashKey) {
				updateBatchMetricsFromBody(c, m, syncWrite, filePath, hashKey)
			} else {
				log.Logger.Info("Problem with hashkey")
			}
		})
		r.POST("/update/", func(c *gin.Context) {
			if checkHash(c, hashKey) {
				updateMetricsFromBody(c, m, syncWrite, filePath, hashKey)
			} else {
				log.Logger.Info("Problem with hashkey")
			}
		})
	}

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigint
		log.Logger.Info("Shutting down the server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = server.Shutdown(ctx); err != nil {
			log.Logger.Error("Error shutting down the server:", zap.Error(err))
		}
	}()
	if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Logger.Error("Error starting the server:", zap.Error(err))
	}
}
