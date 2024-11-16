[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs.
- Fork-Join request pattern, for sending many requests concurrently, getting better client
  performance (deprecated).
- Response Caching, based on response headers (cache-control, last-modified, etag, expires)
    - SFCC uses caching strategies to avoid making an HTTP request if it's not necessary; however,
      this will consume more memory in your app until the validation time expires.
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default JSON.
- Request Body can be `string`, `[]byte`, `struct` & `map`
- File sending
- Default and custom connection pool isolation.
- Trace connection if available

## Table of contents

- [RESTClient](#rest-client)
- [Metrics](#metrics)

## Rest Client

# Installation

```shell
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@latest
```

# ⚡️ Quickstart

- [Examples](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/tree/main/examples?ref_type=heads)

```go
package main

import (
    "context"
    "net/http"
    "time"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
    // Create a new context with a timeout of 5 seconds
    // This will automatically cancel the request if it takes longer than 5 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
    defer cancel()

    // Create a new REST client with custom settings
    client := &rest.Client{
        BaseURL:        "https://gorest.co.in/public/v2",
        Timeout:        time.Millisecond * 1000,
        ConnectTimeout: time.Millisecond * 5000,
        ContentType:    rest.JSON,
        Name:           "example-client",
        // EnableTrace:    true,
        // CustomPool:     &...,
        // BasicAuth:      &...,
        // Client:         &...,
        // OAuth:          &...,
        // UserAgent:      "<Your User Agent>",
        // DisableCache:   false,
        // DisableTimeout: false,
        // FollowRedirect: false,
    }

    // Set headers for the request (optional)
    headers := make(http.Header)
    headers.Add("My-Custom-Header", "My-Custom-Value")

    // Make a GET request (context optional)
    response := client.GetWithContext(ctx, "/users", headers)
    if response.Err != nil {
        log.Fatal(response.Err)
    }

    // Check status code and handle errors accordingly or response.IsOk()
    if response.StatusCode != http.StatusOK {
        log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
    }

    // Untyped fill up
    var users []struct {
        ID     int    `json:"id"`
        Name   string `json:"name"`
        Email  string `json:"email"`
        Gender string `json:"gender"`
        Status string `json:"status"`
    }

    // Untyped fill up or typed with rest.Deserialize[struct | []struct](response)
    err := response.FillUp(&users)
    if err != nil {
        log.Fatal(err)
    }

    // Print the users
    for i := range users {
        log.Infof("User: %v", users[i])
    }
}

```

## Metrics

![prometheus]
![otel]

Requisites

- Make sure you have **prometheus collector endpoint** turned on in your application
- **ENV** variable (dev|uat|pro|any)
- **APP_NAME** variable (repository name)

We do not have a unified dashboard, which can filter by environment, due to this, you have to enter
the specific
environment

Dashboard

- [dev](https://monitoring.dev.dp.iskaypet.com/d/6shkc-L4kk/http-clients?orgId=1)

[prometheus]: images/metrics.png

## Benchmarks

```go
func BenchmarkGet(b *testing.B) {
	client := &rest.Client{}

	for i := 0; i < b.N; i++ {
		resp := client.Get("https://gorest.co.in/public/v2/users")
		if resp.Err != nil {
			log.Info("f[" + strconv.Itoa(i) + "] Error")
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Info("f[" + strconv.Itoa(i) + "] Status != OK (200)")
		}
	}
}

func BenchmarkResty_Get(b *testing.B) {
	client := resty.New()
	for i := 0; i < b.N; i++ {
		resp, err := client.R().Get(fmt.Sprintf("https://gorest.co.in/public/v2/users"))
		if err != nil {
			log.Info("f[" + strconv.Itoa(i) + "] Error")
			continue
		}
		if resp.StatusCode() != http.StatusOK {
			log.Info("f[" + strconv.Itoa(i) + "] Status != OK (200)")
		}
	}
}
```

### go-restclient
    goos: darwin
    goarch: arm64
    pkg: gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest
    cpu: Apple M3 Pro
    BenchmarkGet 
    BenchmarkGet-10    	       3	 405502875 ns/op
    PASS

### resty
    goos: darwin
    goarch: arm64
    pkg: gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest
    cpu: Apple M3 Pro
    BenchmarkResty_Get
    BenchmarkResty_Get-10    	       1	1099176917 ns/op
    PASS
