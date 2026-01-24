package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/arielsrv/go-restclient/rest"
)

func init() {
	numCPU := runtime.NumCPU() - 1
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("using %d CPU cores\n", numCPU)
}

func main() {
	ctx := context.Background()
	http.Handle("/metrics", promhttp.Handler())

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:        "https://httpbin.org",
		ContentType:    rest.JSON,
		Name:           "gorest-client",
		ConnectTimeout: time.Duration(2000) * time.Millisecond,
		Timeout:        time.Duration(1000) * time.Millisecond,
		EnableCache:    true,
		EnableTrace:    true,
	}

	random := func(minValue int64, maxValue int64) int64 {
		z := maxValue - minValue + 1
		n, rErr := rand.Int(rand.Reader, big.NewInt(z))
		if rErr != nil {
			return 100
		}

		return n.Int64() + minValue
	}

	go func() {
		fmt.Println("simulating API requests...")
		for {
			apiURL := fmt.Sprintf("/cache/%d", random(100, 1000))
			response := client.GetWithContext(ctx, apiURL)
			if response.Err != nil {
				fmt.Println(response.Err)
				continue
			}
			fmt.Printf("GET %s, Status: %d\n", apiURL, response.StatusCode)
		}
	}()

	server := &http.Server{Addr: ":8081", ReadHeaderTimeout: 5000 * time.Millisecond}

	fmt.Println("server started, metrics on http://localhost:8081/metrics")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
