package handler

import (
	"log"
	"net/http"
)

func (cfg *APIConfig) Reset(w http.ResponseWriter, r *http.Request) {
	cfg.FileserverHits.Store(0)
	if cfg.Platform != "dev" {
		log.Printf("Invalid %s platform !", cfg.Platform)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	_, err := cfg.DB.DeleteUser(r.Context())
	if err != nil {
		log.Printf("Error deleting users: %v", err)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}
