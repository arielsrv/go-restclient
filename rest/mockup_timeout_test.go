package rest_test

import (
	"net/http"
	"testing"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

// This test exercises the timeout branch in mockup.go where the handler
// intentionally delays sending headers causing the client to time out.
func TestMockup_RequestTimeout(t *testing.T) {
	rest.StartMockupServer()
	t.Cleanup(rest.StopMockupServer)

	url := "http://timeout.test/resource"
	mock := rest.Mock{
		URL:          url,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     "will not be read",
		// Request a small artificial server-side delay. The handler adds DefaultTimeout
		// to this to increase likelihood of client timeout.
		Timeout: 200 * time.Millisecond,
	}
	err := rest.AddMockups(&mock)
	if err != nil {
		t.Fatalf("unexpected error adding mock: %v", err)
	}

	// Build a client with short timeouts so we fail fast
	c := rest.Client{
		Timeout:        150 * time.Millisecond,
		ConnectTimeout: 50 * time.Millisecond,
	}

	start := time.Now()
	resp := c.Get(url)
	if resp.Err == nil {
		t.Fatalf("expected timeout/network error, got none and body=%q", resp.String())
	}
	// Sanity check that it didn't wait too long
	if time.Since(start) > 2*time.Second {
		t.Fatalf("request took too long, likely did not time out as expected")
	}
}
