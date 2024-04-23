package helpers

import (
	"bytes"
	"compress/gzip"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"go.uber.org/zap"
)

func CompressResp(metricsByte []byte) bytes.Buffer {
	var compressBody bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressBody)
	_, err := gzipWriter.Write(metricsByte)
	if err != nil {
		log.Logger.Info("Error convert to gzip.Writer:", zap.Error(err))
		return *bytes.NewBuffer(nil)
	}

	err = gzipWriter.Close()
	if err != nil {
		log.Logger.Info("Error closing compressed:", zap.Error(err))
		return *bytes.NewBuffer(nil)
	}
	return compressBody
}
