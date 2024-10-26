package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/finchrelia/chirpy-server/internal/auth"
	"github.com/finchrelia/chirpy-server/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) chirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Content string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %s", err)
		w.WriteHeader(401)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		w.WriteHeader(401)
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	} else if len(params.Content) > 140 {
		type errorVals struct {
			Data string `json:"error"`
		}

		respBody := errorVals{
			Data: "Chirp is too long",
		}
		dat, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(dat)
	} else {
		cleanedData := cleanText(params.Content)
		chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   cleanedData,
			UserID: userId,
		})
		if err != nil {
			log.Printf("Error creating chirp: %s", err)
			w.WriteHeader(500)
			return
		}
		dat, err := json.Marshal(Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		w.Write(dat)
	}
}

func cleanText(s string) string {
	splittedString := strings.Split(s, " ")
	for idx, element := range splittedString {
		switch strings.ToLower(element) {
		case
			"kerfuffle",
			"sharbert",
			"fornax":
			splittedString[idx] = "****"
		}
	}
	return strings.Join(splittedString, " ")
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps(r.Context())
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		w.WriteHeader(500)
		return
	}
	newChirps := []Chirp{}
	for _, chirp := range chirps {
		newChirps = append(newChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	dat, err := json.Marshal(newChirps)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	idFromQuery := r.PathValue("chirpID")
	id, err := uuid.Parse(idFromQuery)
	if err != nil {
		log.Printf("Not a valid ID: %s", err)
		w.WriteHeader(500)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("Error getting chirp: %s", err)
		w.WriteHeader(500)
		return
	}
	dat, err := json.Marshal(Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %s", err)
		w.WriteHeader(401)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Invalid JWT: %s", err)
		w.WriteHeader(401)
		return
	}
	idFromQuery := r.PathValue("chirpID")

	id, err := uuid.Parse(idFromQuery)
	if err != nil {
		log.Printf("Not a valid ID: %s", err)
		w.WriteHeader(500)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("Error getting chirp: %s", err)
		w.WriteHeader(500)
		return
	}
	if chirp.UserID != userId {
		log.Printf("User %s not allowed to delete chirp owned by %s", userId, chirp.UserID)
		w.WriteHeader(403)
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		log.Printf("Error deleting chirp: %v", err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}
