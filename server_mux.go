package main

import (
	"fmt"
	"net/http"
)

func ServerMux() error {
	mux := http.NewServeMux()
	apiCfg := apiConfig{}

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)

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
