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

```sh
go get -u github.com/arielsrv/ikp_go-restclient
```

# ⚡️ Quickstart

```go
package main

import (
	"go.dp.iskaypet.com/tools/dev/go-restclient"
	"log"
	"net/http"
	"strconv"
	"time"
)

type UserDTO struct {
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
	response.FillUp(&usersDto)

	// or typed filled up
	_, err := rest.TypedFillUp[[]UserDTO](response)
	if err != nil {
		log.Fatal(err)
	}

	var futures []*rest.FutureResponse

	requestBuilder.ForkJoin(func(c *rest.Concurrent) {
		for i := 0; i < len(usersDto); i++ {
			futures = append(futures, c.Get("/users/"+strconv.Itoa(usersDto[i].ID)))
		}
	})

	log.Println("Wait all ...")
	startTime := time.Now()
	for i := range futures {
		if futures[i].Response().StatusCode == http.StatusOK {
			var userDto UserDTO
			futures[i].Response().FillUp(&userDto)
			log.Println("\t" + userDto.Name)
		}
	}
	elapsedTime := time.Since(startTime)
	log.Printf("Elapsed time: %d", elapsedTime)
}

```
