# Go RESTClient

[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)
[![Quality Gate Status](https://sonarqube.tooling.dp.iskaypet.com/api/project_badges/measure?project=iskaypetcom_digital_sre_tools_dev_go-restclient_44a86603-3e76-44e9-b025-4472c8491e3c&metric=alert_status&token=sqb_e80727065210b0976bf7ac69df58870f05717559)](https://sonarqube.tooling.dp.iskaypet.com/dashboard?id=iskaypetcom_digital_sre_tools_dev_go-restclient_44a86603-3e76-44e9-b025-4472c8491e3c)


A high-performance HTTP client library for Go with advanced features including caching,
authentication, metrics, and comprehensive request/response handling.

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

- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Features](#-features)
- [Examples](#-examples)
- [Configuration](#-configuration)
- [Caching](#-caching)
- [Authentication](#-authentication)
- [Metrics & Monitoring](#-metrics--monitoring)
- [Benchmarks](#-benchmarks)
- [Connection Pooling](#-connection-pooling)
- [Roadmap](#%EF%B8%8F-roadmap)

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

### Basic Usage

#### Simple JSON Requests
```go
// examples/json/basic/main.go
client := &rest.Client{
    Name:        "example-client",
    BaseURL:     "https://gorest.co.in/public/v2",
    ContentType: rest.JSON,
    Timeout:     2 * time.Second,
}

response := client.Get("/users")
if response.Err != nil {
    log.Fatal(response.Err)
}

var users []User
if err := response.FillUp(&users); err != nil {
    log.Fatal(err)
}
```

#### Using Generics for Type Safety
```go
// examples/json/generics/main.go
type UserResponse struct {
    Name   string `json:"name"`
    Email  string `json:"email"`
    Gender string `json:"gender"`
    Status string `json:"status"`
    ID     int    `json:"id"`
}

// Typed deserialization with generics
usersResponse, err := rest.Deserialize[[]UserResponse](response)
if err != nil {
    log.Fatal(err)
}
```

#### Custom Headers and Default Headers
```go
// examples/json/dfltheaders/main.go
client := &rest.Client{
    Name:           "example-client",
    BaseURL:        "https://gorest.co.in/public/v2",
    ContentType:    rest.JSON,
    DefaultHeaders: http.Header{
        "X-Static-Header": {"My-Static-Value"},
    },
}

// Dynamic headers for specific requests
headers := make(http.Header)
headers.Set("My-Dynamic-Header-1", "My-Dynamic-Value-1")
headers.Set("My-Dynamic-Header-2", "My-Dynamic-Value-2")

response := client.GetWithContext(ctx, "/users", headers)
```

### Authentication

#### OAuth2 Client Credentials
```go
// examples/json/oauth/main.go
client := &rest.Client{
    Name:        "ocapi-client",
    BaseURL:     "https://www.kiwoko.com/s/-/dw/data/v22_6",
    ContentType: rest.JSON,
    OAuth: &rest.OAuth{
        ClientID:     "your_client_id",
        ClientSecret: "your_client_secret",
        TokenURL:     "https://account.demandware.com/dw/oauth2/access_token",
        AuthStyle:    rest.AuthStyleInHeader,
    },
    EnableTrace: true,
}
```

### Content Types

#### XML Requests
```go
// examples/xml/main.go
client := &rest.Client{
    BaseURL:     "https://gorest.co.in/public/v2",
    ContentType: rest.XML,
}

var usersResponse struct {
    XMLName string `xml:"objects"`
    List    []struct {
        Name   string `xml:"name"`
        Email  string `xml:"email"`
        Gender string `xml:"gender"`
        Status string `xml:"status"`
        ID     int    `xml:"id"`
    } `xml:"object"`
}

response := client.Get("/users.xml")
err := response.FillUp(&usersResponse)
```

#### Form Data Submission
```go
// examples/form/main.go
client := &rest.Client{
    Name:        "example-client",
    BaseURL:     "https://httpbin.org",
    ContentType: rest.FORM,
}

values := url.Values{}
values.Set("key1", "value1")
values.Set("key2", "value2")

response := client.PostWithContext(ctx, "/post", values)
```

### Advanced Features

#### Gzip Compression
```go
// examples/gzip/main.go
client := &rest.Client{
    Name:           "httpbin-client",
    BaseURL:        "https://httpbin.org",
    ContentType:    rest.JSON,
    EnableGzip:     true,
    DefaultHeaders: http.Header{"Accept-Encoding": []string{"gzip"}},
}

headers := make(http.Header)
headers.Add("Accept-Encoding", "gzip")

response := client.GetWithContext(ctx, "/gzip", headers)
```

#### File Upload and Binary Data
```go
// examples/bytes/main.go
client := &rest.Client{
    Name:    "example-client",
    BaseURL: "https://httpbin.org",
}

// Download large files
apiURL := fmt.Sprintf("/bytes/%d", 1*rest.MB)
response := client.GetWithContext(ctx, apiURL)

// Stream bytes
apiURL = fmt.Sprintf("/stream-bytes/%d", 1*rest.MB)
response = client.GetWithContext(ctx, apiURL)
```

#### Redirect Handling
```go
// examples/redirect/main.go
client := &rest.Client{
    Name:           "example-client",
    ContentType:    rest.JSON,
    FollowRedirect: true, // Enable automatic redirect following
}

response := client.Get("https://tinyurl.com/39da2yt4")
```

### Caching and Performance

#### Response Caching
```go
// examples/json/iskaypet/main.go
client := &rest.Client{
    Name:        "sites-client",
    BaseURL:     "https://api.prod.dp.iskaypet.com",
    ContentType: rest.JSON,
    EnableCache: true,
}

// First request - hits the server
response1 := client.Get("/sites")
log.Infof("Cache-Control: %v", response1.Header.Get("Cache-Control"))

// Second request - served from cache (if valid)
response2 := client.Get("/sites")
log.Infof("Cache-Control: %v", response2.Header.Get("Cache-Control"))
```

### Monitoring and Observability

#### Metrics Collection
```go
// examples/metrics/main.go
import "github.com/prometheus/client_golang/prometheus/promhttp"

func main() {
    http.Handle("/metrics", promhttp.Handler())
    
    client := &rest.Client{
        BaseURL:     "https://httpbin.org",
        ContentType: rest.JSON,
        Name:        "gorest-client",
        EnableCache: true,
    }
    
    // Simulate API requests
    go func() {
        for {
            apiURL := fmt.Sprintf("/cache/%d", random(1, 100))
            response := client.GetWithContext(ctx, apiURL)
            log.Infof("GET %s, Status: %d", apiURL, response.StatusCode)
        }
    }()
    
    server := &http.Server{Addr: ":8081"}
    server.ListenAndServe()
}
```

#### Request Tracing
```go
// examples/trace/main.go
import "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-relic/otel/tracing"

func main() {
    app, err := tracing.New(ctx, tracing.WithAppName("MyExample"))
    defer app.Shutdown(ctx)
    
    client := &rest.Client{
        BaseURL:     "https://httpbin.org",
        ContentType: rest.JSON,
        Name:        "gorest-client",
        EnableTrace: true,
    }
    
    // Create transaction for tracing
    txnCtx, txn := tracing.NewTransaction(ctx, "MyHTTPRequest")
    response := client.GetWithContext(txnCtx, "/cache/123")
    if response.Err != nil {
        txn.NoticeError(response.Err)
    }
    txn.End()
}
```

### Design Patterns

#### Dependency Injection (IoC)
```go
// examples/ioc/main.go
type ISitesClient interface {
    GetSites(ctx context.Context) ([]SiteResponse, error)
}

type SitesClient struct {
    httpClient rest.HTTPClient
}

func NewSitesClient(httpClient rest.HTTPClient) *SitesClient {
    return &SitesClient{httpClient: httpClient}
}

func (r SitesClient) GetSites(ctx context.Context) ([]SiteResponse, error) {
    headers := make(http.Header)
    headers.Set("X-Api-Key", "your-api-key")
    
    response := r.httpClient.GetWithContext(ctx, "/sites", headers)
    if response.Err != nil {
        return nil, response.Err
    }
    
    var sitesResponse []SiteResponse
    err := response.FillUp(&sitesResponse)
    return sitesResponse, err
}

// Usage
httpClient := &rest.Client{
    Name:        "sitesResponse-httpClient",
    BaseURL:     "https://api.prod.dp.iskaypet.com",
    ContentType: rest.JSON,
}

sitesClient := NewSitesClient(httpClient)
sites, err := sitesClient.GetSites(context.Background())
```

### Testing with Mock Server

The library provides a built-in mock server for testing HTTP clients without making real network requests.

#### Basic Mock Setup
```go
// examples/mock/main.go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
    // Start the mock server
    rest.StartMockupServer()
    defer rest.StopMockupServer()

    // Define mock responses
    userMock := &rest.Mock{
        URL:          "https://api.example.com/users",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusOK,
        RespBody:     `[{"id":1,"name":"John Doe","email":"john@example.com"}]`,
        RespHeaders: http.Header{
            "Content-Type": {"application/json"},
            "Cache-Control": {"max-age=3600"},
        },
    }

    // Add mock to the server
    err := rest.AddMockups(userMock)
    if err != nil {
        fmt.Printf("Error adding mock: %v\n", err)
        return
    }

    // Create client and make request
    client := &rest.Client{
        Name:        "test-client",
        BaseURL:     "https://api.example.com",
        ContentType: rest.JSON,
        Timeout:     5 * time.Second,
    }

    response := client.GetWithContext(context.Background(), "/users")
    if response.Err != nil {
        fmt.Printf("Error: %v\n", response.Err)
        return
    }

    fmt.Printf("Status: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", response.String())
    fmt.Printf("Cache-Control: %s\n", response.Header.Get("Cache-Control"))
}
```

#### Advanced Mock with Headers and Timeout
```go
// Mock with request headers validation and timeout
func testWithHeaders() {
    rest.StartMockupServer()
    defer rest.StopMockupServer()

    // Mock that validates request headers
    authMock := &rest.Mock{
        URL:          "https://api.example.com/protected",
        HTTPMethod:   http.MethodGet,
        ReqHeaders: http.Header{
            "Authorization": {"Bearer valid-token"},
            "X-API-Key":     {"test-key"},
        },
        RespHTTPCode: http.StatusOK,
        RespBody:     `{"message":"Access granted"}`,
        Timeout:      100 * time.Millisecond, // Simulate network delay
    }

    rest.AddMockups(authMock)

    client := &rest.Client{
        Name:        "auth-client",
        BaseURL:     "https://api.example.com",
        ContentType: rest.JSON,
    }

    // Add required headers
    headers := make(http.Header)
    headers.Set("Authorization", "Bearer valid-token")
    headers.Set("X-API-Key", "test-key")

    response := client.GetWithContext(context.Background(), "/protected", headers)
    fmt.Printf("Response: %s\n", response.String())
}
```

#### Mock for Different HTTP Methods
```go
func testMultipleMethods() {
    rest.StartMockupServer()
    defer rest.StopMockupServer()

    // GET mock
    getMock := &rest.Mock{
        URL:          "https://api.example.com/users/1",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusOK,
        RespBody:     `{"id":1,"name":"John Doe"}`,
    }

    // POST mock
    postMock := &rest.Mock{
        URL:          "https://api.example.com/users",
        HTTPMethod:   http.MethodPost,
        ReqBody:      `{"name":"Jane Doe","email":"jane@example.com"}`,
        RespHTTPCode: http.StatusCreated,
        RespBody:     `{"id":2,"name":"Jane Doe","email":"jane@example.com"}`,
    }

    // PUT mock
    putMock := &rest.Mock{
        URL:          "https://api.example.com/users/1",
        HTTPMethod:   http.MethodPut,
        ReqBody:      `{"name":"John Updated"}`,
        RespHTTPCode: http.StatusOK,
        RespBody:     `{"id":1,"name":"John Updated"}`,
    }

    // DELETE mock
    deleteMock := &rest.Mock{
        URL:          "https://api.example.com/users/1",
        HTTPMethod:   http.MethodDelete,
        RespHTTPCode: http.StatusNoContent,
        RespBody:     "",
    }

    rest.AddMockups(getMock, postMock, putMock, deleteMock)

    client := &rest.Client{
        Name:        "crud-client",
        BaseURL:     "https://api.example.com",
        ContentType: rest.JSON,
    }

    // Test GET
    response := client.Get("/users/1")
    fmt.Printf("GET Response: %s\n", response.String())

    // Test POST
    userData := map[string]string{
        "name":  "Jane Doe",
        "email": "jane@example.com",
    }
    response = client.Post("/users", userData)
    fmt.Printf("POST Response: %s\n", response.String())

    // Test PUT
    updateData := map[string]string{"name": "John Updated"}
    response = client.Put("/users/1", updateData)
    fmt.Printf("PUT Response: %s\n", response.String())

    // Test DELETE
    response = client.Delete("/users/1")
    fmt.Printf("DELETE Status: %d\n", response.StatusCode)
}
```

#### Mock for Error Scenarios
```go
func testErrorScenarios() {
    rest.StartMockupServer()
    defer rest.StopMockupServer()

    // 404 Not Found mock
    notFoundMock := &rest.Mock{
        URL:          "https://api.example.com/users/999",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusNotFound,
        RespBody:     `{"error":"User not found"}`,
    }

    // 500 Internal Server Error mock
    serverErrorMock := &rest.Mock{
        URL:          "https://api.example.com/error",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusInternalServerError,
        RespBody:     `{"error":"Internal server error"}`,
    }

    // Timeout mock
    timeoutMock := &rest.Mock{
        URL:          "https://api.example.com/slow",
        HTTPMethod:   http.MethodGet,
        RespHTTPCode: http.StatusOK,
        RespBody:     `{"message":"Slow response"}`,
        Timeout:      10 * time.Second, // Very slow response
    }

    rest.AddMockups(notFoundMock, serverErrorMock, timeoutMock)

    client := &rest.Client{
        Name:        "error-test-client",
        BaseURL:     "https://api.example.com",
        ContentType: rest.JSON,
        Timeout:     2 * time.Second, // Client timeout
    }

    // Test 404
    response := client.Get("/users/999")
    fmt.Printf("404 Status: %d, Body: %s\n", response.StatusCode, response.String())

    // Test 500
    response = client.Get("/error")
    fmt.Printf("500 Status: %d, Body: %s\n", response.StatusCode, response.String())

    // Test timeout
    response = client.Get("/slow")
    if response.Err != nil {
        fmt.Printf("Timeout Error: %v\n", response.Err)
    }
}
```

#### Running Mock Examples

```bash
# Run basic mock example
go run examples/mock/main.go

# Run with mock flag for testing
go test -mock ./...

# Programmatically start mock server
rest.StartMockupServer()
defer rest.StopMockupServer()
```

### HTML Content Handling

```go
client := &rest.Client{
    Name:           "html-client",
    EnableGzip:     true,
    EnableCache:    true,
}

response1 := client.Get("https://example.com/page.html")
fmt.Printf("Response cached: %t\n", response1.Cached())

// Second request may be served from cache
response2 := client.Get("https://example.com/page.html")
fmt.Printf("Response cached: %t\n", response2.Cached())
```

### Running Examples

Execute any example directly:

```bash
# Basic JSON example
go run examples/json/basic/main.go

# OAuth2 example
go run examples/json/oauth/main.go

# Mock server example
go run examples/mock/main.go

# Metrics example with Prometheus
ENV=local APP_NAME=example go run examples/metrics/main.go

# Tracing example
go run examples/trace/main.go
```

### Live Example

```bash
ENV=local APP_NAME=example go run gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/examples/metrics@latest
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
