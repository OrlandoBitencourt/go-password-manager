package vault

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/orlan/go-password-manager/internal/domain"
)

const (
	VaultExtension = ".vault"
	DefaultVaultDir = "./vaults"
)

// FileRepository implements VaultRepository using the filesystem
type FileRepository struct {
	vaultDir string
}

// NewFileRepository creates a new file-based vault repository
func NewFileRepository(vaultDir string) (*FileRepository, error) {
	if vaultDir == "" {
		vaultDir = DefaultVaultDir
	}

	// Create vault directory if it doesn't exist
	if err := os.MkdirAll(vaultDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create vault directory: %w", err)
	}

	return &FileRepository{
		vaultDir: vaultDir,
	}, nil
}

// Save persists vault metadata to disk
func (r *FileRepository) Save(ctx context.Context, name string, metadata *domain.VaultMetadata) error {
	filePath := r.getVaultPath(name)

	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal vault metadata: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write vault file: %w", err)
	}

	return nil
}

// Load retrieves vault metadata from disk
func (r *FileRepository) Load(ctx context.Context, name string) (*domain.VaultMetadata, error) {
	filePath := r.getVaultPath(name)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to read vault file: %w", err)
	}

	var metadata domain.VaultMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vault metadata: %w", err)
	}

	return &metadata, nil
}

// Exists checks if a vault exists
func (r *FileRepository) Exists(ctx context.Context, name string) (bool, error) {
	filePath := r.getVaultPath(name)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check vault existence: %w", err)
	}
	return true, nil
}

// List returns all available vault names
func (r *FileRepository) List(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(r.vaultDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault directory: %w", err)
	}

	var vaults []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), VaultExtension) {
			name := strings.TrimSuffix(entry.Name(), VaultExtension)
			vaults = append(vaults, name)
		}
	}

	return vaults, nil
}

// getVaultPath constructs the full path to a vault file
func (r *FileRepository) getVaultPath(name string) string {
	return filepath.Join(r.vaultDir, name+VaultExtension)
}
