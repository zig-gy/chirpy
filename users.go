package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zig-gy/chirpy/internal/auth"
	"github.com/zig-gy/chirpy/internal/database"
)

func (cfg *apiConfig) createUsers(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	req := reqBody{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error decoding parameters: %v", err))
		return
	}

	hashPass, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error hashing password: %v", err))
		return
	}

	response, err := cfg.queries.CreateUser(context.Background(), database.CreateUserParams{
		Email: req.Email,
		HashedPassword: hashPass,
	})
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error saving user: %v", err))
		return
	}

	res := resUser{
		ID: response.ID.String(),
		CreatedAt: response.CreatedAt.String(),
		UpdatedAt: response.UpdatedAt.String(),
		Email: response.Email,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("Error encoding response: %v", err)
		return
	}

	writer.WriteHeader(201)
	writer.Write(resBytes)
}

func (cfg *apiConfig) login(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	decoder := json.NewDecoder(request.Body)
	req := reqBody{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error decoding request: %v", err))
		return
	}

	dbUser, err := cfg.queries.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		respondWithError(writer, 401, "Incorrect email or password")
		return
	}

	if err := auth.CheckPasswordHash(req.Password, dbUser.HashedPassword); err != nil {
		respondWithError(writer, 401, "Incorrect email or password")
		return
	}

	timeToExpire, err := time.ParseDuration("1h")
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error parsing time: %v", err))
		return
	}

	jwtToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, timeToExpire)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating token: %v", err))
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating refresh token: %v", err))
		return
	}
	
	refreshExpire, err := time.ParseDuration("1440h")
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error parsing refresh expire time: %v", err))
		return
	}

	refreshObject, err := cfg.queries.CreateRefreshToken(context.Background(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: dbUser.ID,
		ExpiresAt: time.Now().Add(refreshExpire),
	})
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating refresh token: %v", err))
		return
	}

	res := resUser{
		ID: dbUser.ID.String(),
		CreatedAt: dbUser.CreatedAt.String(),
		UpdatedAt: dbUser.UpdatedAt.String(),
		Email: dbUser.Email,
		Token: jwtToken,
		RefreshToken: refreshObject.Token,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error encoding response: %v", err))
		return
	}

	writer.WriteHeader(200)
	writer.Write(resBytes)
}

func (cfg *apiConfig) refresh(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	type resRefresh struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 400, fmt.Sprintf("Header authorization not found: %v", err))
		return
	}

	expiringAt, err := cfg.queries.GetRefreshTokenExpiresAt(context.Background(), refreshToken)
	if err != nil {
		respondWithError(writer, 401, fmt.Sprintf("Refresh token does not exist: %v", err))
		return
	}

	if time.Now().Compare(expiringAt) > 0 {
		if err := cfg.queries.RevokeRefreshToken(context.Background(), refreshToken); err != nil {
			respondWithError(writer, 500, fmt.Sprintf("Error revoking token: %v", err))
			return
		}
		respondWithError(writer, 401, fmt.Sprintf("Refresh token revoked: %v", err))
		return
	}	

	userID, err := cfg.queries.GetUserFromRefreshToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error getting user id: %v", err))
		return
	}

	expiringTime, err := time.ParseDuration("1h")
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error parsing time: %v", err))
		return
	}

	jwtToken, err := auth.MakeJWT(userID, cfg.jwtSecret, expiringTime)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating jwt token: %v", err))
		return
	}

	res := resRefresh{
		Token: jwtToken,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error encoding: %v", err))
		return
	}

	writer.WriteHeader(200)
	writer.Write(resBytes)
}

func (cfg *apiConfig) revoke(writer http.ResponseWriter, request *http.Request) {
	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 400, fmt.Sprintf("Authorization header not found: %v", err))
		return
	}

	if err := cfg.queries.RevokeRefreshToken(context.Background(), token); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error revoking token: %v", err))
		return
	}

	writer.WriteHeader(204)
}

func (cfg *apiConfig) updateUser(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	type reqBody struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	req := reqBody{}
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(&req); err != nil {
		respondWithError(writer, 400, fmt.Sprintf("Could not read the request body: %v", err))
		return
	}

	hashedPass, err := auth.HashPassword(req.Password)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error hashing new password: %v", err))
		return
	}

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 401, fmt.Sprintf("Access denied: %v", err))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(writer, 401, fmt.Sprintf("Access denied: %v", err))
		return
	}

	newUser, err := cfg.queries.UpdateUserEmailAndPassword(context.Background(), database.UpdateUserEmailAndPasswordParams{
		ID: userID,
		Email: req.Email,
		HashedPassword: hashedPass,
	})
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error updating user: %v", err))
		return
	}
	
	res := resUser{
		ID: newUser.ID.String(),
		CreatedAt: newUser.CreatedAt.String(),
		UpdatedAt: newUser.UpdatedAt.String(),
		Email: newUser.Email,
		Token: token,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error encoding response: %v", err))
		return
	}

	writer.WriteHeader(200)
	writer.Write(resBytes)
}

type resUser struct {
	ID string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email string `json:"email"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type reqBody struct {
	Email string `json:"email"`
	Password string `json:"password"`
}