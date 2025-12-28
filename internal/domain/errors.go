package domain

import "errors"

var (
	// ErrVaultNotFound indicates the requested vault does not exist
	ErrVaultNotFound = errors.New("vault not found")

	// ErrVaultAlreadyExists indicates a vault with the given name already exists
	ErrVaultAlreadyExists = errors.New("vault already exists")

	// ErrInvalidMasterPassword indicates authentication failed
	ErrInvalidMasterPassword = errors.New("invalid master password")

	// ErrRecordNotFound indicates the requested password record does not exist
	ErrRecordNotFound = errors.New("password record not found")

	// ErrRecordAlreadyExists indicates a record with the given name already exists
	ErrRecordAlreadyExists = errors.New("password record already exists")

	// ErrEncryptionFailed indicates encryption operation failed
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrDecryptionFailed indicates decryption operation failed
	ErrDecryptionFailed = errors.New("decryption failed")
)
