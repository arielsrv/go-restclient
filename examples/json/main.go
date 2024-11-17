package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	// Create a new context with a timeout of 5 seconds
	// This will automatically cancel the request if it takes longer than 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
	defer cancel()

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.JSON,
		Name:           "example-client",
		// EnableTrace:    true,
		// CustomPool:     &...,
		// BasicAuth:      &...,
		// Client:         &...,
		// OAuth:          &...,
		// UserAgent:      "<Your User Agent>",
		// DisableCache:   false,
		// DisableTimeout: false,
		// FollowRedirect: false,
	}

	// Set headers for the request (optional)
	headers := make(http.Header)
	headers.Add("My-Custom-Header", "My-Custom-Value")

	// Make a GET request (context optional)
	response := client.GetWithContext(ctx, "/users", headers)
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
	}

	// Untyped fill up
	var users []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Gender string `json:"gender"`
		Status string `json:"status"`
	}

	// Untyped fill up or typed with rest.Deserialize[struct | []struct](response)
	err := response.FillUp(&users)
	if err != nil {
		log.Fatal(err)
	}

	// Print the users
	for i := range users {
		log.Infof("User: %v", users[i])
	}
}
