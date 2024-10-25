package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/finchrelia/chirpy-server/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	Platform       string
	JWT            string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalf("Empty dbURL !")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatalf("Empty PLATFORM env var!")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatalf("Empty JWT_SECRET env var!")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to db: %s", err)
	}
	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		DB:             database.New(db),
		Platform:       platform,
		JWT:            jwtSecret,
	}
	mux := http.NewServeMux()
	fsHandler := apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fsHandler)
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, req *http.Request) {
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.serveMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.serveReset))
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirps)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsCreate)
	mux.HandleFunc("POST /api/users", apiCfg.createUsers)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirp)
	mux.HandleFunc("POST /api/login", apiCfg.Login)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeToken)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
