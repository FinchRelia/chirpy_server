package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/finchrelia/chirpy-server/internal/auth"
	"github.com/finchrelia/chirpy-server/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) Login(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	p := params{}
	err := decoder.Decode(&p)
	if err != nil {
		log.Printf("Incorrect email or password")
		w.WriteHeader(401)
		return
	}

	loggedUser, err := cfg.DB.GetUserByEmail(r.Context(), p.Email)
	if err != nil {
		log.Printf("Error retrieving user: %v", err)
	}

	err = auth.CheckPasswordHash(p.Password, loggedUser.HashedPassword)
	if err != nil {
		log.Printf("Incorrect email or password")
		w.WriteHeader(401)
		return
	}

	newJwt, err := auth.MakeJWT(loggedUser.ID, cfg.JWT)
	if err != nil {
		log.Printf("Error creating JWT: %s", newJwt)
		w.WriteHeader(500)
		return
	}
	newRefreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		w.WriteHeader(500)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     newRefreshToken,
		UserID:    loggedUser.ID,
		ExpiresAt: sql.NullTime{Time: time.Now().AddDate(0, 0, 60), Valid: true},
	}
	_, err = cfg.DB.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		log.Printf("Error adding refresh token to db: %s", err)
		w.WriteHeader(500)
		return
	}
	type loginResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		AccessToken  string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		ChirpyRed    bool      `json:"is_chirpy_red"`
	}

	data, err := json.Marshal(loginResponse{
		ID:           loggedUser.ID,
		CreatedAt:    loggedUser.CreatedAt,
		UpdatedAt:    loggedUser.UpdatedAt,
		Email:        loggedUser.Email,
		AccessToken:  newJwt,
		RefreshToken: newRefreshToken,
		ChirpyRed:    loggedUser.IsChirpyRed,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}
