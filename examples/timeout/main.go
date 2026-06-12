package main

import (
	"context"
	"fmt"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:    "timeout-client",
		BaseURL: "https://10.255.255.1", // NOSONAR: non-routable IP intentionally used to force a connection timeout in this example
		Timeout: 1 * time.Second,
	}

	response := client.GetWithContext(context.Background(), "/will-timeout")
	if response.Err != nil {
		fmt.Printf("Network error: %v\n", response.Err)
	} else {
		fmt.Println("Status:", response.StatusCode)
	}
}
