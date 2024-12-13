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

var (
	readVerbs                = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	contentVerbs             = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	defaultCheckRedirectFunc func(request *http.Request, via []*http.Request) error
)

var maxAge = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)

// timeFormats is a list of time.Parse formats to use when parsing HTTP date headers.
var timeFormats = []string{
	time.RFC1123, // "Mon, 02 Jan 2006 15:04:05 GMT"
	time.RFC850,  // "Monday, 02-Jan-06 15:04:05 GMT"
	time.ANSIC,   // "Mon Jan  2 15:04:05 2006"
}

const (
	UserAgentHeader       = "User-Agent"
	ConnectionHeader      = "Connection"
	CacheControlHeader    = "Cache-Control"
	XOriginalURLHeader    = "X-Original-Url"
	ETagHeader            = "ETag"
	LastModifiedHeader    = "Last-Modified"
	ExpiresHeader         = "Expires"
	AcceptEncodingHeader  = "Accept-Encoding"
	ContentEncodingHeader = "Content-Encoding"
	IfModifiedSinceHeader = "If-Modified-Since"
	IfNoneMatchHeader     = "If-None-Match"
)

// newRequest creates a new REST client with default configuration.
func (r *Client) newRequest(ctx context.Context, verb string, apiURL string, body any, headers ...http.Header) *Response {
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
			cacheResponse.Hit()
			if !cacheResponse.revalidate {
				return cacheResponse
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
	httpClient := r.onceHTTPClient(ctx)

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
			IncrementCounter("__go_restclient_request_error",
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
		if cErr := Body.Close(); cErr != nil {
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
	if r.EnableCache && slices.Contains(readVerbs, verb) && (cacheHeaders.TTL || cacheHeaders.LastModified || cacheHeaders.ETag) {
		resourceCache.setNX(cacheURL, response, r.CacheBlockingWrites)
	}

	return response
}

// handleGZip checks if GZip compression is enabled for the given request and response.
func (r *Client) handleGZip(request *http.Request, response *http.Response) bool {
	return (r.EnableGzip ||
		request.Header.Get(AcceptEncodingHeader) == "gzip") && response.Header.Get(ContentEncodingHeader) == "gzip"
}

// setContentReader creates a reader from the given body and content type.
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
func (r *Client) setRespReader(request *http.Request, response *http.Response) (io.ReadCloser, error) {
	if !r.handleGZip(request, response) {
		return response.Body, nil
	}

	reader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		if cErr := gzipReader.Close(); cErr != nil {
			return
		}
	}(reader)

	return reader, nil
}

// setProblems sets the problem field in the response if the response content type is a problem.
func setProblem(result *Response) {
	contentType := result.Header.Get(CanonicalContentTypeHeader)
	if strings.Contains(contentType, "problem") {
		if err := result.FillUp(&result.Problem); err != nil {
			return
		}
	}
}

// checkMockup checks if the request URL should be mocked up.
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

// onceHTTPClient sets up the HTTP client for the given request builder.
func (r *Client) onceHTTPClient(ctx context.Context) *http.Client {
	// This will be executed only once
	// per request builder
	r.clientMtxOnce.Do(func() {
		tr := r.setupTransport()

		r.Client = &http.Client{
			Transport: tr,
		}

		if r.EnableTrace {
			r.Client = &http.Client{
				Transport: otelhttp.NewTransport(r.Client.Transport),
			}
		}

		if r.OAuth != nil {
			oauth := &clientcredentials.Config{
				ClientID:       r.OAuth.ClientID,
				ClientSecret:   r.OAuth.ClientSecret,
				TokenURL:       r.OAuth.TokenURL,
				AuthStyle:      oauth2.AuthStyle(r.OAuth.AuthStyle),
				Scopes:         r.OAuth.Scopes,
				EndpointParams: r.OAuth.EndpointParams,
			}

			r.Client = oauth.Client(context.WithValue(ctx, oauth2.HTTPClient, r.Client))
		}

		for key, value := range r.DefaultHeaders {
			r.defaultHeaders.Store(key, value)
		}
	})

	if !r.FollowRedirect {
		r.Client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return errors.New("avoided redirect attempt")
		}
	} else {
		r.Client.CheckRedirect = defaultCheckRedirectFunc
	}

	if r.Name == "" {
		r.Name = func() string {
			if hostname, err := os.Hostname(); err == nil {
				return hostname
			}
			return "undefined"
		}()
	}

	return r.Client
}

// setupTransport sets up the transport for the client.
func (r *Client) setupTransport() http.RoundTripper {
	transportOnce.Do(func() {
		if dfltTransport == nil {
			dfltTransport = &http.Transport{
				MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
				Proxy:                 http.ProxyFromEnvironment,
				DialContext:           r.getDialContext(),
				ResponseHeaderTimeout: r.getRequestTimeout(),
			}
		}

		defaultCheckRedirectFunc = http.Client{}.CheckRedirect
	})

	currentTransport := dfltTransport

	if customPool := r.CustomPool; customPool != nil {
		if customPool.Transport == nil {
			currentTransport = &http.Transport{
				MaxIdleConnsPerHost:   r.CustomPool.MaxIdleConnsPerHost,
				DialContext:           r.getDialContext(),
				ResponseHeaderTimeout: r.getRequestTimeout(),
			}

			// Set Proxy
			if customPool.Proxy != "" {
				if proxy, err := url.Parse(customPool.Proxy); err == nil {
					if transport, ok := currentTransport.(*http.Transport); ok {
						transport.Proxy = http.ProxyURL(proxy)
					}
				}
			}
			customPool.Transport = currentTransport
		} else {
			customPoolTransport, ok := customPool.Transport.(*http.Transport)
			if !ok {
				// If custom dfltTransport is not http.Transport, timeouts will not be overwritten.
				currentTransport = customPool.Transport
			} else {
				customPoolTransport.DialContext = r.getDialContext()
				customPoolTransport.ResponseHeaderTimeout = r.getRequestTimeout()
				currentTransport = customPoolTransport
			}
		}
	}

	return currentTransport
}

// getDialContext returns a context.DialContext that applies the configured timeouts.
func (r *Client) getDialContext() func(ctx context.Context, network string, address string) (net.Conn, error) {
	return (&net.Dialer{Timeout: r.getConnectionTimeout()}).DialContext
}

// getRequestTimeout returns the configured request timeout.
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

// getConnectionTimeout returns the configured connection timeout.
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

// setParams sets the request parameters and headers.
func (r *Client) setParams(request *http.Request, cacheResponse *Response, cacheURL string, paramHeaders ...http.Header) {
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

		return fmt.Sprintf("%s/%s", r.Name, "go-restclient")
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

// setTTL sets the TTL (Time To Live) for the response.
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

// setLastModified sets the Last-Modified header in the response.
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

// setETag sets the ETag header in the response.
func setETag(response *Response) bool {
	response.etag = response.Header.Get(ETagHeader)

	return response.etag != ""
}

// buildTags builds the tags for the Prometheus metrics.
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
