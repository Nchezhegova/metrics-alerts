package main

import (
	"flag"
	"github.com/Nchezhegova/metrics-alerts/internal/handlers"
	"github.com/gin-gonic/gin"
	"os"
)

var globalMemory = handlers.MemStorage{}

func main() {
	globalMemory.Counter = make(map[string]int64)
	globalMemory.Gauge = make(map[string]float64)

	r := gin.Default()
	r.POST("/update/:type/:name/:value", func(c *gin.Context) {
		globalMemory.UpdateMetrics(c)
	})
	r.GET("/value/:type/:name/", func(c *gin.Context) {
		globalMemory.GetMetric(c)
	})
	r.GET("/", func(c *gin.Context) {
		globalMemory.PrintMetrics(c)
	})

	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	flag.Parse()
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	err := r.Run(addr)
	if err != nil {
		panic(err)
	}
}
