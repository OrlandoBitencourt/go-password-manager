package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters (can be tuned based on requirements)
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4
	Argon2KeyLen  = 32 // 256 bits for AES-256

	// Salt length in bytes
	SaltLength = 32

	// AES-GCM nonce size
	NonceSize = 12
)

// Service implements the CryptoService interface
type Service struct{}

// NewService creates a new crypto service instance
func NewService() *Service {
	return &Service{}
}

// DeriveKey derives an encryption key from a master password using Argon2id
func (s *Service) DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(password) == 0 {
		return nil, fmt.Errorf("password cannot be empty")
	}
	if len(salt) == 0 {
		return nil, fmt.Errorf("salt cannot be empty")
	}

	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLen,
	)

	return key, nil
}

// GenerateSalt creates a cryptographically secure random salt
func (s *Service) GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *Service) Encrypt(plaintext, key []byte) (nonce, ciphertext []byte, err error) {
	if len(key) != Argon2KeyLen {
		return nil, nil, fmt.Errorf("invalid key length: expected %d, got %d", Argon2KeyLen, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *Service) Decrypt(nonce, ciphertext, key []byte) ([]byte, error) {
	if len(key) != Argon2KeyLen {
		return nil, fmt.Errorf("invalid key length: expected %d, got %d", Argon2KeyLen, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: expected %d, got %d", gcm.NonceSize(), len(nonce))
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
