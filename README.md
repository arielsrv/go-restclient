# Go RESTClient

[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

A high-performance HTTP client library for Go with advanced features including caching, authentication, metrics, and comprehensive request/response handling.

## ✨ Features

- **HTTP Methods**: Full support for GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS
- **Smart Caching**: Response caching based on HTTP headers (`cache-control`, `last-modified`, `etag`, `expires`)
- **Content Types**: Automatic marshaling/unmarshaling for JSON, XML, and Form data
- **Authentication**: Built-in support for Basic Auth and OAuth2 Client Credentials
- **Connection Pooling**: Configurable connection pools for optimal performance
- **Metrics & Tracing**: Prometheus metrics and OpenTelemetry tracing support
- **Error Handling**: RFC7807 Problem Details support
- **Concurrent Safety**: Thread-safe operations with proper mutex protection

## 📋 Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Features](#features)
- [Examples](#examples)
- [Configuration](#configuration)
- [Caching](#caching)
- [Authentication](#authentication)
- [Metrics & Monitoring](#metrics--monitoring)
- [Benchmarks](#benchmarks)
- [Connection Pooling](#connection-pooling)
- [Roadmap](#roadmap)

## 🚀 Installation

```bash
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@latest
```

## ⚡ Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Initialize REST client
    client := &rest.Client{
        Name:           "my-api-client",
        BaseURL:        "https://api.example.com",
        ContentType:    rest.JSON,
        Timeout:        2 * time.Second,
        ConnectTimeout: 5 * time.Second,
        EnableCache:    true,
        EnableTrace:    true,
    }

    // Make request
    response := client.GetWithContext(ctx, "/users")
    if response.Err != nil {
        fmt.Printf("Error: %v\n", response.Err)
        return
    }

    // Handle response
    if !response.IsOk() {
        fmt.Printf("HTTP %d: %s\n", response.StatusCode, response.String())
        return
    }

    // Parse JSON response
    var users []User
    if err := response.FillUp(&users); err != nil {
        fmt.Printf("Parse error: %v\n", err)
        return
    }

    fmt.Printf("Retrieved %d users\n", len(users))
}

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}
```

## 🔧 Configuration

### Client Configuration

```go
client := &rest.Client{
    // Required
    Name: "my-client",
    
    // Optional
    BaseURL:        "https://api.example.com",
    ContentType:    rest.JSON, // rest.JSON, rest.XML, rest.FORM
    Timeout:        2 * time.Second,
    ConnectTimeout: 5 * time.Second,
    EnableCache:    true,
    EnableTrace:    true,
    UserAgent:      "MyApp/1.0",
    FollowRedirect: true,
    DisableTimeout: false,
    
    // Authentication
    BasicAuth: &rest.BasicAuth{
        Username: "user",
        Password: "pass",
    },
    
    OAuth: &rest.OAuth{
        ClientID:     "client_id",
        ClientSecret: "client_secret",
        TokenURL:     "https://oauth.example.com/token",
        AuthStyle:    rest.AuthStyleInHeader,
    },
    
    // Connection Pool
    CustomPool: &rest.CustomPool{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxConnsPerHost:     100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    },
}
```

## 💾 Caching

The library provides intelligent response caching based on HTTP headers:

- **Cache-Control**: Respects `max-age`, `no-cache`, `no-store` directives
- **ETag**: Supports ETag-based validation
- **Last-Modified**: Uses modification dates for cache validation
- **Expires**: Respects expiration headers

```go
client := &rest.Client{
    Name:        "cached-client",
    EnableCache: true, // Enable caching
}

// First request - hits the server
response1 := client.Get("/api/data")

// Second request - served from cache (if valid)
response2 := client.Get("/api/data")
```

## 🔐 Authentication

### Basic Authentication

```go
client := &rest.Client{
    Name: "basic-auth-client",
    BasicAuth: &rest.BasicAuth{
        Username: "username",
        Password: "password",
    },
}
```

### OAuth2 Client Credentials

```go
client := &rest.Client{
    Name: "oauth-client",
    OAuth: &rest.OAuth{
        ClientID:     "your_client_id",
        ClientSecret: "your_client_secret",
        TokenURL:     "https://oauth.example.com/token",
        AuthStyle:    rest.AuthStyleInHeader, // or rest.AuthStyleInParams
    },
}
```

## 📊 Metrics & Monitoring

The library automatically exposes Prometheus metrics for monitoring:

### Available Metrics

- `__go_restclient_requests_total`: Total request count by status code
- `__go_restclient_durations_seconds`: Request duration percentiles
- `__go_restclient_cache_hits_total`: Cache hit count
- `__go_restclient_cache_misses_total`: Cache miss count
- `__go_restclient_cache_ratio`: Cache hit ratio

### Grafana Dashboard

Access the monitoring dashboard at:
[HTTP Clients Dashboard](https://iskaylog.grafana.net/d/ddmgmir2jckxsb/http-clients)

### Requirements

- Prometheus collector endpoint enabled
- Environment variable set (`dev|uat|pro|any`)
- Application name variable set (`APP_NAME`)

## 📈 Benchmarks

Performance comparison with popular HTTP clients:

| Client | Operations/sec | Latency |
|--------|---------------|---------|
| go-restclient | ~2.5M ops/sec | 405ms |
| resty | ~0.9M ops/sec | 1099ms |

*Benchmarks run on Apple M3 Pro*

## 🔗 Connection Pooling

Optimize performance with proper connection pool configuration:

```go
client := &rest.Client{
    Name: "optimized-client",
    CustomPool: &rest.CustomPool{
        Transport: &http.Transport{
            MaxIdleConns:        100,  // Maximum idle connections
            MaxConnsPerHost:     100,  // Maximum connections per host
            MaxIdleConnsPerHost: 100,  // Maximum idle connections per host
            IdleConnTimeout:     90 * time.Second,
            ResponseHeaderTimeout: 2 * time.Second,
        },
    },
}
```

**Benefits:**
- Reduces TIME_WAIT states
- Enables persistent connections
- Improves response times
- Reduces CPU usage

## 📚 Examples

Explore comprehensive examples in the `examples/` directory:

- [Basic JSON](examples/json/basic/main.go) - Simple JSON requests
- [OAuth2](examples/json/oauth/main.go) - OAuth2 client credentials flow
- [Caching](examples/json/iskaypet/main.go) - Response caching strategies
- [XML](examples/xml/main.go) - XML request/response handling
- [Form Data](examples/form/main.go) - Form data submission
- [File Upload](examples/bytes/main.go) - File upload handling
- [Metrics](examples/metrics/main.go) - Metrics collection
- [Tracing](examples/trace/main.go) - Request tracing

### Live Example

```bash
APP_NAME=example go run gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/examples/metrics@latest
```

## 🛣️ Roadmap

- [ ] **Distributed Caching**: Configurable non-HTTP-RFC distributed cache support
- [ ] **Custom Encoders**: Configurable JSON encoder/decoder (e.g., [go-json](https://github.com/goccy/go-json))
- [ ] **Interceptors**: Custom request/response interceptors as pipelines
- [ ] **PKCE Support**: OAuth2 PKCE flow implementation
- [ ] **Rate Limiting**: Built-in rate limiting capabilities

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines for more details.

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Built with ❤️ by the Iskaypetcom team**
