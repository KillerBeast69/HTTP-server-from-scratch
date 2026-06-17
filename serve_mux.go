package main

import (
	"fmt"
	"net/http"
)

func ServerMux() error {
	mux := http.NewServeMux()

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