package main

import (
	"context"
	"fmt"
	"net/http"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:           "gzip-headers-client",
		BaseURL:        "https://httpbin.org",
		EnableGzip:     true,
		DefaultHeaders: http.Header{"Accept-Encoding": []string{"gzip"}},
	}

	headers := make(http.Header)
	headers.Set("X-Custom", "demo")

	response := client.GetWithContext(context.Background(), "/gzip", headers)
	fmt.Println("Content-Encoding:", response.Header.Get("Content-Encoding"))
	fmt.Println("Body:", response.String())
}
