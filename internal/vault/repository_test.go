package vault

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/orlan/go-password-manager/internal/domain"
)

func TestNewFileRepository(t *testing.T) {
	t.Run("creates repository with custom directory", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, err := NewFileRepository(tempDir)
		if err != nil {
			t.Fatalf("NewFileRepository() failed: %v", err)
		}
		if repo == nil {
			t.Fatal("NewFileRepository() returned nil")
		}
		if repo.vaultDir != tempDir {
			t.Errorf("expected vaultDir %q, got %q", tempDir, repo.vaultDir)
		}

		// Verify directory was created
		info, err := os.Stat(tempDir)
		if err != nil {
			t.Fatalf("vault directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("vault path is not a directory")
		}
	})

	t.Run("creates repository with default directory", func(t *testing.T) {
		// Clean up default directory before and after test
		defer os.RemoveAll(DefaultVaultDir)
		os.RemoveAll(DefaultVaultDir)

		repo, err := NewFileRepository("")
		if err != nil {
			t.Fatalf("NewFileRepository() failed: %v", err)
		}
		if repo == nil {
			t.Fatal("NewFileRepository() returned nil")
		}
		if repo.vaultDir != DefaultVaultDir {
			t.Errorf("expected vaultDir %q, got %q", DefaultVaultDir, repo.vaultDir)
		}

		// Verify directory was created
		_, err = os.Stat(DefaultVaultDir)
		if err != nil {
			t.Fatalf("default vault directory was not created: %v", err)
		}
	})

	t.Run("creates nested directory structure", func(t *testing.T) {
		tempDir := t.TempDir()
		nestedDir := filepath.Join(tempDir, "level1", "level2", "vaults")

		repo, err := NewFileRepository(nestedDir)
		if err != nil {
			t.Fatalf("NewFileRepository() failed: %v", err)
		}
		if repo == nil {
			t.Fatal("NewFileRepository() returned nil")
		}

		// Verify nested directory was created
		info, err := os.Stat(nestedDir)
		if err != nil {
			t.Fatalf("nested vault directory was not created: %v", err)
		}
		if !info.IsDir() {
			t.Error("vault path is not a directory")
		}
	})

	t.Run("verifies directory is created", func(t *testing.T) {
		tempDir := t.TempDir()
		vaultDir := filepath.Join(tempDir, "secure-vaults")

		repo, err := NewFileRepository(vaultDir)
		if err != nil {
			t.Fatalf("NewFileRepository() failed: %v", err)
		}
		if repo == nil {
			t.Fatal("NewFileRepository() returned nil")
		}

		info, err := os.Stat(vaultDir)
		if err != nil {
			t.Fatalf("failed to stat vault directory: %v", err)
		}

		if !info.IsDir() {
			t.Error("path is not a directory")
		}
	})
}

