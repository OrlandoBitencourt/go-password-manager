package application

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/orlan/go-password-manager/internal/crypto"
	"github.com/orlan/go-password-manager/internal/domain"
	"github.com/orlan/go-password-manager/internal/vault"
)

func setupTestService(t *testing.T) (*VaultService, string) {
	t.Helper()
	vaultDir := t.TempDir()
	repo, err := vault.NewFileRepository(vaultDir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	cryptoSvc := crypto.NewService()
	service := NewVaultService(repo, cryptoSvc)
	return service, vaultDir
}

func TestNewVaultService(t *testing.T) {
	service, _ := setupTestService(t)

	if service == nil {
		t.Fatal("NewVaultService() returned nil")
	}
	if service.repo == nil {
		t.Error("service repository is nil")
	}
	if service.crypto == nil {
		t.Error("service crypto is nil")
	}
	if service.sessions == nil {
		t.Error("service sessions map is nil")
	}
}

func TestCreateVault(t *testing.T) {
	t.Run("creates vault successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		// Verify vault exists
		vaults, err := service.ListVaults(ctx)
		if err != nil {
			t.Fatalf("ListVaults() failed: %v", err)
		}
		if len(vaults) != 1 {
			t.Errorf("expected 1 vault, got %d", len(vaults))
		}
		if len(vaults) > 0 && vaults[0] != "test-vault" {
			t.Errorf("expected vault name %q, got %q", "test-vault", vaults[0])
		}
	})

	t.Run("returns error for duplicate vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "password1")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.CreateVault(ctx, "test-vault", "password2")
		if err != domain.ErrVaultAlreadyExists {
			t.Errorf("expected ErrVaultAlreadyExists, got %v", err)
		}
	})

	t.Run("creates multiple vaults", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		vaults := []string{"vault1", "vault2", "vault3"}
		for _, name := range vaults {
			err := service.CreateVault(ctx, name, "password")
			if err != nil {
				t.Fatalf("CreateVault() failed for %q: %v", name, err)
			}
		}

		list, err := service.ListVaults(ctx)
		if err != nil {
			t.Fatalf("ListVaults() failed: %v", err)
		}
		if len(list) != len(vaults) {
			t.Errorf("expected %d vaults, got %d", len(vaults), len(list))
		}
	})

	t.Run("creates vault with empty password", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		// Should fail due to crypto service validation
		err := service.CreateVault(ctx, "test-vault", "")
		if err == nil {
			t.Error("CreateVault() should fail with empty password")
		}
	})
}

func TestUnlockVault(t *testing.T) {
	t.Run("unlocks vault with correct password", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		// Verify vault is unlocked
		if !service.IsVaultUnlocked(ctx, "test-vault") {
			t.Error("vault should be unlocked")
		}
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "correct-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "wrong-password")
		if err != domain.ErrInvalidMasterPassword {
			t.Errorf("expected ErrInvalidMasterPassword, got %v", err)
		}

		// Verify vault is not unlocked
		if service.IsVaultUnlocked(ctx, "test-vault") {
			t.Error("vault should not be unlocked")
		}
	})

	t.Run("returns error for non-existent vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.UnlockVault(ctx, "non-existent", "password")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("unlocks multiple vaults independently", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "vault1", "password1")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.CreateVault(ctx, "vault2", "password2")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "vault1", "password1")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "vault2", "password2")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		if !service.IsVaultUnlocked(ctx, "vault1") {
			t.Error("vault1 should be unlocked")
		}
		if !service.IsVaultUnlocked(ctx, "vault2") {
			t.Error("vault2 should be unlocked")
		}
	})

	t.Run("re-unlocking already unlocked vault succeeds", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("first UnlockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("second UnlockVault() failed: %v", err)
		}
	})
}

