package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-relic/otel/tracing"

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
	ctx := context.Background()
	app, err := tracing.New(
		ctx,
		tracing.WithAppName("MyExample"),
		tracing.WithProtocol(tracing.
			NewGRPCProtocol("localhost:4317")))
	if err != nil {
		log.Fatal(err)
	}

	defer func(app *tracing.App, ctx context.Context) {
		shutdownErr := app.Shutdown(ctx)
		if shutdownErr != nil {
			log.Fatal(err)
		}
	}(app, ctx)

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
		log.Infof("simulating API requests...")
		for {
			apiURL := fmt.Sprintf("/cache/%d", random(100, 1000))
			txnCtx, txn := tracing.NewTransaction(ctx, "MyHTTPRequest")
			response := client.GetWithContext(txnCtx, apiURL)
			if response.Err != nil {
				txn.NoticeError(response.Err)
				log.Error(response.Err)
				txn.End()
				continue
			}
			log.Infof("GET %s, Status: %d", apiURL, response.StatusCode)
			txn.End()
		}
	}()

	server := &http.Server{Addr: ":8081", ReadHeaderTimeout: 5000 * time.Millisecond}

	log.Info("server started, metrics on http://localhost:8081/metrics")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
