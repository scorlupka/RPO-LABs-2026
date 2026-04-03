package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"payment-auth-system/internal/models"
)

type TerminalHandler struct {
	DB *sql.DB
}

type CreateTerminalRequest struct {
	SerialNumber string `json:"serial_number"`
	Address      string `json:"address"`
	Name         string `json:"name"`
	Status       bool   `json:"status"`
}

func NewTerminalHandler(db *sql.DB) *TerminalHandler {
	return &TerminalHandler{DB: db}
}

func (h *TerminalHandler) HandleTerminals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAll(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetAll godoc
// @Summary Получить список терминалов
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Terminal
// @Failure 401 {string} string
// @Router /terminals [get]
func (h *TerminalHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, serial_number, address, name, status, created_at
		FROM terminals
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var terminals []models.Terminal

	for rows.Next() {
		var terminal models.Terminal

		err := rows.Scan(
			&terminal.ID,
			&terminal.SerialNumber,
			&terminal.Address,
			&terminal.Name,
			&terminal.Status,
			&terminal.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		terminals = append(terminals, terminal)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terminals)
}

// Create godoc
// @Summary Создать терминал
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTerminalRequest true "Create terminal request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /terminals [post]
func (h *TerminalHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTerminalRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.SerialNumber == "" || req.Address == "" || req.Name == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO terminals (serial_number, address, name, status, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, req.SerialNumber, req.Address, req.Name, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":            id,
		"serial_number": req.SerialNumber,
		"address":       req.Address,
		"name":          req.Name,
		"status":        req.Status,
	})
}

// GetByID godoc
// @Summary Получить терминал по ID
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Success 200 {object} models.Terminal
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /terminals/{id} [get]
func (h *TerminalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/terminals/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid terminal id", http.StatusBadRequest)
		return
	}

	var terminal models.Terminal

	err = h.DB.QueryRow(`
		SELECT id, serial_number, address, name, status, created_at
		FROM terminals
		WHERE id = ?
	`, id).Scan(
		&terminal.ID,
		&terminal.SerialNumber,
		&terminal.Address,
		&terminal.Name,
		&terminal.Status,
		&terminal.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "terminal not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(terminal)
}

// DeleteByID godoc
// @Summary Удалить терминал
// @Tags terminals
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /terminals/{id} [delete]
func (h *TerminalHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/terminals/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid terminal id", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		DELETE FROM terminals
		WHERE id = ?
	`, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "terminal not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "terminal deleted",
	})
}

func (h *TerminalHandler) HandleTerminalByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetByID(w, r)
	case http.MethodPut:
		h.UpdateByID(w, r)
	case http.MethodDelete:
		h.DeleteByID(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type UpdateTerminalRequest struct {
	SerialNumber string `json:"serial_number"`
	Address      string `json:"address"`
	Name         string `json:"name"`
	Status       bool   `json:"status"`
}

// UpdateByID godoc
// @Summary Обновить терминал
// @Tags terminals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Terminal ID"
// @Param request body UpdateTerminalRequest true "Update terminal request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /terminals/{id} [put]
func (h *TerminalHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/terminals/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid terminal id", http.StatusBadRequest)
		return
	}

	var req UpdateTerminalRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.SerialNumber == "" || req.Address == "" || req.Name == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE terminals
		SET serial_number = ?, address = ?, name = ?, status = ?
		WHERE id = ?
	`, req.SerialNumber, req.Address, req.Name, req.Status, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "terminal not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":            id,
		"serial_number": req.SerialNumber,
		"address":       req.Address,
		"name":          req.Name,
		"status":        req.Status,
	})
}
