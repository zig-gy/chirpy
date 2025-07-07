package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/zig-gy/chirpy/internal/database"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("JWTSECRET")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		fmt.Printf("Error connecting to database: %v", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		queries: dbQueries,
		platform: platform,
		jwtSecret: secret,
	}

	fmt.Println("Server starting on http://localhost:8080")
	svMux := http.NewServeMux()
	sv := http.Server {
		Handler: svMux,
		Addr: ":8080",
	}

	fileserver := http.StripPrefix("/app",http.FileServer(http.Dir(".")))

	svMux.Handle("/app/", cfg.middlewareMetricsInc(fileserver))
	svMux.HandleFunc("GET /api/healthz", healthz)
	svMux.HandleFunc("GET /admin/metrics", cfg.metrics)
	svMux.HandleFunc("POST /admin/reset", cfg.reset)
	svMux.HandleFunc("POST /api/users", cfg.createUsers)
	svMux.HandleFunc("POST /api/chirps", cfg.createChirp)
	svMux.HandleFunc("GET /api/chirps", cfg.getChirps)
	svMux.HandleFunc("GET /api/chirps/{ChirpID}", cfg.getOneChirp)
	svMux.HandleFunc("POST /api/login", cfg.login)
	svMux.HandleFunc("POST /api/refresh", cfg.refresh)
	svMux.HandleFunc("POST /api/revoke", cfg.revoke)
	
	sv.ListenAndServe()
}

type apiConfig struct {
	fileserverHits atomic.Int32
	queries *database.Queries
	platform string
	jwtSecret string
}
