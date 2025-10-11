package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
)

var (
	key []byte // AES-256 key
)

func init() {
	// Load key from env or generate for dev
	envKey := os.Getenv("ENCRYPTION_KEY")
	if envKey != "" {
		key = []byte(envKey)
		// Ensure key is 32 bytes for AES-256
		if len(key) != 32 {
			panic("ENCRYPTION_KEY must be exactly 32 bytes")
		}
	} else {
		// Generate a random key for development
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic(fmt.Sprintf("failed to generate dev encryption key: %v", err))
		}
	}
}

// Encrypt encrypts data using AES-GCM.
func Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Nonce is prepended to ciphertext
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts AES-GCM encrypted data.
func Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
