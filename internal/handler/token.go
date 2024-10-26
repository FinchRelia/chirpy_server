package handler

import (
	"log"
	"net/http"

	"github.com/finchrelia/chirpy-server/internal/auth"
)

func (cfg *APIConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	dbUser, err := cfg.DB.GetUserFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	newToken, err := auth.MakeJWT(dbUser, cfg.JWT)
	if err != nil {
		log.Printf("Error creating new JWT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type tokenResponse struct {
		AccessToken string `json:"token"`
	}
	JsonResponse(w, http.StatusOK, tokenResponse{
		AccessToken: newToken,
	})
}

func (cfg *APIConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	err = cfg.DB.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Printf("Error revoking token in database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
