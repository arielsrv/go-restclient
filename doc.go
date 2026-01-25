// Package go_restclient is a Go HTTP client library for making RESTful API requests.
//
// This package is a wrapper for the rest subpackage. It is recommended to use the
// github.com/arielsrv/go-restclient/rest package directly for most use cases.
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
// Example usage (using the rest subpackage):
//
//	import "github.com/arielsrv/go-restclient/rest"
//
//	// Simple GET request
//	response := rest.Get("https://api.example.com/users")
//	if response.IsOk() {
//	    var users []User
//	    err := response.FillUp(&users)
//	    // Handle users...
//	}
//
// For more detailed documentation and features, please see the [rest] package.
//
// [rest]: https://pkg.go.dev/github.com/arielsrv/go-restclient/rest
package go_restclient
