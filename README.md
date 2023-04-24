[![pipeline status](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/-/commits/main)
[![release](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/-/badges/release.svg)](https://gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/-/releases)

> This package provides a http client adapter with some features

- GET, POST, PUT, PATCH, DELETE, HEAD & OPTIONS HTTP verbs
- Fork-Join request pattern, for sending many requests concurrently, getting better client performance.
- Async request pattern.
- Automatic marshal and unmarshal for JSON and XML Content-Type. Default JSON.
- Request Body can be `string`, `[]byte`, `struct` & `map`

## Developer tools

- [Golang Lint](https://golangci-lint.run/)
- [Golang Task](https://taskfile.dev/)
- [Golang Dependencies Update](https://github.com/oligot/go-mod-upgrade)
- [jq](https://stedolan.github.io/jq/)

### For macOs

```shell
$ brew install go-task/tap/go-task
$ brew install golangci-lint
$ go install github.com/oligot/go-mod-upgrade@latest
$ brew install jq
```

## Table of contents

* [RESTClient](#rest-client)


## Rest Client

# Installation

go.mod

```go
require gitlab.com/iskaypetcom/digital/tools/dev/go-restclient vX.Y.Z
replace gitlab.com/iskaypetcom/digital/tools/dev/go-restclient => gitlab.com/iskaypetcom/digital/tools/dev/go-restclient.git vX.Y.Z
```

```shell
export GONOSUMDB=gitlab.com
```

# ⚡️ Quickstart

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/rest"
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
