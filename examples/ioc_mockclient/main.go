package main

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type MyMockClient struct{}

func (m *MyMockClient) GetWithContext(ctx context.Context, url string, headers ...http.Header) *rest.Response {
	return &rest.Response{StatusCode: 200, Response: &http.Response{}, Err: nil}
}

func main() {
	client := &rest.Client{
		Name:    "ioc-mock-client",
		BaseURL: "https://api.example.com",
	}
	client.HTTPClient = &MyMockClient{}

	response := client.GetWithContext(context.Background(), "/any")
	fmt.Println("Status:", response.StatusCode)
}
