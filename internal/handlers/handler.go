package handlers

import (
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
)

var mu sync.Mutex

func updateMetrics(c *gin.Context, m storage.MStorage) {
	mu.Lock()
	defer mu.Unlock()
	switch c.Param("type") {
	case "gauge":
		k := c.Param("name")
		v, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		m.GaugeStorage(k, v)

	case "counter":
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
}
func getMetric(c *gin.Context, m storage.MStorage) {
	switch c.Param("type") {
	case "counter":
		v, exists := m.GetCount(c.Param("name"))
		if exists {
			c.JSON(http.StatusOK, v)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case "gauge":
		//v, exists := m.Gauge[c.Param("name")]
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

func printMetrics(c *gin.Context, m storage.MStorage) {
	res := m.GetStorage()
	c.JSON(http.StatusOK, res)
}

func StartServ(m storage.MStorage, addr string) {
	r := gin.Default()
	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		updateMetrics(c, m)
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		getMetric(c, m)
	})
	r.GET("/", func(c *gin.Context) {
		printMetrics(c, m)
	})

	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}
