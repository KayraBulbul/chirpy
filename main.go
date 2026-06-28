package main

import (
	"log"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()

	h1 := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal("error writing body")
		}
	}

	serverMux.HandleFunc("/healthz", h1)

	serverMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen and server error")
	}
}
