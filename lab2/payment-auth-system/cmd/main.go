package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Сервер атворизации пдатежей запущен")
	// TODO: добавить инициализацию БД, маршруты

	http.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
