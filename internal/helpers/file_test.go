package helpers

import (
	"encoding/json"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestWriteFile(t *testing.T) {
	memStorage := storage.MemStorage{
		Gauge: map[string]float64{"test": 1.0},
	}
	filePath := "test.json"

	WriteFile(&memStorage, filePath)

	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer file.Close()

	var decodedData storage.MemStorage
	err = json.NewDecoder(file).Decode(&decodedData)
	assert.NoError(t, err)
	assert.Equal(t, memStorage, decodedData)
	os.Remove(filePath)
}

func TestSetWriterFile(t *testing.T) {
	memStorage := storage.MemStorage{
		Gauge: map[string]float64{"test": 1.0},
	}
	filePath := "test.json"
	restore := false
	storeInterval := 1
	go SetWriterFile(&memStorage, storeInterval, filePath, restore)
	time.Sleep(2 * time.Second)
	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer file.Close()
	var decodedData storage.MemStorage
	err = json.NewDecoder(file).Decode(&decodedData)
	assert.NoError(t, err)
	assert.Equal(t, memStorage, decodedData)

	os.Remove(filePath)
}
