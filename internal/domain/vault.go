package domain

import "time"

// PasswordRecord represents a single password entry in the vault
type PasswordRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Vault represents the encrypted vault structure
type Vault struct {
	Name    string           `json:"name"`
	Records []PasswordRecord `json:"records"`
}

// VaultMetadata contains unencrypted vault information
type VaultMetadata struct {
	Version   string `json:"version"`
	Salt      []byte `json:"salt"`
	Nonce     []byte `json:"nonce"`
	Encrypted []byte `json:"encrypted"`
}
