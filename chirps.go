package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zig-gy/chirpy/internal/auth"
	"github.com/zig-gy/chirpy/internal/database"
)

func (cfg *apiConfig) createChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	token, err := auth.GetBearerToken(request.Header)
	if err != nil {
		respondWithError(writer, 400, fmt.Sprintf("Header authorization not found: %v", err))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(writer, 401, fmt.Sprintf("Not authorized: %v", err))
		return
	}

	type reqChirp struct {
		Body string `json:"body"`
	}

	type resChirp struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(request.Body)
	req := reqChirp{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error decoding paramters: %v", err))
		return
	}

	cleanBody, err := validateChirp(req.Body)
	if err != nil {
		respondWithError(writer, 401, fmt.Sprintf("Error in chirp body: %v", err))
		return
	}

	responseChirp, err := cfg.queries.CreateChirp(context.Background(), database.CreateChirpParams{
		UserID: userID,
		Body: cleanBody,
	})
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating chirp in database: %v", err))
		return
	}

	res := resChirp{
		ID: responseChirp.ID,
		CreatedAt: responseChirp.CreatedAt,
		UpdatedAt: responseChirp.UpdatedAt,
		Body: responseChirp.Body,
		UserID: responseChirp.UserID,
	}
	resBytes, err := json.Marshal(res)
	if err != nil {
		fmt.Printf("Error encoding response: %v", err)
		return
	}

	writer.WriteHeader(201)
	writer.Write(resBytes)
}

func (cfg *apiConfig) getChirps(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	chirps, err := cfg.queries.GetChirps(context.Background())
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error getting from database: %v", err))
		return
	}

	resChirps := make([]resChirp, 0)
	for _, chirp := range chirps {
		resChirps = append(resChirps, resChirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	resBytes, err := json.Marshal(resChirps)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating json: %v", err))
		return
	}

	writer.WriteHeader(200)
	writer.Write(resBytes)
}

func (cfg *apiConfig) getOneChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	param := request.PathValue("ChirpID")
	chirpID, err := uuid.Parse(param)
	if err != nil {
		respondWithError(writer, 400, fmt.Sprintf("Can't parse ID: %v", err))
		return
	}

	dbChirp, err := cfg.queries.GetChirpByID(context.Background(), chirpID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			respondWithError(writer, 404, "Chirp not found")
		} else {
		respondWithError(writer, 500, fmt.Sprintf("Error getting chirp from database: %v", err))
		}
		return
	}

	res := resChirp{
		ID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
	}

	resBytes, err := json.Marshal(res)
	if err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error creating response: %v", err))
		return
	}

	writer.WriteHeader(200)
	writer.Write(resBytes)
}

type resChirp struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

func validateChirp(chirp string) (string, error) {
		if len(chirp) > 140 {
		return "", fmt.Errorf("chirp too long")
	}

	clean := replaceBadWords(chirp)
	return clean, nil
}