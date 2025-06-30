package main

import "net/http"


func (cfg *apiConfig) reset(writer http.ResponseWriter, request *http.Request) {
	cfg.fileserverHits.Store(0)
	writer.Header().Add("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(200)
	writer.Write([]byte("Metrics reset"))
}