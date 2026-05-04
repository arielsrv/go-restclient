package rest

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var Version = "1.0.0"

// hdrKeepAlive and hdrNoCache are package-level shared slices for static header values.
// They are assigned directly into request.Header (map entry), which is safe because:
//   - Header.Set/Add replaces the slice reference, never mutating the backing array
//   - cap == 1 forces append to create a new backing array if anyone calls Header.Add
var (
	hdrKeepAlive = []string{"keep-alive"}
	hdrNoCache   = []string{"no-cache"}
)

// HTTP method categorization for internal use.
var (

	// defaultCheckRedirectFunc is the default function used to handle HTTP redirects.
	defaultCheckRedirectFunc func(request *http.Request, via []*http.Request) error
)

// maxAge is a regular expression used to extract the max-age or s-maxage value
// from a Cache-Control header.
var maxAge = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)

// timeFormats is a list of [time.Parse] formats to use when parsing HTTP date headers.
var timeFormats = []string{
	time.RFC1123, // "Mon, 02 Jan 2006 15:04:05 GMT"
	time.RFC850,  // "Monday, 02-Jan-06 15:04:05 GMT"
	time.ANSIC,   // "Mon Jan  2 15:04:05 2006"
}

// HTTP header constants used throughout the package.
const (
	// UserAgentHeader is the header name for the User-Agent.
	UserAgentHeader = "User-Agent"

	// ConnectionHeader is the header name for the Connection.
	ConnectionHeader = "Connection"

	// CacheControlHeader is the header name for Cache-Control directives.
	CacheControlHeader = "Cache-Control"

	// XOriginalURLHeader is a custom header used to track the original URL when using a mockup server.
	XOriginalURLHeader = "X-Original-Url"

	// ETagHeader is the header name for the ETag value.
	ETagHeader = "ETag"

	// LastModifiedHeader is the header name for the Last-Modified timestamp.
	LastModifiedHeader = "Last-Modified"

	// ExpiresHeader is the header name for the Expires timestamp.
	ExpiresHeader = "Expires"

	// AcceptEncodingHeader is the header name for the Accept-Encoding value.
	AcceptEncodingHeader = "Accept-Encoding"

	// ContentEncodingHeader is the header name for the Content-Encoding value.
	ContentEncodingHeader = "Content-Encoding"

	// IfModifiedSinceHeader is the header name for the If-Modified-Since timestamp.
	IfModifiedSinceHeader = "If-Modified-Since"

	// IfNoneMatchHeader is the header name for the If-None-Match value.
	IfNoneMatchHeader = "If-None-Match"
)

const (
	GZip = "gzip"
)

