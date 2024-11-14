package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
	defer cancel()

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

	var users []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Gender string `json:"gender"`
		Status string `json:"status"`
	}

	headers := make(http.Header)
	headers.Add("Accept", "application/json")
	headers.Add("Content-Type", "application/json")

	response := client.GetWithContext(ctx, "/users", headers)
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
	}

	// Untyped fill up
	err := response.FillUp(&users)
	if err != nil {
		log.Fatal(err)
	}

	for i := range users {
		log.Infof("User: %v", users[i])
	}
}
