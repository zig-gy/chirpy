package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) metrics(writer http.ResponseWriter, request *http.Request) {
	hits := cfg.fileserverHits.Load()
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	
	message := fmt.Sprintf("Hits: %d", hits)
	writer.Write([]byte(message))
}