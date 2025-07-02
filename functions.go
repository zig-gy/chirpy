package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)		
	})
}

func replaceBadWords(chirp string) string {
	words := strings.Split(chirp, " ")
	badWords := []string{
		"kerfuffle",
		"sharbert",
		"fornax",
	}

	for i, word := range words {
		lowered := strings.ToLower(word)
		for _, badWord := range badWords {
			if lowered == badWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	errorJson, innerError := json.Marshal(errorResponse{Error: msg})
	if innerError != nil {
		fmt.Printf("Error encoding response: %v", innerError)
		return
	}
	w.WriteHeader(code)
	w.Write(errorJson)
}