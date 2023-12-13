package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"sync"
)

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

type Storage interface {
	PrintMetrics()
	UpdateMetrics()
	GetMetric()
}

var mu sync.Mutex

func (ms *MemStorage) UpdateMetrics(c *gin.Context) {
	switch c.Param("type") {
	case "gauge":
		k := c.Param("name")
		v, err := strconv.ParseFloat(c.Param("value"), 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		ms.Gauge[k] = v
	case "counter":
		k := c.Param("name")
		v, err := strconv.ParseInt(c.Param("value"), 10, 64)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		ms.Counter[k] += v
	default:
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
}
func (ms *MemStorage) GetMetric(c *gin.Context) {
	switch c.Param("type") {
	case "counter":
		v, exists := ms.Counter[c.Param("name")]
		if exists {
			c.JSON(http.StatusOK, v)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	case "gauge":
		v, exists := ms.Gauge[c.Param("name")]
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

func (ms *MemStorage) PrintMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, ms)
}
