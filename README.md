[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs
- Fork-Join request pattern, for sending many requests concurrently, getting better client performance.
- Response Caching, based on response headers (cache-control, last-modified, etag, expires)
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default JSON.
- Request Body can be `string`, `[]byte`, `struct` & `map`

## Table of contents

* [RESTClient](#rest-client)
* [Metrics](#metrics)

## Rest Client

# Installation


```shell
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@latest
```

# ⚡️ Quickstart

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
    client := &rest.Client{
        BaseURL:        "https://gorest.co.in/public/v2",
        Timeout:        time.Millisecond * 1000,
        ConnectTimeout: time.Millisecond * 5000,
        ContentType:    rest.JSON,
        Name:           "example-client",
        // EnableTrace:    true,
        // CustomPool:     nil,
        // BasicAuth:      nil,
        // Client:         nil,
        // OAuth:          nil,
        // BaseURL:        baseURL,
        // UserAgent:      "",
        // DisableCache:   false,
        // DisableTimeout: false,
        // FollowRedirect: false,
    }

    var users []struct {
        ID     int    `json:"id"`
        Name   string `json:"name"`
        Email  string `json:"email"`
        Gender string `json:"gender"`
        Status string `json:"status"`
    }

    headers := make(http.Header)
    headers.Add("Accept", "application/json")
    headers.Add("Content-Type", "application/json")

    response := client.GetWithContext(context.Background(), "/users", headers)
    if response.Err != nil {
        log.Fatal(response.Err)
    }

    if response.StatusCode != http.StatusOK {
        log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
    }

    // Untyped fill up
    err := response.FillUp(&users)
    if err != nil {
        log.Fatal(err)
    }

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

We do not have a unified dashboard, which can filter by environment, due to this, you have to enter the specific
environment

Dashboard
- [dev](https://monitoring.dev.dp.iskaypet.com/d/6shkc-L4kk/http-clients?orgId=1)

[prometheus]: images/metrics.png
