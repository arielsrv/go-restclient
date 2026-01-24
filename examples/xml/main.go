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
	httpClient := &rest.Client{
		BaseURL:        "https://gorest.co.in/public/v2",
		Timeout:        time.Millisecond * 2000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.XML,
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
		fmt.Println(response.Err)
		os.Exit(1)
	}

	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s\n", response.StatusCode, response.String())
		os.Exit(1)
	}

	err := response.FillUp(&usersResponse)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for i := range usersResponse.List {
		fmt.Printf("User: %v\n", usersResponse.List[i])
	}
}
