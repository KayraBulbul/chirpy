package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/KayraBulbul/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	PLATFORM       string
}

func (cfg *apiConfig) readPlatform() {
	cfg.PLATFORM = os.Getenv("PLATFORM")
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getHits() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		_, err := w.Write([]byte(fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileserverHits.Load())))
		if err != nil {
			log.Fatal("error writing hits body")
		}
	})
}

func (cfg *apiConfig) reset() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.readPlatform()
		if cfg.PLATFORM == "dev" {
			err := cfg.dbQueries.DeleteUsers(r.Context())
			if err != nil {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(200)
			return
		} else {
			w.WriteHeader(403)
			return
		}
	})
}
