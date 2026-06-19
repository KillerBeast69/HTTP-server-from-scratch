package main

import (
	"chirpy/internal/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
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

	if cfg.Platform != "dev" {
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
		return
	}

	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		respond_with_error(w, 500, "Error deleting users")
		return
	}

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

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err := decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		respond_with_error(w, 500, "Error creating user")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	respond_with_json(w, 201, response{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	})
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {

	type payload struct {
		Body    string    `json:"body"`
		User_Id uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := payload{}

	err := decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "Error decoding jason response")
		return
	}

	if len(params.Body) > 140 {
		respond_with_error(w, 400, "Chirp is too long")
		return
	}

	censored_body := censor(params.Body)

	//respond_with_json(w, 200, response{CleanedBody: censored_body})

	chirp_resp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   censored_body,
		UserID: params.User_Id,
	})
	if err != nil {
		respond_with_error(w, 500, "failed to create chirp")
		return
	}

	type response struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	respond_with_json(w, 201, response{
		ID:        chirp_resp.ID,
		CreatedAt: chirp_resp.CreatedAt,
		UpdatedAt: chirp_resp.UpdatedAt,
		Body:      chirp_resp.Body,
		UserID:    chirp_resp.UserID,
	})

}
