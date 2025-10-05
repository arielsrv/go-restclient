package rest

import (
	"net/http"
	"testing"
	"time"
)

// Test helpers in net.go that are unexported. We keep tests in package rest
// to increase coverage for branches not covered by higher-level tests.
func Test_setTTL_fromCacheControlAndExpires(t *testing.T) {
	resp := &Response{Response: &http.Response{Header: http.Header{}}}

	// Case 1: Cache-Control max-age
	resp.Header.Set(CacheControlHeader, "public, max-age=2")
	if !setTTL(resp) {
		t.Fatalf("expected TTL to be set from max-age")
	}
	if resp.ttl == nil || time.Until(*resp.ttl) <= 0 {
		t.Fatalf("ttl should be in the future; got %v", resp.ttl)
	}

	// Case 2: Expires header when Cache-Control not present
	resp2 := &Response{Response: &http.Response{Header: http.Header{}}}
	resp2.Header.Set(ExpiresHeader, time.Now().Add(2*time.Second).Format(time.RFC1123))
	if !setTTL(resp2) {
		t.Fatalf("expected TTL to be set from Expires")
	}
	if resp2.ttl == nil || time.Until(*resp2.ttl) <= 0 {
		t.Fatalf("ttl should be in the future; got %v", resp2.ttl)
	}

	// Case 3: Invalid max-age value
	resp3 := &Response{Response: &http.Response{Header: http.Header{}}}
	resp3.Header.Set(CacheControlHeader, "max-age=abc")
	if setTTL(resp3) {
		t.Fatalf("expected TTL NOT to be set for invalid max-age")
	}
}

func Test_setLastModified_and_setETag(t *testing.T) {
	resp := &Response{Response: &http.Response{Header: http.Header{}}}

	// Last-Modified
	lm := time.Now().Add(-time.Hour).Format(time.RFC1123)
	resp.Header.Set(LastModifiedHeader, lm)
	if !setLastModified(resp) {
		t.Fatalf("expected Last-Modified to be parsed")
	}
	if resp.lastModified == nil || resp.lastModified.IsZero() {
		t.Fatalf("expected lastModified to be set; got %v", resp.lastModified)
	}

	// ETag
	resp.Header.Set(ETagHeader, "\"abc123\"")
	if !setETag(resp) {
		t.Fatalf("expected ETag to be set")
	}
	if resp.etag == "" {
		t.Fatalf("expected etag to be set")
	}
}

func Test_setParams_headersAndOptions(t *testing.T) {
	c := &Client{
		ContentType: JSON,
		EnableGzip:  true,
		UserAgent:   "my-agent",
		BasicAuth:   &BasicAuth{Username: "u", Password: "p"},
	}

	req, err := http.NewRequest(http.MethodPost, "http://example.com", http.NoBody)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	// Simulate mock mode to assert X-Original-Url header is set
	oldMock := *mockUpEnv
	*mockUpEnv = true
	defer func() { *mockUpEnv = oldMock }()

	// Provide a cache response with validators to exercise If-* headers
	lm := time.Now().Add(-time.Minute)
	cr := &Response{Response: &http.Response{Header: http.Header{}}, revalidate: true}
	cr.lastModified = &lm
	cr.etag = "\"tag\""

	c.setParams(req, cr, "http://origin.example.com/resource?a=1")

	// Default headers
	if req.Header.Get(ConnectionHeader) != "keep-alive" {
		t.Errorf("Connection header not set")
	}
	if req.Header.Get(CacheControlHeader) != "no-cache" {
		t.Errorf("Cache-Control header not set")
	}

	// Mock original url header
	if req.Header.Get(XOriginalURLHeader) == "" {
		t.Errorf("X-Original-Url header expected when mock mode is active")
	}

	// Basic auth should be set when OAuth is nil
	u, p, ok := req.BasicAuth()
	if !ok || u != "u" || p != "p" {
		t.Errorf("basic auth not set correctly")
	}

	// User-Agent
	if got := req.Header.Get(UserAgentHeader); got != "my-agent" {
		t.Errorf("unexpected UA: %s", got)
	}

	// Content negotiation
	if req.Header.Get(CanonicalAcceptHeader) == "" {
		t.Errorf("Accept header should be set by content marshaler")
	}
	if req.Header.Get(CanonicalContentTypeHeader) == "" {
		t.Errorf("Content-Type header should be set for content verbs")
	}

	// Gzip
	if req.Header.Get(AcceptEncodingHeader) != "gzip" {
		t.Errorf("Accept-Encoding should be gzip when enabled")
	}

	// Validators from cacheResponse when revalidate=true prefer If-None-Match first
	if req.Header.Get(IfNoneMatchHeader) == "" && req.Header.Get(IfModifiedSinceHeader) == "" {
		t.Errorf("expected some revalidation header to be set")
	}
}
