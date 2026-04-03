package handlers

import (
	"database/sql"
	"net/http"
)

func RegisterRoutes(db *sql.DB) {
	terminalHandler := NewTerminalHandler(db)
	cardHandler := NewCardHandler(db)
	keyHandler := NewKeyHandler(db)
	userHandler := NewUserHandler(db)
	authHandler := NewAuthHandler(db)

	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/api/v1/auth/login", authHandler.Login)

	http.HandleFunc("/api/v1/terminals", JWTMiddleware(terminalHandler.HandleTerminals))
	http.HandleFunc("/api/v1/terminals/", JWTMiddleware(terminalHandler.HandleTerminalByID))

	http.HandleFunc("/api/v1/cards", JWTMiddleware(cardHandler.HandleCards))
	http.HandleFunc("/api/v1/cards/", JWTMiddleware(cardHandler.HandleCardByID))

	http.HandleFunc("/api/v1/keys", JWTMiddleware(keyHandler.HandleKeys))
	http.HandleFunc("/api/v1/keys/", JWTMiddleware(keyHandler.HandleKeyByID))

	http.HandleFunc("/api/v1/users", JWTMiddleware(userHandler.HandleUsers))
	http.HandleFunc("/api/v1/users/", JWTMiddleware(userHandler.HandleUserByID))

}
