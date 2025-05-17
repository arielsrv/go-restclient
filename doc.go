// Package go_restclient is a Go HTTP client library for making RESTful API requests.
//
// This package provides a simple and flexible way to make HTTP requests to RESTful APIs.
// It supports various HTTP methods (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS),
// different content types (JSON, XML, FORM), and features like caching, authentication,
// compression, and tracing.
//
// Key features:
//   - Simple API for making HTTP requests
//   - Support for synchronous and asynchronous requests
//   - Content type handling (JSON, XML, FORM)
//   - Response caching with TTL, ETag, and Last-Modified support
//   - Basic authentication and OAuth2 support
//   - Customizable connection pooling
//   - Gzip compression
//   - OpenTelemetry tracing
//   - Metrics collection
//   - Mockup server support for testing
//
// The package is organized into several components:
//   - Client: The main client for making HTTP requests
//   - Response: Represents an HTTP response with utility methods
//   - Cache: Caches responses for improved performance
//   - Media: Handles different content types (JSON, XML, FORM)
//
// Example usage:
//
//	// Simple GET request
//	response := rest.Get("https://api.example.com/users")
//	if response.IsOk() {
//	    var users []User
//	    err := response.FillUp(&users)
//	    // Handle users...
//	}
//
//	// Custom client with configuration
//	client := rest.Client{
//	    BaseURL: "https://api.example.com",
//	    Timeout: 5 * time.Second,
//	    ContentType: rest.JSON,
//	    EnableCache: true,
//	}
//	response := client.Get("/users")
//
// For more examples, see the examples directory.
package go_restclient
