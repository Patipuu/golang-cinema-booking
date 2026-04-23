package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := "5173"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	frontendDir := "frontend"
	if len(os.Args) > 2 {
		frontendDir = os.Args[2]
	}

	// Get absolute path to frontend directory
	absPath, err := filepath.Abs(frontendDir)
	if err != nil {
		log.Fatalf("Failed to get absolute path: %v", err)
	}

	// Create file server
	fs := http.FileServer(http.Dir(absPath))

	// Add CORS headers
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve files
		fs.ServeHTTP(w, r)
	})

	fmt.Printf("Serving frontend at http://localhost:%s\n", port)
	fmt.Printf("Frontend directory: %s\n", absPath)

	log.Fatal(http.ListenAndServe(":"+port, handler))
}
