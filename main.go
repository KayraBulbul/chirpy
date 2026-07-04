package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/KayraBulbul/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVal struct {
		Error string `json:"error"`
	}
	respBody := returnVal{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing error response: %s", err)
		return
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(data)
	if err != nil {
		log.Printf("Error writing JSON response: %s", err)
		return
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment file")
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening database")
	}

	serverMux := http.NewServeMux()

	h1 := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal("error writing body")
		}
	}

	dbQueries := database.New(db)
	apiCfg := &apiConfig{dbQueries: dbQueries}

	serverMux.HandleFunc("GET /api/healthz", h1)
	serverMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	serverMux.Handle("GET /admin/metrics", apiCfg.getHits())
	serverMux.Handle("POST /admin/reset", apiCfg.reset())
	serverMux.Handle("POST /api/users", apiCfg.createUser())
	serverMux.Handle("POST /api/chirps", apiCfg.createChirp())
	serverMux.Handle("GET /api/chirps", apiCfg.getChirps())
	serverMux.Handle("GET /api/chirps/{chirpID}", apiCfg.getChirpByID())

	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen and server error")
	}
}
