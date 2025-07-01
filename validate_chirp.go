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

	type errorResponse struct {
		Error string `json:"error"`
	}

	type okResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	decoder := json.NewDecoder(request.Body)
	params := requestBody{}
	if err := decoder.Decode(&params); err != nil {
		errorJson, innerError := json.Marshal(errorResponse{Error: fmt.Sprintf("Error decoding parameters: %v", err)})
		if innerError != nil {
			fmt.Printf("Error encoding response: %v", innerError)
			return
		}
		writer.WriteHeader(500)
		writer.Write(errorJson)
		return
	}

	if len(params.Body) > 140 {
		errorJson, innerError := json.Marshal(errorResponse{Error: "Chirp is too long"})
		if innerError != nil {
			fmt.Printf("Error encoding response: %v", innerError)
			return
		}
		writer.WriteHeader(400)
		writer.Write(errorJson)
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