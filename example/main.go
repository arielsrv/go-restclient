package main

import (
	"fmt"
	"log"
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
		Name:           "example_client",
		BaseURL:        "https://gorest.co.in/public/v2",
		CustomPool: &rest.CustomPool{
			MaxIdleConnsPerHost: 20,
			Transport:           &http.Transport{},
		},
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
