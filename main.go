package main

import (
	"log"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()

	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen and server error")
	}
}
