package rest_test

import (
	"fmt"

	"github.com/arielsrv/go-restclient/rest"
)

type ExampleUser struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func ExampleGet() {
	// Simple GET request using the default client
	response := rest.Get("https://jsonplaceholder.typicode.com/users/1")
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
	// Create a custom client
	client := &rest.Client{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Timeout: 5000, // 5 seconds
	}

	// Make a request
	response := client.Get("/users/1")
	if response.IsOk() {
		fmt.Println("Success!")
	}
	// Output:
	// Success!
}

func ExampleResponse_FillUp() {
	response := rest.Get("https://jsonplaceholder.typicode.com/users/1")

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
	response := rest.Get("https://jsonplaceholder.typicode.com/users/1")

	user, err := rest.Deserialize[ExampleUser](response)
	if err != nil {
		// handle error
	}
	fmt.Println(user.Name)
	// Output:
	// Leanne Graham
}
