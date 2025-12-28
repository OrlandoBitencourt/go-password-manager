package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orlan/go-password-manager/internal/application"
	"github.com/orlan/go-password-manager/internal/crypto"
	"github.com/orlan/go-password-manager/internal/domain"
	"github.com/orlan/go-password-manager/internal/vault"
)

func setupTestHandler(t *testing.T) *Handler {
	t.Helper()
	vaultDir := t.TempDir()
	repo, err := vault.NewFileRepository(vaultDir)
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	cryptoSvc := crypto.NewService()
	service := application.NewVaultService(repo, cryptoSvc)
	return NewHandler(service)
}

func TestNewHandler(t *testing.T) {
	handler := setupTestHandler(t)
	if handler == nil {
		t.Fatal("NewHandler() returned nil")
	}
	if handler.service == nil {
		t.Error("handler service is nil")
	}
}

func TestRegisterRoutes(t *testing.T) {
	handler := setupTestHandler(t)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Test that routes are registered (basic smoke test)
	routes := []string{
		"/api/vaults",
		"/api/vaults/create",
		"/api/vaults/unlock",
		"/api/vaults/lock",
		"/api/records",
		"/api/records/add",
		"/api/records/get",
		"/api/records/update",
		"/api/records/delete",
		"/health",
	}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// Should get a response (not 404)
		if w.Code == http.StatusNotFound {
			t.Errorf("route %q was not registered", route)
		}
	}
}

