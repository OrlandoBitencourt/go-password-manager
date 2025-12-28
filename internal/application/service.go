package application

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/orlan/go-password-manager/internal/domain"
)

// VaultService handles vault operations and session management
type VaultService struct {
	repo   domain.VaultRepository
	crypto domain.CryptoService

	// Session management
	sessions map[string]*session
	mu       sync.RWMutex
}

// session holds the decrypted vault and encryption key in memory
type session struct {
	vault *domain.Vault
	key   []byte
}

// NewVaultService creates a new vault service instance
func NewVaultService(repo domain.VaultRepository, crypto domain.CryptoService) *VaultService {
	return &VaultService{
		repo:     repo,
		crypto:   crypto,
		sessions: make(map[string]*session),
	}
}

// CreateVault creates a new encrypted vault
func (s *VaultService) CreateVault(ctx context.Context, name, masterPassword string) error {
	// Check if vault already exists
	exists, err := s.repo.Exists(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check vault existence: %w", err)
	}
	if exists {
		return domain.ErrVaultAlreadyExists
	}

	// Generate salt
	salt, err := s.crypto.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key
	key, err := s.crypto.DeriveKey(masterPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Create empty vault
	vault := &domain.Vault{
		Name:    name,
		Records: []domain.PasswordRecord{},
	}

	// Serialize vault
	vaultData, err := json.Marshal(vault)
	if err != nil {
		return fmt.Errorf("failed to marshal vault: %w", err)
	}

	// Encrypt vault
	nonce, ciphertext, err := s.crypto.Encrypt(vaultData, key)
	if err != nil {
		return domain.ErrEncryptionFailed
	}

	// Create metadata
	metadata := &domain.VaultMetadata{
		Version:   "1.0",
		Salt:      salt,
		Nonce:     nonce,
		Encrypted: ciphertext,
	}

	// Save to disk
	if err := s.repo.Save(ctx, name, metadata); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// UnlockVault authenticates and loads a vault into memory
func (s *VaultService) UnlockVault(ctx context.Context, name, masterPassword string) error {
	// Load vault metadata
	metadata, err := s.repo.Load(ctx, name)
	if err != nil {
		return err
	}

	// Derive key from master password
	key, err := s.crypto.DeriveKey(masterPassword, metadata.Salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Decrypt vault
	vaultData, err := s.crypto.Decrypt(metadata.Nonce, metadata.Encrypted, key)
	if err != nil {
		return domain.ErrInvalidMasterPassword
	}

	// Deserialize vault
	var vault domain.Vault
	if err := json.Unmarshal(vaultData, &vault); err != nil {
		return fmt.Errorf("failed to unmarshal vault: %w", err)
	}

	// Store session
	s.mu.Lock()
	s.sessions[name] = &session{
		vault: &vault,
		key:   key,
	}
	s.mu.Unlock()

	return nil
}

// LockVault removes the vault from memory
func (s *VaultService) LockVault(ctx context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[name]; !exists {
		return domain.ErrVaultNotFound
	}

	delete(s.sessions, name)
	return nil
}

// AddPasswordRecord adds a new password record to the vault
func (s *VaultService) AddPasswordRecord(ctx context.Context, vaultName, recordName, username, password string) error {
	s.mu.Lock()
	sess, exists := s.sessions[vaultName]
	s.mu.Unlock()

	if !exists {
		return domain.ErrVaultNotFound
	}

	// Check if record already exists
	for _, record := range sess.vault.Records {
		if record.Name == recordName {
			return domain.ErrRecordAlreadyExists
		}
	}

	// Create new record
	record := domain.PasswordRecord{
		ID:        uuid.New().String(),
		Name:      recordName,
		Username:  username,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add to vault
	s.mu.Lock()
	sess.vault.Records = append(sess.vault.Records, record)
	s.mu.Unlock()

	// Save to disk
	if err := s.saveVault(ctx, vaultName, sess); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// GetPasswordRecord retrieves a password record by name
func (s *VaultService) GetPasswordRecord(ctx context.Context, vaultName, recordName string) (*domain.PasswordRecord, error) {
	s.mu.RLock()
	sess, exists := s.sessions[vaultName]
	s.mu.RUnlock()

	if !exists {
		return nil, domain.ErrVaultNotFound
	}

	for _, record := range sess.vault.Records {
		if record.Name == recordName {
			return &record, nil
		}
	}

	return nil, domain.ErrRecordNotFound
}

// ListPasswordRecords returns all password records in the vault
func (s *VaultService) ListPasswordRecords(ctx context.Context, vaultName string) ([]domain.PasswordRecord, error) {
	s.mu.RLock()
	sess, exists := s.sessions[vaultName]
	s.mu.RUnlock()

	if !exists {
		return nil, domain.ErrVaultNotFound
	}

	return sess.vault.Records, nil
}

// UpdatePasswordRecord updates an existing password record
func (s *VaultService) UpdatePasswordRecord(ctx context.Context, vaultName, recordName, username, password string) error {
	s.mu.Lock()
	sess, exists := s.sessions[vaultName]
	s.mu.Unlock()

	if !exists {
		return domain.ErrVaultNotFound
	}

	// Find and update record
	found := false
	s.mu.Lock()
	for i := range sess.vault.Records {
		if sess.vault.Records[i].Name == recordName {
			if username != "" {
				sess.vault.Records[i].Username = username
			}
			if password != "" {
				sess.vault.Records[i].Password = password
			}
			sess.vault.Records[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}
	s.mu.Unlock()

	if !found {
		return domain.ErrRecordNotFound
	}

	// Save to disk
	if err := s.saveVault(ctx, vaultName, sess); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// DeletePasswordRecord removes a password record from the vault
func (s *VaultService) DeletePasswordRecord(ctx context.Context, vaultName, recordName string) error {
	s.mu.Lock()
	sess, exists := s.sessions[vaultName]
	s.mu.Unlock()

	if !exists {
		return domain.ErrVaultNotFound
	}

	// Find and delete record
	found := false
	s.mu.Lock()
	for i, record := range sess.vault.Records {
		if record.Name == recordName {
			sess.vault.Records = append(sess.vault.Records[:i], sess.vault.Records[i+1:]...)
			found = true
			break
		}
	}
	s.mu.Unlock()

	if !found {
		return domain.ErrRecordNotFound
	}

	// Save to disk
	if err := s.saveVault(ctx, vaultName, sess); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// ListVaults returns all available vault names
func (s *VaultService) ListVaults(ctx context.Context) ([]string, error) {
	return s.repo.List(ctx)
}

// IsVaultUnlocked checks if a vault is currently unlocked
func (s *VaultService) IsVaultUnlocked(ctx context.Context, vaultName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.sessions[vaultName]
	return exists
}

// saveVault encrypts and persists the vault to disk
func (s *VaultService) saveVault(ctx context.Context, name string, sess *session) error {
	// Serialize vault
	vaultData, err := json.Marshal(sess.vault)
	if err != nil {
		return fmt.Errorf("failed to marshal vault: %w", err)
	}

	// Encrypt vault
	nonce, ciphertext, err := s.crypto.Encrypt(vaultData, sess.key)
	if err != nil {
		return domain.ErrEncryptionFailed
	}

	// Load existing metadata to preserve salt
	metadata, err := s.repo.Load(ctx, name)
	if err != nil {
		return err
	}

	// Update encrypted data
	metadata.Nonce = nonce
	metadata.Encrypted = ciphertext

	// Save to disk
	return s.repo.Save(ctx, name, metadata)
}
