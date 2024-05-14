package middleware_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"github.com/Nchezhegova/metrics-alerts/internal/helpers"
	"github.com/Nchezhegova/metrics-alerts/internal/http/middleware"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecryptBody(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating RSA key pair: %v", err)
	}

	r := gin.New()
	r.Use(middleware.DecryptBody(key))
	{
		r.POST("/test", func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.String(http.StatusInternalServerError, "error reading body")
				return
			}
			c.String(http.StatusOK, string(body))
		})
	}

	originalBody := []byte("Hello, world!")
	encryptedBody, err := helpers.EncryptData(originalBody, &key.PublicKey)
	if err != nil {
		t.Fatalf("error encrypting body: %v", err)
	}

	req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(encryptedBody))
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != string(originalBody) {
		t.Errorf("expected body %q; got %q", string(originalBody), w.Body.String())
	}
}

func TestDecryptBodyNoKey(t *testing.T) {
	r := gin.New()
	r.Use(middleware.DecryptBody(nil))
	{
		r.POST("/test", func(c *gin.Context) {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.String(http.StatusInternalServerError, "error reading body")
				return
			}
			c.String(http.StatusOK, string(body))
		})
	}

	originalBody := []byte("Hello, world!")

	req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(originalBody))
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != string(originalBody) {
		t.Errorf("expected body %q; got %q", string(originalBody), w.Body.String())
	}
}

func TestDecryptBodyErrorDecryptData(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("error generating RSA key pair: %v", err)
	}

	r := gin.New()
	r.Use(middleware.DecryptBody(key))
	{
		r.POST("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
	}

	encryptedBody := []byte("invalid data")

	req, err := http.NewRequest("POST", "/test", bytes.NewBuffer(encryptedBody))
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, w.Code)
	}
}
