package main

import (
	"flag"
	"github.com/Nchezhegova/metrics-alerts/internal/handlers"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"os"
)

func main() {
	var globalMemory = storage.MemStorage{}
	globalMemory.Counter = make(map[string]int64)
	globalMemory.Gauge = make(map[string]float64)

	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	flag.Parse()
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}

	// перенесла старт сервака и обработку url в handlers
	handlers.StartServ(&globalMemory, addr)
}
