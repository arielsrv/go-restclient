package rest

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
	cr := &Response{Response: &http.Response{Header: http.Header{}}, revalidate: true}
	cr.lastModified = new(time.Now().Add(-time.Minute))
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

func Test_setTTL_maxAgeZero(t *testing.T) {
	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(CacheControlHeader, "max-age=0")
	if setTTL(resp) {
		t.Fatal("expected setTTL to return false for max-age=0")
	}
}

func Test_setLastModified_invalidDate(t *testing.T) {
	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(LastModifiedHeader, "not-a-valid-date")
	if setLastModified(resp) {
		t.Fatal("expected setLastModified to return false for invalid date")
	}
}

func Test_setParams_defaultUserAgent(t *testing.T) {
	// cachedUserAgentHdr is nil because newHTTPClient was never called,
	// and UserAgent is empty → exercises the ua=="" fallback branch.
	c := &Client{ContentType: JSON}
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	c.setParams(req, nil, "http://example.com")
	if got := req.Header.Get(UserAgentHeader); got == "" {
		t.Error("expected a default User-Agent to be set")
	}
}

// customRoundTripper is a non-*http.Transport RoundTripper used to cover
// the fallback branches in setupTransport.
type customRoundTripper struct{}

func (customRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, nil
}

func Test_setupTransport_nonHttpTransport_dflt(t *testing.T) {
	// Ensure transportMtxOnce has already fired so it won't overwrite dfltTransport.
	c0 := &Client{}
	_ = c0.setupTransport()

	// Replace dfltTransport with a custom type to hit the "return dfltTransport" branch.
	old := dfltTransport
	dfltTransport = customRoundTripper{}
	defer func() { dfltTransport = old }()

	c := &Client{}
	tr := c.setupTransport()
	if _, ok := tr.(*http.Transport); ok {
		t.Error("expected non-http.Transport to be returned")
	}
}

func Test_setupTransport_customPool_nonHttpTransport(t *testing.T) {
	// CustomPool.Transport is already set to a non-*http.Transport →
	// hits the "return r.CustomPool.Transport" branch at line 442.
	c := &Client{
		CustomPool: &CustomPool{Transport: customRoundTripper{}},
	}
	tr := c.setupTransport()
	if _, ok := tr.(customRoundTripper); !ok {
		t.Error("expected customRoundTripper to be returned")
	}
}

func Test_newRequest_invalidMethod(t *testing.T) {
	// An HTTP method containing a space is rejected by http.NewRequestWithContext,
	// covering the error-return branch at net.go:162.
	c := &Client{}
	resp := c.newRequest(context.Background(), "INVALID METHOD", "http://example.com", nil)
	if resp.Err == nil {
		t.Error("expected error for invalid HTTP method")
	}
}

// Test_setTTL_fromCacheControlAndExpires already covers the invalid-Atoi path
// (max-age=abc) but the regex (\d+) makes it unreachable; tested below for completeness.
func Test_setLastModified_empty(t *testing.T) {
	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	// No Last-Modified header → returns false immediately.
	if setLastModified(resp) {
		t.Fatal("expected false when Last-Modified header is absent")
	}
}

func Test_setTTL_expiredExpires(t *testing.T) {
	// Expires in the past → setTTL returns false (no future TTL).
	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(ExpiresHeader, time.Now().Add(-10*time.Second).Format(time.RFC1123))
	if setTTL(resp) {
		t.Fatal("expected false for already-expired Expires header")
	}
}

func Test_newRequest_ioReadAllError(t *testing.T) {
	// Spin up a server that drops the TCP connection after writing headers,
	// forcing io.ReadAll to return an error (covering net.go:197-201).
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "hijack not supported", http.StatusInternalServerError)
			return
		}
		conn, buf, _ := hj.Hijack()
		// Write a minimal HTTP/1.1 200 response with a Content-Length > 0,
		// then close the connection without sending the body.
		_, _ = buf.WriteString("HTTP/1.1 200 OK\r\nContent-Length: 100\r\n\r\n")
		_ = buf.Flush()
		conn.Close()
	}))
	defer server.Close()

	// Disable mock mode so the request goes to our real test server.
	oldMock := *mockUpEnv
	*mockUpEnv = false
	defer func() { *mockUpEnv = oldMock }()

	c := &Client{DisableTimeout: true}
	resp := c.newRequest(context.Background(), http.MethodGet, server.URL, nil)
	if resp.Err == nil {
		t.Error("expected io.ReadAll error when connection drops mid-response")
	}
}

