package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"payment-auth-system/internal/models"
)

type TransactionHandler struct {
	DB *sql.DB
}

type CreateTransactionRequest struct {
	Amount     int64 `json:"amount"`
	CardID     int64 `json:"card_id"`
	TerminalID int64 `json:"terminal_id"`
}

type UpdateTransactionRequest struct {
	Amount     int64 `json:"amount"`
	CardID     int64 `json:"card_id"`
	TerminalID int64 `json:"terminal_id"`
}

func NewTransactionHandler(db *sql.DB) *TransactionHandler {
	return &TransactionHandler{DB: db}
}

func (h *TransactionHandler) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetAll(w, r)
	case http.MethodPost:
		h.Create(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TransactionHandler) HandleTransactionByID(w http.ResponseWriter, r *http.Request) {
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

// GetAll godoc
// @Summary Получить список транзакций
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Transaction
// @Failure 401 {string} string
// @Router /transactions [get]
func (h *TransactionHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.DB.Query(`
		SELECT id, amount, card_id, terminal_id, created_at
		FROM transactions
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	transactions := make([]models.Transaction, 0)

	for rows.Next() {
		var transaction models.Transaction

		err := rows.Scan(
			&transaction.ID,
			&transaction.Amount,
			&transaction.CardID,
			&transaction.TerminalID,
			&transaction.CreatedAt,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		transactions = append(transactions, transaction)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// GetByID godoc
// @Summary Получить транзакцию по ID
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.Transaction
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /transactions/{id} [get]
func (h *TransactionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	var transaction models.Transaction

	err = h.DB.QueryRow(`
		SELECT id, amount, card_id, terminal_id, created_at
		FROM transactions
		WHERE id = ?
	`, id).Scan(
		&transaction.ID,
		&transaction.Amount,
		&transaction.CardID,
		&transaction.TerminalID,
		&transaction.CreatedAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

// Create godoc
// @Summary Создать транзакцию
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTransactionRequest true "Create transaction request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Router /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 || req.CardID == 0 || req.TerminalID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		INSERT INTO transactions (amount, card_id, terminal_id, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, req.Amount, req.CardID, req.TerminalID)
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
		"amount":      req.Amount,
		"card_id":     req.CardID,
		"terminal_id": req.TerminalID,
	})
}

// UpdateByID godoc
// @Summary Обновить транзакцию
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Transaction ID"
// @Param request body UpdateTransactionRequest true "Update transaction request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /transactions/{id} [put]
func (h *TransactionHandler) UpdateByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	var req UpdateTransactionRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 || req.CardID == 0 || req.TerminalID == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		UPDATE transactions
		SET amount = ?, card_id = ?, terminal_id = ?
		WHERE id = ?
	`, req.Amount, req.CardID, req.TerminalID, id)
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
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":          id,
		"amount":      req.Amount,
		"card_id":     req.CardID,
		"terminal_id": req.TerminalID,
	})
}

// DeleteByID godoc
// @Summary Удалить транзакцию
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path int true "Transaction ID"
// @Success 200 {object} map[string]string
// @Failure 400 {string} string
// @Failure 401 {string} string
// @Failure 404 {string} string
// @Router /transactions/{id} [delete]
func (h *TransactionHandler) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/v1/transactions/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid transaction id", http.StatusBadRequest)
		return
	}

	result, err := h.DB.Exec(`
		DELETE FROM transactions
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
		http.Error(w, "transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "transaction deleted",
	})
}
