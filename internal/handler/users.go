package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/finchrelia/chirpy-server/internal/auth"
	"github.com/finchrelia/chirpy-server/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	ChirpyRed      bool      `json:"is_chirpy_red"`
}

func (cfg *APIConfig) CreateUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
	}
	newDBUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user %s: %v", params.Email, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	newId := newDBUser.ID
	JsonResponse(w, http.StatusCreated, User{
		ID:        newId,
		CreatedAt: newDBUser.CreatedAt,
		UpdatedAt: newDBUser.UpdatedAt,
		Email:     newDBUser.Email,
		ChirpyRed: newDBUser.IsChirpyRed,
	})
}

func (cfg *APIConfig) UpdateUsers(w http.ResponseWriter, r *http.Request) {
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

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	credentialsQueryParams := database.UpdateUserCredentialsParams{
		ID:             userId,
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	updatedCredentials, err := cfg.DB.UpdateUserCredentials(r.Context(), credentialsQueryParams)
	if err != nil {
		log.Printf("Error updating user credentials: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JsonResponse(w, http.StatusOK, User{
		ID:        userId,
		CreatedAt: updatedCredentials.CreatedAt,
		UpdatedAt: updatedCredentials.UpdatedAt,
		Email:     updatedCredentials.Email,
		ChirpyRed: updatedCredentials.IsChirpyRed,
	})
}

func (cfg *APIConfig) SubscribeUser(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error extracting apiKey: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	type parameters struct {
		Event string            `json:"event"`
		Data  map[string]string `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	paramsUserIdString := params.Data["user_id"]
	paramsUserId, err := uuid.Parse(paramsUserIdString)
	if err != nil {
		log.Printf("Specified user_id is not a valid UUID: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = cfg.DB.UpgradeUser(r.Context(), paramsUserId)
	if err != nil {
		log.Printf("No user matches user_id given: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
