package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:           "httpbin-client", // required for logging and tracing
		ContentType:    rest.JSON,        // rest.JSON by default
		Timeout:        time.Duration(5000) * time.Millisecond,
		BaseURL:        "https://httpbin.org",
		EnableGzip:     true,                                             // enable gzip compression by default (optionally enable it)
		DefaultHeaders: http.Header{"Accept-Encoding": []string{"gzip"}}, // default Accept-Encoding header (optionally customize it)
	}

	// Enable gzip compression for the request (or use EnableGzip option in the client settings)
	headers := make(http.Header)
	headers.Add("Accept-Encoding", "gzip")
	headers.Add("My-Header-1", "value1")
	headers.Add("My-Header-1", "value2")
	headers.Add("My-Header-2", "value2")

	// Make a GET request (context optional)
	response := client.GetWithContext(context.Background(), "/gzip", headers)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		os.Exit(1)
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	var httpResponse HTTPResponse
	err := response.FillUp(&httpResponse)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %+v\n", httpResponse)
}

type HTTPResponse struct {
	Headers struct {
		Accept          string `json:"Accept"`
		AcceptEncoding  string `json:"Accept-Encoding"`
		AcceptLanguage  string `json:"Accept-Language"`
		Host            string `json:"Host"`
		Priority        string `json:"Priority"`
		Referer         string `json:"Referer"`
		SecChUa         string `json:"Sec-Ch-Ua"`
		SecChUaMobile   string `json:"Sec-Ch-Ua-Mobile"`
		SecChUaPlatform string `json:"Sec-Ch-Ua-Platform"`
		SecFetchDest    string `json:"Sec-Fetch-Dest"`
		SecFetchMode    string `json:"Sec-Fetch-Mode"`
		SecFetchSite    string `json:"Sec-Fetch-Site"`
		UserAgent       string `json:"User-Agent"`
		XAmznTraceId    string `json:"X-Amzn-Trace-Id"`
	} `json:"headers"`
	Method  string `json:"method"`
	Origin  string `json:"origin"`
	Gzipped bool   `json:"gzipped"`
}
