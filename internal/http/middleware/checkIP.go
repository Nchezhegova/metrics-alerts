package middleware

import (
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
)

func CheckIP(trustedSubnet string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if trustedSubnet == "" {
			c.Next()
			return
		}
		ipStr := c.GetHeader("X-Real-IP")
		if ipStr == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "X-Real-IP header is required"})
			return
		}
		ip := net.ParseIP(ipStr)
		_, ipNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "error parsing trusted subnet"})
			return
		}
		if !ipNet.Contains(ip) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "IP is not in the trusted subnet"})
			return
		}
		c.Next()
		return
	}

}
