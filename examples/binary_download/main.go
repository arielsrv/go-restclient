package main

import (
	"context"
	"fmt"
	"os"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func main() {
	client := &rest.Client{
		Name:    "binary-download-client",
		BaseURL: "https://httpbin.org",
	}

	response := client.GetWithContext(context.Background(), "/bytes/1024")
	if response.Err == nil {
		err := os.WriteFile("output.bin", response.Bytes(), 0o644)
		if err != nil {
			fmt.Println("Error saving file:", err)
		} else {
			fmt.Println("File saved as output.bin")
		}
	} else {
		fmt.Println("Error in download:", response.Err)
	}
}