func TestSave(t *testing.T) {
	t.Run("saves vault metadata successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// Verify file exists
		filePath := filepath.Join(tempDir, "test-vault"+VaultExtension)
		_, err = os.Stat(filePath)
		if err != nil {
			t.Fatalf("vault file was not created: %v", err)
		}
	})

	t.Run("saves vault file successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		filePath := filepath.Join(tempDir, "test-vault"+VaultExtension)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("failed to stat vault file: %v", err)
		}

		if info.IsDir() {
			t.Error("vault file should not be a directory")
		}
	})

	t.Run("saves vault with correct JSON format", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		filePath := filepath.Join(tempDir, "test-vault"+VaultExtension)
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("failed to read vault file: %v", err)
		}

		var loaded domain.VaultMetadata
		err = json.Unmarshal(data, &loaded)
		if err != nil {
			t.Fatalf("failed to unmarshal vault file: %v", err)
		}

		if loaded.Version != metadata.Version {
			t.Errorf("expected version %q, got %q", metadata.Version, loaded.Version)
		}
	})

	t.Run("overwrites existing vault", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata1 := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", metadata1)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		metadata2 := &domain.VaultMetadata{
			Version:   "2.0",
			Salt:      []byte{13, 14, 15, 16},
			Nonce:     []byte{17, 18, 19, 20},
			Encrypted: []byte{21, 22, 23, 24},
		}

		err = repo.Save(ctx, "test-vault", metadata2)
		if err != nil {
			t.Fatalf("Save() failed on overwrite: %v", err)
		}

		loaded, err := repo.Load(ctx, "test-vault")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.Version != "2.0" {
			t.Errorf("vault was not overwritten: expected version %q, got %q", "2.0", loaded.Version)
		}
	})

	t.Run("saves vault with special characters in name", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		vaultName := "my-vault_2024"
		err := repo.Save(ctx, vaultName, metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		filePath := filepath.Join(tempDir, vaultName+VaultExtension)
		_, err = os.Stat(filePath)
		if err != nil {
			t.Fatalf("vault file was not created: %v", err)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("loads vault metadata successfully", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		expected := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", expected)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		loaded, err := repo.Load(ctx, "test-vault")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if loaded.Version != expected.Version {
			t.Errorf("expected version %q, got %q", expected.Version, loaded.Version)
		}
		if string(loaded.Salt) != string(expected.Salt) {
			t.Errorf("salt mismatch")
		}
		if string(loaded.Nonce) != string(expected.Nonce) {
			t.Errorf("nonce mismatch")
		}
		if string(loaded.Encrypted) != string(expected.Encrypted) {
			t.Errorf("encrypted data mismatch")
		}
	})

	t.Run("returns ErrVaultNotFound for non-existent vault", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		_, err := repo.Load(ctx, "non-existent-vault")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("returns error for corrupted JSON", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		// Create corrupted vault file
		filePath := filepath.Join(tempDir, "corrupted"+VaultExtension)
		err := os.WriteFile(filePath, []byte("not valid json {"), 0600)
		if err != nil {
			t.Fatalf("failed to create corrupted file: %v", err)
		}

		_, err = repo.Load(ctx, "corrupted")
		if err == nil {
			t.Error("Load() should return error for corrupted JSON")
		}
		if err == domain.ErrVaultNotFound {
			t.Error("Load() should not return ErrVaultNotFound for corrupted JSON")
		}
	})

	t.Run("loads vault with empty encrypted data", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{},
		}

		err := repo.Save(ctx, "empty-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		loaded, err := repo.Load(ctx, "empty-vault")
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		if len(loaded.Encrypted) != 0 {
			t.Errorf("expected empty encrypted data, got %d bytes", len(loaded.Encrypted))
		}
	})
}

func TestExists(t *testing.T) {
	t.Run("returns true for existing vault", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		err := repo.Save(ctx, "test-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		exists, err := repo.Exists(ctx, "test-vault")
		if err != nil {
			t.Fatalf("Exists() failed: %v", err)
		}
		if !exists {
			t.Error("Exists() returned false for existing vault")
		}
	})

	t.Run("returns false for non-existent vault", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		exists, err := repo.Exists(ctx, "non-existent-vault")
		if err != nil {
			t.Fatalf("Exists() failed: %v", err)
		}
		if exists {
			t.Error("Exists() returned true for non-existent vault")
		}
	})

	t.Run("handles vault with special characters", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		vaultName := "my-vault_2024"
		err := repo.Save(ctx, vaultName, metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		exists, err := repo.Exists(ctx, vaultName)
		if err != nil {
			t.Fatalf("Exists() failed: %v", err)
		}
		if !exists {
			t.Error("Exists() returned false for vault with special characters")
		}
	})
}

