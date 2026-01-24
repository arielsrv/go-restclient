package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

// This example demonstrates using ETag with the go-restclient.
// It performs two consecutive GET requests to an endpoint that returns a stable ETag.
// The first request gets the resource from the server and stores it in the internal cache.
// The second request automatically sends If-None-Match and, upon 304 from the server,
// serves the cached response (StatusCode remains 200) and marks it as Cached().
func main() {
	client := &rest.Client{
		Name:        "etag-example",
		BaseURL:     "https://httpbin.org",
		ContentType: rest.JSON,
		Timeout:     5 * time.Second,
		EnableGzip:  true,
		EnableCache: true, // required to leverage ETag/Last-Modified cache behavior
	}

	const path = "/etag/foobar-etag"
	ctx := context.Background()

	fmt.Println("-- First request --")
	resp1 := client.GetWithContext(ctx, path)
	check(resp1)
	fmt.Printf(
		"Status: %d, Cached: %t, ETag: %q\n",
		resp1.StatusCode,
		resp1.Cached(),
		resp1.Header.Get(rest.ETagHeader),
	)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("-- Second request (should revalidate with If-None-Match and use cache) --")
	resp2 := client.GetWithContext(ctx, path)
	check(resp2)
	fmt.Printf(
		"Status: %d, Cached: %t, ETag: %q\n",
		resp2.StatusCode,
		resp2.Cached(),
		resp2.Header.Get(rest.ETagHeader),
	)

	// Note: On the second call, the library sends If-None-Match automatically.
	// If the server replies 304 Not Modified, the go-restclient returns the cached response
	// (StatusCode stays 200) and resp2.Cached() will be true.
}

func check(r *rest.Response) {
	if r == nil {
		fmt.Println("nil response")
		os.Exit(1)
	}
	if r.Err != nil {
		fmt.Printf("error: %v\n", r.Err)
		os.Exit(1)
	}
	if !r.IsOk() {
		fmt.Printf("unexpected status: %d body=%s\n", r.StatusCode, r.String())
		os.Exit(1)
	}
}
