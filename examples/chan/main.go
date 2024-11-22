package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	// Create a new context with a timeout of 5 seconds
	// This will automatically cancel the request if it takes longer than 500 milliseconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1000)*time.Millisecond)
	defer cancel()

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:     "https://gorest.co.in/public/v2",
		ContentType: rest.JSON,
		Name:        "gorest-client",
		Timeout:     time.Duration(5000) * time.Millisecond,
	}

	// Create a channel to collect the response asynchronously.
	rChan := client.ChanGetWithContext(ctx, "/users")

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

	uChan := make(chan *rest.Response, 1)

	var wg sync.WaitGroup

	for i := range usersResponse {
		wg.Add(1)
		go func(userResponse UserResponse) {
			defer wg.Done()
			apiURL := fmt.Sprintf("/users/%d", userResponse.ID)
			uChan <- client.GetWithContext(ctx, apiURL)
		}(usersResponse[i])
	}

	go func() {
		defer wg.Done()
		for u := range uChan {
			if u.Err != nil {
				fmt.Printf("Error fetching user data: %v\n", u.Err)
				continue
			}
			if u.StatusCode != http.StatusOK {
				fmt.Printf("Received non-200 status code for user: %d\n", u.StatusCode)
				continue
			}
			fmt.Printf("User: %+v\n", u)
		}
	}()

	wg.Wait()
}

type UserResponse struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    uint   `json:"id"`
}
