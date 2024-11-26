package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:     "https://httpbin.org",
		ContentType: rest.JSON,
		Name:        "gorest-client",
		Timeout:     time.Duration(5000) * time.Millisecond,
	}

	random := func(min, max int64) int64 {
		z := max - min + 1
		n, err := rand.Int(rand.Reader, big.NewInt(z))
		if err != nil {
			return 10
		}

		return n.Int64() + min
	}

	go func() {
		for {
			response := client.GetWithContext(context.Background(), fmt.Sprintf("/cache/%d", random(1, 100)))
			if response.Err != nil {
				fmt.Printf("Error: %v\n", response.Err)
				continue
			}
		}
	}()

	fmt.Printf("server started, metrics on http://localhost:8081/metrics")
	if err := http.ListenAndServe(":8081", router); err != nil {
		panic(err)
	}
}
