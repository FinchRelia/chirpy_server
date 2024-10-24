package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/finchrelia/chirpy-server/internal/auth"
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
		log.Printf("Error retrieving user: %s", err)
	}

	err = auth.CheckPasswordHash(p.Password, loggedUser.HashedPassword)
	if err != nil {
		log.Printf("Incorrect email or password")
		w.WriteHeader(401)
		return
	}

	data, err := json.Marshal(User{
		ID:        loggedUser.ID,
		CreatedAt: loggedUser.CreatedAt,
		UpdatedAt: loggedUser.UpdatedAt,
		Email:     loggedUser.Email,
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