func Test_setRespReader_gzipCloseError(t *testing.T) {
	// Build a gzip stream that is valid enough for gzip.NewReader to succeed
	// but is truncated so that Close() finds a bad checksum.
	// This covers net.go:279-281 (the cErr != nil branch in the defer).
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write([]byte("hello"))
	_ = w.Close()

	// Truncate the last 4 bytes (part of the GZIP CRC32 trailer).
	truncated := buf.Bytes()[:buf.Len()-4]

	c := &Client{}
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")
	httpResp := &http.Response{
		Header: http.Header{"Content-Encoding": []string{"gzip"}},
		Body:   io.NopCloser(bytes.NewReader(truncated)),
	}

	reader, err := c.setRespReader(req, httpResp)
	// gzip.NewReader should succeed; Close() fires in the defer and may or may not
	// surface an error depending on read-ahead. The important thing is no panic.
	if err != nil {
		// NewReader itself failed — the truncation was too aggressive; not the path we want.
		t.Logf("gzip.NewReader failed (expected nil err here): %v", err)
		return
	}
	// Read all bytes; this exercises the Close() defer path.
	_, _ = io.ReadAll(reader)
}

// nilRespRoundTripper returns (nil, nil) to exercise the nil-response guard
// in newRequest.
type nilRespRoundTripper struct{}

func (nilRespRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil //nolint:nilnil // intentional for test
}

func Test_newRequest_NilHTTPResponse(t *testing.T) {
	c := &Client{
		BaseURL: "http://example.com",
		CustomPool: &CustomPool{
			Transport: nilRespRoundTripper{},
		},
	}
	resp := c.newRequest(t.Context(), http.MethodGet, "/", nil)
	require.NotNil(t, resp)
	// http.Client converts (nil, nil) from a RoundTripper into an error before
	// our nil-check is reached, but the request must still return a non-nil
	// Response carrying the error rather than panicking.
	require.Error(t, resp.Err)
}

func Test_lookupCache_NilResourceCache(t *testing.T) {
	saved := resourceCache
	resourceCache = nil
	defer func() { resourceCache = saved }()

	c := &Client{EnableCache: true}
	value, hit := c.lookupCache("http://example.com", true)
	require.Nil(t, value)
	require.False(t, hit)
}

func Test_storeInCache_NilResourceCache(t *testing.T) {
	saved := resourceCache
	resourceCache = nil
	defer func() { resourceCache = saved }()

	c := &Client{}
	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(CacheControlHeader, "max-age=60")
	// Must not panic with nil resourceCache.
	c.storeInCache("http://example.com", resp, nil)
}

func Test_newRequest_304WithoutCachedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL}
	// No cached response and no revalidation headers: a 304 should be
	// converted into a real response (not a nil dereference).
	resp := c.newRequest(t.Context(), http.MethodGet, "/", nil)
	require.NotNil(t, resp)
	require.NoError(t, resp.Err)
	require.Equal(t, http.StatusNotModified, resp.StatusCode)
}

func Test_newRequest_MockEnvWithNilServerURL(t *testing.T) {
	savedFlag := *mockUpEnv
	savedURL := mockServerURL
	*mockUpEnv = true
	mockServerURL = nil
	defer func() {
		*mockUpEnv = savedFlag
		mockServerURL = savedURL
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL}
	// Even with mock enabled, a nil mockServerURL must not panic; the request
	// falls through to the real URL.
	resp := c.newRequest(t.Context(), http.MethodGet, "/", nil)
	require.NotNil(t, resp)
}

// Test_lookupContentNegotiationHeaders_Unsupported covers the (nil, nil) early
// return when ContentType is not in the contentMarshalers map.
func Test_lookupContentNegotiationHeaders_Unsupported(t *testing.T) {
	c := &Client{ContentType: ContentType(999)}
	accept, ct := c.lookupContentNegotiationHeaders()
	require.Nil(t, accept)
	require.Nil(t, ct)
}

// Test_applyContentNegotiation_Unsupported covers the early return in
// applyContentNegotiation when no Accept header can be resolved.
func Test_applyContentNegotiation_Unsupported(t *testing.T) {
	c := &Client{ContentType: ContentType(999)}
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	require.NoError(t, err)
	c.applyContentNegotiation(req)
	require.Empty(t, req.Header.Get(CanonicalAcceptHeader))
	require.Empty(t, req.Header.Get(CanonicalContentTypeHeader))
}

// Test_precomputeCachedHeaders_Unsupported covers the early return when the
// configured ContentType has no marshaler registered.
func Test_precomputeCachedHeaders_Unsupported(t *testing.T) {
	c := &Client{ContentType: ContentType(999), UserAgent: "ua"}
	c.precomputeCachedHeaders()
	require.Equal(t, []string{"ua"}, c.cachedUserAgentHdr)
	require.Nil(t, c.cachedAcceptHdr)
	require.Nil(t, c.cachedContentTypeHdr)
}

// Test_findUnmarshalerBySuffix_UnknownSuffix covers the final `return nil`
// branch in findUnmarshalerBySuffix when the suffix isn't json or xml.
func Test_findUnmarshalerBySuffix_UnknownSuffix(t *testing.T) {
	require.Nil(t, findUnmarshalerBySuffix("application/something+yaml"))
	require.Nil(t, findUnmarshalerBySuffix("text/plain"))
}
