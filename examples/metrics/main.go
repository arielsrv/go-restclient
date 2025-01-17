package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	http.Handle("/metrics", promhttp.Handler())

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:        "https://httpbin.org",
		ContentType:    rest.JSON,
		Name:           "gorest-client",
		ConnectTimeout: time.Duration(1000) * time.Millisecond,
		Timeout:        time.Duration(500) * time.Millisecond,
		EnableCache:    true,
	}

	random := func(minValue int64, maxValue int64) int64 {
		z := maxValue - minValue + 1
		n, err := rand.Int(rand.Reader, big.NewInt(z))
		if err != nil {
			return 100
		}

		return n.Int64() + minValue
	}

	go func() {
		for {
			apiURL := fmt.Sprintf("/cache/%d", random(1, 100))
			response := client.GetWithContext(context.Background(), apiURL)
			if response.Err != nil {
				fmt.Printf("error: %v\n", response.Err)
				continue
			}
		}
	}()

	server := &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 5000 * time.Millisecond,
	}

	fmt.Printf("server started, metrics on http://localhost:8081/metrics\n")
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
