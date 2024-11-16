package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
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

	"github.com/pkg/errors"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-sdk-config/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
)

var (
	readVerbs                = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	contentVerbs             = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	defaultCheckRedirectFunc func(req *http.Request, via []*http.Request) error
)

var (
	maxAge        = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)
	dfltUserAgent = "go-restclient"
)

func (r *Client) newRequest(ctx context.Context, verb string, url string, body any, headers ...http.Header) *Response {
	var cacheURL string
	var cacheResp *Response

	result := new(Response)
	url = r.BaseURL + url

	// If Cache enable && operation is read: Cache GET
	if !r.DisableCache && slices.Contains(readVerbs, verb) {
		if cacheResp = resourceCache.get(url); cacheResp != nil {
			cacheResp.cacheHit.Store(true)
			if !cacheResp.revalidate {
				return cacheResp
			}
		}
	}

	// Marshal request to JSON or XML
	reader, err := r.marshalBody(body)
	if err != nil {
		result.Err = err
		return result
	}

	// Change URL to point to Mockup server
	url, cacheURL, err = checkMockup(url)
	if err != nil {
		result.Err = err
		return result
	}

	if r.EnableTrace {
		ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))
	}

	httpClient := r.onceHTTPClient(ctx)

	request, err := http.NewRequestWithContext(ctx, verb, url, reader)
	if err != nil {
		result.Err = err
		return result
	}

	// Set extra parameters
	r.setParams(request, cacheResp, cacheURL, headers...)

	// Make the request
	start := time.Now()
	response, err := httpClient.Do(request)

	metrics.Collector.Prometheus().RecordExecutionTime("services_dashboard_services_timers", time.Since(start), metrics.Tags{
		"application":   env.GetString("APP_NAME", "undefined"),
		"environment":   env.GetString("ENV", "undefined"),
		"client_name":   r.Name,
		"event_type":    "http_connection",
		"event_subtype": "response_time",
		"service_type":  "http_client",
	})

	if err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			metrics.Collector.Prometheus().IncrementCounter("services_dashboard_services_counters_total", metrics.Tags{
				"application":   env.GetString("APP_NAME", "undefined"),
				"environment":   env.GetString("ENV", "undefined"),
				"client_name":   r.Name,
				"event_type":    "http_connection_error",
				"event_subtype": "timeout",
				"service_type":  "http_client",
			})
			result.Err = err
			return result
		}

		metrics.Collector.Prometheus().IncrementCounter("services_dashboard_services_counters_total", metrics.Tags{
			"application":   env.GetString("APP_NAME", "undefined"),
			"environment":   env.GetString("ENV", "undefined"),
			"client_name":   r.Name,
			"event_type":    "http_connection_error",
			"event_subtype": "network",
			"service_type":  "http_client",
		})
		result.Err = err
		return result
	}
	defer response.Body.Close()

	// Read response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		result.Err = err
		return result
	}

	metrics.Collector.Prometheus().IncrementCounter("services_dashboard_services_counters_total", metrics.Tags{
		"application":   env.GetString("APP_NAME", "undefined"),
		"environment":   env.GetString("ENV", "undefined"),
		"client_name":   r.Name,
		"event_type":    "http_status",
		"event_subtype": strconv.Itoa(response.StatusCode),
		"service_type":  "http_client",
	})

	// If we get a 304, return response from cache
	if response.StatusCode == http.StatusNotModified {
		result = cacheResp
		return result
	}

	result.Response = response
	result.byteBody = respBody

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

func (r *Client) marshalBody(body any) (io.Reader, error) {
	switch r.ContentType {
	case JSON:
		b, err := json.Marshal(body)
		return bytes.NewBuffer(b), err
	case XML:
		b, err := xml.Marshal(body)
		return bytes.NewBuffer(b), err
	case FORM:
		b, ok := body.(url.Values)
		if !ok {
			return nil, errors.New("body must be of type url.Values or map[string]interface{}")
		}
		return strings.NewReader(b.Encode()), nil
	case BYTES:
		var ok bool
		b, ok := body.([]byte)
		if !ok {
			return nil, errors.New("body must be of type []byte or map[string]interface{}")
		}
		return bytes.NewBuffer(b), nil
	}

	return nil, errors.New("invalid content type")
}