// newRequest creates a new HTTP request and returns the response.
// It handles URL validation, caching, content type marshaling, mockup server redirection,
// tracing, metrics collection, and response processing.
//
// Parameters:
//   - ctx: The context for the request, which can be used for cancellation and tracing.
//   - verb: The HTTP method to use (GET, POST, etc.).
//   - apiURL: The URL to request, which will be appended to the client's BaseURL.
//   - body: The request body, which will be marshaled according to the client's ContentType.
//   - headers: Optional additional headers to include in the request.
//
// Returns a Response object containing the response or any error that occurred.
func (r *Client) newRequest(
	ctx context.Context,
	verb string,
	apiURL string,
	body any,
	headers ...http.Header,
) *Response {
	// Fast URL build: simple concatenation avoids a redundant url.Parse + String() round-trip.
	// http.NewRequestWithContext already calls url.Parse internally, so URL validation
	// (and any resulting error) is handled there — no need to do it twice.
	apiURL = r.BaseURL + apiURL

	// Pre-compute read-verb check once; reused for both cache lookup and cache store.
	isReadVerb := verb == http.MethodGet || verb == http.MethodHead || verb == http.MethodOptions

	var cacheResponse *Response
	// If Cache enable && operation is read: Cache GET
	if r.EnableCache && isReadVerb {
		if value, hit := resourceCache.get(apiURL); hit {
			cacheResponse = value
			if cacheResponse != nil {
				cacheResponse.Hit()
				if !cacheResponse.revalidate {
					return cacheResponse
				}
			}
		}
	}

	// Prepare contentReader for the body
	contentReader, err := setContentReader(body, r.ContentType)
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	// Inline checkMockup: on the hot path (mock disabled) this is a single atomic load —
	// no function call, no string copy.
	cacheURL := apiURL
	if *mockUpEnv {
		rURL, mErr := url.Parse(apiURL)
		if mErr != nil {
			return &Response{Err: mErr}
		}
		rURL.Scheme = mockServerURL.Scheme
		rURL.Host = mockServerURL.Host
		apiURL = rURL.String()
	}

	// Enable trace if enabled
	if r.EnableTrace {
		ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	}

	// Create a new HTTP client
	httpClient := r.newHTTPClient(ctx)

	// Create a new HTTP request
	request, err := http.NewRequestWithContext(ctx, verb, apiURL, contentReader)
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	// Set extra parameters
	r.setParams(request, cacheResponse, cacheURL, headers...)

	// Make the request
	httpResponse, err := httpClient.Do(request)
	// Error handling
	if err != nil {
		return &Response{
			Err: err,
		}
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(httpResponse.Body)

	// If we get a 304, return httpResponse from cache
	if httpResponse.StatusCode == http.StatusNotModified {
		return cacheResponse
	}

	respReader, err := r.setRespReader(request, httpResponse)
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	// Read httpResponse
	respBody, err := io.ReadAll(respReader)
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	// Create a new response
	response := &Response{
		Response: httpResponse,
		bytes:    respBody,
	}

	setProblem(response)

	// Only compute cache-validation headers when caching is enabled.
	// setTTL / setLastModified / setETag each perform header lookups and regexp
	// matching; skipping them on the non-cache path saves ~10 calls per request.
	if r.EnableCache && isReadVerb {
		ttl := setTTL(response)
		lastModified := setLastModified(response)
		etag := setETag(response)

		response.revalidate = !ttl && (lastModified || etag)

		if ttl || lastModified || etag {
			// Content changed after revalidation: force-update the existing entry so
			// the new ETag/body replaces the stale one (fixes the setNX-skip bug).
			// First-time insertion uses setNX to avoid races on concurrent requests.
			if cacheResponse != nil {
				resourceCache.set(cacheURL, response)
			} else {
				resourceCache.setNX(cacheURL, response)
			}
		}
	}

	return response
}

// handleGZip checks if GZip compression is enabled for the given request and response.
// Returns true if the response is gzip-encoded and the client is configured to handle it.
func (r *Client) handleGZip(request *http.Request, response *http.Response) bool {
	return (r.EnableGzip ||
		request.Header.Get(AcceptEncodingHeader) == GZip) && response.Header.Get(ContentEncodingHeader) == GZip
}

// setContentReader creates a reader from the given body and content type.
// It marshals the body according to the specified content type and returns an [io.Reader].
// If body is nil, it returns [http.NoBody].
// Returns an error if the content type is not supported or if marshaling fails.
func setContentReader(body any, contentType ContentType) (io.Reader, error) {
	if body != nil {
		mediaContent, found := contentMarshalers[contentType]
		if !found {
			return nil, fmt.Errorf("marshal fail, unsupported content type: %d", contentType)
		}

		reader, err := mediaContent.Marshal(body)
		if err != nil {
			return nil, err
		}

		return reader, nil
	}

	return http.NoBody, nil
}

// setRespReader creates a reader from the given request and response.
// It handles gzip decompression if necessary.
// Returns an [io.ReadCloser] for reading the response body.
func (r *Client) setRespReader(request *http.Request, response *http.Response) (io.ReadCloser, error) {
	if !r.handleGZip(request, response) {
		return response.Body, nil
	}

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		cErr := gzipReader.Close()
		if cErr != nil {
			return
		}
	}(reader)

	return reader, nil
}

// setProblem sets the Problem field in the response if the response content type
// indicates it's a problem response (contains "problem" in the Content-Type).
// It attempts to deserialize the response body into the Problem field.
func setProblem(result *Response) {
	contentType := result.Header.Get(CanonicalContentTypeHeader)
	if strings.Contains(contentType, "problem") {
		err := result.FillUp(&result.Problem)
		if err != nil {
			return
		}
	}
}

