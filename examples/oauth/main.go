package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
)

func main() {
	client := &rest.Client{
		Name:           "ocapi-client",
		BaseURL:        "https://www.kiwoko.com/s/-/dw/data/v22_6",
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.JSON,
		OAuth: &rest.OAuth{
			ClientID:     "a11d0149-687e-452e-9c94-783d489d4f72",
			ClientSecret: "Kiwoko@1234",
			TokenURL:     "https://account.demandware.com/dw/oauth2/access_token",
			AuthStyle:    rest.AuthStyleInHeader,
		},
		EnableTrace: true,
	}

	type SitesResponse struct {
		V    string `json:"_v"`
		Type string `json:"_type"`
		Data []struct {
			Type          string `json:"_type"`
			ResourceState string `json:"_resource_state"`
			ID            string `json:"id"`
			Link          string `json:"link"`
		} `json:"data"`
		Count int `json:"count"`
		Start int `json:"start"`
		Total int `json:"total"`
	}

	ctx := context.Background()

	for {
		response := client.GetWithContext(ctx, "/sites")
		if response.Err != nil {
			log.Fatal(response.Err)
		}

		if response.StatusCode != http.StatusOK {
			log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.String())
		}

		var sitesResponse SitesResponse
		err := response.FillUp(&sitesResponse)
		if err != nil {
			log.Fatal(err)
		}

		for i := range sitesResponse.Data {
			siteResponse := sitesResponse.Data[i]
			log.Infof("Site: %s, Link: %s", siteResponse.ID, siteResponse.Link)
		}

		log.Info("Waiting for 10 seconds before exiting...  (Ctrl+C to stop)")
		time.Sleep(time.Duration(10) * time.Second)
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
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)
	otel.Tracer("example")
}
