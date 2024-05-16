package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
)

const Counter = "counter"
const Gauge = "gauge"

const DATEBASE = "postgres://user:password@localhost/metrics"
const MaxRetries = 3

const IP = "127.0.0.1"

type Config struct {
	Addr          string `json:"address"`
	StoreInterval int    `json:"store_interval"`
	FilePath      string `json:"file_storage_path"`
	Restore       bool   `json:"restore"`
	KeyPath       string `json:"crypto_key"`
	AddrDB        string `json:"database_dsn"`
	Hash          string `json:"hash"`
	ConfigFile    string `json:"config_file"`
	TrustedSubnet string `json:"trusted_subnet"`
	//agent's config
	PollInterval   int `json:"poll_interval"`
	ReportInterval int `json:"report_interval"`
	RateLimit      int `json:"rate_limit"`
}

// NewConfig returns a new Config with default values
func NewConfig() *Config {
	return &Config{
		Addr:           "localhost:8080",
		StoreInterval:  0,
		FilePath:       "/tmp/metrics-db.json",
		Restore:        true,
		KeyPath:        "",
		AddrDB:         "",
		Hash:           "",
		ConfigFile:     "",
		PollInterval:   2,
		ReportInterval: 10,
		RateLimit:      5,
		TrustedSubnet:  "127.0.0.1/32",
	}
}

// SetConfigFromFlags sets the Config fields from the command line flags
func (c *Config) SetConfigFromFlags() {
	flag.StringVar(&c.Addr, "a", c.Addr, "Address to listen on")
	flag.IntVar(&c.StoreInterval, "i", c.StoreInterval, "Interval to store metrics")
	flag.StringVar(&c.FilePath, "f", c.FilePath, "Path to store metrics")
	flag.BoolVar(&c.Restore, "r", c.Restore, "Restore metrics from file")
	flag.StringVar(&c.KeyPath, "crypto_key", c.KeyPath, "Path to key for encryption")
	flag.StringVar(&c.AddrDB, "d", c.AddrDB, "Database DSN")
	flag.StringVar(&c.Hash, "k", c.Hash, "Hash for password")
	flag.StringVar(&c.ConfigFile, "c", c.ConfigFile, "Path to config file")
	flag.IntVar(&c.PollInterval, "p", c.PollInterval, "Poll interval")
	//в задании у ReportInterval флаг -r, но тогда пересекалось бы с restore
	flag.IntVar(&c.ReportInterval, "ri", c.ReportInterval, "Report interval")
	flag.IntVar(&c.RateLimit, "l", c.RateLimit, "Rate limit")
	flag.StringVar(&c.TrustedSubnet, "t", c.TrustedSubnet, "Trusted subnet")
	flag.Parse()
}

// SetConfigFromEnv sets the Config fields from the environment variables
func (c *Config) SetConfigFromEnv() {
	if addr := os.Getenv("ADDRESS"); addr != "" {
		c.Addr = addr
	}
	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		storeIntervalInt, err := strconv.Atoi(storeInterval)
		if err != nil {
			return
		}
		c.StoreInterval = storeIntervalInt
	}
	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		c.FilePath = filePath
	}
	if restore := os.Getenv("RESTORE"); restore != "" {
		restoreValue, err := strconv.ParseBool(restore)
		if err != nil {
			return
		}
		c.Restore = restoreValue
	}
	if keyPath := os.Getenv("CRYPTO_KEY"); keyPath != "" {
		c.KeyPath = keyPath
	}
	if addrDB := os.Getenv("DATABASE_DSN"); addrDB != "" {
		c.AddrDB = addrDB
	}
	if hash := os.Getenv("HASH"); hash != "" {
		c.Hash = hash
	}
	if configFile := os.Getenv("CONFIG"); configFile != "" {
		c.ConfigFile = configFile
	}
	if pollInterval := os.Getenv("POLL_INTERVAL"); pollInterval != "" {
		pollIntervalInt, err := strconv.Atoi(pollInterval)
		if err != nil {
			return
		}
		c.PollInterval = pollIntervalInt
	}
	if reportInterval := os.Getenv("REPORT_INTERVAL"); reportInterval != "" {
		reportIntervalInt, err := strconv.Atoi(reportInterval)
		if err != nil {
			return
		}
		c.ReportInterval = reportIntervalInt

	}
	if rateLimit := os.Getenv("RATE_LIMIT"); rateLimit != "" {
		rateLimitInt, err := strconv.Atoi(rateLimit)
		if err != nil {
			return
		}
		c.RateLimit = rateLimitInt
	}
	if trustedSubnet := os.Getenv("TRUSTED_SUBNET"); trustedSubnet != "" {
		c.TrustedSubnet = trustedSubnet
	}
}

// SetConfigFromJSON sets the Config fields from the JSON file
func (c *Config) SetConfigFromJSON() error {
	if c.ConfigFile == "" {
		return nil
	}
	file, err := os.OpenFile(c.ConfigFile, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	var config Config
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		return err
	}
	if c.Addr == "" {
		c.Addr = config.Addr
	}
	if c.StoreInterval == 0 {
		c.StoreInterval = config.StoreInterval
	}
	if c.FilePath == "" {
		c.FilePath = config.FilePath
	}
	if c.Restore {
		c.Restore = config.Restore
	}
	if c.KeyPath == "" {
		c.KeyPath = config.KeyPath
	}
	if c.AddrDB == "" {
		c.AddrDB = config.AddrDB
	}
	if c.Hash == "" {
		c.Hash = config.Hash
	}
	if c.PollInterval == 0 {
		c.PollInterval = config.PollInterval
	}
	if c.ReportInterval == 0 {
		c.ReportInterval = config.ReportInterval
	}
	if c.RateLimit == 0 {
		c.RateLimit = config.RateLimit
	}
	if c.TrustedSubnet == "" {
		c.TrustedSubnet = config.TrustedSubnet
	}
	return nil
}
