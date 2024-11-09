[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs
- Fork-Join request pattern, for sending many requests concurrently, getting better client performance.
- Response Caching, based on response headers (cache-control, last-modified, etag, expires)
- Async request pattern.
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default JSON.
- Request Body can be `string`, `[]byte`, `struct` & `map`

## Table of contents

* [RESTClient](#rest-client)
* [Metrics](#metrics)

## Rest Client

# Installation

go.mod

```go
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@latest
```

```shell
export GOPRIVATE=gitlab.com/iskaypetcom
```

# ⚡️ Quickstart

```go
package main

import (
    "net/http"
    "time"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"

    "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
    baseURL := "https://gorest.co.in/public/v2"

    httpClient := &rest.RequestBuilder{
        Timeout:        time.Millisecond * 1000,
        ConnectTimeout: time.Millisecond * 5000,
        BaseURL:        baseURL,
        // OAuth: 		...
        // CustomPool:  ...
    }

    var users []struct {
        ID     int    `json:"id"`
        Name   string `json:"name"`
        Email  string `json:"email"`
        Gender string `json:"gender"`
        Status string `json:"status"`
    }

    response := httpClient.Get("/users")
    if response.Err != nil {
        log.Fatal(response.Err)
    }

    if response.StatusCode != http.StatusOK {
        log.Fatalf("Status: %d, Body: %s", response.StatusCode, response.Body)
    }

    // Typed fill up
    result, err := rest.Unmarshal[[]UserDTO](response)
    if err != nil {
        log.Fatal(err)
    }

    for i := range result {
        log.Infof("User: %v", result[i])
    }

    // Untyped fill up
    err = response.FillUp(&users)
    if err != nil {
        log.Fatal(err)
    }

    for i := range users {
        log.Infof("User: %v", users[i])
    }
}

type UserDTO struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Gender string `json:"gender"`
    Status string `json:"status"`
}
```
## Metrics
![prometheus]

Requisites
* Make sure you have **prometheus collector endpoint** turned on in your application
* **ENV** variable (dev|uat|pro|any)
* **APP_NAME** variable (repository name)

We do not have a unified dashboard, which can filter by environment, due to this, you have to enter the specific
environment

Dashboard
* [dev](https://monitoring.dev.dp.iskaypet.com/d/6shkc-L4kk/http-clients?orgId=1)

[prometheus]: images/metrics.png
