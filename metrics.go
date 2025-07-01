package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) metrics(writer http.ResponseWriter, request *http.Request) {
	hits := cfg.fileserverHits.Load()
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(200)
	
	message := fmt.Sprintf(`
	<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>
	`, hits)
	writer.Write([]byte(message))
}