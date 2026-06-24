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
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	respond_with_json(w, 201, response{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
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

	authorIDString := r.URL.Query().Get("author_id")
	SortType := r.URL.Query().Get("sort")

	var allChirps []database.Chirp
	var err error

	if authorIDString != "" {
		authorID, err := uuid.Parse(authorIDString)
		if err != nil {
			respond_with_error(w, 400, "invalid author ID")
			return
		}

		allChirps, err = cfg.DB.GetChirpsByAuthor(r.Context(), authorID)
	} else {
		if SortType == "desc" {
			allChirps, err = cfg.DB.GetAllChirpsDesc(r.Context())
		} else {
			allChirps, err = cfg.DB.GetAllChirps(r.Context())
		}
	}
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
		//Expires  int    `json:"expires_in_seconds"`
	}

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

	/*if params.Expires == 0 || params.Expires > 3600 {
		params.Expires = 3600
	}*/

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.secret,
		time.Hour,
	)
	if err != nil {
		respond_with_error(w, 500, "could not create JWT")
		return
	}

	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour),
		UserID:    user.ID,
	})
	if err != nil {
		respond_with_error(w, 500, "could not save refresh token")
		return
	}

	type user_resource struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	respond_with_json(w, 200, user_resource{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed.Bool,
	})

	// how do I store the expiration time in the data base?

}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond_with_error(w, 401, "missing token")
		return
	}

	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respond_with_error(w, 401, "invalid or expired refresh token")
		return
	}

	newAccessToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respond_with_error(w, 500, "could not create new JWT")
		return
	}

	respond_with_json(w, 200, map[string]string{
		"token": newAccessToken,
	})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond_with_error(w, 401, "missing token")
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respond_with_error(w, 500, "could not revoke refresh token")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond_with_error(w, 401, "missing token")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respond_with_error(w, 401, "invalid or expired token")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err = decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "failed to decode json")
		return
	}

	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respond_with_error(w, 500, "failed to hash password")
		return
	}

	user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashed_password,
		ID:             userID,
	})
	if err != nil {
		respond_with_error(w, 500, "failed to update user in database")
		return
	}

	type response struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	//user.HashedPassword = hashed_password
	respond_with_json(w, 200, response{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed.Bool,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respond_with_error(w, 401, "missing token")
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.secret)
	if err != nil {
		respond_with_error(w, 401, "invalid or expired token")
		return
	}

	chirpIDString := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respond_with_error(w, 400, "invalid chirp ID")
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)
	if err != nil {
		respond_with_error(w, 404, "chirp not found")
		return
	}

	if chirp.UserID != userID {
		respond_with_error(w, 403, "you are not authorized to delete this chirp")
		return
	}

	err = cfg.DB.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respond_with_error(w, 500, "failed to delete chirp")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerChirpyRed(w http.ResponseWriter, r *http.Request) {

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != cfg.PolkaKey {
		respond_with_error(w, 401, "unauthorized access")
		return
	}

	type request struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := request{}

	err = decoder.Decode(&params)
	if err != nil {
		respond_with_error(w, 500, "failed to decode json")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	_, err = cfg.DB.Chirpyred(r.Context(), params.Data.UserID)
	if err != nil {
		respond_with_error(w, 404, "user not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
