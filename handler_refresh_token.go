package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/BlochLior/chirpy/internal/auth"
)

// handlerCheckRefreshToken
func (cfg *apiConfig) handlerCheckRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := extractAuthorizationFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", auth.ErrNoAuthHeaderIncluded)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil || user.ExpiresAt.Before(time.Now()) || user.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", auth.ErrNoAuthHeaderIncluded)
		return
	}
	jwtToken, _ := auth.MakeJWT(user.ID, cfg.jwtSecret)

	respondWithJSON(w, http.StatusOK, response{Token: jwtToken})
}

// handlerRevokeRefreshToken -
func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := extractAuthorizationFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized access", auth.ErrNoAuthHeaderIncluded)
		return
	}
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unsuccessful revoke", err)
		return
	}
	respondWithStatus(w, http.StatusNoContent)
}

func extractAuthorizationFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", auth.ErrNoAuthHeaderIncluded
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return token, nil
}
