package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

var (
	masterKey     []byte
	encryptionErr error
)

func initEncryption() {
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		encryptionErr = errors.New("ENCRYPTION_KEY environment variable is not set")
		return
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		encryptionErr = errors.New("ENCRYPTION_KEY must be a valid hex string")
		return
	}
	if len(key) != 32 {
		encryptionErr = errors.New("ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
		return
	}
	masterKey = key
}

func encrypt(plaintext string) (string, error) {
	if encryptionErr != nil {
		return "", encryptionErr
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func decrypt(cipherHex string) (string, error) {
	if encryptionErr != nil {
		return "", encryptionErr
	}
	ciphertext, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
