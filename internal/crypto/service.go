package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters - configured per OWASP recommendations
	// These parameters are intentionally resource-intensive to defend against
	// brute-force and dictionary attacks on the master password.
	//
	// OWASP Guidelines for Argon2id:
	// - Time (iterations): Minimum 3, recommended 3-4 for production
	// - Memory: Minimum 128 MB, recommended 128-256 MB for high security
	// - Threads: 4 is optimal for most systems
	// - Key length: 32 bytes (256 bits) for AES-256-GCM
	//
	// Performance impact: ~300-500ms on modern hardware (acceptable for password
	// manager authentication). Existing vaults remain compatible as each vault
	// stores its own derivation parameters.
	Argon2Time    = 3          // 3 iterations - balances security and performance
	Argon2Memory  = 128 * 1024 // 128 MB - OWASP minimum for high-value secrets
	Argon2Threads = 4          // Parallelism factor
	Argon2KeyLen  = 32         // 256 bits for AES-256-GCM

	// Salt length in bytes - 32 bytes provides 256 bits of entropy
	// Cryptographically secure random salt prevents rainbow table attacks
	SaltLength = 32

	// AES-GCM nonce size - 12 bytes is standard for GCM mode
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
