package http

import (
	"encoding/json"
	"net/http"

	"github.com/orlan/go-password-manager/internal/application"
	"github.com/orlan/go-password-manager/internal/domain"
)

// Handler handles HTTP requests for the password manager API
type Handler struct {
	service     *application.VaultService
	csrfManager *CSRFManager
}

// NewHandler creates a new HTTP handler
func NewHandler(service *application.VaultService) *Handler {
	return &Handler{
		service:     service,
		csrfManager: NewCSRFManager(),
	}
}

// RegisterRoutes sets up the HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// CSRF token endpoint (no CSRF protection needed)
	mux.HandleFunc("/api/csrf-token", h.handleCSRFToken)

	mux.HandleFunc("/api/vaults", h.handleVaults)
	mux.HandleFunc("/api/vaults/create", h.handleCreateVault)
	mux.HandleFunc("/api/vaults/unlock", h.handleUnlockVault)
	mux.HandleFunc("/api/vaults/lock", h.handleLockVault)
	mux.HandleFunc("/api/records", h.handleRecords)
	mux.HandleFunc("/api/records/add", h.handleAddRecord)
	mux.HandleFunc("/api/records/get", h.handleGetRecord)
	mux.HandleFunc("/api/records/update", h.handleUpdateRecord)
	mux.HandleFunc("/api/records/delete", h.handleDeleteRecord)
	mux.HandleFunc("/health", h.handleHealth)
}

// GetCSRFMiddleware returns the CSRF middleware for this handler
func (h *Handler) GetCSRFMiddleware() func(http.Handler) http.Handler {
	return CSRFMiddleware(h.csrfManager)
}

// CreateVaultRequest represents a request to create a new vault
type CreateVaultRequest struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

// UnlockVaultRequest represents a request to unlock a vault
type UnlockVaultRequest struct {
	Name           string `json:"name"`
	MasterPassword string `json:"master_password"`
}

// LockVaultRequest represents a request to lock a vault
type LockVaultRequest struct {
	Name string `json:"name"`
}

// AddRecordRequest represents a request to add a password record
type AddRecordRequest struct {
	VaultName string `json:"vault_name"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

// GetRecordRequest represents a request to retrieve a password record
type GetRecordRequest struct {
	VaultName string `json:"vault_name"`
	Name      string `json:"name"`
}

// UpdateRecordRequest represents a request to update a password record
type UpdateRecordRequest struct {
	VaultName string `json:"vault_name"`
	Name      string `json:"name"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
}

// DeleteRecordRequest represents a request to delete a password record
type DeleteRecordRequest struct {
	VaultName string `json:"vault_name"`
	Name      string `json:"name"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
}

// handleCSRFToken generates and returns a new CSRF token
func (h *Handler) handleCSRFToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token, err := h.csrfManager.generateToken()
	if err != nil {
		h.sendError(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	// Set token in cookie (HttpOnly for security)
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(CSRFTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	// Also return in response body for client-side storage
	h.sendJSON(w, map[string]string{"token": token})
}

// handleHealth returns the health status of the service
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleVaults lists all available vaults
func (h *Handler) handleVaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vaults, err := h.service.ListVaults(r.Context())
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, map[string]interface{}{"vaults": vaults})
}

// handleCreateVault creates a new vault
func (h *Handler) handleCreateVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateVaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.MasterPassword == "" {
		h.sendError(w, "name and master_password are required", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateVault(r.Context(), req.Name, req.MasterPassword); err != nil {
		if err == domain.ErrVaultAlreadyExists {
			h.sendError(w, err.Error(), http.StatusConflict)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "vault created successfully"})
}

// handleUnlockVault unlocks a vault
func (h *Handler) handleUnlockVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UnlockVaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.MasterPassword == "" {
		h.sendError(w, "name and master_password are required", http.StatusBadRequest)
		return
	}

	if err := h.service.UnlockVault(r.Context(), req.Name, req.MasterPassword); err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		if err == domain.ErrInvalidMasterPassword {
			h.sendError(w, err.Error(), http.StatusUnauthorized)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "vault unlocked successfully"})
}

// handleLockVault locks a vault
func (h *Handler) handleLockVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LockVaultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		h.sendError(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := h.service.LockVault(r.Context(), req.Name); err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "vault locked successfully"})
}

// handleRecords lists all records in a vault
func (h *Handler) handleRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vaultName := r.URL.Query().Get("vault_name")
	if vaultName == "" {
		h.sendError(w, "vault_name query parameter is required", http.StatusBadRequest)
		return
	}

	records, err := h.service.ListPasswordRecords(r.Context(), vaultName)
	if err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, "vault not found or not unlocked", http.StatusNotFound)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, map[string]interface{}{"records": records})
}

// handleAddRecord adds a new password record
func (h *Handler) handleAddRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.VaultName == "" || req.Name == "" || req.Username == "" || req.Password == "" {
		h.sendError(w, "vault_name, name, username, and password are required", http.StatusBadRequest)
		return
	}

	if err := h.service.AddPasswordRecord(r.Context(), req.VaultName, req.Name, req.Username, req.Password); err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, "vault not found or not unlocked", http.StatusNotFound)
			return
		}
		if err == domain.ErrRecordAlreadyExists {
			h.sendError(w, err.Error(), http.StatusConflict)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "password record added successfully"})
}

// handleGetRecord retrieves a password record
func (h *Handler) handleGetRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vaultName := r.URL.Query().Get("vault_name")
	recordName := r.URL.Query().Get("name")

	if vaultName == "" || recordName == "" {
		h.sendError(w, "vault_name and name query parameters are required", http.StatusBadRequest)
		return
	}

	record, err := h.service.GetPasswordRecord(r.Context(), vaultName, recordName)
	if err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, "vault not found or not unlocked", http.StatusNotFound)
			return
		}
		if err == domain.ErrRecordNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, record)
}

// handleUpdateRecord updates a password record
func (h *Handler) handleUpdateRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.VaultName == "" || req.Name == "" {
		h.sendError(w, "vault_name and name are required", http.StatusBadRequest)
		return
	}

	if req.Username == "" && req.Password == "" {
		h.sendError(w, "at least username or password must be provided", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePasswordRecord(r.Context(), req.VaultName, req.Name, req.Username, req.Password); err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, "vault not found or not unlocked", http.StatusNotFound)
			return
		}
		if err == domain.ErrRecordNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "password record updated successfully"})
}

// handleDeleteRecord deletes a password record
func (h *Handler) handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.VaultName == "" || req.Name == "" {
		h.sendError(w, "vault_name and name are required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePasswordRecord(r.Context(), req.VaultName, req.Name); err != nil {
		if err == domain.ErrVaultNotFound {
			h.sendError(w, "vault not found or not unlocked", http.StatusNotFound)
			return
		}
		if err == domain.ErrRecordNotFound {
			h.sendError(w, err.Error(), http.StatusNotFound)
			return
		}
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.sendJSON(w, SuccessResponse{Message: "password record deleted successfully"})
}

// sendJSON sends a JSON response
func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// sendError sends an error response
func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
