package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:    "binary-download-client",
		BaseURL: "https://httpbin.org",
		Timeout: time.Second * 10,
	}

	response := client.GetWithContext(context.Background(), "/bytes/1024")
	if response.Err == nil {
		err := os.WriteFile("output.bin", []byte(response.String()), 0o600)
		if err != nil {
			fmt.Println("Error saving file:", err)
		} else {
			fmt.Println("File saved as output.bin")
		}
	} else {
		fmt.Println("Error in download:", response.Err)
	}
}
