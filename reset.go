package main

import (
	"context"
	"fmt"
	"net/http"
)


func (cfg *apiConfig) reset(writer http.ResponseWriter, request *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(writer, 403, "Can't reset data in production platform")
		return
	}
	
	if err := cfg.queries.ResetUsers(context.Background()); err != nil {
		respondWithError(writer, 500, fmt.Sprintf("Error resetting users in database: %v", err))
		return
	}

	cfg.fileserverHits.Store(0)
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("Metrics reset"))
}