# Developer Guide for go-restclient

This document provides essential information for developers working on the `go-restclient` project.

## Build and Configuration

The project uses a standard Go layout and includes a `Taskfile.yml` for common development tasks.

### Requirements
- Go 1.26 or higher.
- [Task](https://taskfile.dev/) (optional, but recommended).

### Common Tasks
- **Build**: `go build ./...` or `task build`
- **Lint**: `task lint` (runs `golangci-lint`, `gofumpt`, and `betteralign`)
- **Tidy Dependencies**: `task download`
- **Generate Mocks**: `task mock` (uses `mockery`)

## Testing

The project uses the standard `testing` package along with `testify/assert` and `testify/require`.

### Running Tests
- **All tests**: `go test ./...` or `task test`
- **With Race Detector**: `task race`
- **Coverage**: `task coverage` (generates `coverage.out` and `coverage.html`)

### Test Architecture
Tests for the `rest` package are located in the `rest` directory and use the `rest_test` package to ensure they only use the public API.
- `rest/allsetup_test.go`: Contains a global `httptest.NewServer` and common setup logic (e.g., `TestMain`, `User` struct).
- `rest/rest_test.go`: Contains the main suite of integration tests.

### Adding New Tests
When adding tests, prefer using the existing test server infrastructure.

#### Example Test
Create a file like `rest/demo_test.go`:

```go
package rest_test

import (
	"net/http"
	"testing"

	"github.com/arielsrv/go-restclient/rest"
	"github.com/stretchr/testify/assert"
)

func TestMyNewFeature(t *testing.T) {
	// 'server' is a global httptest.NewServer defined in allsetup_test.go
	resp := rest.Get(server.URL + "/user")
	
	assert.Nil(t, resp.Err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## Development Guidelines

### Code Style
- Follow modern Go idioms (Go 1.26+).
- Use `any` instead of `interface{}`.
- Use `omitzero` instead of `omitempty` for modern JSON tags where appropriate.
- Use `new(val)` for pointer initialization.
- Use `slices` and `maps` standard library packages for collection manipulations.

### Mocking
The project uses `mockery`. If you change interfaces that are mocked, run `task mock` to update the mocks in the `mocks/` directory.

### Error Handling
- Use `errors.Is` and `errors.As` for error checking.
- Prefer `errors.AsType[T](err)` (Go 1.26+) when checking for specific error types.
