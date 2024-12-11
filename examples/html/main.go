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
	httpClient := &rest.Client{
		BaseURL:        "https://syndicate.synthrone.com",
		Timeout:        time.Millisecond * 2000,
		ConnectTimeout: time.Millisecond * 5000,
	}

	response := httpClient.GetWithContext(context.Background(), "/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
	}

	fmt.Printf("Response: %s", response.String())
}
