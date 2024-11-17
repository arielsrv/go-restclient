package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	httpClient := &rest.Client{
		BaseURL:           "https://gorest.co.in/public/v2",
		Timeout:           time.Millisecond * 2000,
		ConnectionTimeout: time.Millisecond * 5000,
		ContentType:       rest.XML,
	}

	var usersResponse struct {
		XMLName string `xml:"objects"`
		List    []struct {
			Name   string `xml:"name"`
			Email  string `xml:"email"`
			Gender string `xml:"gender"`
			Status string `xml:"status"`
			ID     int    `xml:"id"`
		} `xml:"object"`
	}

	response := httpClient.GetWithContext(context.Background(), "/users.xml")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
	}

	err := response.FillUp(&usersResponse)
	if err != nil {
		log.Fatal(err)
	}

	for i := range usersResponse.List {
		log.Infof("User: %v", usersResponse.List[i])
	}
}
