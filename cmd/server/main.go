package main

import (
	"github.com/Nchezhegova/metrics-alerts/cmd/grpcProtocol"
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

	printBuildInfo()
	conf := config.NewConfig()
	conf.SetConfigFromFlags()
	conf.SetConfigFromEnv()
	err := conf.SetConfigFromJSON()
	if err != nil {
		log.Logger.Info("Error loading configuration:", zap.Error(err))
		return
	}

	if conf.AddrDB != "" {
		DBMemory := storage.DBStorage{}
		storage.OpenDB(conf.AddrDB)
		handlers.StartServ(&DBMemory, conf.Addr, conf.StoreInterval, conf.FilePath, conf.Restore, conf.Hash, conf.KeyPath, conf.TrustedSubnet)
		grpcProtocol.StartGRPCServer(&DBMemory, conf.TrustedSubnet)
		defer storage.DB.Close()
	} else {
		globalMemory := storage.MemStorage{}
		globalMemory.Counter = make(map[string]int64)
		globalMemory.Gauge = make(map[string]float64)
		grpcProtocol.StartGRPCServer(&globalMemory, conf.TrustedSubnet)
		handlers.StartServ(&globalMemory, conf.Addr, conf.StoreInterval, conf.FilePath, conf.Restore, conf.Hash, conf.KeyPath, conf.TrustedSubnet)
	}
	defer log.Logger.Sync()
}
