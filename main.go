package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

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

func validateChirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}

		if len(params.Body) > 140 {
			respondWithError(w, 400, "Chirp is too long")
		} else {
			words := strings.Split(params.Body, " ")

			for i, word := range words {
				switch strings.ToLower(word) {
				case "kerfuffle":
					fallthrough
				case "sharbert":
					fallthrough
				case "fornax":
					words[i] = "****"
				}
				continue
			}

			type returnVals struct {
				CleanedBody string `json:"cleaned_body"`
			}
			respBody := returnVals{
				CleanedBody: strings.Join(words, " "),
			}
			respondWithJSON(w, 200, respBody)

		}
	})
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
	serverMux.Handle("POST /admin/reset", apiCfg.resetHits())
	serverMux.Handle("POST /api/validate_chirp", validateChirp())
	serverMux.Handle("POST /api/users")

	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("Listen and server error")
	}
}
