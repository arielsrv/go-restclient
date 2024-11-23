package rest

import (
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

func (r *Client) newRequest(ctx context.Context, verb string, apiURL string, body any, headers ...http.Header) *Response {
	var cacheURL string
	var cacheResponse *Response

	result := new(Response)
	apiURL = r.BaseURL + apiURL

	// If Cache enable && operation is read: Cache GET
	if !r.DisableCache && slices.Contains(readVerbs, verb) {
		if cacheResponse = resourceCache.get(apiURL); cacheResponse != nil {
			cacheResponse.cacheHit.Store(true)
			if !cacheResponse.revalidate {
				return cacheResponse
			}
		}
	}

	// Prepare reader for the body
	var bodyReader io.Reader
	bodyReader = http.NoBody
	if body != nil {
		media, found := mediaMarshaler[r.ContentType]
		if !found {
			result.Err = fmt.Errorf("marshal fail, unsupported content type: %d", r.ContentType)
			return result
		}

		reader, err := media.Marshal(body)
		if err != nil {
			result.Err = err
			return result
		}

		bodyReader = reader
	}

	// Change URL to point to Mockup server
	apiURL, cacheURL, err := checkMockup(apiURL)
	if err != nil {
		result.Err = err
		return result
	}

	// Enable trace if enabled
	if r.EnableTrace {
		ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	}

	// Create a new HTTP client
	httpClient := r.onceHTTPClient(ctx)

	// Create a new HTTP request
	request, err := http.NewRequestWithContext(ctx, verb, apiURL, bodyReader)
	if err != nil {
		result.Err = err
		return result
	}

	// Set extra parameters
	r.setParams(request, cacheResponse, cacheURL, headers...)

	// Make the request
	start := time.Now()
	response, err := httpClient.Do(request)
	duration := time.Since(start)

	// Metrics
	metrics.Collector.Prometheus().RecordExecutionTime("go_restclient_durations_seconds", duration, metrics.Tags{"client_name": r.Name})

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
			IncrementCounter("go_restclient_request_error",
				metrics.Tags{
					"client_name": r.Name,
					"error_type":  errorType,
				})

		// Deprecated
		metrics.Collector.Prometheus().
			IncrementCounter("services_dashboard_services_counters_total",
				buildTags(r.Name, "http_connection_error", errorType))

		result.Err = err
		return result
	}
	defer response.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		result.Err = err
		return result
	}

	// Metrics
	metrics.Collector.Prometheus().
		IncrementCounter("go_restclient_requests_total",
			metrics.Tags{
				"client_name": r.Name,
				"status_code": strconv.Itoa(response.StatusCode),
			})

	// Deprecated
	metrics.Collector.Prometheus().IncrementCounter("services_dashboard_services_counters_total",
		buildTags(r.Name, "http_status", strconv.Itoa(response.StatusCode)))

	// If we get a 304, return response from cache
	if response.StatusCode == http.StatusNotModified {
		result = cacheResponse
		return result
	}

	result.Response = response
	result.bytes = responseBody
	setProblem(result)

	// Cache headers
	ttl := setTTL(result)
	lastModified := setLastModified(result)
	etag := setETag(result)

	if !ttl && (lastModified || etag) {
		result.revalidate = true
	}

	// If Cache enable: Cache SENA
	if !r.DisableCache && slices.Contains(readVerbs, verb) && (ttl || lastModified || etag) {
		resourceCache.setNX(cacheURL, result)
	}

	return result
}

func setProblem(result *Response) {
	contentType := result.Header.Get("Content-Type")
	if strings.Contains(contentType, "problem") {
		err := result.FillUp(&result.Problem)
		if err != nil {
			return
		}
	}
}

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

		for k, v := range r.DefaultHeaders {
			r.defaultHeaders.Store(k, v)
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
		hostname, err := os.Hostname()
		if err == nil {
			r.Name = hostname
		} else {
			r.Name = "undefined"
		}
	}

	return r.Client
}

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
					if dfltTransport, ok := currentTransport.(*http.Transport); ok {
						dfltTransport.Proxy = http.ProxyURL(proxy)
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

func (r *Client) setParams(request *http.Request, cacheResponse *Response, cacheURL string, headers ...http.Header) {
	// Default headers
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Cache-Control", "no-cache")

	// If mockup
	if *mockUpEnv {
		request.Header.Set("X-Original-Url", cacheURL)
	}

	// Basic Auth
	if r.BasicAuth != nil && r.OAuth == nil {
		request.SetBasicAuth(r.BasicAuth.Username, r.BasicAuth.Password)
	}

	// User Agent
	request.Header.Set("User-Agent", func() string {
		if r.UserAgent != "" {
			return r.UserAgent
		}

		return fmt.Sprintf("%s/%s", r.Name, "go-restclient")
	}())

	// Encoding
	content, found := mediaMarshaler[r.ContentType]
	if found {
		request.Header.Set("Accept", content.DefaultHeaders().Get("Accept"))
		if slices.Contains(contentVerbs, request.Method) {
			request.Header.Set("Content-Type", content.DefaultHeaders().Get("Content-Type"))
		}
	}

	if cacheResponse != nil && cacheResponse.revalidate {
		switch {
		case cacheResponse.etag != "":
			request.Header.Set("If-None-Match", cacheResponse.etag)
		case cacheResponse.lastModified != nil:
			request.Header.Set("If-Modified-Since", cacheResponse.lastModified.Format(time.RFC1123))
		}
	}

	r.defaultHeaders.Range(func(key, value any) bool {
		request.Header[key.(string)] = value.([]string)
		return true
	})

	for _, h := range headers {
		for k, v := range h {
			request.Header[k] = v
		}
	}
}

func setTTL(response *Response) bool {
	// Cache-Control Header
	cacheControl := maxAge.FindStringSubmatch(response.Header.Get("Cache-Control"))

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

	// Expires Header
	expires, err := time.Parse(time.RFC1123, response.Header.Get("Expires"))
	if err != nil {
		return false
	}

	if expires.Sub(now) > 0 {
		response.ttl = &expires
		return true
	}

	return false
}

func setLastModified(resp *Response) bool {
	lastModified, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	if err != nil {
		return false
	}

	resp.lastModified = &lastModified
	return true
}

func setETag(response *Response) bool {
	response.etag = response.Header.Get("Etag")

	return response.etag != ""
}

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
