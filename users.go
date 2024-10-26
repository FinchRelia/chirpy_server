package main

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

func (cfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
	}
	newDBUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user %s: %s", params.Email, err)
		w.WriteHeader(500)
		return
	}
	newId := newDBUser.ID
	newUser := User{
		ID:        newId,
		CreatedAt: newDBUser.CreatedAt,
		UpdatedAt: newDBUser.UpdatedAt,
		Email:     newDBUser.Email,
		ChirpyRed: newDBUser.IsChirpyRed,
	}
	dat, err := json.Marshal(newUser)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)
}

func (cfg *apiConfig) updateUsers(w http.ResponseWriter, r *http.Request) {
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

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	defer r.Body.Close()

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
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
		w.WriteHeader(500)
		return
	}
	data, err := json.Marshal(User{
		ID:        userId,
		CreatedAt: updatedCredentials.CreatedAt,
		UpdatedAt: updatedCredentials.UpdatedAt,
		Email:     updatedCredentials.Email,
		ChirpyRed: updatedCredentials.IsChirpyRed,
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

func (cfg *apiConfig) subscribeUser(w http.ResponseWriter, r *http.Request) {
	_, err := auth.GetAPIKey(r.Header)
	if err != nil {
		log.Printf("Error extracting apiKey: %s", err)
		w.WriteHeader(401)
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
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(400)
		return
	}
	defer r.Body.Close()

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	paramsUserIdString := params.Data["user_id"]
	paramsUserId, err := uuid.Parse(paramsUserIdString)
	if err != nil {
		log.Printf("Specified user_id is not a valid UUID: %v", err)
		w.WriteHeader(400)
		return
	}
	err = cfg.DB.UpgradeUser(r.Context(), paramsUserId)
	if err != nil {
		log.Printf("No user matches user_id given: %v", err)
		w.WriteHeader(404)
		return
	}
	w.WriteHeader(204)
}
