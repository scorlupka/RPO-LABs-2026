package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"payment-auth-system/internal/models"
)

type CardHandler struct {
	DB *sql.DB
}

type CreateCardRequest struct {
	Number     string `json:"number"`
	Balance    int64  `json:"balance"`
	IsBlocked  bool   `json:"is_blocked"`
	OwnerName  string `json:"owner_name"`
	ExpireDate string `json:"expire_date"`
	KeyID      int64  `json:"key_id"`
}

type UpdateCardRequest struct {
	Number     string `json:"number"`
	Balance    int64  `json:"balance"`
	IsBlocked  bool   `json:"is_blocked"`
	OwnerName  string `json:"owner_name"`
	ExpireDate string `json:"expire_date"`
	KeyID      int64  `json:"key_id"`
}

func NewCardHandler(db *sql.DB) *CardHandler {
	return &CardHandler{DB: db}
}

func (h *CardHandler) HandleCards(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAll(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *CardHandler) HandleCardByID(w http.ResponseWriter, r *http.Request) {
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

func (h *CardHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, number, balance, is_blocked, owner_name, expire_date, key_id, created_at
		FROM cards
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cards := make([]models.Card, 0)

	for rows.Next() {
		var card models.Card

		err := rows.Scan(
			&card.ID,
			&card.Number,
			&card.Balance,
			&card.IsBlocked,
			&card.OwnerName,
			&card.ExpireDate,
			&card.KeyID,
			&card.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cards = append(cards, card)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

func (h *CardHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cards/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	var card models.Card

	err = h.DB.QueryRow(`
		SELECT id, number, balance, is_blocked, owner_name, expire_date, key_id, created_at
		FROM cards
		WHERE id = ?
	`, id).Scan(
		&card.ID,
		&card.Number,
		&card.Balance,
		&card.IsBlocked,
		&card.OwnerName,
		&card.ExpireDate,
		&card.KeyID,
		&card.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCardRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Number == "" || req.OwnerName == "" || req.ExpireDate == "" || req.KeyID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO cards (number, balance, is_blocked, owner_name, expire_date, key_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, req.Number, req.Balance, req.IsBlocked, req.OwnerName, req.ExpireDate, req.KeyID)
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
		"id":          id,
		"number":      req.Number,
		"balance":     req.Balance,
		"is_blocked":  req.IsBlocked,
		"owner_name":  req.OwnerName,
		"expire_date": req.ExpireDate,
		"key_id":      req.KeyID,
	})
}

func (h *CardHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cards/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	var req UpdateCardRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Number == "" || req.OwnerName == "" || req.ExpireDate == "" || req.KeyID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE cards
		SET number = ?, balance = ?, is_blocked = ?, owner_name = ?, expire_date = ?, key_id = ?
		WHERE id = ?
	`, req.Number, req.Balance, req.IsBlocked, req.OwnerName, req.ExpireDate, req.KeyID, id)
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
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":          id,
		"number":      req.Number,
		"balance":     req.Balance,
		"is_blocked":  req.IsBlocked,
		"owner_name":  req.OwnerName,
		"expire_date": req.ExpireDate,
		"key_id":      req.KeyID,
	})
}

func (h *CardHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/cards/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid card id", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		DELETE FROM cards
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
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "card deleted",
	})
}
