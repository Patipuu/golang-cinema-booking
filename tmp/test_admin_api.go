package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	// 1. Login to get token
	loginResp, err := http.Post("http://localhost:8082/api/v1/login", "application/json", io.NopCloser(bytes.NewBuffer([]byte(`{"email":"admin@gmail.com","password":"password123"}`))))
	if err != nil {
		log.Fatal(err)
	}
	defer loginResp.Body.Close()

	var loginData struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.NewDecoder(loginResp.Body).Decode(&loginData)
	token := loginData.Data.Token
	if token == "" {
		log.Fatal("Failed to get token")
	}
	fmt.Println("Token acquired:", token[:10]+"...")

	// 2. Call admin showtimes
	req, _ := http.NewRequest("GET", "http://localhost:8082/api/v1/admin/showtimes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(body))
}
