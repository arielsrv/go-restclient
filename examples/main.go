package main

import (
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	baseURL := "https://gorest.co.in/public/v2"

	rb := rest.RequestBuilder{
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		BaseURL:        baseURL,
	}

	var users []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Email  string `json:"email"`
		Gender string `json:"gender"`
		Status string `json:"status"`
	}

	response := rb.Get("/users")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
	}

	err := response.FillUp(&users)
	if err != nil {
		log.Fatal(err)
	}

	for i := range users {
		log.Infof("User: %v", users[i])
	}
}
