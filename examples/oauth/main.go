package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

// This example demonstrates OAuth2 Client Credentials flow.
//
// A local httptest.Server acts as the token endpoint so the example runs
// without real credentials or external network access.
//
// In production, replace tokenURL and credentials with your real values:
//
//	OAuth: &rest.OAuth{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    TokenURL:     "https://auth.example.com/oauth/token",
//	    Scopes:       []string{"api:read"},
//	    AuthStyle:    rest.AuthStyleInHeader,
//	},
func main() {
	// ── Token server (simulates the OAuth2 authorization server) ─────────────
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
  "access_token": "mock-bearer-token-abc123",
  "token_type":   "Bearer",
  "expires_in":   3600
}`)
	}))
	defer tokenServer.Close()

	// ── API server (simulates the protected resource server) ──────────────────
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	const protectedURL = "http://api.example.com/products"

	err := rest.AddMockups(&rest.Mock{
		URL:          protectedURL,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `[{"id":1,"name":"Widget"},{"id":2,"name":"Gadget"}]`,
		RespHeaders:  http.Header{"Content-Type": {"application/json"}},
	})
	if err != nil {
		fmt.Printf("mock setup error: %v\n", err)
		os.Exit(1)
	}

	// ── Client configured with OAuth2 ─────────────────────────────────────────
	client := &rest.Client{
		Name:           "oauth-example",
		ContentType:    rest.JSON,
		Timeout:        5 * time.Second,
		ConnectTimeout: 5 * time.Second,
		OAuth: &rest.OAuth{
			ClientID:     "demo-client-id",
			ClientSecret: "demo-client-secret",
			TokenURL:     tokenServer.URL + "/token", // local test server
			Scopes:       []string{"products:read"},
			AuthStyle:    rest.AuthStyleInHeader,
			EndpointParams: url.Values{
				"audience": {"http://api.example.com"},
			},
		},
	}

	resp := client.GetWithContext(context.Background(), protectedURL)
	if resp.Err != nil {
		fmt.Printf("request error: %v\n", resp.Err)
		os.Exit(1)
	}

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Body:   %s\n", resp.String())
}
