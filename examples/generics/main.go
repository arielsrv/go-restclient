package main

import (
	"fmt"
	"os"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

type UserResponse struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
	Status string `json:"status"`
	ID     int    `json:"id"`
}

func main() {
	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.JSON,
		Name:           "example-client",
	}

	// Make a GET request
	response := client.Get("/users")
	if response.Err != nil {
		fmt.Println(response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if !response.IsOk() {
		fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
		os.Exit(1)
	}

	// Typed fill up
	usersResponse, err := rest.Deserialize[[]UserResponse](response)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Print the users
	for i := range usersResponse {
		fmt.Printf("User: %v\n", usersResponse[i])
	}
}
