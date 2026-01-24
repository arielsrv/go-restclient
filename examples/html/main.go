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
	ctx := context.Background()

	client := &rest.Client{
		Name:           "html-client",
		Timeout:        1000 * time.Millisecond,
		ConnectTimeout: 2000 * time.Millisecond,
		EnableGzip:     true,
		EnableCache:    true,
	}

	response1 := client.GetWithContext(
		ctx,
		"https://syndicate.synthrone.com/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html",
	)
	switch {
	case response1.Err != nil:
		fmt.Println(response1.Err)
		os.Exit(1)
	case response1.StatusCode != http.StatusOK:
		fmt.Printf("status_code: %d, reason: %s\n", response1.StatusCode, response1.String())
		os.Exit(1)
	}
	fmt.Printf("Response cached: %t\n", response1.Cached())

	// Simulate a delay to allow the cache to expire
	time.Sleep(time.Millisecond * 100)

	// server response with 304 Not Modified so client should not make a new request and return cached response
	response2 := client.GetWithContext(
		ctx,
		"https://syndicate.synthrone.com/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html",
	)
	switch {
	case response2.Err != nil:
		fmt.Println(response2.Err)
		os.Exit(1)
	case response2.StatusCode != http.StatusOK:
		fmt.Printf("status_code: %d, reason: %s\n", response2.StatusCode, response2.String())
		os.Exit(1)
	}

	fmt.Printf("Response cached: %t\n", response2.Cached())
}
