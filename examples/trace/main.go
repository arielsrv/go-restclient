package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
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
		EnableTrace:    true,
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
			apiURL := fmt.Sprintf("/cache/%d", random(100, 1000))
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

func init() {
	spanExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		return
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(spanExporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("example"),
			)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	otel.Tracer("example")
}
