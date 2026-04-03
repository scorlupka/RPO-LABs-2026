package handlers

import (
	"database/sql"
	"net/http"

	_ "payment-auth-system/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RegisterRoutes(db *sql.DB) {
	terminalHandler := NewTerminalHandler(db)
	cardHandler := NewCardHandler(db)
	keyHandler := NewKeyHandler(db)
	userHandler := NewUserHandler(db)
	authHandler := NewAuthHandler(db)
	transactionHandler := NewTransactionHandler(db)
	terminalAPIHandler := NewTerminalAPIHandler(db)

	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	http.Handle("/api/v1/swagger/", httpSwagger.WrapHandler)

	http.HandleFunc("/api/v1/auth/login", authHandler.Login)

	http.HandleFunc("/api/v1/terminal/authorize", terminalAPIHandler.Authorize)

	http.HandleFunc("/api/v1/terminal/keys", terminalAPIHandler.GetKeys)

	http.HandleFunc("/api/v1/transactions", JWTMiddleware(transactionHandler.HandleTransactions))
	http.HandleFunc("/api/v1/transactions/", JWTMiddleware(transactionHandler.HandleTransactionByID))

	http.HandleFunc("/api/v1/terminals", JWTMiddleware(terminalHandler.HandleTerminals))
	http.HandleFunc("/api/v1/terminals/", JWTMiddleware(terminalHandler.HandleTerminalByID))

	http.HandleFunc("/api/v1/cards", JWTMiddleware(cardHandler.HandleCards))
	http.HandleFunc("/api/v1/cards/", JWTMiddleware(cardHandler.HandleCardByID))

	http.HandleFunc("/api/v1/keys", JWTMiddleware(AdminOnly(keyHandler.HandleKeys)))
	http.HandleFunc("/api/v1/keys/", JWTMiddleware(AdminOnly(keyHandler.HandleKeyByID)))

	http.HandleFunc("/api/v1/users", JWTMiddleware(userHandler.HandleUsers))
	http.HandleFunc("/api/v1/users/", JWTMiddleware(userHandler.HandleUserByID))

}
