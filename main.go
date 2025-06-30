package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Server starting on http://localhost:8080")
	svMux := http.NewServeMux()
	sv := http.Server {
		Handler: svMux,
		Addr: ":8080",
	}

	svMux.Handle("/", http.FileServer(http.Dir(".")))

	sv.ListenAndServe()
}
