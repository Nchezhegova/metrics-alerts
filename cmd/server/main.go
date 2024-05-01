package main

import (
	"github.com/Nchezhegova/metrics-alerts/internal/config"
	"github.com/Nchezhegova/metrics-alerts/internal/http/handlers"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"go.uber.org/zap"
	_ "net/http/pprof"
	"os/exec"
	"strings"
)

// link flags
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

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

func main() {
	//var addr string
	//var storeInterval int
	//var filePath string
	//var restore bool
	//var hash string
	//var keyPath string

	printBuildInfo()
	conf := config.NewConfig()
	conf.SetConfigFromFlags()
	conf.SetConfigFromEnv()
	err := conf.SetConfigFromJSON()
	if err != nil {
		log.Logger.Info("Error loading configuration:", zap.Error(err))
		return
	}

	//
	//flag.StringVar(&addr, "a", "localhost:8080", "input addr serv")
	//flag.IntVar(&storeInterval, "i", 0, "input addr serv")
	//flag.StringVar(&filePath, "f", "/tmp/metrics-db.json", "input addr serv")
	//flag.BoolVar(&restore, "r", true, "input addr serv")
	//flag.StringVar(&hash, "k", "", "input hash")
	//flag.StringVar(&keyPath, "crypto-key", "", "path to the key file")
	//flag.Parse()
	//if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
	//	addr = envRunAddr
	//}
	//if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
	//	storeIntervalInt, err := strconv.Atoi(envStoreInterval)
	//	if err != nil {
	//		log.Logger.Info("Error convert STORE_INTERVAL to int:", zap.Error(err))
	//		return
	//	}
	//	storeInterval = storeIntervalInt
	//}
	//if envFilePath := os.Getenv("FILE_STORAGE_PATH"); envFilePath != "" {
	//	filePath = envFilePath
	//}
	//if envRestore := os.Getenv("RESTORE"); envRestore != "" {
	//	restoreValue, err := strconv.ParseBool(envRestore)
	//	if err != nil {
	//		log.Logger.Info("Error convert RESTORE to bool:", zap.Error(err))
	//		return
	//	}
	//	restore = restoreValue
	//}
	//if envHashKey := os.Getenv("KEY"); envHashKey != "" {
	//	hash = envHashKey
	//}
	//if envKeyPath := os.Getenv("CRYPTO_KEY"); envKeyPath != "" {
	//	keyPath = envKeyPath
	//}
	//
	//var addrDB string
	//
	//flag.StringVar(&addrDB, "d", "", "input addr db")
	//if envDBaddr := os.Getenv("DATABASE_DSN"); envDBaddr != "" {
	//	addrDB = envDBaddr
	//}

	if conf.AddrDB != "" {
		DBMemory := storage.DBStorage{}
		storage.OpenDB(conf.AddrDB)
		//handlers.StartServ(&DBMemory, addr, storeInterval, filePath, restore, hash, keyPath)
		handlers.StartServ(&DBMemory, conf.Addr, conf.StoreInterval, conf.FilePath, conf.Restore, conf.Hash, conf.KeyPath)
	} else {
		globalMemory := storage.MemStorage{}
		globalMemory.Counter = make(map[string]int64)
		globalMemory.Gauge = make(map[string]float64)
		//handlers.StartServ(&globalMemory, addr, storeInterval, filePath, restore, hash, keyPath)
		handlers.StartServ(&globalMemory, conf.Addr, conf.StoreInterval, conf.FilePath, conf.Restore, conf.Hash, conf.KeyPath)
	}
	defer log.Logger.Sync()
	defer storage.DB.Close()
}