func TestLockVault(t *testing.T) {
	t.Run("locks unlocked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.LockVault(ctx, "test-vault")
		if err != nil {
			t.Fatalf("LockVault() failed: %v", err)
		}

		if service.IsVaultUnlocked(ctx, "test-vault") {
			t.Error("vault should be locked")
		}
	})

	t.Run("returns error for non-existent vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.LockVault(ctx, "non-existent")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("locks specific vault without affecting others", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "vault1", "password1")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.CreateVault(ctx, "vault2", "password2")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "vault1", "password1")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "vault2", "password2")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.LockVault(ctx, "vault1")
		if err != nil {
			t.Fatalf("LockVault() failed: %v", err)
		}

		if service.IsVaultUnlocked(ctx, "vault1") {
			t.Error("vault1 should be locked")
		}
		if !service.IsVaultUnlocked(ctx, "vault2") {
			t.Error("vault2 should still be unlocked")
		}
	})
}

func TestAddPasswordRecord(t *testing.T) {
	t.Run("adds password record successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("expected 1 record, got %d", len(records))
		}
		if len(records) > 0 {
			if records[0].Name != "gmail" {
				t.Errorf("expected name %q, got %q", "gmail", records[0].Name)
			}
			if records[0].Username != "user@gmail.com" {
				t.Errorf("expected username %q, got %q", "user@gmail.com", records[0].Username)
			}
			if records[0].Password != "secret123" {
				t.Errorf("expected password %q, got %q", "secret123", records[0].Password)
			}
			if records[0].ID == "" {
				t.Error("record ID should not be empty")
			}
			if records[0].CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
			if records[0].UpdatedAt.IsZero() {
				t.Error("UpdatedAt should not be zero")
			}
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("returns error for duplicate record name", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user1@gmail.com", "pass1")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user2@gmail.com", "pass2")
		if err != domain.ErrRecordAlreadyExists {
			t.Errorf("expected ErrRecordAlreadyExists, got %v", err)
		}
	})

	t.Run("adds multiple records", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		records := []struct {
			name, username, password string
		}{
			{"gmail", "user@gmail.com", "pass1"},
			{"github", "user@github.com", "pass2"},
			{"twitter", "user@twitter.com", "pass3"},
		}

		for _, r := range records {
			err = service.AddPasswordRecord(ctx, "test-vault", r.name, r.username, r.password)
			if err != nil {
				t.Fatalf("AddPasswordRecord() failed for %q: %v", r.name, err)
			}
		}

		list, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(list) != len(records) {
			t.Errorf("expected %d records, got %d", len(records), len(list))
		}
	})

	t.Run("persists record after unlock/lock cycle", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.LockVault(ctx, "test-vault")
		if err != nil {
			t.Fatalf("LockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("expected 1 record after unlock, got %d", len(records))
		}
	})
}

