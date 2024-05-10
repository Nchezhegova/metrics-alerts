package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"go.uber.org/zap"
	"os"
)

func ConvertPublicKey(key string) (*rsa.PublicKey, error) {
	keyFile, err := os.ReadFile(key)
	if err != nil {
		log.Logger.Error("Error reading key file", zap.Error(err))
		return nil, err
	}
	block, _ := pem.Decode(keyFile)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Logger.Error("Invalid public key", zap.String("type", block.Type))
		return nil, fmt.Errorf("invalid public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Logger.Error("Error parsing public key", zap.Error(err))
		return nil, err
	}
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		log.Logger.Error("Error converting to RSA public key")
		return nil, err
	}
	return rsaPublicKey, nil
}

func ConvertPrivateKey(key string) (*rsa.PrivateKey, error) {
	keyFile, err := os.ReadFile(key)
	if err != nil {
		log.Logger.Error("Error reading key file", zap.Error(err))
		return nil, err
	}

	block, _ := pem.Decode(keyFile)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Logger.Error("Invalid private key", zap.String("type", block.Type))
		return nil, fmt.Errorf("invalid private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Logger.Error("Error parsing private key", zap.Error(err))
		return nil, err
	}

	return privateKey, nil
}

func EncryptData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}

func DecryptData(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	data, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, err
	}
	return data, nil
}
