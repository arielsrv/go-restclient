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
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/oauth2"
)

var (
	readVerbs                = []string{http.MethodGet, http.MethodHead, http.MethodOptions}
	contentVerbs             = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
	defaultCheckRedirectFunc func(req *http.Request, via []*http.Request) error
)

var maxAge = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)

const HTTPDateFormat string = "Mon, 01 Jan 2006 15:04:05 GMT"

func (r *Client) newRequest(ctx context.Context, verb string, reqURL string, reqBody any, headers ...http.Header) (result *Response) {
	var cacheURL string
	var cacheResp *Response

	result = new(Response)
	reqURL = r.BaseURL + reqURL

	// If Cache enable && operation is read: Cache GET
	if !r.DisableCache && slices.Contains(readVerbs, verb) {
		if cacheResp = resourceCache.get(reqURL); cacheResp != nil {
			cacheResp.cacheHit.Store(true)
			if !cacheResp.revalidate {
				return cacheResp
			}
		}
	}

	// Marshal request to JSON or XML
	reader, err := r.marshalReqBody(reqBody)
	if err != nil {
		result.Err = err
		return
	}

	// Change URL to point to Mockup server
	reqURL, cacheURL, err = checkMockup(reqURL)
	if err != nil {
		result.Err = err
		return
	}

	parentCtx := ctx
	if r.EnableTrace {
		clientTrace := otelhttptrace.NewClientTrace(ctx)
		parentCtx = httptrace.WithClientTrace(ctx, clientTrace)
	}

	client := r.getClient(parentCtx)

	request, err := http.NewRequestWithContext(parentCtx, verb, reqURL, reader)
	if err != nil {
		result.Err = err
		return
	}

	// Set extra parameters
	r.setParams(request, cacheResp, cacheURL, headers...)

	startTime := time.Now()
	// Make the request
	httpResp, err := client.Do(request)
	elapsedTime := time.Since(startTime)

	HTTPCollector.RecordExecutionTime(r.Name, "http_connection", "response_time", elapsedTime)

	if err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			HTTPCollector.IncrementCounter(r.Name, "http_connection_error", "timeout")
		}
		HTTPCollector.IncrementCounter(r.Name, "http_connection_error", "network")
		result.Err = err
		return
	}

	// Read response
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		result.Err = err
		return
	}

	HTTPCollector.IncrementCounter(r.Name, "http_status", strconv.Itoa(httpResp.StatusCode))

	// If we get a 304, return response from cache
	if httpResp.StatusCode == http.StatusNotModified {
		result = cacheResp
		return
	}

	result.Response = httpResp
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

	return
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

func (r *Client) marshalReqBody(body any) (io.Reader, error) {
	if body != nil {
		switch r.ContentType {
		case JSON:
			b, err := json.Marshal(body)
			return bytes.NewBuffer(b), err
		case XML:
			b, err := xml.Marshal(body)
			return bytes.NewBuffer(b), err
		case FORM:
			return strings.NewReader(body.(url.Values).Encode()), nil
		case BYTES:
			var ok bool
			b, ok := body.([]byte)
			if !ok {
				return nil, fmt.Errorf("bytes: body is %T(%v) not a byte slice", body, body)
			}
			return bytes.NewBuffer(b), nil
		}
	}
	return nil, nil
}

func (r *Client) getClient(ctx context.Context) *http.Client {
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
						tr.(*http.Transport).Proxy = http.ProxyURL(proxy)
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
		return "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient"
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
			req.Header.Set("If-Modified-Since", cacheResp.lastModified.Format(HTTPDateFormat))
		}
	}

	if len(headers) > 0 {
		for _, h := range headers {
			for k, v := range h {
				req.Header[k] = v
			}
		}
	}
}

func setTTL(resp *Response) (set bool) {
	now := time.Now()

	// Cache-Control Header
	cacheControl := maxAge.FindStringSubmatch(resp.Header.Get("Cache-Control"))

	if len(cacheControl) > 1 {
		ttl, err := strconv.Atoi(cacheControl[1])
		if err != nil {
			return
		}

		if ttl > 0 {
			t := now.Add(time.Duration(ttl) * time.Second)
			resp.ttl = &t
			set = true
		}

		return
	}

	// Expires Header
	// Date format from RFC-2616, Section 14.21
	expires, err := time.Parse(HTTPDateFormat, resp.Header.Get("Expires"))
	if err != nil {
		return
	}

	if expires.Sub(now) > 0 {
		resp.ttl = &expires
		set = true
	}

	return
}

func setLastModified(resp *Response) bool {
	lastModified, err := time.Parse(HTTPDateFormat, resp.Header.Get("Last-Modified"))
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
