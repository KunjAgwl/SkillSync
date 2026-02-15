package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080"
	client := &http.Client{Timeout: 5 * time.Second}

	// Wait for server to be up
	time.Sleep(2 * time.Second)

	email := fmt.Sprintf("test%d@example.com", time.Now().Unix())
	password := "password123"

	// 1. Register
	fmt.Println("Testing /register...")
	regBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err := client.Post(baseURL+"/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		fmt.Printf("Register failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Register Status: %s\nResponse: %s\n", resp.Status, string(body))

	if resp.StatusCode != http.StatusCreated {
		fmt.Println("Register failed!")
		return
	}

	// CHECK TOKEN IN REGISTER RESPONSE
	var regResp map[string]interface{}
	json.Unmarshal(body, &regResp)
	if token, ok := regResp["token"].(string); ok && token != "" {
		fmt.Println("SUCCESS: Register response contains valid JWT token")
	} else {
		fmt.Println("FAILURE: No token in register response")
	}

	// 1.5 Register Duplicate (Security Check)
	fmt.Println("\nTesting duplicate /register (Security Check)...")
	respDup, err := client.Post(baseURL+"/register", "application/json", bytes.NewBuffer(regBody))
	if err != nil {
		fmt.Printf("Duplicate Register failed (network error): %v\n", err)
		return
	}
	defer respDup.Body.Close()
	fmt.Printf("Duplicate Register Status: %s\n", respDup.Status)

	if respDup.StatusCode == http.StatusConflict {
		fmt.Println("SUCCESS: Duplicate registration rejected (Account Verified Safe)")
	} else {
		fmt.Printf("FAILURE: Duplicate registration should fail but got %s\n", respDup.Status)
	}

	// 2. Login
	fmt.Println("\nTesting /login...")
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	resp, err = client.Post(baseURL+"/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		fmt.Printf("Login failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	// fmt.Printf("Login Status: %s\nResponse: %s\n", resp.Status, string(body)) // Reduce noise

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Login failed!")
		return
	}

	// Check for token
	var loginResp map[string]interface{}
	json.Unmarshal(body, &loginResp)
	if token, ok := loginResp["token"].(string); ok && token != "" {
		fmt.Println("SUCCESS: Login received valid JWT token")
	} else {
		fmt.Println("FAILURE: No token in login response")
	}
}
