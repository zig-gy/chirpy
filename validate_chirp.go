package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func validateChirp(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Add("Content-Type", "application/json")

	type requestBody struct {
		Body string `json:"body"`
	}

	type okResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestBody{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(writer, 500,  fmt.Sprintf("Error decoding parameters: %v", err))
		return
	}

	if len(params.Body) > 140 {
		respondWithError(writer, 400, "Chirp is too long")
		return
	}

	clean := replaceBadWords(params.Body)
	res, err := json.Marshal(okResponse{CleanedBody: clean})
	if err != nil {
		fmt.Printf("Error encoding response: %v", err)
		return
	}

	writer.WriteHeader(200)
	writer.Write(res)
}