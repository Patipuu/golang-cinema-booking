package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Serve static files from the "frontend" directory at project root.
	frontendDir := filepath.Join(".", "frontend")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(frontendDir)))
	log.Printf("Serving frontend from %s\n", frontendDir)
	addr := ":5173"
	log.Printf("Serving frontend from %s on http://localhost%s\n", frontendDir, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

