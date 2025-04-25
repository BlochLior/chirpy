package main

import (
	"net/http"
	"time"

	"github.com/BlochLior/chirpy/internal/auth"
)

// handlerRefresh
func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't find token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "couldn't get user for refresh token", err)
		return
	}
	jwtToken, _ := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)

	respondWithJSON(w, http.StatusOK, response{Token: jwtToken})
}

// handlerRevoke -
func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't find token", err)
		return
	}
	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke session", err)
		return
	}
	respondWithStatus(w, http.StatusNoContent)
}
