package main

import (
	"flag"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/handlers"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"os"
	"strconv"
)

func main() {
	var globalMemory = storage.MemStorage{}
	globalMemory.Counter = make(map[string]int64)
	globalMemory.Gauge = make(map[string]float64)

	var addr string
	var storeInterval int
	var filePath string
	var restore bool
	flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	flag.IntVar(&storeInterval, "i", 0, "input addr serv")
	flag.StringVar(&filePath, "f", "/tmp/metrics-db.json", "input addr serv")
	flag.BoolVar(&restore, "r", true, "input addr serv")
	flag.Parse()
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		storeIntervalInt, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			fmt.Println("Error convert STORE_INTERVAL to int:", err)
			return
		} else {
			storeInterval = storeIntervalInt
		}
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		filePath = envFilePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restoreValue, err := strconv.ParseBool(envRestore)
		if err != nil {
			fmt.Println("Error convert RESTORE to bool:", err)
			return
		} else {
			restore = restoreValue
		}
	}

	// перенесла старт сервака и обработку url в handlers
	handlers.StartServ(&globalMemory, addr, storeInterval, filePath, restore)
}