// The newHTTPClient sets up the HTTP client for the given request builder.
// It initializes the client only once per [http.Client] instance using [sync.Once],
// configuring transport, tracing, OAuth, and default headers.
//
// The client is configured with:
//   - Custom transport settings
//   - OpenTelemetry tracing if enabled
//   - OAuth2 client credentials if provided
//   - Default headers
//   - Redirect handling based on FollowRedirect setting
//
// Returns the configured [http.Client].
func (r *Client) newHTTPClient(ctx context.Context) *http.Client {
	r.clientMtxOnce.Do(func() {
		r.clientMtx.Lock()
		defer r.clientMtx.Unlock()

		tr := r.setupTransport()
		if r.EnableTrace {
			tr = otelhttp.NewTransport(tr)
		}
		r.Client = &http.Client{Transport: tr}

		// Redirect handling
		if !r.FollowRedirect {
			r.Client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
				return errors.New("avoided redirect attempt")
			}
		} else {
			r.Client.CheckRedirect = defaultCheckRedirectFunc
		}

		// Default name
		if r.Name == "" {
			if hostname, err := os.Hostname(); err == nil {
				r.Name = hostname
			} else {
				r.Name = "undefined"
			}
		}

		for key, value := range r.DefaultHeaders {
			r.defaultHeaders.Store(key, value)
		}

		// Pre-compute User-Agent as a []string once so setParams can assign it
		// directly into the header map without a per-request string concat or
		// []string allocation.
		ua := r.UserAgent
		if ua == "" {
			ua = "go-restclient/" + Version + " (github; +https://github.com/arielsrv/go-restclient)"
		}
		r.cachedUserAgentHdr = []string{ua}

		// Pre-compute Accept and Content-Type header values from the configured
		// ContentType marshaler. DefaultHeaders() creates a new http.Header map
		// (3 allocs) on every call; caching the []string avoids that entirely.
		if marshaler, found := contentMarshalers[r.ContentType]; found {
			hdrs := marshaler.DefaultHeaders()
			if vals := hdrs[CanonicalAcceptHeader]; len(vals) > 0 {
				r.cachedAcceptHdr = []string{vals[0]}
			}
			if vals := hdrs[CanonicalContentTypeHeader]; len(vals) > 0 {
				r.cachedContentTypeHdr = []string{vals[0]}
			}
		}
	})

	if r.OAuth != nil {
		oauthConfig := &clientcredentials.Config{
			ClientID:       r.OAuth.ClientID,
			ClientSecret:   r.OAuth.ClientSecret,
			TokenURL:       r.OAuth.TokenURL,
			AuthStyle:      oauth2.AuthStyle(r.OAuth.AuthStyle),
			Scopes:         r.OAuth.Scopes,
			EndpointParams: r.OAuth.EndpointParams,
		}

		return oauthConfig.Client(context.WithValue(ctx, oauth2.HTTPClient, r.Client))
	}

	return r.Client
}

// setupTransport sets up the HTTP transport for the client.
// It configures connection pooling, timeouts, and proxy settings.
//
// If a CustomPool is provided, it uses that for transport configuration.
// Otherwise, it uses the default transport shared across all clients.
//
// Returns the configured [http.RoundTripper] to use for HTTP requests.
func (r *Client) setupTransport() http.RoundTripper {
	timeout := r.getRequestTimeout()
	// If there's no CustomPool, use the default transport
	if r.CustomPool == nil {
		transportMtxOnce.Do(func() {
			dfltTransport = &http.Transport{
				MaxIdleConnsPerHost: http.DefaultMaxIdleConnsPerHost,
				Proxy:               http.ProxyFromEnvironment,
				DialContext:         r.getDialContext(),
			}
			defaultCheckRedirectFunc = http.Client{}.CheckRedirect
		})

		// Use a local copy of the transport to avoid modifying the shared dfltTransport's
		// ResponseHeaderTimeout, which could cause race conditions or incorrect timeouts
		// for other clients sharing it.
		if transport, ok := dfltTransport.(*http.Transport); ok {
			tCopy := transport.Clone()
			tCopy.ResponseHeaderTimeout = timeout
			return tCopy
		}
		return dfltTransport
	}

	// If the CustomPool already has a transport, update timeouts if it's *http.Transport
	if transport, ok := r.CustomPool.Transport.(*http.Transport); ok {
		transport.DialContext = r.getDialContext()
		transport.ResponseHeaderTimeout = r.getRequestTimeout()
		return transport
	}

	// Create a new custom transport if none is set yet
	if r.CustomPool.Transport == nil {
		transport := &http.Transport{
			MaxIdleConnsPerHost:   r.CustomPool.MaxIdleConnsPerHost,
			DialContext:           r.getDialContext(),
			ResponseHeaderTimeout: r.getRequestTimeout(),
		}

		// If a proxy is defined, parse and set it
		if proxyURL := r.CustomPool.Proxy; proxyURL != "" {
			if parsed, err := url.Parse(proxyURL); err == nil {
				transport.Proxy = http.ProxyURL(parsed)
			}
		}

		r.CustomPool.Transport = transport
		return transport
	}

	// If a non-http.Transport is already set
	return r.CustomPool.Transport
}

