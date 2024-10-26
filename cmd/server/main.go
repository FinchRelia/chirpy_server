package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/finchrelia/chirpy-server/internal/database"
	"github.com/finchrelia/chirpy-server/internal/handler"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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
	polkaKey := os.Getenv("POLKA_KEY")
	if polkaKey == "" {
		log.Fatalf("Empty POLKA_KEY env var!")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to db: %v", err)
	}
	apiCfg := &handler.APIConfig{
		FileserverHits: atomic.Int32{},
		DB:             database.New(db),
		Platform:       platform,
		JWT:            jwtSecret,
		PolkaKey:       polkaKey,
	}

	mux := http.NewServeMux()
	fsHandler := apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fsHandler)
	mux.HandleFunc("GET /api/healthz", handler.Readiness)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.SubscribeUser)

	mux.HandleFunc("POST /api/login", apiCfg.Login)
	mux.HandleFunc("POST /api/refresh", apiCfg.RefreshToken)
	mux.HandleFunc("POST /api/revoke", apiCfg.RevokeToken)

	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.Metrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.Reset))

	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirps)
	mux.HandleFunc("POST /api/chirps", apiCfg.ChirpsCreate)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.DeleteChirp)

	mux.HandleFunc("POST /api/users", apiCfg.CreateUsers)
	mux.HandleFunc("PUT /api/users", apiCfg.UpdateUsers)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
