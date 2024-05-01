package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/Nchezhegova/metrics-alerts/internal/log"
	"go.uber.org/zap"
	"os"
)

func ConvertPublicKey(key string) (*rsa.PublicKey, error) {
	keyFile, err := os.ReadFile(key)
	if err != nil {
		log.Logger.Error("Error reading key file", zap.Error(err))
		os.Exit(1)
	}
	// Parse key
	block, _ := pem.Decode(keyFile)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Logger.Error("Invalid public key", zap.String("type", block.Type))
		os.Exit(1)
	}
	// Parse public key
	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Logger.Error("Error parsing public key", zap.Error(err))
		os.Exit(1)
	}
	// Convert to RSA public key
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		log.Logger.Error("Error converting to RSA public key")
		os.Exit(1)
	}
	return rsaPublicKey, nil
}

func ConvertPrivateKey(key string) (*rsa.PrivateKey, error) {
	// Read key from file
	keyFile, err := os.ReadFile(key)
	if err != nil {
		log.Logger.Error("Error reading key file", zap.Error(err))
		os.Exit(1)
	}

	// Parse key
	block, _ := pem.Decode(keyFile)
	if block == nil || block.Type != "PRIVATE KEY" {
		log.Logger.Error("Invalid private key", zap.String("type", block.Type))
		os.Exit(1)
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		log.Logger.Error("Error parsing private key", zap.Error(err))
		os.Exit(1)
	}

	return privateKey, nil
}

func EncryptData(data []byte, publicKey *rsa.PublicKey) ([]byte, error) {
	// Encrypt data
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, data)
	if err != nil {
		return nil, err
	}
	return encryptedData, nil
}

func DecryptData(encryptedData []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	// Decrypt data
	data, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, encryptedData)
	if err != nil {
		return nil, err
	}
	return data, nil
}
