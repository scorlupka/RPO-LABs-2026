package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"payment-auth-system/internal/models"
)

type KeyHandler struct {
	DB *sql.DB
}

type CreateKeyRequest struct {
	KeyValue string `json:"key_value"`
}

type UpdateKeyRequest struct {
	KeyValue string `json:"key_value"`
}

func NewKeyHandler(db *sql.DB) *KeyHandler {
	return &KeyHandler{DB: db}
}

func (h *KeyHandler) HandleKeys(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAll(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *KeyHandler) HandleKeyByID(w http.ResponseWriter, r *http.Request) {
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

func (h *KeyHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, key_value, created_at
		FROM keys
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	keys := make([]models.Key, 0)

	for rows.Next() {
		var key models.Key

		err := rows.Scan(
			&key.ID,
			&key.KeyValue,
			&key.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		keys = append(keys, key)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *KeyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/keys/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	var key models.Key

	err = h.DB.QueryRow(`
		SELECT id, key_value, created_at
		FROM keys
		WHERE id = ?
	`, id).Scan(
		&key.ID,
		&key.KeyValue,
		&key.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}

func (h *KeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateKeyRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.KeyValue == "" {
		http.Error(w, "missing key_value", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO keys (key_value, created_at)
		VALUES (?, CURRENT_TIMESTAMP)
	`, req.KeyValue)
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
		"id":        id,
		"key_value": req.KeyValue,
	})
}

func (h *KeyHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/keys/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	var req UpdateKeyRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.KeyValue == "" {
		http.Error(w, "missing key_value", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE keys
		SET key_value = ?
		WHERE id = ?
	`, req.KeyValue, id)
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
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":        id,
		"key_value": req.KeyValue,
	})
}

func (h *KeyHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/keys/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid key id", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		DELETE FROM keys
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
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "key deleted",
	})
}
