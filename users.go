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

	var timeToExpire time.Duration
	if req.ExpiresInSeconds == 0 || req.ExpiresInSeconds > 3600 {
		timeToExpire, err = time.ParseDuration("1h")
		if err != nil {
			respondWithError(writer, 500, fmt.Sprintf("Error parsing time: %v", err))
			return
		}
	} else {
		timeToExpire, err = time.ParseDuration(fmt.Sprintf("%ds", req.ExpiresInSeconds))
		if err != nil {
			respondWithError(writer, 500, fmt.Sprintf("Error parsing passed time: %v", err))
			return
		}
	}

	jwtToken, err := auth.MakeJWT(dbUser.ID, cfg.jwtSecret, timeToExpire)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating token: %v", err))
		return
	}

	res := resUser{
		ID: dbUser.ID.String(),
		CreatedAt: dbUser.CreatedAt.String(),
		UpdatedAt: dbUser.UpdatedAt.String(),
		Email: dbUser.Email,
		Token: jwtToken,
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
}

type reqBody struct {
	Email string `json:"email"`
	Password string `json:"password"`
	ExpiresInSeconds int `json:"expires_in_seconds"`
}