// getDialContext returns a context.DialContext function that applies the configured connection timeout.
// This is used by the HTTP transport to establish network connections.
func (r *Client) getDialContext() func(ctx context.Context, network string, address string) (net.Conn, error) {
	return (&net.Dialer{Timeout: r.getConnectionTimeout()}).DialContext
}

// getRequestTimeout returns the configured request timeout duration.
// It considers the DisableTimeout flag and the Timeout setting, falling back to DefaultTimeout if needed.
// Returns:
//   - 0 if timeouts are disabled
//   - r.Timeout if it's greater than 0
//   - DefaultTimeout otherwise
func (r *Client) getRequestTimeout() time.Duration {
	switch {
	case r.DisableTimeout:
		return 0
	case r.Timeout > 0:
		return r.Timeout
	default:
		return DefaultTimeout
	}
}

// getConnectionTimeout returns the configured connection timeout duration.
// It considers the DisableTimeout flag and the ConnectTimeout setting, falling back to DefaultConnectTimeout if needed.
// Returns:
//   - 0 if timeouts are disabled
//   - r.ConnectTimeout if it's greater than 0
//   - DefaultConnectTimeout otherwise
func (r *Client) getConnectionTimeout() time.Duration {
	switch {
	case r.DisableTimeout:
		return 0
	case r.ConnectTimeout > 0:
		return r.ConnectTimeout
	default:
		return DefaultConnectTimeout
	}
}

// asyncNewRequest performs an asynchronous HTTP request and returns a channel that will receive the response.
// This allows for non-blocking HTTP requests where the response can be processed later.
//
// Parameters:
//   - ctx: The context for the request
//   - verb: The HTTP method to use
//   - url: The URL to request
//   - body: The request body
//   - headers: Optional additional headers
//
// Returns a channel that will receive the Response when the request completes.
// The channel is buffered with size 1 and will be closed after the response is sent.
func (r *Client) asyncNewRequest(
	ctx context.Context,
	verb string,
	url string,
	body any,
	headers ...http.Header,
) <-chan *Response {
	rChan := make(chan *Response, 1)
	go func() {
		defer close(rChan)
		rChan <- r.newRequest(ctx, verb, url, body, headers...)
	}()

	return rChan
}

