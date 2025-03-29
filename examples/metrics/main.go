package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func init() {
	numCPU := runtime.NumCPU() - 1
	runtime.GOMAXPROCS(numCPU)
	log.Infof("using %d CPU cores", numCPU)
}

func main() {
	http.Handle("/metrics", promhttp.Handler())

	// Create a new REST client with custom settings
	client := &rest.Client{
		BaseURL:        "https://httpbin.org",
		ContentType:    rest.JSON,
		Name:           "gorest-client",
		ConnectTimeout: time.Duration(2000) * time.Millisecond,
		Timeout:        time.Duration(1000) * time.Millisecond,
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
		log.Infof("simulating API requests...")
		for {
			apiURL := fmt.Sprintf("/cache/%d", random(1, 100))
			response := client.GetWithContext(context.Background(), apiURL)
			if response.Err != nil {
				log.Error(response.Err)
				continue
			}
			log.Infof("GET %s, Status: %d", apiURL, response.StatusCode)
		}
	}()

	server := &http.Server{Addr: ":8081", ReadHeaderTimeout: 5000 * time.Millisecond}

	log.Info("server started, metrics on http://localhost:8081/metrics")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
