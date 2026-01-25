package rest_test

import (
	"fmt"
	"net/http"

	"github.com/arielsrv/go-restclient/rest"
)

type ExampleUser struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func ExampleGet() {
	// Start the mockup server for a reproducible example
	rest.StartMockupServer()
	defer rest.StopMockupServer()
	rest.FlushMockups()

	// Configure a mock for the specific URL
	_ = rest.AddMockups(&rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   "GET",
		RespHTTPCode: 200,
		RespBody:     `{"name": "Leanne Graham", "id": 1}`,
		RespHeaders:  http.Header{"Content-Type": []string{"application/json"}},
	})

	// Simple GET request using the default client
	response := rest.Get("https://api.example.com/users/1")
	if !response.IsOk() {
		fmt.Printf("Error: %v\n", response.Err)
		return
	}

	var user ExampleUser
	if err := response.FillUp(&user); err != nil {
		fmt.Printf("Error deserializing: %v\n", err)
		return
	}

	fmt.Printf("User: %s\n", user.Name)
	// Output:
	// User: Leanne Graham
}

func ExampleClient() {
	// Start the mockup server for a reproducible example
	rest.StartMockupServer()
	defer rest.StopMockupServer()
	rest.FlushMockups()

	// Configure a mock for the specific URL
	_ = rest.AddMockups(&rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   "GET",
		RespHTTPCode: 200,
		RespBody:     `{"name": "Leanne Graham", "id": 1}`,
		RespHeaders:  http.Header{"Content-Type": []string{"application/json"}},
	})

	// Create a custom client
	client := &rest.Client{
		BaseURL: "https://api.example.com",
		Timeout: 5000, // 5 seconds
	}

	// Make a request
	response := client.Get("/users/1")
	if response.IsOk() {
		fmt.Println("Success!")
	} else {
		fmt.Printf("Error: %v\n", response.Err)
		if response.Response != nil {
			fmt.Printf("Status: %d\n", response.StatusCode)
		}
	}
	// Output:
	// Success!
}

func ExampleResponse_FillUp() {
	// Start the mockup server for a reproducible example
	rest.StartMockupServer()
	defer rest.StopMockupServer()
	rest.FlushMockups()

	// Configure a mock
	_ = rest.AddMockups(&rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   "GET",
		RespHTTPCode: 200,
		RespBody:     `{"name": "Leanne Graham", "id": 1}`,
		RespHeaders:  http.Header{"Content-Type": []string{"application/json"}},
	})

	response := rest.Get("https://api.example.com/users/1")

	var user ExampleUser
	err := response.FillUp(&user)
	if err != nil {
		// handle error
	}
	fmt.Println(user.Name)
	// Output:
	// Leanne Graham
}

func ExampleDeserialize() {
	// Start the mockup server for a reproducible example
	rest.StartMockupServer()
	defer rest.StopMockupServer()
	rest.FlushMockups()

	// Configure a mock
	_ = rest.AddMockups(&rest.Mock{
		URL:          "https://api.example.com/users/1",
		HTTPMethod:   "GET",
		RespHTTPCode: 200,
		RespBody:     `{"name": "Leanne Graham", "id": 1}`,
		RespHeaders:  http.Header{"Content-Type": []string{"application/json"}},
	})

	response := rest.Get("https://api.example.com/users/1")

	user, err := rest.Deserialize[ExampleUser](response)
	if err != nil {
		// handle error
	}
	fmt.Println(user.Name)
	// Output:
	// Leanne Graham
}
