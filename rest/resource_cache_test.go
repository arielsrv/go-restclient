package rest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func TestCacheGetLowCacheMaxSize(t *testing.T) {
	mcs := rest.MaxCacheSize
	defer func() { rest.MaxCacheSize = mcs }()

	rest.MaxCacheSize = 500

	var f [1000]*rest.Response

	for i := range f {
		f[i] = rb.Get("/cache/user")

		if f[i].Response.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGet(t *testing.T) {
	c := &rest.Client{BaseURL: server.URL, EnableCache: true}

	for i := range 1000 {
		t.Log(i)
		r := c.Get("/cache/user")

		if r.Err != nil {
			t.Fatal("Error:", r.Err)
		}

		if r.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGet_NotModified(t *testing.T) {
	client := &rest.Client{
		BaseURL: server.URL,
	}

	response := client.Get("/cache/user/not_modified")
	require.NotNil(t, response)
	require.NoError(t, response.Err)
	require.NotNil(t, response.Response)
	assert.Equal(t, http.StatusNotModified, response.Response.StatusCode)

	response = client.Get("/cache/user/not_modified")
	require.NotNil(t, response)
	require.NoError(t, response.Err)
	require.NotNil(t, response.Response)
	assert.Equal(t, http.StatusNotModified, response.Response.StatusCode)
}

func TestCacheGetEtag(t *testing.T) {
	c := &rest.Client{BaseURL: server.URL, EnableCache: true, Timeout: 10 * time.Second, ConnectTimeout: 10 * time.Second}
	response := c.Get("/cache/etag/user")
	if response.Err != nil {
		t.Fatal(response.Err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatal("Error getting response: ", response.Err)
	}

	etag := response.Header.Get("ETag")
	require.NotEmpty(t, etag)

	response = c.Get("/cache/etag/user", http.Header{"If-None-Match": []string{etag}})
	if response.Err != nil {
		t.Fatal(response.Err)
	}

	// Should be Not Modified (304) when the response has not been modified in cURL
	if response.StatusCode != http.StatusOK {
		t.Fatal("Expected Status Not Modified")
	}
}

func TestCacheGetLastModified(t *testing.T) {
	var f [100]*rest.Response

	for i := range f {
		f[i] = rb.Get("/cache/lastmodified/user")

		if f[i].Response.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheGetExpires(t *testing.T) {
	var f [100]*rest.Response

	for i := range f {
		f[i] = rb.Get("/cache/expires/user")

		if f[i].Response.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestCacheSlowGet(t *testing.T) {
	var f [1000]*rest.Response

	for i := range f {
		f[i] = rb.Get("/cache/user")

		if f[i].Response.StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}

		// Wait for so we get cache eviction
		time.Sleep(3 * time.Millisecond)
	}
}