func TestGetPasswordRecord(t *testing.T) {
	t.Run("retrieves password record successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		record, err := service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("GetPasswordRecord() failed: %v", err)
		}

		if record.Name != "gmail" {
			t.Errorf("expected name %q, got %q", "gmail", record.Name)
		}
		if record.Username != "user@gmail.com" {
			t.Errorf("expected username %q, got %q", "user@gmail.com", record.Username)
		}
		if record.Password != "secret123" {
			t.Errorf("expected password %q, got %q", "secret123", record.Password)
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		_, err = service.GetPasswordRecord(ctx, "test-vault", "non-existent")
		if err != domain.ErrRecordNotFound {
			t.Errorf("expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		_, err = service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})
}

func TestUpdatePasswordRecord(t *testing.T) {
	t.Run("updates password successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "old@gmail.com", "oldpass")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond) // Ensure UpdatedAt is different

		err = service.UpdatePasswordRecord(ctx, "test-vault", "gmail", "", "newpass")
		if err != nil {
			t.Fatalf("UpdatePasswordRecord() failed: %v", err)
		}

		record, err := service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("GetPasswordRecord() failed: %v", err)
		}

		if record.Password != "newpass" {
			t.Errorf("expected password %q, got %q", "newpass", record.Password)
		}
		if record.Username != "old@gmail.com" {
			t.Errorf("username should not change, got %q", record.Username)
		}
		if record.UpdatedAt.Equal(record.CreatedAt) {
			t.Error("UpdatedAt should be different from CreatedAt")
		}
	})

	t.Run("updates username successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "old@gmail.com", "password")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.UpdatePasswordRecord(ctx, "test-vault", "gmail", "new@gmail.com", "")
		if err != nil {
			t.Fatalf("UpdatePasswordRecord() failed: %v", err)
		}

		record, err := service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("GetPasswordRecord() failed: %v", err)
		}

		if record.Username != "new@gmail.com" {
			t.Errorf("expected username %q, got %q", "new@gmail.com", record.Username)
		}
		if record.Password != "password" {
			t.Errorf("password should not change, got %q", record.Password)
		}
	})

	t.Run("updates both username and password", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "old@gmail.com", "oldpass")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.UpdatePasswordRecord(ctx, "test-vault", "gmail", "new@gmail.com", "newpass")
		if err != nil {
			t.Fatalf("UpdatePasswordRecord() failed: %v", err)
		}

		record, err := service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("GetPasswordRecord() failed: %v", err)
		}

		if record.Username != "new@gmail.com" {
			t.Errorf("expected username %q, got %q", "new@gmail.com", record.Username)
		}
		if record.Password != "newpass" {
			t.Errorf("expected password %q, got %q", "newpass", record.Password)
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.UpdatePasswordRecord(ctx, "test-vault", "non-existent", "user", "pass")
		if err != domain.ErrRecordNotFound {
			t.Errorf("expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UpdatePasswordRecord(ctx, "test-vault", "gmail", "user", "pass")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("persists update after unlock/lock cycle", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "old@gmail.com", "oldpass")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.UpdatePasswordRecord(ctx, "test-vault", "gmail", "new@gmail.com", "newpass")
		if err != nil {
			t.Fatalf("UpdatePasswordRecord() failed: %v", err)
		}

		err = service.LockVault(ctx, "test-vault")
		if err != nil {
			t.Fatalf("LockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		record, err := service.GetPasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("GetPasswordRecord() failed: %v", err)
		}

		if record.Username != "new@gmail.com" || record.Password != "newpass" {
			t.Error("update was not persisted")
		}
	})
}

func TestDeletePasswordRecord(t *testing.T) {
	t.Run("deletes password record successfully", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.DeletePasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("DeletePasswordRecord() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 0 {
			t.Errorf("expected 0 records after delete, got %d", len(records))
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.DeletePasswordRecord(ctx, "test-vault", "non-existent")
		if err != domain.ErrRecordNotFound {
			t.Errorf("expected ErrRecordNotFound, got %v", err)
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.DeletePasswordRecord(ctx, "test-vault", "gmail")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})

	t.Run("deletes specific record without affecting others", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "pass1")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "github", "user@github.com", "pass2")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.DeletePasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("DeletePasswordRecord() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 1 {
			t.Errorf("expected 1 record, got %d", len(records))
		}
		if len(records) > 0 && records[0].Name != "github" {
			t.Errorf("expected remaining record %q, got %q", "github", records[0].Name)
		}
	})

	t.Run("persists deletion after unlock/lock cycle", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		err = service.AddPasswordRecord(ctx, "test-vault", "gmail", "user@gmail.com", "secret123")
		if err != nil {
			t.Fatalf("AddPasswordRecord() failed: %v", err)
		}

		err = service.DeletePasswordRecord(ctx, "test-vault", "gmail")
		if err != nil {
			t.Fatalf("DeletePasswordRecord() failed: %v", err)
		}

		err = service.LockVault(ctx, "test-vault")
		if err != nil {
			t.Fatalf("LockVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 0 {
			t.Errorf("expected 0 records after unlock, got %d", len(records))
		}
	})
}

func TestListPasswordRecords(t *testing.T) {
	t.Run("lists all records", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		expected := []string{"gmail", "github", "twitter"}
		for _, name := range expected {
			err = service.AddPasswordRecord(ctx, "test-vault", name, "user@"+name+".com", "pass")
			if err != nil {
				t.Fatalf("AddPasswordRecord() failed: %v", err)
			}
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != len(expected) {
			t.Errorf("expected %d records, got %d", len(expected), len(records))
		}
	})

	t.Run("returns empty list for new vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		_, err = service.ListPasswordRecords(ctx, "test-vault")
		if err != domain.ErrVaultNotFound {
			t.Errorf("expected ErrVaultNotFound, got %v", err)
		}
	})
}

