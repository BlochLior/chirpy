package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/BlochLior/chirpy/internal/auth"
	"github.com/BlochLior/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds,omitempty"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserFromEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "email does not exist in database", err)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect password or email", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate token", err)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	refreshTokenExpiry := time.Now().Add(time.Hour * 24 * 60)
	err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		ExpiresAt: refreshTokenExpiry,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "encountered error inserting refresh token to database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerUpdateLoginInfo(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		NewEmail    string `json:"email"`
		NewPassword string `json:"password"`
	}
	type response struct {
		User
	}

	jwtToken, err := extractAuthorizationFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no token found in request", err)
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid jwtToken", err)
		return
	}

	var params parameters
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't hash new password", err)
		return
	}

	updatedUser, err := cfg.db.UpdateUserEmailAndPassword(r.Context(), database.UpdateUserEmailAndPasswordParams{
		ID:             userID,
		Email:          params.NewEmail,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "unable to update user in the database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        updatedUser.ID,
			CreatedAt: updatedUser.CreatedAt,
			UpdatedAt: updatedUser.UpdatedAt,
			Email:     updatedUser.Email,
		},
	})

}
