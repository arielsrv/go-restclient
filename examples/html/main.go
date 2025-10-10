package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
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
		log.Fatal(response1.Err)
	case response1.StatusCode != http.StatusOK:
		log.Fatalf("status_code: %d, reason: %s", response1.StatusCode, response1.String())
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
		log.Fatal(response2.Err)
	case response2.StatusCode != http.StatusOK:
		log.Fatalf("status_code: %d, reason: %s", response2.StatusCode, response2.String())
	}

	fmt.Printf("Response cached: %t\n", response2.Cached())
}
