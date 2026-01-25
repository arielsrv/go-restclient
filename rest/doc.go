// Package rest provides a simple and flexible Go HTTP client for making RESTful API requests.
//
// It supports various HTTP methods, content types (JSON, XML, FORM), and advanced features
// like caching, authentication, compression, and tracing.
//
// # Basic Usage
//
// The package provides package-level functions for quick requests using a default client:
//
//	response := rest.Get("https://api.example.com/users")
//	if response.IsOk() {
//	    var users []User
//	    err := response.FillUp(&users)
//	    // ...
//	}
//
// # Advanced Configuration
//
// For more control, create and configure a [Client]:
//
//	client := &rest.Client{
//	    BaseURL: "https://api.example.com",
//	    Timeout: 5 * time.Second,
//	    ContentType: rest.JSON,
//	    EnableCache: true,
//	}
//	response := client.Post("/users", newUser)
//
// # Features
//
//   - Synchronous and Asynchronous requests.
//   - Built-in support for JSON, XML, and Form data.
//   - Response caching with TTL and ETag support.
//   - Authentication: Basic Auth and OAuth2.
//   - Observability: OpenTelemetry tracing and Prometheus metrics.
//   - Testing: Built-in mockup server for unit testing.
package rest
