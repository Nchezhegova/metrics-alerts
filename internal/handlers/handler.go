package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
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
		m.GaugeStorage(k, v)

	case config.Counter:
		k := c.Param("name")
		v, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		m.CountStorage(k, v)
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if syncWrite {
		helpers.WriteFile(m, filePath)
	}
}
func getMetric(c *gin.Context, m storage.MStorage) {
	switch c.Param("type") {
	case config.Counter:
		v, exists := m.GetCount(c.Param("name"))
		if exists {
			c.JSON(http.StatusOK, v)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case config.Gauge:
		v, exists := m.GetGauge(c.Param("name"))
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
func updateMetricsFromBody(c *gin.Context, m storage.MStorage, syncWrite bool, filePath string) {
	mu.Lock()
	defer mu.Unlock()

	var metrics storage.Metrics
	var b io.ReadCloser

	//if c.GetHeader("Content-Encoding") == "gzip" {
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
		m.GaugeStorage(k, *v)

	case config.Counter:
		k := metrics.ID
		v := metrics.Delta
		m.CountStorage(k, *v)
		vNew, _ := m.GetCount(metrics.ID)
		metrics.Delta = &vNew

	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//if c.GetHeader("Accept-Encoding") == "gzip" {
	if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		metricsByte, err := json.Marshal(metrics)
		if err != nil {
			fmt.Println("Error convert to JSON:", err)
			return
		}

		_, err = gzipWriter.Write(metricsByte)
		if err != nil {
			fmt.Println("Error convert to gzip.Writer:", err)
			return
		}

		err = gzipWriter.Close()
		if err != nil {
			fmt.Println("Error closing compressed:", err)
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
	case "counter":
		v, exists := m.GetCount(metrics.ID)
		if exists {
			metrics.Delta = &v
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case "gauge":
		v, exists := m.GetGauge(metrics.ID)
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
			fmt.Println("Error convert to JSON:", err)
			return
		}

		_, err = gzipWriter.Write(metricsByte)
		if err != nil {
			fmt.Println("Error convert to gzip.Writer:", err)
			return
		}

		err = gzipWriter.Close()
		if err != nil {
			fmt.Println("Error closing compressed:", err)
			return
		}

		compressedData := compressBody.String()
		c.Header("Content-Encoding", "gzip")
		c.Header("Content-Type", "application/json")

		c.Data(http.StatusOK, "application/json", []byte(compressedData))
	} else {
		c.JSON(http.StatusOK, metrics)
	}
}

func printMetrics(c *gin.Context, m storage.MStorage) {
	res := m.GetStorage()
	metricsByte, err := json.Marshal(res)
	if err != nil {
		fmt.Println("Error convert to JSON:", err)
		return
	}
	if c.GetHeader("Accept-Encoding") == "gzip" {
		var compressBody bytes.Buffer
		gzipWriter := gzip.NewWriter(&compressBody)

		_, err = gzipWriter.Write(metricsByte)
		if err != nil {
			fmt.Println("Error convert to gzip.Writer:", err)
			return
		}

		err = gzipWriter.Close()
		if err != nil {
			fmt.Println("Error closing compressed:", err)
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

func StartServ(m storage.MStorage, addr string, storeInterval int, filePath string, restore bool) {
	r := gin.Default()

	// и тут инициализация.
	Logger := helpers.InitLogger()
	r.Use(helpers.GinLogger(Logger), gin.Recovery())

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

	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}
