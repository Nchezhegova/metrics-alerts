package config

import (
	"os"
	"testing"
)

func TestSetConfigFromFlags(t *testing.T) {
	os.Args = []string{"cmd", "-a", "test_addr", "-i", "5", "-f", "/test/path", "-crypto_key", "/test/key", "-d", "test_db_dsn", "-k", "test_hash", "-c", "/test/config_file", "-p", "3", "-ri", "15", "-l", "10"}

	conf := NewConfig()

	conf.SetConfigFromFlags()

	if conf.Addr != "test_addr" {
		t.Errorf("expected Addr=test_addr, got %s", conf.Addr)
	}
	if conf.StoreInterval != 5 {
		t.Errorf("expected StoreInterval=5, got %d", conf.StoreInterval)
	}
	if conf.FilePath != "/test/path" {
		t.Errorf("expected FilePath=/test/path, got %s", conf.FilePath)
	}
	if conf.Restore != true {
		t.Errorf("expected Restore=false, got %t", conf.Restore)
	}
	if conf.KeyPath != "/test/key" {
		t.Errorf("expected KeyPath=/test/key, got %s", conf.KeyPath)
	}
	if conf.AddrDB != "test_db_dsn" {
		t.Errorf("expected AddrDB=test_db_dsn, got %s", conf.AddrDB)
	}
	if conf.Hash != "test_hash" {
		t.Errorf("expected Hash=test_hash, got %s", conf.Hash)
	}
	if conf.ConfigFile != "/test/config_file" {
		t.Errorf("expected ConfigFile=/test/config_file, got %s", conf.ConfigFile)
	}
	if conf.PollInterval != 3 {
		t.Errorf("expected PollInterval=3, got %d", conf.PollInterval)
	}
	if conf.ReportInterval != 15 {
		t.Errorf("expected ReportInterval=15, got %d", conf.ReportInterval)
	}
	if conf.RateLimit != 10 {
		t.Errorf("expected RateLimit=10, got %d", conf.RateLimit)
	}
}

func TestSetConfigFromEnv(t *testing.T) {
	os.Setenv("ADDRESS", "test_addr")
	os.Setenv("STORE_INTERVAL", "5")
	os.Setenv("FILE_STORAGE_PATH", "/test/path")
	os.Setenv("RESTORE", "false")
	os.Setenv("CRYPTO_KEY", "/test/key")
	os.Setenv("DATABASE_DSN", "test_db_dsn")
	os.Setenv("HASH", "test_hash")
	os.Setenv("CONFIG", "/test/config_file")
	os.Setenv("POLL_INTERVAL", "3")
	os.Setenv("REPORT_INTERVAL", "15")
	os.Setenv("RATE_LIMIT", "10")

	conf := NewConfig()

	conf.SetConfigFromEnv()

	if conf.Addr != "test_addr" {
		t.Errorf("expected Addr=test_addr, got %s", conf.Addr)
	}
	if conf.StoreInterval != 5 {
		t.Errorf("expected StoreInterval=5, got %d", conf.StoreInterval)
	}
	if conf.FilePath != "/test/path" {
		t.Errorf("expected FilePath=/test/path, got %s", conf.FilePath)
	}
	if conf.Restore != false {
		t.Errorf("expected Restore=false, got %t", conf.Restore)
	}
	if conf.KeyPath != "/test/key" {
		t.Errorf("expected KeyPath=/test/key, got %s", conf.KeyPath)
	}
	if conf.AddrDB != "test_db_dsn" {
		t.Errorf("expected AddrDB=test_db_dsn, got %s", conf.AddrDB)
	}
	if conf.Hash != "test_hash" {
		t.Errorf("expected Hash=test_hash, got %s", conf.Hash)
	}
	if conf.ConfigFile != "/test/config_file" {
		t.Errorf("expected ConfigFile=/test/config_file, got %s", conf.ConfigFile)
	}
	if conf.PollInterval != 3 {
		t.Errorf("expected PollInterval=3, got %d", conf.PollInterval)
	}
	if conf.ReportInterval != 15 {
		t.Errorf("expected ReportInterval=15, got %d", conf.ReportInterval)
	}
	if conf.RateLimit != 10 {
		t.Errorf("expected RateLimit=10, got %d", conf.RateLimit)
	}
}

func TestSetConfigFromJSON(t *testing.T) {
	tempFile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatalf("failed to create temporary config file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	_, err = tempFile.WriteString(`{
		"address": "test_addr",
		"store_interval": 5,
		"file_storage_path": "/test/path",
		"restore": false,
		"crypto_key": "/test/key",
		"database_dsn": "test_db_dsn",
		"hash": "test_hash",
		"poll_interval": 3,
		"report_interval": 15,
		"rate_limit": 10
	}`)
	if err != nil {
		t.Fatalf("failed to write to temporary config file: %v", err)
	}

	conf := NewConfig()

	conf.ConfigFile = tempFile.Name()
	err = conf.SetConfigFromJSON()
	if err != nil {
		t.Fatalf("error setting config from JSON: %v", err)
	}

	if conf.Addr != "localhost:8080" {
		t.Errorf("expected Addr=localhost:8080, got %s", conf.Addr)
	}
	if conf.StoreInterval != 5 {
		t.Errorf("expected StoreInterval=5, got %d", conf.StoreInterval)
	}
	if conf.FilePath != "/tmp/metrics-db.json" {
		t.Errorf("expected FilePath=/tmp/metrics-db.json, got %s", conf.FilePath)
	}
	if conf.Restore != false {
		t.Errorf("expected Restore=false, got %t", conf.Restore)
	}
	if conf.KeyPath != "/test/key" {
		t.Errorf("expected KeyPath=/test/key, got %s", conf.KeyPath)
	}
	if conf.AddrDB != "test_db_dsn" {
		t.Errorf("expected AddrDB=test_db_dsn, got %s", conf.AddrDB)
	}
	if conf.Hash != "test_hash" {
		t.Errorf("expected Hash=test_hash, got %s", conf.Hash)
	}
	if conf.PollInterval != 2 {
		t.Errorf("expected PollInterval=3, got %d", conf.PollInterval)
	}
	if conf.ReportInterval != 10 {
		t.Errorf("expected ReportInterval=15, got %d", conf.ReportInterval)
	}
	if conf.RateLimit != 5 {
		t.Errorf("expected RateLimit=10, got %d", conf.RateLimit)
	}
}
