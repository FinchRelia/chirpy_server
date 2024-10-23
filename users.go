package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
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
	newDBUser, err := cfg.DB.CreateUser(r.Context(), params.Email)
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
