package main

import (
	"encoding/json"
	"net/http"

	"github.com/BlochLior/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpgrade(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No api key found in header", err)
		return
	}
	if cfg.polkaKey != apiKey {
		respondWithError(w, http.StatusUnauthorized, "unmatching api key", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to decode parameters", err)
		return
	}
	if params.Event != "user.upgraded" {
		respondWithStatus(w, http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "userID not valid", err)
		return
	}
	_, err = cfg.db.GetUserFromID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "userID not found in db", err)
		return
	}

	err = cfg.db.UpgradeUser(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "unable to upgrade user", err)
		return
	}
	respondWithStatus(w, http.StatusNoContent)
}
