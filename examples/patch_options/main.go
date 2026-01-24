package main

import (
	"context"
	"fmt"

	"github.com/arielsrv/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:        "patch-options-client",
		BaseURL:     "https://httpbin.org",
		ContentType: rest.JSON,
	}

	// PATCH example
	patchData := map[string]string{"field": "new value"}
	response := client.PatchWithContext(context.Background(), "/patch", patchData)
	fmt.Println("PATCH status:", response.StatusCode)
	fmt.Println("PATCH body:", response.String())

	// OPTIONS example
	response = client.OptionsWithContext(context.Background(), "/patch")
	fmt.Println("OPTIONS Allow methods:", response.Header.Get("Allow"))
}
