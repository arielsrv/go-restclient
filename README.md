[![pipeline status](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs
- Fork-Join request pattern, for sending many requests concurrently, getting better client performance.
- Async request pattern.
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default JSON.
- Request Body can be `string`, `[]byte`, `struct` & `map`

## Local environment

You don't need **VPN**, **Vanity Gateway Server** or **SSH** protocol to use internal Iskaypet packages for Go.

**$HOME/.gitconfig**

> [url "https://oauth2:**{$GITLAB_TOKEN}**@gitlab.com"]\
> &emsp;insteadOf = https://gitlab.com

**$HOME/.netrc** (macOs/Unix)

> machine gitlab.com\
> &emsp;login **your_gitlab_account**\
> &emsp;password **your_gitlab_token**

**%USERPROFILE%/_netrc** (Windows)

> machine gitlab.com\
> &emsp;login **your_gitlab_account**\
> &emsp;password **your_gitlab_token**

**GOPRIVATE**
> export GOPRIVATE=gitlab.com/iskaypetcom

## Developer tools

* [Local environment](#environment)
* [Golang Lint](https://golangci-lint.run/)
* [Golang Task](https://taskfile.dev/)
* [Golang Dependencies Update](https://github.com/oligot/go-mod-upgrade)
* [jq](https://stedolan.github.io/jq/)

### For macOs

```shell
$ brew install go-task/tap/go-task
$ brew install golangci-lint
$ go install github.com/oligot/go-mod-upgrade@latest
$ brew install jq
```

## Table of contents

* [RESTClient](#rest-client)
* [Metrics](#metrics)

## Rest Client

# Installation

go.mod

```go
go get gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient@v0.0.4
```

```shell
export GOPRIVATE=gitlab.com/iskaypetcom
```

# ⚡️ Quickstart

```go
package main

import (
	"fmt"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type UserDTO struct {
	ID   int64
	Name string
}

func main() {
	requestBuilder := rest.RequestBuilder{
		Timeout:        time.Millisecond * 3000,
		ConnectTimeout: time.Millisecond * 5000,
		BaseURL:        "https://gorest.co.in/public/v2",
        Name: "example_client",                           // for metrics
	}

	// This won't be blocked.
	requestBuilder.AsyncGet("/users", func(response *rest.Response) {
		if response.StatusCode == http.StatusOK {
			log.Println(response)
		}
	})

	response := requestBuilder.Get("/users")
	if response.StatusCode != http.StatusOK {
		log.Fatal(response.Err.Error())
	}

	var usersDto []UserDTO
	err := response.FillUp(&usersDto)
	if err != nil {
		log.Fatal(err)
	}

	// or typed filled up
	_, err = rest.TypedFillUp[[]UserDTO](response)
	if err != nil {
		log.Fatal(err)
	}

	var futures []*rest.FutureResponse

	requestBuilder.ForkJoin(func(c *rest.Concurrent) {
		for i := 0; i < len(usersDto); i++ {
			futures = append(futures, c.Get(fmt.Sprintf("/users/%d", usersDto[i].ID)))
		}
	})

	log.Println("Wait all ...")
	startTime := time.Now()
	for i := range futures {
		if futures[i].Response().StatusCode == http.StatusOK {
			var userDto UserDTO
			convertionErr := futures[i].Response().FillUp(&userDto)
			if convertionErr != nil {
				log.Fatal(convertionErr)
			}
			log.Println("\t" + userDto.Name)
		}
	}
	elapsedTime := time.Since(startTime)
	log.Printf("Elapsed time: %d", elapsedTime)
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
