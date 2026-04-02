package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"payment-auth-system/internal/models"
)

type TerminalHandler struct {
	DB *sql.DB
}

func NewTerminalHandler(db *sql.DB) *TerminalHandler {
	return &TerminalHandler{DB: db}
}

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
