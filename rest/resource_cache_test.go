package rest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/arielsrv/go-restclient/rest"
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

// TestCacheGetEtagContentChanged verifies that when a server returns 200 (content changed)
// during an ETag revalidation, the cache entry is updated with the new response.
// This is a regression test for the bug where setNX would skip the update because
// the old entry was still present, leaving stale content in the cache indefinitely.
func TestCacheGetEtagContentChanged(t *testing.T) {
	c := &rest.Client{
		BaseURL:        server.URL,
		EnableCache:    true,
		Timeout:        10 * time.Second,
		ConnectTimeout: 10 * time.Second,
	}

	// Request 1: fresh fetch → server returns 200 with ETag "etag-v1".
	r1 := c.Get("/cache/etag/user/changed")
	require.NoError(t, r1.Err)
	require.Equal(t, http.StatusOK, r1.StatusCode)
	assert.False(t, r1.Cached(), "first response should not be cached")
	assert.Equal(t, "etag-v1", r1.Header.Get(rest.ETagHeader))

	// Give ristretto time to asynchronously commit the Set operation before the
	// next Get reads from the cache.
	time.Sleep(10 * time.Millisecond)

	// Request 2: revalidation → server sees If-None-Match: etag-v1 and returns
	// 200 with a new ETag "etag-v2" (content changed).
	// Bug (before fix): setNX would skip the update; cache would still hold "etag-v1".
	// Fix: set() forces the update so the cache now holds "etag-v2".
	r2 := c.Get("/cache/etag/user/changed")
	require.NoError(t, r2.Err)
	require.Equal(t, http.StatusOK, r2.StatusCode)
	assert.Equal(t, "etag-v2", r2.Header.Get(rest.ETagHeader),
		"cache must be updated with the new ETag after content changed")

	// Give ristretto time to commit the force-updated entry before the next Get.
	time.Sleep(10 * time.Millisecond)

	// Request 3: revalidation → server sees If-None-Match: etag-v2 and returns 304.
	// The client must return the cached r2 response (Cached() == true).
	r3 := c.Get("/cache/etag/user/changed")
	require.NoError(t, r3.Err)
	require.Equal(t, http.StatusOK, r3.StatusCode)
	assert.True(t, r3.Cached(), "third response should come from cache after 304")
}
