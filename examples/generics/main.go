package main

import (
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
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
		log.Fatal(response.Err)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if !response.IsOk() {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
	}

	// Typed fill up
	usersResponse, err := rest.Deserialize[[]UserResponse](response)
	if err != nil {
		log.Fatal(err)
	}

	// Print the users
	for i := range usersResponse {
		log.Infof("User: %v", usersResponse[i])
	}
}
