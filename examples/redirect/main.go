package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

func main() {
	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:           "example-client", // required for logging and tracing
		ContentType:    rest.JSON,        // rest.JSON by default
		Timeout:        time.Duration(5000) * time.Millisecond,
		FollowRedirect: true, // false by default
	}

	// Make a GET request (context optional)
	response := client.GetWithContext(context.Background(), "https://tinyurl.com/39da2yt4")
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
