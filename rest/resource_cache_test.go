package rest_test

import (
	"net/http"
	"testing"
	"time"

	"gitlab.com/arielsrv/go-restclient/rest"
)

func TestCacheGetLowCacheMaxSize(t *testing.T) {
	mcs := rest.MaxCacheSize
	defer func() { rest.MaxCacheSize = mcs }()

	rest.MaxCacheSize = 500

	var f [1000]*rest.Response

	for i := range f {
		f[i] = rb.Get("/cache/user")

		if f[i].StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGet(t *testing.T) {
	c := &rest.Client{BaseURL: server.URL, EnableCache: true}

	for range 1000 {
		r := c.Get("/cache/user")

		if r.Err != nil {
			t.Fatal("Error:", r.Err)
		}

		if r.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGetEtag(t *testing.T) {
	c := &rest.Client{
		BaseURL:        server.URL,
		EnableCache:    true,
		Timeout:        10 * time.Second,
		ConnectTimeout: 10 * time.Second,
	}

	for range 1000 {
		response := c.Get("/cache/etag/user")
		if response.Err != nil {
			t.Fatal(response.Err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatal("Error getting response: ", response.Err)
		}
	}
}

func TestCacheGetLastModified(t *testing.T) {
	c := &rest.Client{
		BaseURL:        server.URL,
		EnableCache:    true,
		Timeout:        10 * time.Second,
		ConnectTimeout: 10 * time.Second,
	}

	for range 1000 {
		response := c.Get("/cache/lastmodified/user")
		if response.Err != nil {
			t.Fatal(response.Err)
		}

		if response.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGetExpires(t *testing.T) {
	c := &rest.Client{
		BaseURL:        server.URL,
		EnableCache:    true,
		Timeout:        10 * time.Second,
		ConnectTimeout: 10 * time.Second,
	}

	for range 1000 {
		response := c.Get("/cache/expires/user")
		if response.Err != nil {
			t.Fatal(response.Err)
		}
		if response.StatusCode != http.StatusOK {
			t.Fatal("Error getting response: ", response.Err)
		}
	}
}

func TestCacheSlowGet(t *testing.T) {
	c := &rest.Client{BaseURL: server.URL, EnableCache: true}

	for range 1000 {
		r := c.Get("/cache/user")

		if r.Err != nil {
			t.Fatal("Error:", r.Err)
		}

		if r.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}

		time.Sleep(3 * time.Millisecond)
	}
}
