package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app", http.StripPrefix("/app", fileServer))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets/", http.FileServer(http.Dir("./assets/"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
