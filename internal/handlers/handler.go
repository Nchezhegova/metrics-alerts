package handlers

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var mu sync.Mutex

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
func updateMetricsFromBody(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string) {
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

	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		metricsByte, err := json.Marshal(metrics)
		if err != nil {
			log.Logger.Info("Error convert to JSON:", zap.Error(err))
			return
		}

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
		c.Header("Content-Type", "application/json")

		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		c.JSON(http.StatusOK, metrics)
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}
func updateBatchMetricsFromBody(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string) {
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
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		metricsByte, err := json.Marshal(metricsList)
		if err != nil {
			log.Logger.Info("Error convert to JSON:", zap.Error(err))
			return
		}

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
		c.Header("Content-Type", "application/json")

		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		c.JSON(http.StatusOK, metricsList)
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}

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
func getMetricFromBody(c *gin.Context, m storage.MStorage) {

	var metrics storage.Metrics
	var buf bytes.Buffer
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
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

	if c.GetHeader("Accept-Encoding") == "gzip" {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		metricsByte, err := json.Marshal(metrics)
		if err != nil {
			log.Logger.Info("Error convert to JSON:", zap.Error(err))
			return
		}

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
		c.Header("Content-Type", "application/json")

		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		c.JSON(http.StatusOK, metrics)
		//c.String(http.StatusOK, metrics)
	}
}

func printMetrics(c *gin.Context, m storage.MStorage) {
	res := m.GetStorage(c)
	metricsByte, err := json.Marshal(res)
	if err != nil {
		log.Logger.Info("Error convert to JSON:", zap.Error(err))
		return
	}
	if c.GetHeader("Accept-Encoding") == "gzip" {
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
	} else {
		c.String(http.StatusOK, string(metricsByte))
	}
}

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

func StartServ(m storage.MStorage, addr string, storeInterval int, filePath string, restore bool) {
	r := gin.Default()
	r.ContextWithFallback = true

	r.Use(log.GinLogger(log.Logger), gin.Recovery())

	syncWrite := helpers.SetWriterFile(m, storeInterval, filePath, restore)

	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		updateMetrics(c, m, syncWrite, filePath)
	})
	r.POST("/update/", func(c *gin.Context) {
		updateMetricsFromBody(c, m, syncWrite, filePath)
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		getMetric(c, m)
	})
	r.POST("/value/", func(c *gin.Context) {
		getMetricFromBody(c, m)
	})
	r.GET("/", func(c *gin.Context) {
		printMetrics(c, m)
	})
	r.GET("/ping", func(c *gin.Context) {
		checkDB(c, storage.DB)
	})
	r.POST("/updates/", func(c *gin.Context) {
		updateBatchMetricsFromBody(c, m, syncWrite, filePath)
	})

	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}