// setParams sets the request parameters and headers.
// It configures various HTTP headers for the request, including:
//   - Default headers (Connection, Cache-Control)
//   - Mockup server headers if enabled
//   - Authentication headers (Basic Auth)
//   - User-Agent
//   - Content negotiation headers (Accept, Content-Type)
//   - Compression headers (Accept-Encoding)
//   - Cache validation headers (If-None-Match, If-Modified-Since)
//   - Client default headers
//   - Custom headers provided as parameters
func (r *Client) setParams(
	request *http.Request,
	cacheResponse *Response,
	cacheURL string,
	paramHeaders ...http.Header,
) {
	// Direct map assignment with pre-allocated shared slices avoids the []string{val}
	// allocation that Header.Set creates on every call.
	request.Header[ConnectionHeader] = hdrKeepAlive
	request.Header[CacheControlHeader] = hdrNoCache

	// If mockup
	if *mockUpEnv {
		request.Header.Set(XOriginalURLHeader, cacheURL)
	}

	// Basic Auth
	if r.BasicAuth != nil && r.OAuth == nil {
		request.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	// User-Agent: use pre-allocated []string from client init — no closure, no concat.
	// Fall back to on-the-fly computation if newHTTPClient hasn't run yet (e.g. in tests).
	userAgentHdr := r.cachedUserAgentHdr
	if userAgentHdr == nil {
		ua := r.UserAgent
		if ua == "" {
			ua = "go-restclient/" + Version + " (github; +https://github.com/arielsrv/go-restclient)"
		}
		userAgentHdr = []string{ua}
	}
	request.Header[UserAgentHeader] = userAgentHdr

	// Accept / Content-Type: use cached []string values from client init.
	// This replaces two DefaultHeaders() calls that each allocated a new http.Header map
	// (1 map + 2 slices = 3 allocs) per request.
	// Fall back to DefaultHeaders() if the client hasn't been initialized yet.
	acceptHdr := r.cachedAcceptHdr
	ctHdr := r.cachedContentTypeHdr
	if acceptHdr == nil {
		if marshaler, found := contentMarshalers[r.ContentType]; found {
			hdrs := marshaler.DefaultHeaders()
			if vals := hdrs[CanonicalAcceptHeader]; len(vals) > 0 {
				acceptHdr = []string{vals[0]}
			}
			if vals := hdrs[CanonicalContentTypeHeader]; len(vals) > 0 {
				ctHdr = []string{vals[0]}
			}
		}
	}
	if acceptHdr != nil {
		request.Header[CanonicalAcceptHeader] = acceptHdr
		switch request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			if ctHdr != nil {
				request.Header[CanonicalContentTypeHeader] = ctHdr
			}
		}
	}

	// Gzip Encoding
	if r.EnableGzip {
		request.Header.Set(AcceptEncodingHeader, GZip)
	}

	if cacheResponse != nil && cacheResponse.revalidate {
		switch {
		case cacheResponse.etag != "":
			request.Header.Set(IfNoneMatchHeader, cacheResponse.etag)
		case cacheResponse.lastModified != nil:
			request.Header.Set(IfModifiedSinceHeader, cacheResponse.lastModified.Format(time.RFC1123))
		}
	}

	r.defaultHeaders.Range(func(key, value any) bool {
		values := value.([]string)
		for _, v := range values {
			request.Header.Add(key.(string), v)
		}
		return true
	})

	for _, headers := range paramHeaders {
		for k, values := range headers {
			for _, v := range values {
				request.Header.Add(k, v)
			}
		}
	}
}

// setTTL sets the TTL (Time To Live) for the response based on cache headers.
// It checks for:
//   - max-age or s-maxage in Cache-Control header
//   - Expires header
//
// Returns true if a TTL was successfully set, false otherwise.
func setTTL(response *Response) bool {
	// Cache-Control Header
	cacheControl := maxAge.FindStringSubmatch(response.Header.Get(CacheControlHeader))

	now := time.Now()
	if len(cacheControl) > 1 {
		ttl, err := strconv.Atoi(cacheControl[1])
		if err != nil {
			return false
		}

		if ttl > 0 {
			response.ttl = new(now.Add(time.Duration(ttl) * time.Second))
			return true
		}

		return false
	}

	for i := range timeFormats {
		format := timeFormats[i]
		if expires, err := time.Parse(format, response.Header.Get(ExpiresHeader)); err == nil && expires.Sub(now) > 0 {
			response.ttl = &expires
			return true
		}
	}

	return false
}

// setLastModified parses and sets the Last-Modified timestamp from the response headers.
// It tries to parse the timestamp using various time formats.
// Returns true if the Last-Modified header was successfully parsed and set, false otherwise.
func setLastModified(response *Response) bool {
	lastModifiedValue := response.Header.Get(LastModifiedHeader)
	if lastModifiedValue == "" {
		return false
	}

	for i := range timeFormats {
		format := timeFormats[i]
		if lastModified, err := time.Parse(format, lastModifiedValue); err == nil {
			response.lastModified = &lastModified
			return true
		}
	}

	return false
}

// setETag extracts and sets the ETag value from the response headers.
// Returns true if an ETag was found, false otherwise.
func setETag(response *Response) bool {
	response.etag = response.Header.Get(ETagHeader)

	return response.etag != ""
}
