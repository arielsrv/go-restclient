package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

func main() {
	fmt.Println("=== Go RESTClient Mock Server Examples ===")

	// Example 1: Basic Mock Setup
	fmt.Println("1. Basic Mock Setup:")
	basicMockExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 2: Mock with Headers and Timeout
	fmt.Println("2. Mock with Headers and Timeout:")
	mockWithHeadersExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 3: Multiple HTTP Methods
	fmt.Println("3. Multiple HTTP Methods:")
	multipleMethodsExample()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 4: Error Scenarios
	fmt.Println("4. Error Scenarios:")
	errorScenariosExample()
}

func basicMockExample() {
	// Start the mock server
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	// Define mock responses
	userMock := &rest.Mock{
		URL:          "https://api.example.com/users",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `[{"id":1,"name":"John Doe","email":"john@example.com"}]`,
		RespHeaders: http.Header{
			"Content-Type":  {"application/json"},
			"Cache-Control": {"max-age=3600"},
		},
	}

	// Add mock to the server
	err := rest.AddMockups(userMock)
	if err != nil {
		fmt.Printf("Error adding mock: %v\n", err)
		return
	}

	// Create client and make request
	client := &rest.Client{
		Name:        "test-client",
		BaseURL:     "https://api.example.com",
		ContentType: rest.JSON,
		Timeout:     5 * time.Second,
	}

	response := client.GetWithContext(context.Background(), "/users")
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		return
	}

	fmt.Printf("Status: %d\n", response.StatusCode)
	fmt.Printf("Body: %s\n", response.String())
	fmt.Printf("Cache-Control: %s\n", response.Header.Get("Cache-Control"))
}

func mockWithHeadersExample() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	// Mock that validates request headers
	authMock := &rest.Mock{
		URL:        "https://api.example.com/protected",
		HTTPMethod: http.MethodGet,
		ReqHeaders: http.Header{
			"Authorization": {"Bearer valid-token"},
			"X-API-Key":     {"test-key"},
		},
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"message":"Access granted"}`,
		Timeout:      100 * time.Millisecond, // Simulate network delay
	}

	err := rest.AddMockups(authMock)
	if err != nil {
		fmt.Printf("Error adding mock: %v\n", err)
		return
	}

	client := &rest.Client{
		Name:        "auth-client",
		BaseURL:     "https://api.example.com",
		ContentType: rest.JSON,
	}

	// Add required headers
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer valid-token")
	headers.Set("X-Api-Key", "test-key")

	response := client.GetWithContext(context.Background(), "/protected", headers)
	fmt.Printf("Response: %s\n", response.String())
}

func multipleMethodsExample() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	// GET mock
	getMock := &rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"id":1,"name":"John Doe"}`,
	}

	// POST mock
	postMock := &rest.Mock{
		URL:          "https://api.example.com/users",
		HTTPMethod:   http.MethodPost,
		ReqBody:      `{"name":"Jane Doe","email":"jane@example.com"}`,
		RespHTTPCode: http.StatusCreated,
		RespBody:     `{"id":2,"name":"Jane Doe","email":"jane@example.com"}`,
	}

	// PUT mock
	putMock := &rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   http.MethodPut,
		ReqBody:      `{"name":"John Updated"}`,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"id":1,"name":"John Updated"}`,
	}

	// DELETE mock
	deleteMock := &rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   http.MethodDelete,
		RespHTTPCode: http.StatusNoContent,
		RespBody:     "",
	}

	err := rest.AddMockups(getMock, postMock, putMock, deleteMock)
	if err != nil {
		fmt.Printf("Error adding mock: %v\n", err)
		return
	}

	client := &rest.Client{
		Name:        "crud-client",
		BaseURL:     "https://api.example.com",
		ContentType: rest.JSON,
	}

	// Test GET
	response := client.Get("/users/1")
	fmt.Printf("GET Response: %s\n", response.String())

	// Test POST
	userData := map[string]string{
		"name":  "Jane Doe",
		"email": "jane@example.com",
	}
	response = client.Post("/users", userData)
	fmt.Printf("POST Response: %s\n", response.String())

	// Test PUT
	updateData := map[string]string{"name": "John Updated"}
	response = client.Put("/users/1", updateData)
	fmt.Printf("PUT Response: %s\n", response.String())

	// Test DELETE
	response = client.Delete("/users/1")
	fmt.Printf("DELETE Status: %d\n", response.StatusCode)
}

func errorScenariosExample() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	// 404 Not Found mock
	notFoundMock := &rest.Mock{
		URL:          "https://api.example.com/users/999",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusNotFound,
		RespBody:     `{"error":"User not found"}`,
	}

	// 500 Internal Server Error mock
	serverErrorMock := &rest.Mock{
		URL:          "https://api.example.com/error",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusInternalServerError,
		RespBody:     `{"error":"Internal server error"}`,
	}

	// Timeout mock
	timeoutMock := &rest.Mock{
		URL:          "https://api.example.com/slow",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"message":"Slow response"}`,
		Timeout:      10 * time.Second, // Very slow response
	}

	err := rest.AddMockups(notFoundMock, serverErrorMock, timeoutMock)
	if err != nil {
		fmt.Printf("Error adding mock: %v\n", err)
		return
	}

	client := &rest.Client{
		Name:        "error-test-client",
		BaseURL:     "https://api.example.com",
		ContentType: rest.JSON,
		Timeout:     2 * time.Second, // Client timeout
	}

	// Test 404
	response := client.Get("/users/999")
	fmt.Printf("404 Status: %d, Body: %s\n", response.StatusCode, response.String())

	// Test 500
	response = client.Get("/error")
	fmt.Printf("500 Status: %d, Body: %s\n", response.StatusCode, response.String())

	// Test timeout
	response = client.Get("/slow")
	if response.Err != nil {
		fmt.Printf("Timeout Error: %v\n", response.Err)
	}
}
