package domain

import "context"

// VaultRepository defines the interface for vault persistence
type VaultRepository interface {
	// Save persists vault metadata to disk
	Save(ctx context.Context, name string, metadata *VaultMetadata) error

	// Load retrieves vault metadata from disk
	Load(ctx context.Context, name string) (*VaultMetadata, error)

	// Exists checks if a vault exists
	Exists(ctx context.Context, name string) (bool, error)

	// List returns all available vault names
	List(ctx context.Context) ([]string, error)
}
