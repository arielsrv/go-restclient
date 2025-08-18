package main

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:    "timeout-client",
		BaseURL: "https://10.255.255.1", // IP no enrutable para forzar timeout
		Timeout: 1 * time.Second,
	}

	response := client.GetWithContext(context.Background(), "/will-timeout")
	if response.Err != nil {
		fmt.Printf("Network error: %v\n", response.Err)
	} else {
		fmt.Println("Status:", response.StatusCode)
	}
}
