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
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	client := &rest.Client{
		Name:       "html-client",
		EnableGzip: true,
		Timeout:    100 * time.Millisecond, ConnectTimeout: 100 * time.Millisecond,
	}

	fmt.Printf("[go-restclient] example started, MaxCacheSize: %d, NumCounters: %f, BufferItems: %d\n",
		rest.MaxCacheSize,
		rest.NumCounters,
		rest.BufferItems,
	)

	response := client.GetWithContext(ctx, "https://syndicate.synthrone.com/df9g5m2kxcv7/ROY153637_M/latest/ROY153637_M.html")

	switch {
	case response.Err != nil:
		log.Fatal(response.Err)
	case response.StatusCode != http.StatusOK:
		log.Fatalf("status_code: %d, reason: %s", response.StatusCode, response.String())
	default:
		fmt.Printf("%s\n", response.String())
	}
}
