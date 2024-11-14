package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.JSON,
		Name:           "example-client",
		OAuth: &clientcredentials.Config{
			ClientID:     "client_id",
			ClientSecret: "client_secret",
			TokenURL:     "https://account.demandware.com/dw/oauth2/access_token",
			AuthStyle:    oauth2.AuthStyleInHeader,
		},
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

	response := client.GetWithContext(context.Background(), "/users", headers)
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
