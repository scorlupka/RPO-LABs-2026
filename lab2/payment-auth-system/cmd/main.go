package main

import (
	"log"
	"net/http"
	appdb "payment-auth-system/internal/db"
	handler "payment-auth-system/internal/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.Println("Сервер атворизации пдатежей запущен")
	db, err := appdb.New("./data/app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("База данных успещно подключена")

	handler.RegisterRoutes(db)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
