package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/finchrelia/chirpy-server/internal/auth"
)

func (cfg *apiConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %s", err)
		w.WriteHeader(401)
		return
	}

	dbUser, err := cfg.DB.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		w.WriteHeader(401)
		return
	}
	newToken, err := auth.MakeJWT(dbUser, cfg.JWT)
	if err != nil {
		log.Printf("Error creating new JWT: %v", err)
		w.WriteHeader(500)
		return
	}
	type tokenResponse struct {
		AccessToken string `json:"token"`
	}

	data, err := json.Marshal(tokenResponse{
		AccessToken: newToken,
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

func (cfg *apiConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %s", err)
		w.WriteHeader(401)
		return
	}
	err = cfg.DB.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error revoking token in database: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(204)
}
