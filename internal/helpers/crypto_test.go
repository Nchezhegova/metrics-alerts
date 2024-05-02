package helpers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"testing"
)

func TestEncryptionAndDecryption(t *testing.T) {
	keySize := 2048
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		t.Fatalf("Error generating RSA key pair: %v", err)
	}

	publicKey := &privateKey.PublicKey

	originalData := []byte("Hello, world!")

	encryptedData, err := helpers.EncryptData(originalData, publicKey)
	if err != nil {
		t.Fatalf("Error encrypting data: %v", err)
	}

	decryptedData, err := helpers.DecryptData(encryptedData, privateKey)
	if err != nil {
		t.Fatalf("Error decrypting data: %v", err)
	}

	if string(decryptedData) != string(originalData) {
		t.Fatalf("Decrypted data doesn't match original data. Expected: %s, Got: %s", originalData, decryptedData)
	}
}
