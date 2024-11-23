[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs.
- Response Caching, based on response headers (`cache-control`, `last-modified`, `etag`, `expires`)
    - SFCC uses caching strategies to avoid making an HTTP request if it's not necessary; however,
      this will consume more memory in your app until the validation time expires.
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default `JSON`.
    - Including HTTP `RFC7807` [Problems](https://datatracker.ietf.org/doc/html/rfc7807)
- Content-Type can be `JSON`, `XML` & `FORM`
- Request Body can be `string`, `[]byte`, `struct` & `map`
- FORM sending
- File sending
- Default and custom connection pool isolation.
- Trace connection if available
- Deprecated. Fork-Join request pattern, for sending many requests concurrently, getting better client
  performance. Use AsyncAPI instead.
- RX Channels coming soon

## Table of contents

- [RESTClient](#rest-client)
- [Metrics](#metrics)
- [Benchmarks](#ben)

## Rest Client

# Installation

```shell
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@latest
```

# Examples
- [json](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/json/basic/main.go?ref_type=heads)
- [oauth](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/json/oauth/main.go?ref_type=heads)
- [ioc](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/ioc/main.go)
- [caching](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/json/iskaypet/main.go?ref_type=heads)
- [xml](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/xml/main.go?ref_type=heads) 
- [bytes](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/bytes/main.go?ref_type=heads)
- [form](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/form/main.go?ref_type=heads)
- [redirect](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/blob/main/examples/redirect/main.go?ref_type=heads)

# Quickstart

```go
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
    // Create a new context with a timeout of 5 seconds
    // This will automatically cancel the request if it takes longer than 5 seconds to complete
    ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(5000))
    defer cancel()

    // Create a new REST client with custom settings
    client := &rest.Client{
        Name:           "example-client",                       // required for logging and tracing
        BaseURL:        "https://gorest.co.in/public/v2",       // optional parameters
        ContentType:    rest.JSON,                              // rest.JSON by default
        Timeout:        time.Millisecond * time.Duration(2000), // transmission timeout
        ConnectTimeout: time.Millisecond * time.Duration(5000), // socket timeout
        /*DisableCache:   false,                                  // Last-Modified and ETag headers are enabled by default
          CustomPool: &rest.CustomPool{ // for fine-tuning the connection pool
          	Transport: &http.Transport{
          		IdleConnTimeout:       time.Duration(2000) * time.Millisecond,
          		ResponseHeaderTimeout: time.Duration(2000) * time.Millisecond,
          		MaxIdleConnsPerHost:   10,
          	},
          },
          BasicAuth: &rest.BasicAuth{
          	Username: "your_username",
          	Password: "your_password",
          },
          OAuth: &rest.OAuth{
          	ClientID:     "your_client_id",
          	ClientSecret: "your_client_secret",
          	TokenURL:     "https://oauth.gorest.co.in/oauth/token",
          	AuthStyle:    rest.AuthStyleInHeader,
          },
          EnableTrace:    true,
          UserAgent:      "<Your User Agent>",
          DisableTimeout: false,
          FollowRedirect: false,*/
    }

    // Set headers for the request (optional)
    headers := make(http.Header)
    headers.Add("My-Custom-Header", "My-Custom-Value")

    // Make a GET request (context optional)
    response := client.GetWithContext(ctx, "/users", headers)
    if response.Err != nil {
        fmt.Printf("Error: %v\n", response.Err)
        os.Exit(1)
    }

    // Check status code and handle errors accordingly or response.IsOk()
    if response.StatusCode != http.StatusOK {
        fmt.Printf("Status: %d, Body: %s", response.StatusCode, response.String())
        os.Exit(1)
    }

    // Untyped fill up
    var users []struct {
        ID    int    `json:"id"`
        Name  string `json:"name"`
        Email string `json:"email"`
    }

    // Untyped fill up or typed with rest.Deserialize[struct | []struct](response)
    if err := response.FillUp(&users); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    // Print the users
    for i := range users {
        fmt.Printf("User: %d, Name: %s, Email: %s\n", users[i].ID, users[i].Name, users[i].Email)
    }
}

```

## Output

```text
User: 7527456, Name: Hiranmay Dhawan IV, Email: dhawan_hiranmay_iv@grant.test
User: 7527454, Name: Daevika Khan, Email: khan_daevika@stark.example
User: 7527452, Name: Msgr. Baalagopaal Dubashi, Email: msgr_baalagopaal_dubashi@beatty.test
User: 7527451, Name: Tanirika Johar, Email: johar_tanirika@okeefe.example
User: 7527450, Name: Gopaal Nehru, Email: nehru_gopaal@white-harris.test
User: 7527449, Name: Bhagwanti Kapoor, Email: kapoor_bhagwanti@hahn.example
User: 7527448, Name: Brahmanandam Reddy, Email: brahmanandam_reddy@corkery-cormier.example
User: 7527447, Name: Bela Bhattathiri, Email: bhattathiri_bela@nicolas.example
User: 7527446, Name: Poornima Tandon, Email: poornima_tandon@collier.test
User: 7527445, Name: Dhyaneshwar Reddy, Email: dhyaneshwar_reddy@brown.test
```

## Metrics

```text
# HELP __go_restclient_cache_buffer_items 
# TYPE __go_restclient_cache_buffer_items gauge
__go_restclient_cache_buffer_items 64
# HELP __go_restclient_cache_cost_added_bytes_total 
# TYPE __go_restclient_cache_cost_added_bytes_total counter
__go_restclient_cache_cost_added_bytes_total 13109
# HELP __go_restclient_cache_cost_evicted_bytes_total 
# TYPE __go_restclient_cache_cost_evicted_bytes_total counter
__go_restclient_cache_cost_evicted_bytes_total 2279
# HELP __go_restclient_cache_gets_dropped_total 
# TYPE __go_restclient_cache_gets_dropped_total counter
__go_restclient_cache_gets_dropped_total 0
# HELP __go_restclient_cache_gets_kept_total 
# TYPE __go_restclient_cache_gets_kept_total counter
__go_restclient_cache_gets_kept_total 0
# HELP __go_restclient_cache_hits_total 
# TYPE __go_restclient_cache_hits_total counter
__go_restclient_cache_hits_total 8
# HELP __go_restclient_cache_keys_added_total 
# TYPE __go_restclient_cache_keys_added_total counter
__go_restclient_cache_keys_added_total 23
# HELP __go_restclient_cache_keys_evicted_total 
# TYPE __go_restclient_cache_keys_evicted_total counter
__go_restclient_cache_keys_evicted_total 4
# HELP __go_restclient_cache_keys_updated_total 
# TYPE __go_restclient_cache_keys_updated_total counter
__go_restclient_cache_keys_updated_total 1
# HELP __go_restclient_cache_max_cost_bytes 
# TYPE __go_restclient_cache_max_cost_bytes gauge
__go_restclient_cache_max_cost_bytes 1.073741824e+09
# HELP __go_restclient_cache_misses_total 
# TYPE __go_restclient_cache_misses_total counter
__go_restclient_cache_misses_total 48
# HELP __go_restclient_cache_num_counters 
# TYPE __go_restclient_cache_num_counters gauge
__go_restclient_cache_num_counters 1e+07
# HELP __go_restclient_cache_ratio 
# TYPE __go_restclient_cache_ratio gauge
__go_restclient_cache_ratio 0.14285714285714285
# HELP __go_restclient_cache_sets_dropped_total 
# TYPE __go_restclient_cache_sets_dropped_total counter
__go_restclient_cache_sets_dropped_total 0
# HELP __go_restclient_cache_sets_rejected_total 
# TYPE __go_restclient_cache_sets_rejected_total counter
__go_restclient_cache_sets_rejected_total 0
# HELP __go_restclient_durations_seconds 
# TYPE __go_restclient_durations_seconds summary
__go_restclient_durations_seconds{client_name="gorest-client",quantile="0.5"} 113
__go_restclient_durations_seconds{client_name="gorest-client",quantile="0.95"} 586
__go_restclient_durations_seconds{client_name="gorest-client",quantile="0.99"} 649
__go_restclient_durations_seconds_sum{client_name="gorest-client"} 4298
__go_restclient_durations_seconds_count{client_name="gorest-client"} 24
# HELP __go_restclient_requests_total 
# TYPE __go_restclient_requests_total counter
__go_restclient_requests_total{client_name="gorest-client",status_code="200"} 24
```

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
package rest_test

import (
    "net/http"
    "strconv"
    "testing"

    "github.com/go-resty/resty/v2"
    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

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
        resp, err := client.R().Get("https://gorest.co.in/public/v2/users")
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

### resty

    goos: darwin
    goarch: arm64
    pkg: gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest
    cpu: Apple M3 Pro
    BenchmarkResty_Get
    BenchmarkResty_Get-10    	       1	1099176917 ns/op
    PASS

### go-restclient

    goos: darwin
    goarch: arm64
    pkg: gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest
    cpu: Apple M3 Pro
    BenchmarkGet 
    BenchmarkGet-10    	       3	 405502875 ns/op
    PASS

    Hug!
