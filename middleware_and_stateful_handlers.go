package main

import (
	"chirpy/internal/auth"
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
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err := decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "Something went wrong")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respond_with_error(w, 500, "failed to hash password")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
	})
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

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond_with_error(w, 401, "Missing authentication token")
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respond_with_error(w, 401, "invalid authentication token")
		return
	}

	type request struct {
		Body string `json:"body"`
	}

	// what do I do here
	decoder := json.NewDecoder(r.Body)
	params := request{}

	err = decoder.Decode(&params)
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
		UserID: userID,
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

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {

	allChirps, err := cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		respond_with_error(w, 500, "failed to get all chirps")
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	chirps := []Chirp{}

	for _, aChirp := range allChirps {
		chirps = append(chirps, Chirp{
			ID:        aChirp.ID,
			CreatedAt: aChirp.CreatedAt,
			UpdatedAt: aChirp.UpdatedAt,
			Body:      aChirp.Body,
			UserId:    aChirp.UserID,
		})
	}

	respond_with_json(w, 200, chirps)

}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {

	chirpIDString := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respond_with_error(w, 400, "invalid chirp ID")
		return
	}

	chirp_resp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		respond_with_error(w, 404, "chirp not found")
		return
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	respond_with_json(w, 200, Chirp{
		ID:        chirp_resp.ID,
		CreatedAt: chirp_resp.CreatedAt,
		UpdatedAt: chirp_resp.UpdatedAt,
		Body:      chirp_resp.Body,
		UserId:    chirp_resp.UserID,
	})

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Password string `json:"password"`
		Email    string `json:"email"`
		Expires  int    `json:"expires_in_seconds"`
	}

	// how do I make expires_in_seconds optional

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err := decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "failed to decode json")
		return
	}

	user, err := cfg.DB.GetUser(r.Context(), params.Email)
	if err != nil {
		respond_with_error(w, 401, "incorrect email or password")
		return
	}

	password_is_valid, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !password_is_valid {
		respond_with_error(w, 401, "incorrect email or password")
		return
	}

	if params.Expires == 0 || params.Expires > 3600 {
		params.Expires = 3600
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Duration(params.Expires)*time.Second,
	)
	if err != nil {
		respond_with_error(w, 500, "could not create JWT")
	}

	type user_resource struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}

	respond_with_json(w, 200, user_resource{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})

}
