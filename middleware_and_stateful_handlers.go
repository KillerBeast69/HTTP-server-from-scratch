package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	html_template := `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`

	w.Write([]byte(fmt.Sprintf(html_template, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
	w.Write([]byte("Hits counter reset to 0"))
}

func (cfg *apiConfig) handlerValidate(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err := decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "Something went wrong")
		return
	}

	if len(params.Body) > 140 {
		respond_with_error(w, 400, "Chirp is too long")
		return
	}

	censored_body := censor(params.Body)

	type response struct {
		CleanedBody string `json:"cleaned_body"`
	}

	respond_with_json(w, 200, response{CleanedBody: censored_body})

}
