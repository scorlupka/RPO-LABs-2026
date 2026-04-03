package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("my-secret-key")

type AuthHandler struct {
	DB *sql.DB
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserAuthData struct {
	ID           int64
	Login        string
	PasswordHash string
	IsAdmin      bool
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{DB: db}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "missing login or password", http.StatusBadRequest)
		return
	}

	var user UserAuthData

	err = h.DB.QueryRow(`
		SELECT id, login, password_hash, is_admin
		FROM users
		WHERE login = ?
	`, req.Login).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.IsAdmin,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "invalid login or password", http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"login":    user.Login,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
	})
}
