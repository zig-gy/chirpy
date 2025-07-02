package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) createUsers(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	type reqBody struct {
		Email string `json:"email"`
	}

	type resUser struct {
		ID string `json:"id"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(request.Body)
	req := reqBody{}
	if err := decoder.Decode(&req); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error decoding parameters: %v", err))
		return
	}

	response, err := cfg.queries.CreateUser(context.Background(), req.Email)
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