func TestListVaults(t *testing.T) {
	t.Run("lists all vaults", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		vaults := []string{"personal", "work", "shared"}
		for _, name := range vaults {
			err := service.CreateVault(ctx, name, "password")
			if err != nil {
				t.Fatalf("CreateVault() failed: %v", err)
			}
		}

		list, err := service.ListVaults(ctx)
		if err != nil {
			t.Fatalf("ListVaults() failed: %v", err)
		}

		if len(list) != len(vaults) {
			t.Errorf("expected %d vaults, got %d", len(vaults), len(list))
		}
	})

	t.Run("returns empty list when no vaults exist", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		list, err := service.ListVaults(ctx)
		if err != nil {
			t.Fatalf("ListVaults() failed: %v", err)
		}

		if len(list) != 0 {
			t.Errorf("expected 0 vaults, got %d", len(list))
		}
	})
}

func TestIsVaultUnlocked(t *testing.T) {
	t.Run("returns true for unlocked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		if !service.IsVaultUnlocked(ctx, "test-vault") {
			t.Error("IsVaultUnlocked() should return true")
		}
	})

	t.Run("returns false for locked vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		if service.IsVaultUnlocked(ctx, "test-vault") {
			t.Error("IsVaultUnlocked() should return false for locked vault")
		}
	})

	t.Run("returns false for non-existent vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		if service.IsVaultUnlocked(ctx, "non-existent") {
			t.Error("IsVaultUnlocked() should return false for non-existent vault")
		}
	})
}

func TestConcurrency(t *testing.T) {
	t.Run("concurrent access to different vaults", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		vaultCount := 10
		var wg sync.WaitGroup
		wg.Add(vaultCount)

		for i := 0; i < vaultCount; i++ {
			go func(index int) {
				defer wg.Done()
				vaultName := fmt.Sprintf("vault-%d", index)

				err := service.CreateVault(ctx, vaultName, "password")
				if err != nil {
					t.Errorf("CreateVault() failed: %v", err)
					return
				}

				err = service.UnlockVault(ctx, vaultName, "password")
				if err != nil {
					t.Errorf("UnlockVault() failed: %v", err)
					return
				}

				err = service.AddPasswordRecord(ctx, vaultName, "record", "user", "pass")
				if err != nil {
					t.Errorf("AddPasswordRecord() failed: %v", err)
					return
				}
			}(i)
		}

		wg.Wait()

		vaults, err := service.ListVaults(ctx)
		if err != nil {
			t.Fatalf("ListVaults() failed: %v", err)
		}

		if len(vaults) != vaultCount {
			t.Errorf("expected %d vaults, got %d", vaultCount, len(vaults))
		}
	})

	t.Run("concurrent read operations on same vault", func(t *testing.T) {
		service, _ := setupTestService(t)
		ctx := context.Background()

		err := service.CreateVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("CreateVault() failed: %v", err)
		}

		err = service.UnlockVault(ctx, "test-vault", "my-password")
		if err != nil {
			t.Fatalf("UnlockVault() failed: %v", err)
		}

		// Add some records first
		for i := 0; i < 5; i++ {
			recordName := fmt.Sprintf("record-%d", i)
			err := service.AddPasswordRecord(ctx, "test-vault", recordName, "user", "pass")
			if err != nil {
				t.Fatalf("AddPasswordRecord() failed: %v", err)
			}
		}

		// Now do concurrent reads
		readCount := 20
		var wg sync.WaitGroup
		wg.Add(readCount)

		for i := 0; i < readCount; i++ {
			go func() {
				defer wg.Done()
				_, err := service.ListPasswordRecords(ctx, "test-vault")
				if err != nil {
					t.Errorf("ListPasswordRecords() failed: %v", err)
				}
			}()
		}

		wg.Wait()

		records, err := service.ListPasswordRecords(ctx, "test-vault")
		if err != nil {
			t.Fatalf("ListPasswordRecords() failed: %v", err)
		}

		if len(records) != 5 {
			t.Errorf("expected 5 records, got %d", len(records))
		}
	})
}
