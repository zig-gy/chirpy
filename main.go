package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	fmt.Println("Server starting on http://localhost:8080")
	svMux := http.NewServeMux()
	sv := http.Server {
		Handler: svMux,
		Addr: ":8080",
	}

	fileserver := http.StripPrefix("/app",http.FileServer(http.Dir(".")))

	svMux.Handle("/app/", cfg.middlewareMetricsInc(fileserver))
	svMux.HandleFunc("GET /healthz", healthz)
	svMux.HandleFunc("GET /metrics", cfg.metrics)
	svMux.HandleFunc("POST /reset", cfg.reset)

	sv.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
}