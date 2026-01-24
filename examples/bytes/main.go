package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func main() {
	ctx := context.Background()

	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:           "example-client",                       // required for logging and tracing
		BaseURL:        "https://httpbin.org",                  // optional parameters
		Timeout:        time.Millisecond * time.Duration(2000), // transmission timeout
		ConnectTimeout: time.Millisecond * time.Duration(5000), // socket timeout
	}

	apiURL := fmt.Sprintf("/bytes/%d", 1*rest.MB)
	response := client.GetWithContext(ctx, apiURL)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	fmt.Printf("Form Response: %+v\n", response.String())

	apiURL = fmt.Sprintf("/stream-bytes/%d", 1*rest.MB)
	response = client.GetWithContext(ctx, apiURL)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	fmt.Printf("Form Response: %+v\n", response.String())
}
