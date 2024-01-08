package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/storage"
	"os"
	"time"
)

func WriteFile(m storage.MStorage, filePath string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Error open file:", err)
		return
	}
	defer file.Close()
	data, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error convert metrics to str:", err)
		return
	}
	_, err = file.Write(data)
	if err != nil {
		fmt.Println("Error write file:", err)
		return
	}
}

func readFile(m storage.MStorage, filePath string) {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Error open file:", err)
		return
	}
	defer file.Close()
	var data storage.MemStorage
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decode data:", err)
		return
	}
	m.SetStartData(data)
}

func SetWriterFile(m storage.MStorage, storeInterval int, filePath string, restore bool) bool {
	if filePath == "" {
		return false
	}

	if restore {
		readFile(m, filePath)
	}
	if storeInterval == 0 {
		return true
	}

	storeIntervalSecond := time.Duration(storeInterval) * time.Second
	go func() {
		for {
			WriteFile(m, filePath)
			time.Sleep(storeIntervalSecond)
		}

	}()

	return false
}
