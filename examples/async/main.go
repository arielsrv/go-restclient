package main

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	// Create a new context with a timeout of 5 seconds
	// This will automatically cancel the request if it takes longer than 500 milliseconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(500))
	defer cancel()

	// Create a new REST client with custom settings
	client := rest.NewClient(
		rest.WithBaseURL("https://gorest.co.in/public/v2"),
		rest.WithContentType(rest.JSON),
		rest.WithName("gorest-client"),
	)

	// Create a channel to collect the response asynchronously.
	rChan := make(chan *rest.Response, 1)

	// Fire and forget the response, rChan is used to collect the response asynchronously
	client.AsyncGetWithContext(ctx, "/users", func(response *rest.Response) {
		select {
		case <-ctx.Done():
			rChan <- &rest.Response{Err: errors.Wrap(ctx.Err(), "global cancelled")} // Only to show the context cancellation error, don't handle it in a real-world scenario
		default:
			rChan <- response
		}
	})

	// Wait for the response and handle errors
	response := <-rChan
	if response.Err != nil {
		log.Fatal(response.Err)
	}

	// Print the response status code and body to the console.
	log.Infof("Response status: %d, Body: %s", response.StatusCode, response.String())
}
