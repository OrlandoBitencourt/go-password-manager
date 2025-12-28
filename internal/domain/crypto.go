package domain

// CryptoService defines the interface for cryptographic operations
type CryptoService interface {
	// DeriveKey derives an encryption key from a master password and salt
	DeriveKey(password string, salt []byte) ([]byte, error)

	// GenerateSalt creates a random salt for key derivation
	GenerateSalt() ([]byte, error)

	// Encrypt encrypts plaintext using AES-256-GCM
	Encrypt(plaintext, key []byte) (nonce, ciphertext []byte, err error)

	// Decrypt decrypts ciphertext using AES-256-GCM
	Decrypt(nonce, ciphertext, key []byte) ([]byte, error)
}
