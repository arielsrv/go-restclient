package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	// Create a new context with a timeout of 5 seconds
	// This will automatically cancel the request if it takes longer than 500 milliseconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5000)*time.Millisecond)
	defer cancel()

	// Create a new REST usersClient with custom settings
	usersClient := &rest.Client{
		BaseURL:     "https://gorest.co.in/public/v2",
		ContentType: rest.JSON,
		Name:        "gorest-usersClient",
		Timeout:     time.Duration(5000) * time.Millisecond,
	}

	rChan := make(chan *rest.Response)

	// Create a channel to collect the response asynchronously.
	usersClient.GetChanWithContext(ctx, "/users", rChan)

	// Wait for the response and handle errors
	response := <-rChan
	if response.Err != nil {
		panic(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Received non-200 status code: %d", response.StatusCode))
	}

	var usersResponse []UserResponse
	err := response.FillUp(&usersResponse)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Users: %+v\n", usersResponse)

	defer close(rChan)
	for i := range usersResponse {
		userResponse := usersResponse[i]
		apiURL := fmt.Sprintf("/users/%d", userResponse.ID)
		usersClient.GetChanWithContext(ctx, apiURL, rChan)
	}

	for range usersResponse {
		if response.Err != nil {
			fmt.Printf("Error fetching user data: %v\n", response.Err)
			continue
		}
		if response.StatusCode != http.StatusOK {
			fmt.Printf("Received non-200 status code for user: %d\n", response.StatusCode)
			continue
		}
		fmt.Printf("User: %+v\n", response)
	}
}

type UserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    uint   `json:"id"`
}
