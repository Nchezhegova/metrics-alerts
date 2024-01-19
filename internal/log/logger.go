package log

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

var Logger *zap.Logger

func init() {
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("Не удалось создать логгер: %v", err))
	}

}

func GinLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		statusCode := c.Writer.Status()
		size := c.Writer.Size()

		logger.Info(
			"HTTP Request",
			zap.String("method", c.Request.Method),
			zap.Duration("duration", duration),
			zap.String("URI", c.Request.RequestURI),
			zap.Int("Response status", statusCode),
			zap.Int("Response size", size),
		)
	}
}

//func InitLogger() *zap.Logger {
//	logger, err := zap.NewProduction()
//
//	//перетащить в мейн
//	defer logger.Sync()
//	return logger
//}
