package main

import (
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	baseURL := "https://gorest.co.in/public/v2"

	httpClient := &rest.RequestBuilder{
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		BaseURL:        baseURL,
		// OAuth: 		...
		// CustomPool:  ...
	}

	var users []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Gender string `json:"gender"`
		Status string `json:"status"`
	}

	response := httpClient.Get("/users")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
	}

	// Typed fill up
	result, err := rest.Unmarshal[[]UserDTO](response)
	if err != nil {
		log.Fatal(err)
	}

	for i := range result {
		log.Infof("User: %v", result[i])
	}

	// Untyped fill up
	err = response.FillUp(&users)
	if err != nil {
		log.Fatal(err)
	}

	for i := range users {
		log.Infof("User: %v", users[i])
	}
}

type UserDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
	Status string `json:"status"`
}
