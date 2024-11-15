package main

import (
	"context"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"

	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	ocapiClient := &rest.Client{
		BaseURL:        "https://www.kiwoko.com/s/-/dw/data/v22_6",
		Timeout:        time.Millisecond * 1000,
		ConnectTimeout: time.Millisecond * 5000,
		ContentType:    rest.JSON,
		Name:           "example-ocapiClient",
		OAuth: &clientcredentials.Config{
			ClientID:     "a11d0149-687e-452e-9c94-783d489d4f72",
			ClientSecret: "Kiwoko@1234",
			TokenURL:     "https://account.demandware.com/dw/oauth2/access_token",
			AuthStyle:    oauth2.AuthStyleInHeader,
		},
		EnableTrace: true,
	}

	type SitesResponse struct {
		V    string `json:"_v"`
		Type string `json:"_type"`
		Data []struct {
			Type          string `json:"_type"`
			ResourceState string `json:"_resource_state"`
			Id            string `json:"id"`
			Link          string `json:"link"`
		} `json:"data"`
		Count int `json:"count"`
		Start int `json:"start"`
		Total int `json:"total"`
	}

	response := ocapiClient.GetWithContext(context.Background(), "/sites")
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	if response.StatusCode != http.StatusOK {
		log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
	}

	var sitesResponse SitesResponse
	err := response.FillUp(&sitesResponse)
	if err != nil {
		log.Fatal(err)
	}

	for i := range sitesResponse.Data {
		siteResponse := sitesResponse.Data[i]
		log.Infof("Site: %s, Link: %s", siteResponse.Id, siteResponse.Link)
	}

	time.Sleep(time.Second * 10)
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
