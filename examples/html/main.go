package main

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:           "html-client",
		EnableGzip:     true,
		EnableCache:    true,
		Timeout:        1000 * time.Millisecond,
		ConnectTimeout: 2000 * time.Millisecond,
	}

	response := client.Get("https://syndicate.synthrone.com/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html")
	switch {
	case response.Err != nil:
		log.Fatal(response.Err)
	case response.StatusCode != http.StatusOK:
		log.Fatalf("status_code: %d, reason: %s", response.StatusCode, response.String())
	}
	fmt.Printf("%s\n", response.String())
	fmt.Printf("Response cached: %t\n", response.Cached())

	response = client.Get("https://syndicate.synthrone.com/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html")
	switch {
	case response.Err != nil:
		log.Fatal(response.Err)
	case response.StatusCode != http.StatusOK:
		log.Fatalf("status_code: %d, reason: %s", response.StatusCode, response.String())
	}
	fmt.Printf("%s\n", response.String())
	fmt.Printf("Response cached: %t\n", response.Cached())
}
