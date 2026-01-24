package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func main() {
	ctx := context.Background()

	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:           "example-client",                       // required for logging and tracing
		BaseURL:        "https://gorest.co.in/public/v2",       // optional parameters
		ContentType:    rest.JSON,                              // rest.JSON by default
		Timeout:        time.Millisecond * time.Duration(2000), // transmission timeout
		ConnectTimeout: time.Millisecond * time.Duration(5000), // socket timeout
		DefaultHeaders: http.Header{
			"X-Static-Header": {"My-Static-Value"}, // Add custom headers to the request
		},
	}

	// Set headers for the request (optional)
	headers := make(http.Header)
	headers.Set("My-Dynamic-Header-1", "My-Dynamic-Value-1")
	headers.Set("My-Dynamic-Header-2", "My-Dynamic-Value-2")

	// Make a GET request (context optional)
	response := client.GetWithContext(ctx, "/users", headers)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	// Untyped fill up
	var users []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Untyped fill up or typed with rest.Deserialize[struct | []struct](response)
	if err := response.FillUp(&users); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print the users
	for i := range users {
		fmt.Printf("User: %d, Name: %s, Email: %s\n", users[i].ID, users[i].Name, users[i].Email)
	}
}
