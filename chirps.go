package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/zig-gy/chirpy/internal/database"
)

func (cfg *apiConfig) createChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	type reqChirp struct {
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
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
		UserID: req.UserID,
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

	type resChirp struct {
		ID uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body string `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

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

func validateChirp(chirp string) (string, error) {
		if len(chirp) > 140 {
		return "", fmt.Errorf("chirp too long")
	}

	clean := replaceBadWords(chirp)
	return clean, nil
}