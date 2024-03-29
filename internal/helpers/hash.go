package helpers

import (
	"crypto/hmac"
	"crypto/sha256"
)

func CalculateHash(body []byte, key string) []byte {
	hmacKey := []byte(key)
	h := hmac.New(sha256.New, hmacKey)
	h.Write(body)
	return h.Sum(nil)
}

//
//func SetHashHeader(metricsByte []byte) string {
//
//}
