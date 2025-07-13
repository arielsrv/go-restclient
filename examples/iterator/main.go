package main

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"os"
	"slices"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

type UserResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
	Status string `json:"status"`
}

type UsersClient struct {
	httpClient rest.HTTPClient
}

func NewUsersClient(httpClient rest.HTTPClient) *UsersClient {
	return &UsersClient{
		httpClient: httpClient,
	}
}

func (r *UsersClient) GetUsers(ctx context.Context) (iter.Seq[UserResponse], error) {
	response := r.httpClient.GetWithContext(ctx, "/public/v2/users")
	if response.Err != nil {
		return nil, response.Err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", response.StatusCode)
	}

	var usersResponse []UserResponse
	err := response.FillUp(&usersResponse)
	if err != nil {
		return nil, err
	}

	return slices.Values(usersResponse), nil
}

type UserDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Gender string `json:"gender"`
	Status string `json:"status"`
}

type UsersService struct {
	usersClient *UsersClient
}

func NewUsersService(usersClient *UsersClient) *UsersService {
	return &UsersService{usersClient: usersClient}
}

func (r *UsersService) GetUsers(ctx context.Context) (iter.Seq[UserDTO], error) {
	usersResponse, err := r.usersClient.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return func(yield func(userDTO UserDTO) bool) {
		for userResponse := range usersResponse {
			yield(UserDTO{
				ID:     userResponse.ID,
				Name:   userResponse.Name,
				Email:  userResponse.Email,
				Gender: userResponse.Gender,
				Status: userResponse.Status,
			})
		}
	}, nil
}

func main() {
	ctx := context.Background()

	usersService := NewUsersService(
		NewUsersClient(&rest.Client{
			Name:        "gorest-co-in",                         // required for logging and tracing
			BaseURL:     "https://gorest.co.in",                 // optional parameters
			ContentType: rest.JSON,                              // rest.JSON by default (rest.XML, rest.FORM, etc.)
			Timeout:     time.Millisecond * time.Duration(2000), // transmission timeout
		}))

	users, err := usersService.GetUsers(ctx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Print the users, lazy loading
	for user := range users {
		fmt.Printf("User: %d, Name: %s, Email: %s\n", user.ID, user.Name, user.Email)
	}
}
