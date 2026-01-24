package main

import (
	"context"
	"fmt"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:    "problem-client",
		BaseURL: "https://httpbin.org",
	}

	response := client.GetWithContext(context.Background(), "/problem")
	if !response.IsOk() {
		var problem rest.Problem
		if err := response.FillUp(&problem); err == nil {
			fmt.Printf("Problem: %s - %s\n", problem.Title, problem.Detail)
		} else {
			fmt.Println("Error parsing problem details:", err)
		}
	} else {
		fmt.Println("Status:", response.StatusCode)
	}
}
