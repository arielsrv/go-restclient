package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

func main() {
	// Create a new REST client with custom settings
	client := &rest.Client{
		Name:           "example-client",                       // required for logging and tracing
		BaseURL:        "https://httpbin.org",                  // optional parameters
		Timeout:        time.Millisecond * time.Duration(2000), // transmission timeout
		ConnectTimeout: time.Millisecond * time.Duration(5000), // socket timeout
		ContentType:    rest.FORM,                              // rest.JSON by default
	}

	values := url.Values{}
	values.Set("key1", "value1")
	values.Set("key2", "value2")

	response := client.PostWithContext(context.Background(), "/post", values)
	if response.Err != nil {
		fmt.Printf("Error: %v\n", response.Err)
		log.Fatal()
	}

	// Check status code and handle errors accordingly or response.IsOk()
	if response.StatusCode != http.StatusOK {
		fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
		os.Exit(1)
	}

	var formResponse FormResponse
	err := response.FillUp(&formResponse)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Form Response: %+v\n", formResponse)
}

type FormResponse struct {
	Args    struct{} `json:"args"`
	Data    string   `json:"data"`
	Files   struct{} `json:"files"`
	Form    struct{} `json:"form"`
	Headers struct {
		Accept         string `json:"Accept"`
		AcceptEncoding string `json:"Accept-Encoding"`
		CacheControl   string `json:"Cache-Control"`
		ContentLength  string `json:"Content-Length"`
		ContentType    string `json:"Content-Type"`
		Host           string `json:"Host"`
		UserAgent      string `json:"User-Agent"`
		XAmznTraceID   string `json:"X-Amzn-Trace-Id"`
	} `json:"headers"`
	JSON   interface{} `json:"json"`
	Origin string      `json:"origin"`
	URL    string      `json:"url"`
}