func TestList(t *testing.T) {
	t.Run("lists all vaults", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		vaultNames := []string{"vault1", "vault2", "vault3"}
		for _, name := range vaultNames {
			err := repo.Save(ctx, name, metadata)
			if err != nil {
				t.Fatalf("Save() failed: %v", err)
			}
		}

		vaults, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(vaults) != len(vaultNames) {
			t.Errorf("expected %d vaults, got %d", len(vaultNames), len(vaults))
		}

		for _, name := range vaultNames {
			found := false
			for _, v := range vaults {
				if v == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("vault %q not found in list", name)
			}
		}
	})

	t.Run("returns empty list when no vaults exist", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		vaults, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(vaults) != 0 {
			t.Errorf("expected empty list, got %d vaults", len(vaults))
		}
	})

	t.Run("ignores non-vault files", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		// Create a vault
		err := repo.Save(ctx, "real-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// Create non-vault files
		os.WriteFile(filepath.Join(tempDir, "readme.txt"), []byte("info"), 0600)
		os.WriteFile(filepath.Join(tempDir, "data.json"), []byte("{}"), 0600)
		os.WriteFile(filepath.Join(tempDir, ".hidden"), []byte("hidden"), 0600)

		vaults, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(vaults) != 1 {
			t.Errorf("expected 1 vault, got %d", len(vaults))
		}
		if len(vaults) > 0 && vaults[0] != "real-vault" {
			t.Errorf("expected vault name %q, got %q", "real-vault", vaults[0])
		}
	})

	t.Run("ignores directories", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		// Create a vault
		err := repo.Save(ctx, "real-vault", metadata)
		if err != nil {
			t.Fatalf("Save() failed: %v", err)
		}

		// Create a directory with .vault extension
		os.Mkdir(filepath.Join(tempDir, "fake-vault"+VaultExtension), 0700)

		vaults, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(vaults) != 1 {
			t.Errorf("expected 1 vault, got %d", len(vaults))
		}
	})

	t.Run("lists vaults with special characters", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)
		ctx := context.Background()

		metadata := &domain.VaultMetadata{
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte{9, 10, 11, 12},
		}

		vaultNames := []string{"my-vault", "vault_2024", "personal-vault"}
		for _, name := range vaultNames {
			err := repo.Save(ctx, name, metadata)
			if err != nil {
				t.Fatalf("Save() failed for %q: %v", name, err)
			}
		}

		vaults, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() failed: %v", err)
		}

		if len(vaults) != len(vaultNames) {
			t.Errorf("expected %d vaults, got %d", len(vaultNames), len(vaults))
		}
	})
}

func TestGetVaultPath(t *testing.T) {
	t.Run("constructs correct path", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)

		path := repo.getVaultPath("test-vault")
		expected := filepath.Join(tempDir, "test-vault"+VaultExtension)

		if path != expected {
			t.Errorf("expected path %q, got %q", expected, path)
		}
	})

	t.Run("handles vault name with special characters", func(t *testing.T) {
		tempDir := t.TempDir()
		repo, _ := NewFileRepository(tempDir)

		path := repo.getVaultPath("my-vault_2024")
		expected := filepath.Join(tempDir, "my-vault_2024"+VaultExtension)

		if path != expected {
			t.Errorf("expected path %q, got %q", expected, path)
		}
	})
}

func TestIntegrationSaveLoadFlow(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewFileRepository(tempDir)
	if err != nil {
		t.Fatalf("NewFileRepository() failed: %v", err)
	}
	ctx := context.Background()

	// Create multiple vaults
	vaults := map[string]*domain.VaultMetadata{
		"personal": {
			Version:   "1.0",
			Salt:      []byte{1, 2, 3, 4},
			Nonce:     []byte{5, 6, 7, 8},
			Encrypted: []byte("personal encrypted data"),
		},
		"work": {
			Version:   "1.0",
			Salt:      []byte{9, 10, 11, 12},
			Nonce:     []byte{13, 14, 15, 16},
			Encrypted: []byte("work encrypted data"),
		},
	}

	// Save all vaults
	for name, metadata := range vaults {
		err := repo.Save(ctx, name, metadata)
		if err != nil {
			t.Fatalf("Save() failed for %q: %v", name, err)
		}

		// Verify exists
		exists, err := repo.Exists(ctx, name)
		if err != nil {
			t.Fatalf("Exists() failed for %q: %v", name, err)
		}
		if !exists {
			t.Errorf("vault %q should exist after save", name)
		}
	}

	// List vaults
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(list) != len(vaults) {
		t.Errorf("expected %d vaults in list, got %d", len(vaults), len(list))
	}

	// Load and verify each vault
	for name, expected := range vaults {
		loaded, err := repo.Load(ctx, name)
		if err != nil {
			t.Fatalf("Load() failed for %q: %v", name, err)
		}

		if loaded.Version != expected.Version {
			t.Errorf("version mismatch for %q", name)
		}
		if string(loaded.Encrypted) != string(expected.Encrypted) {
			t.Errorf("encrypted data mismatch for %q", name)
		}
	}
}
