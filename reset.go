package main

import (
	"net/http"
)

func (cfg *apiConfig) serveReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
}
