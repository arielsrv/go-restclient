# go-restclient

> A high-performance, feature-rich HTTP client for Go — with smart caching, OAuth2, async requests, OpenTelemetry tracing, and a built-in mock server.

[![Go Reference](https://pkg.go.dev/badge/github.com/arielsrv/go-restclient.svg)](https://pkg.go.dev/github.com/arielsrv/go-restclient)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arielsrv/go-restclient)](https://go.dev/)
[![Build Status](https://github.com/arielsrv/go-restclient/actions/workflows/go.yml/badge.svg)](https://github.com/arielsrv/go-restclient/actions/workflows/go.yml)
[![Lint Status](https://github.com/arielsrv/go-restclient/actions/workflows/lint.yml/badge.svg)](https://github.com/arielsrv/go-restclient/actions/workflows/lint.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Coverage](https://img.shields.io/badge/Coverage-1-red)

---

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Client Configuration](#client-configuration)
- [HTTP Methods & Response Handling](#http-methods--response-handling)
- [Caching — How It Works](#caching--how-it-works)
- [Authentication](#authentication)
- [Async Requests](#async-requests)
- [Gzip Compression](#gzip-compression)
- [Connection Pooling](#connection-pooling)
- [OpenTelemetry Tracing](#opentelemetry-tracing)
- [RFC 7807 Problem Details](#rfc-7807-problem-details)
- [Built-in Mock Server](#built-in-mock-server)
- [Benchmarks](#benchmarks)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)

---

## Features

| Category | What you get |
|---|---|
| **HTTP Methods** | `GET` `POST` `PUT` `PATCH` `DELETE` `HEAD` `OPTIONS` — sync and async variants, all with `WithContext` support |
| **Smart Caching** | TTL via `Cache-Control` / `Expires`, ETag + Last-Modified revalidation, Ristretto LFU backend, weak-pointer GC safety |
| **Content Types** | JSON, XML, `application/x-www-form-urlencoded` — auto-detect on response via `Content-Type` header |
| **Authentication** | Basic Auth and OAuth2 Client Credentials (header, params, or auto-detect style) |
| **Async Requests** | Channel-based async API — non-blocking, buffered, closed automatically |
| **Gzip** | `EnableGzip` flag — adds `Accept-Encoding: gzip` and decompresses transparently |
| **Connection Pooling** | Shared default transport or per-client `CustomPool` with proxy support |
| **OpenTelemetry** | `EnableTrace` — integrates `otelhttp` + `otelhttptrace` for full span propagation |
| **RFC 7807** | Automatic `application/problem+json` / `application/problem+xml` deserialization |
| **Mock Server** | Built-in `StartMockupServer()` / `AddMockups()` for unit testing without real HTTP |
| **Concurrency** | Thread-safe — mutex-protected client initialization, sync-atomic cache flags |

> **Requires Go 1.23+** (uses the `weak` package for GC-safe cache pointers)

---

## Installation

```bash
go get github.com/arielsrv/go-restclient
```

---

## Quick Start

### Using the package-level default client

```go
package main

import (
    "fmt"
    "github.com/arielsrv/go-restclient/rest"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    response := rest.Get("https://gorest.co.in/public/v2/users/1")
    if response.Err != nil {
        panic(response.Err)
    }

    var user User
    if err := response.FillUp(&user); err != nil {
        panic(err)
    }
    fmt.Printf("User: %+v\n", user)
}
```

### Using a custom `Client`

```go
client := &rest.Client{
    Name:        "my-service",
    BaseURL:     "https://api.example.com",
    ContentType: rest.JSON,
    Timeout:     2 * time.Second,
    EnableCache: true,
}

// Typed deserialization using generics
response := client.GetWithContext(ctx, "/users/1")
user, err := rest.Deserialize[User](response)
```

---

## Client Configuration

```go
client := &rest.Client{
    // Identity
    Name:    "my-service",        // label used in metrics; defaults to hostname
    BaseURL: "https://api.example.com",

    // Timeouts
    Timeout:        2 * time.Second,  // response-header timeout (default: 500ms)
    ConnectTimeout: 5 * time.Second,  // TCP dial timeout (default: 1500ms)
    DisableTimeout: false,            // set true to disable all timeouts

    // Encoding
    ContentType: rest.JSON,   // rest.JSON | rest.XML | rest.FORM

    // Features
    EnableCache:    true,   // enable HTTP response caching
    EnableGzip:     true,   // send Accept-Encoding: gzip, decompress response
    EnableTrace:    true,   // enable OpenTelemetry tracing
    FollowRedirect: true,   // follow 3xx redirects (default: false)
    UserAgent:      "MyApp/1.0",

    // Authentication (mutually exclusive — OAuth takes precedence)
    BasicAuth: &rest.BasicAuth{
        Username: "user",
        Password: "pass",
    },
    OAuth: &rest.OAuth{
        ClientID:     "client_id",
        ClientSecret: "client_secret",
        TokenURL:     "https://auth.example.com/token",
        Scopes:       []string{"read", "write"},
        AuthStyle:    rest.AuthStyleInHeader, // AuthStyleAutoDetect | AuthStyleInHeader | AuthStyleInParams
    },

    // Default headers applied to every request
    DefaultHeaders: http.Header{
        "X-Tenant-ID": {"acme-corp"},
    },

    // Custom connection pool (optional)
    CustomPool: &rest.CustomPool{
        MaxIdleConnsPerHost: 100,
        Proxy:               "http://proxy.internal:8080",
    },
}
```

### Configuration reference

| Field | Type | Default | Description |
|---|---|---|---|
| `Name` | `string` | hostname | Identifier used in Prometheus metrics |
| `BaseURL` | `string` | `""` | Prefix prepended to every request URL |
| `Timeout` | `Duration` | 500ms | Response-header read timeout |
| `ConnectTimeout` | `Duration` | 1500ms | TCP dial timeout |
| `DisableTimeout` | `bool` | `false` | Disables all timeouts |
| `ContentType` | `ContentType` | `JSON` | Default body encoding for requests |
| `EnableCache` | `bool` | `false` | Enables HTTP response caching |
| `EnableGzip` | `bool` | `false` | Enables gzip compression |
| `EnableTrace` | `bool` | `false` | Enables OpenTelemetry tracing |
| `FollowRedirect` | `bool` | `false` | Follows 3xx redirects automatically |
| `UserAgent` | `string` | `go-restclient/…` | Custom `User-Agent` header |
| `BasicAuth` | `*BasicAuth` | `nil` | Basic Auth credentials |
| `OAuth` | `*OAuth` | `nil` | OAuth2 Client Credentials config |
| `DefaultHeaders` | `http.Header` | `nil` | Headers applied to every request |
| `CustomPool` | `*CustomPool` | `nil` | Per-client transport / connection pool |

---

## HTTP Methods & Response Handling

### Synchronous

```go
// Read
resp := client.Get("/users")
resp := client.GetWithContext(ctx, "/users", optionalHeaders...)
resp := client.Head("/users")
resp := client.Options("/users")

// Write
resp := client.Post("/users", body)
resp := client.Put("/users/1", body)
resp := client.Patch("/users/1", partial)
resp := client.Delete("/users/1")
```

### Working with `*Response`

```go
// Check for transport error
if resp.Err != nil { ... }

// Check HTTP status (2xx–3xx)
if !resp.IsOk() { ... }

// Unified error check (transport + status)
if err := resp.VerifyIsOkOrError(); err != nil { ... }

// Deserialize body (auto-detects Content-Type)
var user User
resp.FillUp(&user)

// Generic helper
user, err := rest.Deserialize[User](resp)

// Raw body
fmt.Println(resp.String())

// Was this served from cache?
fmt.Println(resp.Cached())

// RFC 7807 problem (populated automatically)
fmt.Println(resp.Problem)

// Full request+response dump for debugging
fmt.Println(resp.Debug())
```

---

## Caching — How It Works

Set `EnableCache: true` on the client. Only read operations (`GET`, `HEAD`, `OPTIONS`) are cached.

The cache backend is [Ristretto](https://github.com/dgraph-io/ristretto) — an LFU cache with TTL
support and weak-pointer entries so that the GC can reclaim memory when the system is under
pressure. The default maximum size is **256 MB** (configurable via `rest.MaxCacheSize`).

```go
client := &rest.Client{
    EnableCache: true,
}
```

### Caching modes

The behavior depends entirely on which HTTP cache headers the **server** returns.

---

#### Mode 1 — TTL (`Cache-Control: max-age` or `Expires`)

```text
Server response headers:
  Cache-Control: max-age=60
  (or)
  Expires: Mon, 21 Apr 2026 12:00:00 GMT
```

| Request | What happens |
|---|---|
| 1st | Hits the server → response stored with TTL |
| 2nd … Nth (within TTL) | Served **directly from cache** — no network call |
| After TTL expires | Cache entry evicted → fresh request to the server |

> **Best for**: static or slowly-changing resources. Zero network overhead within the TTL window.

---

#### Mode 2 — ETag (no TTL)

```text
Server response headers:
  ETag: "abc123"
```

| Request | What the client sends | Server responds | Result |
|---|---|---|---|
| 1st | — | `200` + `ETag: "v1"` | Cached with `revalidate=true` |
| 2nd | `If-None-Match: "v1"` | `304 Not Modified` | Cached response returned — no body transfer |
| 2nd (content changed) | `If-None-Match: "v1"` | `200` + `ETag: "v2"` | Cache **force-updated** with new response |
| 3rd (after update) | `If-None-Match: "v2"` | `304 Not Modified` | Updated cached response returned |

> **Key property**: every request hits the network for a cheap conditional check
> (`304` means no body is transferred). If the content has changed, the new response
> always replaces the stale one.

---

#### Mode 3 — Last-Modified (no TTL)

Same as Mode 2 but uses `If-Modified-Since` / `Last-Modified` headers instead of `If-None-Match` / `ETag`.

---

#### Mode 4 — ETag or Last-Modified + TTL (combined)

Within the TTL window, requests are served from cache with **no network call** (same as Mode 1).
After the TTL expires, the cache entry is evicted and a fresh conditional request is made using
`If-None-Match` or `If-Modified-Since`.

---

### The `revalidate` flag

Internally, responses are marked with `revalidate = true` when they carry an ETag or Last-Modified
but **no** TTL. This flag tells the client to always go to the network for a conditional check
rather than returning the cached entry directly.

### Cache decision flow

```text
GET /resource
     │
     ▼
 Cache hit?
     │
   ──┴──
  No    Yes
  │      │
  │    revalidate=false? ──► return cached response
  │      │
  │    revalidate=true
  │      │
  │      ▼
  │   Send conditional request (If-None-Match / If-Modified-Since)
  │      │
  │   ───┴───────────────────
  │   304              200
  │   │                │
  │   return        force-update cache
  │   cached        return new response
  │   response
  │
  ▼
Send fresh request → cache response → return
```

---

## Authentication

### Basic Auth

```go
client := &rest.Client{
    BasicAuth: &rest.BasicAuth{
        Username: "alice",
        Password: "s3cr3t",
    },
}
```

### OAuth2 Client Credentials

```go
client := &rest.Client{
    OAuth: &rest.OAuth{
        ClientID:     "my-client",
        ClientSecret: "my-secret",
        TokenURL:     "https://auth.example.com/oauth/token",
        Scopes:       []string{"api:read"},
        AuthStyle:    rest.AuthStyleInHeader,
        // EndpointParams: url.Values{"audience": {"https://api.example.com"}},
    },
}
```

> OAuth2 takes precedence over Basic Auth. The library handles token acquisition and refresh automatically using `golang.org/x/oauth2`.

---

## Async Requests

Every HTTP method has an async variant that returns a `<-chan *Response`. The channel is buffered (size 1) and closed automatically after the response is sent.

```go
ch := client.AsyncGet("/users")

// Do other work here...

resp := <-ch
if resp.Err != nil {
    // handle error
}

var users []User
resp.FillUp(&users)
```

```go
// Multiple concurrent requests
ch1 := client.AsyncGetWithContext(ctx, "/users")
ch2 := client.AsyncGetWithContext(ctx, "/products")

users, _ := rest.Deserialize[[]User](<-ch1)
products, _ := rest.Deserialize[[]Product](<-ch2)
```

---

## Gzip Compression

```go
client := &rest.Client{
    EnableGzip: true,
}

// The client automatically:
// 1. Adds "Accept-Encoding: gzip" to every request
// 2. Decompresses gzip responses transparently
resp := client.Get("/large-dataset")
```

You can also send the header manually per-request if you prefer not to enable it globally:

```go
headers := http.Header{"Accept-Encoding": {"gzip"}}
resp := client.GetWithContext(ctx, "/data", headers)
```

---

## Connection Pooling

By default, all `Client` instances share a single `*http.Transport` (the global default pool). If you need per-client isolation, a custom pool, or a proxy:

```go
client := &rest.Client{
    CustomPool: &rest.CustomPool{
        MaxIdleConnsPerHost: 100,
        Proxy:               "http://proxy.internal:3128",
    },
}
```

Or bring your own transport:

```go
client := &rest.Client{
    CustomPool: &rest.CustomPool{
        Transport: &http.Transport{
            MaxIdleConns:        200,
            MaxConnsPerHost:     200,
            MaxIdleConnsPerHost: 200,
            IdleConnTimeout:     90 * time.Second,
        },
    },
}
```

---

## OpenTelemetry Tracing

Set `EnableTrace: true` and the client will:

- Hook into `net/http/httptrace` to create OTel spans for DNS, connect, TLS, and request/response lifecycle events
- Use the incoming `ctx` to propagate the active span, so HTTP client spans become children of your service spans automatically

```go
client := &rest.Client{
    Name:        "my-service",
    EnableTrace: true,
}

// Initialize your TracerProvider before making requests
// (see examples/trace/main.go for a full OTLP setup)
resp := client.GetWithContext(ctx, "https://api.example.com/data")
```

### Minimal stdout setup (for local development)

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func initTracing() func(context.Context) error {
    exp, _ := stdouttrace.New(stdouttrace.WithPrettyPrint())
    tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
    otel.SetTracerProvider(tp)
    return tp.Shutdown
}
```

### Common OTLP environment variables

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
export OTEL_SERVICE_NAME=my-service
export OTEL_RESOURCE_ATTRIBUTES=deployment.environment=local
```

---

## RFC 7807 Problem Details

If a response has a `Content-Type` of `application/problem+json` or `application/problem+xml`, the library automatically deserializes the body into `response.Problem`:

```go
resp := client.Get("/resource")

if resp.Problem != nil {
    fmt.Printf("type:   %s\n", resp.Problem.Type)
    fmt.Printf("title:  %s\n", resp.Problem.Title)
    fmt.Printf("detail: %s\n", resp.Problem.Detail)
    fmt.Printf("status: %d\n", resp.Problem.Status)
}
```

The `Problem` struct follows [RFC 7807](https://datatracker.ietf.org/doc/html/rfc7807) and supports both JSON and XML formats.

---

## Built-in Mock Server

The library ships with a complete HTTP mock server, designed for unit testing — no real network calls required.

### Basic usage

```go
func TestMyClient(t *testing.T) {
    rest.StartMockupServer()
    defer rest.StopMockupServer()

    err := rest.AddMockups(&rest.Mock{
        URL:          "https://api.example.com/users/1",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusOK,
        RespBody:     `{"id":1,"name":"Alice"}`,
        RespHeaders:  http.Header{"Content-Type": {"application/json"}},
    })
    require.NoError(t, err)

    client := &rest.Client{Name: "test"}
    resp := client.Get("https://api.example.com/users/1")

    require.NoError(t, resp.Err)
    require.Equal(t, http.StatusOK, resp.StatusCode)
    require.Equal(t, `{"id":1,"name":"Alice"}`, resp.String())
}
```

### Mock fields

| Field | Type | Description |
|---|---|---|
| `URL` | `string` | Full URL to match (query params are order-normalized) |
| `HTTPMethod` | `string` | HTTP method to match (`http.MethodGet`, etc.) |
| `RespHTTPCode` | `int` | Status code to return |
| `RespBody` | `string` | Response body |
| `RespHeaders` | `http.Header` | Response headers |
| `ReqHeaders` | `http.Header` | Expected request headers (informational) |
| `ReqBody` | `string` | Expected request body (informational) |
| `Timeout` | `time.Duration` | Simulate a slow/timeout response |

### Simulating caching headers

```go
rest.AddMockups(&rest.Mock{
    URL:          "https://api.example.com/data",
    HTTPMethod:   http.MethodGet,
    RespHTTPCode: http.StatusOK,
    RespBody:     `{"version":1}`,
    RespHeaders: http.Header{
        "Content-Type":  {"application/json"},
        "Cache-Control": {"max-age=60"},
    },
})
```

### Simulating ETag revalidation

```go
// First call: return 200 + ETag
rest.AddMockups(&rest.Mock{
    URL:          "https://api.example.com/resource",
    HTTPMethod:   http.MethodGet,
    RespHTTPCode: http.StatusOK,
    RespBody:     `{"version":1}`,
    RespHeaders:  http.Header{"ETag": {"v1"}, "Content-Type": {"application/json"}},
})

client := &rest.Client{EnableCache: true}
r1 := client.Get("https://api.example.com/resource") // 200, cached

// Replace mock to simulate 304
rest.FlushMockups()
rest.AddMockups(&rest.Mock{
    URL:          "https://api.example.com/resource",
    HTTPMethod:   http.MethodGet,
    RespHTTPCode: http.StatusNotModified,
})

r2 := client.Get("https://api.example.com/resource") // served from cache
fmt.Println(r2.Cached()) // true
```

### Using the `-mock` CLI flag

```bash
go test -mock ./...
```

### Managing mocks

```go
rest.FlushMockups()       // clear all registered mocks
rest.StopMockupServer()   // stop the server and reset state
```

---

## Benchmarks

All benchmarks run against an **in-process `httptest.Server`** (zero external network latency)
on an Apple M1 Pro. Source: [`rest/benchmark_compare_test.go`](rest/benchmark_compare_test.go).

```bash
go test ./rest/... -bench=. -benchmem -benchtime=5s -count=1 -run=^$
```

### Results

| Scenario | Library | ns/op | B/op | allocs/op | vs resty |
|---|---|---:|---:|---:|---|
| Plain GET | **go-restclient** | **47,888** | **8,452** | **85** | **+1.7% faster, −7% memory, −7 allocs** |
| Plain GET | resty/v2 | 48,700 | 9,821 | 91 | baseline |
| Cached GET (TTL) | **go-restclient** | **125** | **55** | **1** | **390× faster, −99.4% memory** |
| Cached GET (TTL) | resty/v2 *(no cache)* | 49,509 | 9,913 | 92 | baseline |
| Slow GET (100 ms) | **go-restclient** | 101,574,787 | **10,173** | **87** | **−16% memory, −6 allocs** |
| Slow GET (100 ms) | resty/v2 | 101,422,116 | 12,153 | 93 | baseline |

### Takeaways

- **Plain GET**: go-restclient is faster **and** uses less memory. Savings come from
  eliminating a redundant `url.Parse` + `fmt.Sprintf`, caching per-request header
  `[]string` allocations, and inlining the hot path.
- **Cached GET (TTL)**: once a response is cached, subsequent reads cost only **125 ns**
  and **55 B** — a **390× speedup** over resty's uncached call. resty has no built-in
  HTTP cache.
- **Slow GET**: when the bottleneck is handler latency (100 ms), both libraries perform
  identically in time. go-restclient still wins on memory (−16%) and allocations (−6).

---

## Examples

The `examples/` directory contains runnable programs for every major feature:

| Example | Description |
|---|---|
| [`basic/`](examples/basic/main.go) | Simple JSON GET + POST |
| [`generics/`](examples/generics/main.go) | Typed deserialization with `rest.Deserialize[T]` |
| [`etag/`](examples/etag/main.go) | ETag revalidation — unchanged (304) and content-changed (200 → force-update) |
| [`html/`](examples/html/main.go) | Caching with `Last-Modified` |
| [`gzip/`](examples/gzip/main.go) | Gzip transparent decompression |
| [`gzip_headers/`](examples/gzip_headers/main.go) | Per-request `Accept-Encoding` |
| [`oauth/`](examples/oauth/main.go) | OAuth2 Client Credentials (mock token server + mock API) |
| [`dfltheaders/`](examples/dfltheaders/main.go) | Default + per-request headers |
| [`headers/`](examples/headers/main.go) | Custom headers |
| [`xml/`](examples/xml/main.go) | XML content type |
| [`form/`](examples/form/main.go) | `application/x-www-form-urlencoded` |
| [`bytes/`](examples/bytes/main.go) | Binary upload / large downloads |
| [`binary_download/`](examples/binary_download/main.go) | Download a file to disk |
| [`str/`](examples/str/main.go) | Raw string response handling |
| [`patch_options/`](examples/patch_options/main.go) | PATCH and OPTIONS verbs |
| [`redirect/`](examples/redirect/main.go) | Redirect handling |
| [`timeout/`](examples/timeout/main.go) | Timeout configuration |
| [`ioc/`](examples/ioc/main.go) | Dependency injection pattern |
| [`mock/`](examples/mock/main.go) | Mock server usage |
| [`iterator/`](examples/iterator/main.go) | Async/iterator pattern |
| [`problem/`](examples/problem/main.go) | RFC 7807 problem details (auto-deserialized from `application/problem+json`) |
| [`metrics/`](examples/metrics/main.go) | Prometheus metrics — starts an HTTP server on `:8081` |
| [`trace/`](examples/trace/main.go) | OpenTelemetry tracing — starts an HTTP server on `:8081` |

Run any example:

```bash
go run examples/basic/main.go
go run examples/etag/main.go
ENV=local APP_NAME=example go run examples/metrics/main.go
```

---

## Contributing

Issues and pull requests are welcome. Please make sure all tests pass before submitting:

```bash
go test ./rest/... -count=1
```

---

## License

MIT — see [LICENSE](LICENSE) for details.
