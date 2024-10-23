package main

import (
	"log"
	"net/http"
)

func (cfg *apiConfig) serveReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	if cfg.Platform != "dev" {
		log.Printf("Invalid %s platform !", cfg.Platform)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, err := cfg.DB.DeleteUser(r.Context())
	if err != nil {
		log.Printf("Error deleting users: %s", err)
	}
}
