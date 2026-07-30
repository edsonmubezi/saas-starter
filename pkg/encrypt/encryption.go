package secureid

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func loadEnv() {
	// Try to load .env file, but don't fail if it doesn't exist
	// In production (Coolify, Docker, etc.), environment variables are injected directly
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found (this is OK in production environments)")
	}
}

func EncryptID(id string) (string, error) {
	secret := os.Getenv("SECRET_KEY")
	if len(secret) != 32 {
		return "", errors.New("SECRET_KEY must be exactly 32 bytes long")
	}
	secretKey := []byte(secret)

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(id), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptID(encrypted string) (string, error) {
	secret := os.Getenv("SECRET_KEY")
	if len(secret) != 32 {
		return "", errors.New("SECRET_KEY must be exactly 32 bytes long")
	}
	secretKey := []byte(secret)
	data, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
