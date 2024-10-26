package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
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

func (cfg *APIConfig) ChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Content string `json:"body"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Invalid JWT: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cleanedChirp, err := cleanChirp(params.Content)
	if err != nil {
		type errorResponse struct {
			Error error `json:"error"`
		}
		JsonResponse(w, http.StatusBadRequest, errorResponse{Error: err})
	}
	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedChirp,
		UserID: userId,
	})
	if err != nil {
		log.Printf("Error creating chirp: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JsonResponse(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func cleanChirp(s string) (string, error) {
	const maxChirpLength = 140
	if len(s) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

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
	return strings.Join(splittedString, " "), nil
}

func (cfg *APIConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps := []database.Chirp{}
	authorIdString := r.URL.Query().Get("author_id")
	if authorIdString != "" {
		authorId, err := uuid.Parse(authorIdString)
		if err != nil {
			log.Printf("Incorrect author ID: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		dbChirps, err = cfg.DB.GetChirpsByUserid(r.Context(), authorId)
		if err != nil {
			log.Printf("No chirp for user given: %v", err)
		}
	} else {
		var err error
		dbChirps, err = cfg.DB.GetChirps(r.Context())
		if err != nil {
			log.Printf("Error getting chirps: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	chirps := []Chirp{}
	for _, chirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	sortOrder := r.URL.Query().Get("sort")
	sort.Slice(chirps, func(i, j int) bool {
		if sortOrder == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		// Defaults to asc
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})
	JsonResponse(w, http.StatusOK, chirps)
}

func (cfg *APIConfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	idFromQuery := r.PathValue("chirpID")
	id, err := uuid.Parse(idFromQuery)
	if err != nil {
		log.Printf("Not a valid ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("Error getting chirp: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JsonResponse(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *APIConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("Error extracting token: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.JWT)
	if err != nil {
		log.Printf("Invalid JWT: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	idFromQuery := r.PathValue("chirpID")

	id, err := uuid.Parse(idFromQuery)
	if err != nil {
		log.Printf("Not a valid ID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	chirp, err := cfg.DB.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("Error getting chirp: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if chirp.UserID != userId {
		log.Printf("User %s not allowed to delete chirp owned by %s", userId, chirp.UserID)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     id,
		UserID: userId,
	})
	if err != nil {
		log.Printf("Error deleting chirp: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
