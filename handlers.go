package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/KayraBulbul/chirpy/internal/auth"
	"github.com/KayraBulbul/chirpy/internal/database"
	"github.com/google/uuid"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
}

func (cfg *apiConfig) readPlatform() {
	cfg.platform = os.Getenv("PLATFORM")
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
		if cfg.platform == "dev" {
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

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) createChirp() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type reqParams struct {
			Body string `json:"body"`
		}

		decoder := json.NewDecoder(r.Body)
		params := reqParams{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, 500, "Error decoding createChrip body")
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, 500, "Error getting bearer token")
		}
		id, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			respondWithError(w, 401, "Unauthorized")
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
			resParams := database.CreateChirpParams{
				Body:   strings.Join(words, " "),
				UserID: id,
			}

			chirp, err := cfg.dbQueries.CreateChirp(r.Context(), resParams)
			if err != nil {
				respondWithError(w, 500, "Error creating chirp")
			}
			respondWithJSON(w, 201, Chirp{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, id})
		}
	})
}

func (cfg *apiConfig) getChirps() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dbChirps, err := cfg.dbQueries.GetChrips(r.Context())
		if err != nil {
			log.Printf("Error retrieving chirps from db: %s", err)
			w.WriteHeader(500)
			return
		}
		chirps := []Chirp{}
		for _, chirp := range dbChirps {
			chirps = append(chirps, Chirp{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID})
		}

		respondWithJSON(w, 200, chirps)
	})
}

func (cfg *apiConfig) getChirpByID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("chirpID")
		id, err := uuid.Parse(idString)
		if err != nil {
			log.Printf("Error parsing id: %s", err)
			w.WriteHeader(500)
			return
		}
		chirp, err := cfg.dbQueries.GetChirpById(r.Context(), id)
		if err != nil {
			w.WriteHeader(404)
			return
		}
		respondWithJSON(w, 200, Chirp{chirp.ID, chirp.CreatedAt, chirp.UpdatedAt, chirp.Body, chirp.UserID})
	})
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUser() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, 500, "Error decoding createUser request body")
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, 500, "Error hashing password")
		}

		userParams := database.CreateUserParams{
			Email:          params.Email,
			HashedPassword: hashedPassword,
		}

		user, err := cfg.dbQueries.CreateUser(r.Context(), userParams)
		if err != nil {
			log.Printf("createUser error: %v", err)
			respondWithError(w, 500, "Error creating user")
			return
		}
		respondWithJSON(w, 201, User{user.ID, user.CreatedAt, user.UpdatedAt, user.Email})
	})
}

func (cfg *apiConfig) login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type requestParams struct {
			Password string `json:"password"`
			Email    string `json:"email"`
		}
		type response struct {
			User
			Token        string `json:"token"`
			RefreshToken string `json:"refresh_token"`
		}

		decoder := json.NewDecoder(r.Body)
		params := requestParams{}
		err := decoder.Decode(&params)
		if err != nil {
			respondWithError(w, 500, "Error decoding login request body")
		}

		user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			respondWithError(w, 401, "Error getting user from database")
		}

		match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
		if err != nil {
			respondWithError(w, 500, "Error checking password match")
		}

		if !match {
			respondWithError(w, 401, "Unauthorized")
		} else {
			token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
			if err != nil {
				respondWithError(w, 500, "Error making JWT token")
			}
			refreshToken := auth.MakeRefreshToken()
			refreshParams := database.CreateRefreshTokenParams{
				Token:  refreshToken,
				UserID: user.ID,
			}
			_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), refreshParams)
			if err != nil {
				respondWithError(w, 500, "Error creating refresh token in database")
			}
			respondWithJSON(w, 200, response{
				User: User{
					ID:        user.ID,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
					Email:     user.Email,
				},
				Token:        token,
				RefreshToken: refreshToken,
			})
		}
	})
}

func (cfg *apiConfig) validateRefresh() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			Token string `json:"token"`
		}
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, 500, "Error getting bearer token")
		}
		refreshToken, err := cfg.dbQueries.GetRefreshByToken(r.Context(), token)
		if err != nil {
			respondWithError(w, 401, "Refresh token doesn't exist")
		}
		if refreshToken.RevokedAt.Valid {
			respondWithError(w, 401, "Refresh token has been revoked")
		}
		accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
		if err != nil {
			respondWithError(w, 500, "Error making JWT token")
		}
		respondWithJSON(w, 200, response{Token: accessToken})
	})
}

func (cfg *apiConfig) revokeRefresh() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, 500, "Error getting bearer token")
		}
		refreshToken, err := cfg.dbQueries.GetRefreshByToken(r.Context(), token)
		if err != nil {
			respondWithError(w, 401, "Refresh token doesn't exist")
		}
		err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken.Token)
		if err != nil {
			respondWithError(w, 500, "Error revoking refresh token")
		}
		w.WriteHeader(204)
	})
}
