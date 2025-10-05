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
	"slices"
	"strconv"
	"strings"
	"time"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-sdk-config/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// HTTP method categorization for internal use.
var (
	// readVerbs contains HTTP methods that are considered "read" operations.
	// These methods are eligible for caching.
	readVerbs = []string{http.MethodGet, http.MethodHead, http.MethodOptions}

	// contentVerbs contains HTTP methods that typically include a request body.
	contentVerbs = []string{http.MethodPost, http.MethodPut, http.MethodPatch}

	// defaultCheckRedirectFunc is the default function used to handle HTTP redirects.
	defaultCheckRedirectFunc func(request *http.Request, via []*http.Request) error
)

// maxAge is a regular expression used to extract the max-age or s-maxage value
// from a Cache-Control header.
var maxAge = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)

// timeFormats is a list of time.Parse formats to use when parsing HTTP date headers.
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
	validURL, err := url.Parse(fmt.Sprintf("%s%s", r.BaseURL, apiURL))
	if err != nil {
		return &Response{
			Err: err,
		}
	}

	apiURL = validURL.String()

	var cacheResponse *Response
	// If Cache enable && operation is read: Cache GET
	if r.EnableCache && slices.Contains(readVerbs, verb) {
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

	// Change URL to point to Mockup server
	var cacheURL string
	apiURL, cacheURL, err = checkMockup(apiURL)
	if err != nil {
		return &Response{
			Err: err,
		}
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
	start := time.Now()
	httpResponse, err := httpClient.Do(request)
	duration := time.Since(start)

	// Metrics
	metrics.Collector.Prometheus().
		RecordExecutionTime("__go_restclient_durations_seconds", duration, metrics.Tags{
			"client_name": r.Name,
		})

	// Deprecated
	metrics.Collector.Prometheus().RecordExecutionTime("services_dashboard_services_timers", duration,
		buildTags(r.Name, "http_connection", "response_time"))

	// Error handling
	if err != nil {
		var netError net.Error
		errorType := "network"
		if errors.As(err, &netError) && netError.Timeout() {
			errorType = "timeout"
		}

		// Metrics
		metrics.Collector.Prometheus().
			IncrementCounter("__go_restclient_requests_error",
				metrics.Tags{
					"client_name": r.Name,
					"error_type":  errorType,
				})

		// Deprecated
		metrics.Collector.Prometheus().
			IncrementCounter("services_dashboard_services_counters_total",
				buildTags(r.Name, "http_connection_error", errorType))

		return &Response{
			Err: err,
		}
	}
	defer func(Body io.ReadCloser) {
		cErr := Body.Close()
		if cErr != nil {
			log.Errorf("error closing response body: %v", cErr)
		}
	}(httpResponse.Body)

	// Metrics
	metrics.Collector.Prometheus().
		IncrementCounter("__go_restclient_requests_total",
			metrics.Tags{
				"client_name": r.Name,
				"status_code": strconv.Itoa(httpResponse.StatusCode),
			})

	// Deprecated
	metrics.Collector.Prometheus().IncrementCounter("services_dashboard_services_counters_total",
		buildTags(r.Name, "http_status", strconv.Itoa(httpResponse.StatusCode)))

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

	// Cache headers
	cacheHeaders := struct {
		TTL          bool
		LastModified bool
		ETag         bool
	}{
		TTL:          setTTL(response),
		LastModified: setLastModified(response),
		ETag:         setETag(response),
	}

	// Must revalidate response if necessary
	response.revalidate = !cacheHeaders.TTL && (cacheHeaders.LastModified || cacheHeaders.ETag)

	// If Cache enable: Cache SENA
	if r.EnableCache && slices.Contains(readVerbs, verb) &&
		(cacheHeaders.TTL || cacheHeaders.LastModified || cacheHeaders.ETag) {
		resourceCache.setNX(cacheURL, response)
	}

	return response
}

// handleGZip checks if GZip compression is enabled for the given request and response.
// Returns true if the response is gzip-encoded and the client is configured to handle it.
func (r *Client) handleGZip(request *http.Request, response *http.Response) bool {
	return (r.EnableGzip ||
		request.Header.Get(AcceptEncodingHeader) == "gzip") && response.Header.Get(ContentEncodingHeader) == "gzip"
}

// setContentReader creates a reader from the given body and content type.
// It marshals the body according to the specified content type and returns an io.Reader.
// If body is nil, it returns http.NoBody.
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
// Returns an io.ReadCloser for reading the response body.
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

// checkMockup checks if the request URL should be redirected to a mockup server.
// If mockup mode is enabled, it replaces the scheme and host of the URL with
// those of the mockup server, while preserving the original URL for caching.
//
// Returns:
//   - The URL to use for the request (may be modified for mockup)
//   - The original URL to use for caching
//   - Any error that occurred during URL parsing
func checkMockup(reqURL string) (string, string, error) {
	cacheURL := reqURL

	if *mockUpEnv {
		rURL, err := url.Parse(reqURL)
		if err != nil {
			return reqURL, cacheURL, err
		}

		rURL.Scheme = mockServerURL.Scheme
		rURL.Host = mockServerURL.Host

		return rURL.String(), cacheURL, nil
	}

	return reqURL, cacheURL, nil
}

// The newHTTPClient sets up the HTTP client for the given request builder.
// It initializes the client only once per Client instance using sync.Once,
// configuring transport, tracing, OAuth, and default headers.
//
// The client is configured with:
//   - Custom transport settings
//   - OpenTelemetry tracing if enabled
//   - OAuth2 client credentials if provided
//   - Default headers
//   - Redirect handling based on FollowRedirect setting
//
// Returns the configured http.Client.
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
// Returns the configured http.RoundTripper to use for HTTP requests.
func (r *Client) setupTransport() http.RoundTripper {
	transportMtxOnce.Do(func() {
		dfltTransport = &http.Transport{
			MaxIdleConnsPerHost:   http.DefaultMaxIdleConnsPerHost,
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           r.getDialContext(),
			ResponseHeaderTimeout: r.getRequestTimeout(),
		}
		defaultCheckRedirectFunc = http.Client{}.CheckRedirect
	})

	// If there's no CustomPool, use the default transport
	if r.CustomPool == nil {
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
	// Default headers
	request.Header.Set(ConnectionHeader, "keep-alive")
	request.Header.Set(CacheControlHeader, "no-cache")

	// If mockup
	if *mockUpEnv {
		request.Header.Set(XOriginalURLHeader, cacheURL)
	}

	// Basic Auth
	if r.BasicAuth != nil && r.OAuth == nil {
		request.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	// User Agent
	request.Header.Set(UserAgentHeader, func() string {
		if r.UserAgent != "" {
			return r.UserAgent
		}

		return "go-restclient/1.0.0 (iskaypet-sre; +https://gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient)"
	}())

	// Encoding
	if marshaler, found := contentMarshalers[r.ContentType]; found {
		request.Header.Set(CanonicalAcceptHeader, marshaler.DefaultHeaders().Get(CanonicalAcceptHeader))
		if slices.Contains(contentVerbs, request.Method) {
			request.Header.Set(CanonicalContentTypeHeader, marshaler.DefaultHeaders().Get(CanonicalContentTypeHeader))
		}
	}

	// Gzip Encoding
	if r.EnableGzip {
		request.Header.Set(AcceptEncodingHeader, "gzip")
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
			t := now.Add(time.Duration(ttl) * time.Second)
			response.ttl = &t
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

// buildTags builds the tags for the Prometheus metrics.
// It creates a map of tags with client name, event type, event subtype,
// application name, environment, and service type.
// These tags are used to categorize and filter metrics in Prometheus.
func buildTags(clientName, eventType, eventSubType string) metrics.Tags {
	return metrics.Tags{
		"client_name":   clientName,
		"event_type":    eventType,
		"event_subtype": eventSubType,
		"application":   env.GetString("APP_NAME", "undefined"),
		"environment":   env.GetString("ENV", "undefined"),
		"service_type":  "http_client",
	}
}
