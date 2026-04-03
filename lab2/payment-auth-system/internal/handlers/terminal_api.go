package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type TerminalAPIHandler struct {
	DB *sql.DB
}

type AuthorizeTransactionRequest struct {
	CardNumber           string `json:"card_number"`
	Amount               int64  `json:"amount"`
	TerminalSerialNumber string `json:"terminal_serial_number"`
}

type AuthorizeTransactionResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	CardID      int64  `json:"card_id,omitempty"`
	TerminalID  int64  `json:"terminal_id,omitempty"`
	BalanceLeft int64  `json:"balance_left,omitempty"`
}

func NewTerminalAPIHandler(db *sql.DB) *TerminalAPIHandler {
	return &TerminalAPIHandler{DB: db}
}

// Authorize godoc
// @Summary Авторизация платежа терминалом
// @Description Проверяет карту, терминал и баланс, затем создаёт транзакцию
// @Tags terminal
// @Accept json
// @Produce json
// @Param request body AuthorizeTransactionRequest true "Authorize transaction request"
// @Success 200 {object} AuthorizeTransactionResponse
// @Failure 400 {string} string
// @Failure 401 {object} AuthorizeTransactionResponse
// @Router /terminal/authorize [post]
func (h *TerminalAPIHandler) Authorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthorizeTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.CardNumber == "" || req.Amount <= 0 || req.TerminalSerialNumber == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	var cardID int64
	var balance int64
	var isBlocked bool

	err = h.DB.QueryRow(`
		SELECT id, balance, is_blocked
		FROM cards
		WHERE number = ?
	`, req.CardNumber).Scan(&cardID, &balance, &isBlocked)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthorizeTransactionResponse{
			Status:  "declined",
			Message: "card not found",
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if isBlocked {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthorizeTransactionResponse{
			Status:  "declined",
			Message: "card is blocked",
			CardID:  cardID,
		})
		return
	}

	var terminalID int64
	err = h.DB.QueryRow(`
		SELECT id
		FROM terminals
		WHERE serial_number = ?
	`, req.TerminalSerialNumber).Scan(&terminalID)
	if err == sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthorizeTransactionResponse{
			Status:  "declined",
			Message: "terminal not found",
			CardID:  cardID,
		})
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if balance < req.Amount {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthorizeTransactionResponse{
			Status:      "declined",
			Message:     "insufficient funds",
			CardID:      cardID,
			TerminalID:  terminalID,
			BalanceLeft: balance,
		})
		return
	}

	newBalance := balance - req.Amount

	_, err = h.DB.Exec(`
		UPDATE cards
		SET balance = ?
		WHERE id = ?
	`, newBalance, cardID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.DB.Exec(`
		INSERT INTO transactions (amount, card_id, terminal_id, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, req.Amount, cardID, terminalID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthorizeTransactionResponse{
		Status:      "approved",
		Message:     "transaction approved",
		CardID:      cardID,
		TerminalID:  terminalID,
		BalanceLeft: newBalance,
	})
}

// GetKeys godoc
// @Summary Получить ключи для терминала
// @Description Возвращает все ключи для терминала
// @Tags terminal
// @Produce json
// @Success 200 {array} map[string]interface{}
// @Router /terminal/keys [get]
func (h *TerminalAPIHandler) GetKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := h.DB.Query(`
		SELECT id, key_value, created_at
		FROM keys
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	keys := make([]map[string]any, 0)

	for rows.Next() {
		var id int64
		var keyValue string
		var createdAt string

		err := rows.Scan(&id, &keyValue, &createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		keys = append(keys, map[string]any{
			"id":         id,
			"key_value":  keyValue,
			"created_at": createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}
