package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/arielsrv/go-restclient/rest"
)

// UserResponse represents a user resource.
type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

// IUsersClient defines the interface — depends on rest.HTTPClient, not the concrete type.
// This enables dependency injection and easy mocking in tests.
type IUsersClient interface {
	GetUsers(ctx context.Context) ([]UserResponse, error)
}

// UsersClient is the concrete implementation wired to an HTTPClient.
type UsersClient struct {
	httpClient rest.HTTPClient
}

// NewUsersClient constructs a UsersClient.
// In production: pass &rest.Client{...}.
// In unit tests:  pass a mock that implements rest.HTTPClient.
func NewUsersClient(httpClient rest.HTTPClient) *UsersClient {
	return &UsersClient{httpClient: httpClient}
}

func (c *UsersClient) GetUsers(ctx context.Context) ([]UserResponse, error) {
	resp := c.httpClient.GetWithContext(ctx, "/users")
	if resp.Err != nil {
		return nil, resp.Err
	}
	if !resp.IsOk() {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return rest.Deserialize[[]UserResponse](resp)
}

func main() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	err := rest.AddMockups(&rest.Mock{
		URL:          "http://api.example.com/users",
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody: `[
  {"id":1,"name":"Alice","email":"alice@example.com","status":"active"},
  {"id":2,"name":"Bob",  "email":"bob@example.com",  "status":"inactive"}
]`,
		RespHeaders: http.Header{"Content-Type": {"application/json"}},
	})
	if err != nil {
		fmt.Printf("mock error: %v\n", err)
		os.Exit(1)
	}

	// Wire up: inject a real rest.Client — swap for a mock in tests.
	httpClient := &rest.Client{
		Name:        "users-client",
		BaseURL:     "http://api.example.com",
		ContentType: rest.JSON,
	}

	var client IUsersClient = NewUsersClient(httpClient)

	users, err := client.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Retrieved %d users:\n", len(users))
	for _, u := range users {
		fmt.Printf("  [%d] %s <%s> (%s)\n", u.ID, u.Name, u.Email, u.Status)
	}
}
