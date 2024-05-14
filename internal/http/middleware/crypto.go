package middleware

import (
	"bytes"
	"crypto/rsa"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func DecryptBody(key *rsa.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		if key == nil {
			c.Next()
			return
		}
		encryptedBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error reading body"})
			c.Abort()
			return
		}
		decryptedBody, err := helpers.DecryptData(encryptedBody, key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error decrypting body"})
			c.Abort()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(decryptedBody))
		c.Next()
	}
}
