package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	ctx := context.Background()

	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:        "example-client",                                    // required for logging and tracing
		BaseURL:     "https://order-status-oms-sfcc.dev.dp.iskaypet.com", // optional parameters
		ContentType: rest.JSON,                                           // rest.JSON by default (rest.XML, rest.FORM, etc.)
		Timeout:     time.Millisecond * time.Duration(2000),              // transmission timeout
		/*ConnectTimeout: time.Millisecond * time.Duration(5000), // socket timeout
		EnableCache:    true,                                   // Last-Modified, Expires & ETag headers are disabled by default
		CustomPool: &rest.CustomPool{ // for fine-tuning the connection pool
			Transport: &http.Transport{
				IdleConnTimeout:       time.Duration(2000) * time.Millisecond,
				ResponseHeaderTimeout: time.Duration(2000) * time.Millisecond,
				MaxIdleConnsPerHost:   10,
			},
		},
		BasicAuth: &rest.BasicAuth{
			Username: "your_username",
			Password: "your_password",
		},
		OAuth: &rest.OAuth{
			ClientID:     "your_client_id",
			ClientSecret: "your_client_secret",
			TokenURL:     "https://oauth.gorest.co.in/oauth/token",
			AuthStyle:    rest.AuthStyleInHeader,
		},
		EnableTrace:    true,
		UserAgent:      "<Your User Agent>",
		DisableTimeout: false,
		FollowRedirect: false,*/
	}

	// Set headers for the request (optional)
	headers := make(http.Header)
	headers.Add("My-Custom-Header", "My-Custom-Value")

	// Make a GET request (context optional)
	response := client.GetWithContext(ctx, "/purchases/81319403/orders/81319403?site_id=TES", headers)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	// Untyped fill up
	var users []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// Untyped fill up or typed with rest.Deserialize[struct | []struct](response)
	if err := response.FillUp(&users); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print the users
	for i := range users {
		fmt.Printf("User: %d, Name: %s, Email: %s\n", users[i].ID, users[i].Name, users[i].Email)
	}
}
