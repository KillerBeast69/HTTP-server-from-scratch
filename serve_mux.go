package main

import (
	"fmt"
	"net/http"
)

func ServerMux() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fileServerHandler := http.FileServer(http.Dir("."))

	mux.Handle("/app/", http.StripPrefix("/app/", fileServerHandler))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	listenAndServeErr := server.ListenAndServe()
	if listenAndServeErr != nil {
		return fmt.Errorf("error starting server: %v", listenAndServeErr)
	}

	return nil
}
