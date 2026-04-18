package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/arielsrv/go-restclient/rest"
)

// This example demonstrates automatic RFC 7807 Problem Details deserialization.
// When the server returns Content-Type: application/problem+json, the library
// automatically populates response.Problem without any extra code.
func main() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	const problemURL = "http://api.example.com/orders/99"

	err := rest.AddMockups(&rest.Mock{
		URL:          problemURL,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusNotFound,
		RespBody: `{
  "type":     "https://example.com/errors/not-found",
  "title":    "Order Not Found",
  "detail":   "Order 99 does not exist.",
  "status":   404,
  "instance": "/orders/99"
}`,
		RespHeaders: http.Header{
			"Content-Type": {"application/problem+json"},
		},
	})
	if err != nil {
		fmt.Printf("mock setup error: %v\n", err)
		os.Exit(1)
	}

	client := &rest.Client{ContentType: rest.JSON}
	resp := client.GetWithContext(context.Background(), problemURL)

	if resp.Err != nil {
		fmt.Printf("request error: %v\n", resp.Err)
		os.Exit(1)
	}

	fmt.Printf("Status:   %d\n", resp.StatusCode)

	if resp.Problem != nil {
		fmt.Printf("Type:     %s\n", resp.Problem.Type)
		fmt.Printf("Title:    %s\n", resp.Problem.Title)
		fmt.Printf("Detail:   %s\n", resp.Problem.Detail)
		fmt.Printf("Status:   %d\n", resp.Problem.Status)
		fmt.Printf("Instance: %s\n", resp.Problem.Instance)
	} else {
		fmt.Println("No problem details in response")
	}
}
