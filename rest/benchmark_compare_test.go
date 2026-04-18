package rest_test

// Comparative benchmarks: go-restclient vs resty/v2
//
// All benchmarks run against the in-process httptest.Server (server variable)
// defined in allsetup_test.go, so results are free of external network latency
// and are fully reproducible.
//
// Scenarios
// ─────────────────────────────────────────────────────────────────────────────
//  Plain GET           – single JSON response, no cache
//  Cached GET (TTL)    – Cache-Control: max-age, served from cache after 1st hit
//  Slow GET            – handler sleeps 100 ms (latency-dominated)
//
// Run with:
//   go test ./rest/... -bench=. -benchmem -benchtime=3s -count=1

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/arielsrv/go-restclient/rest"
)

// ── Plain GET ────────────────────────────────────────────────────────────────

func BenchmarkRestClient_Get(b *testing.B) {
	c := &rest.Client{BaseURL: server.URL}
	b.ResetTimer()
	for b.Loop() {
		resp := c.Get("/user")
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

func BenchmarkResty_Get(b *testing.B) {
	c := resty.New().SetBaseURL(server.URL)
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.R().Get("/user")
		if err != nil || resp.StatusCode() != http.StatusOK {
			b.Fatalf("unexpected: err=%v status=%d", err, resp.StatusCode())
		}
	}
}

// ── Cached GET (TTL) ─────────────────────────────────────────────────────────

func BenchmarkRestClient_CachedGet(b *testing.B) {
	c := &rest.Client{BaseURL: server.URL, EnableCache: true}
	b.ResetTimer()
	for b.Loop() {
		resp := c.Get("/cache/user")
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

// resty has no built-in HTTP-RFC cache, so we measure plain repeated GETs —
// the fair comparison for the cache scenario is latency without cache.
func BenchmarkResty_CachedGet(b *testing.B) {
	c := resty.New().SetBaseURL(server.URL)
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.R().Get("/cache/user")
		if err != nil || resp.StatusCode() != http.StatusOK {
			b.Fatalf("unexpected: err=%v status=%d", err, resp.StatusCode())
		}
	}
}

// ── Slow GET (100 ms handler) ────────────────────────────────────────────────

func BenchmarkRestClient_SlowGet(b *testing.B) {
	c := &rest.Client{BaseURL: server.URL}
	b.ResetTimer()
	for b.Loop() {
		resp := c.Get("/slow/user")
		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status: %d", resp.StatusCode)
		}
	}
}

func BenchmarkResty_SlowGet(b *testing.B) {
	c := resty.New().SetBaseURL(server.URL)
	b.ResetTimer()
	for b.Loop() {
		resp, err := c.R().Get("/slow/user")
		if err != nil || resp.StatusCode() != http.StatusOK {
			b.Fatalf("unexpected: err=%v status=%d", err, resp.StatusCode())
		}
	}
}
