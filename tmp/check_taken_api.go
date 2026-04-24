package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	showtimeID := "2c83e286-3697-41f0-b5b7-43b3403a2a1b" // From previous test
	resp, err := http.Get("http://localhost:8082/api/v1/seats/showtime/" + showtimeID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Status: %d\nBody: %s\n", resp.StatusCode, string(body))
}
