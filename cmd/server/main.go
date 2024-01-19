package main

import (
	"flag"
	"github.com/Nchezhegova/metrics-alerts/internal/handlers"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"go.uber.org/zap"
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
			log.Logger.Info("Error convert STORE_INTERVAL to int:", zap.Error(err))
			return
		}
		storeInterval = storeIntervalInt
	}
	if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
		filePath = envFilePath
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restoreValue, err := strconv.ParseBool(envRestore)
		if err != nil {
			log.Logger.Info("Error convert RESTORE to bool:", zap.Error(err))
			return
		}
		restore = restoreValue
	}

	handlers.StartServ(&globalMemory, addr, storeInterval, filePath, restore)

	defer log.Logger.Sync()
	defer storage.DB.Close()
}
