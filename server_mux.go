package main

import (
	"chirpy/internal/database"
	"fmt"
	"net/http"
)

func ServerMux(db *database.Queries, platform string) error {
	mux := http.NewServeMux()

	apiCfg := apiConfig{
		DB:       db,
		Platform: platform,
	}

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fileServerHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServerHandler))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	//mux.HandleFunc("POST /api/validate_chirp", apiCfg.handlerValidate)
	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)

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