func TestHandleHealth(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %q", response["status"])
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

func TestHandleCreateVault(t *testing.T) {
	t.Run("creates vault successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := CreateVaultRequest{
			Name:           "test-vault",
			MasterPassword: "my-password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleCreateVault(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response SuccessResponse
		json.NewDecoder(w.Body).Decode(&response)
		if response.Message != "vault created successfully" {
			t.Errorf("unexpected message: %q", response.Message)
		}
	})

	t.Run("returns error for duplicate vault", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := CreateVaultRequest{
			Name:           "test-vault",
			MasterPassword: "my-password",
		}
		body, _ := json.Marshal(reqBody)

		// Create vault first time
		req1 := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBuffer(body))
		w1 := httptest.NewRecorder()
		handler.handleCreateVault(w1, req1)

		// Try to create again
		body2, _ := json.Marshal(reqBody)
		req2 := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBuffer(body2))
		w2 := httptest.NewRecorder()
		handler.handleCreateVault(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w2.Code)
		}
	})

	t.Run("returns error for missing name", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := CreateVaultRequest{
			Name:           "",
			MasterPassword: "my-password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleCreateVault(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for missing password", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := CreateVaultRequest{
			Name:           "test-vault",
			MasterPassword: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleCreateVault(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/create", bytes.NewBufferString("invalid json"))
		w := httptest.NewRecorder()

		handler.handleCreateVault(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/vaults/create", nil)
		w := httptest.NewRecorder()

		handler.handleCreateVault(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleUnlockVault(t *testing.T) {
	t.Run("unlocks vault successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		// Create vault first
		handler.service.CreateVault(nil, "test-vault", "my-password")

		reqBody := UnlockVaultRequest{
			Name:           "test-vault",
			MasterPassword: "my-password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/unlock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUnlockVault(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns error for wrong password", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "correct-password")

		reqBody := UnlockVaultRequest{
			Name:           "test-vault",
			MasterPassword: "wrong-password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/unlock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUnlockVault(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("returns error for non-existent vault", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := UnlockVaultRequest{
			Name:           "non-existent",
			MasterPassword: "password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/unlock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUnlockVault(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for missing fields", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := UnlockVaultRequest{
			Name:           "",
			MasterPassword: "password",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/unlock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUnlockVault(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/vaults/unlock", nil)
		w := httptest.NewRecorder()

		handler.handleUnlockVault(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleLockVault(t *testing.T) {
	t.Run("locks vault successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")

		reqBody := LockVaultRequest{Name: "test-vault"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/lock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleLockVault(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns error for non-existent vault", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := LockVaultRequest{Name: "non-existent"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/lock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleLockVault(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for missing name", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := LockVaultRequest{Name: ""}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults/lock", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleLockVault(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/vaults/lock", nil)
		w := httptest.NewRecorder()

		handler.handleLockVault(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleVaults(t *testing.T) {
	t.Run("lists vaults successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "vault1", "password")
		handler.service.CreateVault(nil, "vault2", "password")

		req := httptest.NewRequest(http.MethodGet, "/api/vaults", nil)
		w := httptest.NewRecorder()

		handler.handleVaults(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string][]string
		json.NewDecoder(w.Body).Decode(&response)

		if len(response["vaults"]) != 2 {
			t.Errorf("expected 2 vaults, got %d", len(response["vaults"]))
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/vaults", nil)
		w := httptest.NewRecorder()

		handler.handleVaults(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleAddRecord(t *testing.T) {
	t.Run("adds record successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")

		reqBody := AddRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
			Username:  "user@gmail.com",
			Password:  "secret123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/records/add", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleAddRecord(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")

		reqBody := AddRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
			Username:  "user@gmail.com",
			Password:  "secret123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/records/add", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleAddRecord(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for duplicate record", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")
		handler.service.AddPasswordRecord(nil, "test-vault", "gmail", "user@gmail.com", "pass")

		reqBody := AddRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
			Username:  "user2@gmail.com",
			Password:  "pass2",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/records/add", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleAddRecord(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
		}
	})

	t.Run("returns error for missing fields", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := AddRecordRequest{
			VaultName: "test-vault",
			Name:      "",
			Username:  "user@gmail.com",
			Password:  "secret123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/records/add", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleAddRecord(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/records/add", nil)
		w := httptest.NewRecorder()

		handler.handleAddRecord(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleGetRecord(t *testing.T) {
	t.Run("gets record successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")
		handler.service.AddPasswordRecord(nil, "test-vault", "gmail", "user@gmail.com", "secret123")

		req := httptest.NewRequest(http.MethodGet, "/api/records/get?vault_name=test-vault&name=gmail", nil)
		w := httptest.NewRecorder()

		handler.handleGetRecord(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var record domain.PasswordRecord
		json.NewDecoder(w.Body).Decode(&record)

		if record.Name != "gmail" {
			t.Errorf("expected name 'gmail', got %q", record.Name)
		}
		if record.Username != "user@gmail.com" {
			t.Errorf("expected username 'user@gmail.com', got %q", record.Username)
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")

		req := httptest.NewRequest(http.MethodGet, "/api/records/get?vault_name=test-vault&name=non-existent", nil)
		w := httptest.NewRecorder()

		handler.handleGetRecord(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for locked vault", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")

		req := httptest.NewRequest(http.MethodGet, "/api/records/get?vault_name=test-vault&name=gmail", nil)
		w := httptest.NewRecorder()

		handler.handleGetRecord(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for missing query parameters", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/records/get?vault_name=test-vault", nil)
		w := httptest.NewRecorder()

		handler.handleGetRecord(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/records/get", nil)
		w := httptest.NewRecorder()

		handler.handleGetRecord(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleRecords(t *testing.T) {
	t.Run("lists records successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")
		handler.service.AddPasswordRecord(nil, "test-vault", "gmail", "user@gmail.com", "pass1")
		handler.service.AddPasswordRecord(nil, "test-vault", "github", "user@github.com", "pass2")

		req := httptest.NewRequest(http.MethodGet, "/api/records?vault_name=test-vault", nil)
		w := httptest.NewRecorder()

		handler.handleRecords(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var response map[string][]domain.PasswordRecord
		json.NewDecoder(w.Body).Decode(&response)

		if len(response["records"]) != 2 {
			t.Errorf("expected 2 records, got %d", len(response["records"]))
		}
	})

	t.Run("returns error for missing vault_name", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/records", nil)
		w := httptest.NewRecorder()

		handler.handleRecords(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/records", nil)
		w := httptest.NewRecorder()

		handler.handleRecords(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleUpdateRecord(t *testing.T) {
	t.Run("updates record successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")
		handler.service.AddPasswordRecord(nil, "test-vault", "gmail", "old@gmail.com", "oldpass")

		reqBody := UpdateRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
			Username:  "new@gmail.com",
			Password:  "newpass",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/records/update", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUpdateRecord(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")

		reqBody := UpdateRecordRequest{
			VaultName: "test-vault",
			Name:      "non-existent",
			Password:  "newpass",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/records/update", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUpdateRecord(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for missing both username and password", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := UpdateRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/records/update", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleUpdateRecord(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/records/update", nil)
		w := httptest.NewRecorder()

		handler.handleUpdateRecord(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

func TestHandleDeleteRecord(t *testing.T) {
	t.Run("deletes record successfully", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")
		handler.service.AddPasswordRecord(nil, "test-vault", "gmail", "user@gmail.com", "pass")

		reqBody := DeleteRecordRequest{
			VaultName: "test-vault",
			Name:      "gmail",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodDelete, "/api/records/delete", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleDeleteRecord(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("returns error for non-existent record", func(t *testing.T) {
		handler := setupTestHandler(t)

		handler.service.CreateVault(nil, "test-vault", "my-password")
		handler.service.UnlockVault(nil, "test-vault", "my-password")

		reqBody := DeleteRecordRequest{
			VaultName: "test-vault",
			Name:      "non-existent",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodDelete, "/api/records/delete", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleDeleteRecord(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("returns error for missing fields", func(t *testing.T) {
		handler := setupTestHandler(t)

		reqBody := DeleteRecordRequest{
			VaultName: "test-vault",
			Name:      "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodDelete, "/api/records/delete", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		handler.handleDeleteRecord(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("returns error for wrong method", func(t *testing.T) {
		handler := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/records/delete", nil)
		w := httptest.NewRecorder()

		handler.handleDeleteRecord(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}
