package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Сервер атворизации пдатежей запущен")
	// TODO: добавить инициализацию БД, маршруты

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Fatal(http.ListenAndServeTLS(":8888",
		"/etc/ssl/certs/server.crt",
		"/etc/ssl/private/server.key",
		nil))
}