func (r *Client) onceHTTPClient(ctx context.Context) *http.Client {
	// This will be executed only once
	// per request builder
	r.clientMtxOnce.Do(func() {
		dfltTransportOnce.Do(func() {
			if dfltTransport == nil {
				dfltTransport = &http.Transport{
					MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
					Proxy:                 http.ProxyFromEnvironment,
					DialContext:           (&net.Dialer{Timeout: r.getConnectionTimeout()}).DialContext,
					ResponseHeaderTimeout: r.getRequestTimeout(),
				}
			}

			defaultCheckRedirectFunc = http.Client{}.CheckRedirect
		})

		tr := dfltTransport

		if cp := r.CustomPool; cp != nil {
			if cp.Transport == nil {
				tr = &http.Transport{
					MaxIdleConnsPerHost:   r.CustomPool.MaxIdleConnsPerHost,
					DialContext:           (&net.Dialer{Timeout: r.getConnectionTimeout()}).DialContext,
					ResponseHeaderTimeout: r.getRequestTimeout(),
				}

				// Set Proxy
				if cp.Proxy != "" {
					if proxy, err := url.Parse(cp.Proxy); err == nil {
						if transport, ok := tr.(*http.Transport); ok {
							transport.Proxy = http.ProxyURL(proxy)
						}
					}
				}
				cp.Transport = tr
			} else {
				ctr, ok := cp.Transport.(*http.Transport)
				if ok {
					ctr.DialContext = (&net.Dialer{Timeout: r.getConnectionTimeout()}).DialContext
					ctr.ResponseHeaderTimeout = r.getRequestTimeout()
					tr = ctr
				} else {
					// If custom transport is not http.Transport, timeouts will not be overwritten.
					tr = cp.Transport
				}
			}
		}

		r.Client = &http.Client{
			Transport: tr,
		}

		if r.EnableTrace {
			r.Client = &http.Client{
				Transport: otelhttp.NewTransport(r.Client.Transport),
			}
		}

		if r.OAuth != nil {
			r.Client = r.OAuth.Client(context.WithValue(ctx, oauth2.HTTPClient, r.Client))
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
		log.Debugf("No name provided for request builder, using hostname")
		hostname, found := os.LookupEnv("HOSTNAME")
		if found {
			r.Name = hostname
		} else {
			r.Name = "unknown"
		}
	}

	return r.Client
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

func (r *Client) setParams(req *http.Request, cacheResp *Response, cacheURL string, headers ...http.Header) {
	// Default headers
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")

	// If mockup
	if *mockUpEnv {
		req.Header.Set("X-Original-Url", cacheURL)
	}

	// Basic Auth
	if r.BasicAuth != nil {
		req.SetBasicAuth(r.BasicAuth.UserName, r.BasicAuth.Password)
	}

	// User Agent
	req.Header.Set("User-Agent", func() string {
		if r.UserAgent != "" {
			return r.UserAgent
		}
		return fmt.Sprintf("%s/%s", r.Name, dfltUserAgent)
	}())

	// Encoding
	var cType string

	switch r.ContentType {
	case JSON:
		cType = "json"
	case XML:
		cType = "xml"
	case FORM:
		cType = "x-www-form-urlencoded"
	case BYTES:
		cType = "octet-stream"
	}

	if cType != "" {
		req.Header.Set("Accept", "application/"+cType)

		if slices.Contains(contentVerbs, req.Method) {
			req.Header.Set("Content-Type", "application/"+cType)
		}
	}

	if cacheResp != nil && cacheResp.revalidate {
		switch {
		case cacheResp.etag != "":
			req.Header.Set("If-None-Match", cacheResp.etag)
		case cacheResp.lastModified != nil:
			req.Header.Set("If-Modified-Since", cacheResp.lastModified.Format(time.RFC1123))
		}
	}

	// User headers
	if len(headers) > 0 {
		for _, h := range headers {
			for k, v := range h {
				req.Header[k] = v
			}
		}
	}
}

func setTTL(resp *Response) bool {
	now := time.Now()

	// Cache-Control Header
	cacheControl := maxAge.FindStringSubmatch(resp.Header.Get("Cache-Control"))

	if len(cacheControl) > 1 {
		ttl, err := strconv.Atoi(cacheControl[1])
		if err != nil {
			return false
		}

		if ttl > 0 {
			t := now.Add(time.Duration(ttl) * time.Second)
			resp.ttl = &t
			return true
		}

		return false
	}

	// Expires Header
	// Date format from RFC-2616, Section 14.21
	expires, err := time.Parse(time.RFC1123, resp.Header.Get("Expires"))
	if err != nil {
		return false
	}

	if expires.Sub(now) > 0 {
		resp.ttl = &expires
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

func setETag(resp *Response) bool {
	resp.etag = resp.Header.Get("Etag")

	return resp.etag != ""
}
