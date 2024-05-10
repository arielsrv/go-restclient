package rest

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"

	"github.com/pkg/errors"

	"github.com/goccy/go-json"
	"golang.org/x/oauth2"
)

var readVerbs = [3]string{http.MethodGet, http.MethodHead, http.MethodOptions}
var contentVerbs = [3]string{http.MethodPost, http.MethodPut, http.MethodPatch}
var defaultCheckRedirectFunc func(req *http.Request, via []*http.Request) error

var maxAge = regexp.MustCompile(`(?:max-age|s-maxage)=(\d+)`)

const HTTPDateFormat string = "Mon, 01 Jan 2006 15:04:05 GMT"

func (rb *RequestBuilder) doRequest(verb string, reqURL string, reqBody interface{}) (result *Response) {
	var cacheURL string
	var cacheResp *Response

	result = new(Response)
	reqURL = rb.BaseURL + reqURL

	// If Cache enable && operation is read: Cache GET
	if !rb.DisableCache && matchVerbs(verb, readVerbs) {
		if cacheResp = resourceCache.get(reqURL); cacheResp != nil {
			cacheResp.cacheHit.Store(true)
			if !cacheResp.revalidate {
				return cacheResp
			}
		}
	}

	// Marshal request to JSON or XML
	reader, err := rb.marshalReqBody(reqBody)
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

	ctx := context.Background()

	// Get Client (client + transport)
	client := rb.getClient(ctx)

	request, err := http.NewRequestWithContext(ctx, verb, reqURL, reader)

	if err != nil {
		result.Err = err
		return
	}

	// Set extra parameters
	rb.setParams(request, cacheResp, cacheURL)

	startTime := time.Now()
	// Make the request
	httpResp, err := client.Do(request)
	elapsedTime := time.Since(startTime)

	HTTPCollector.RecordExecutionTime(rb.Name, "http_connection", "response_time", elapsedTime)

	if err != nil {
		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			HTTPCollector.IncrementCounter(rb.Name, "http_connection_error", "timeout")
		}
		HTTPCollector.IncrementCounter(rb.Name, "http_connection_error", "network")
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

	HTTPCollector.IncrementCounter(rb.Name, "http_status", strconv.Itoa(httpResp.StatusCode))

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

	// If Cache enable: Cache SETNX
	if !rb.DisableCache && matchVerbs(verb, readVerbs) && (ttl || lastModified || etag) {
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

func (rb *RequestBuilder) marshalReqBody(body interface{}) (io.Reader, error) {
	if body != nil {
		switch rb.ContentType {
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
func (rb *RequestBuilder) getClient(ctx context.Context) *http.Client {
	// This will be executed only once
	// per request builder
	rb.clientMtxOnce.Do(func() {
		dTransportMtxOnce.Do(func() {
			if defaultTransport == nil {
				defaultTransport = &http.Transport{
					MaxIdleConnsPerHost:   DefaultMaxIdleConnsPerHost,
					Proxy:                 http.ProxyFromEnvironment,
					DialContext:           (&net.Dialer{Timeout: rb.getConnectionTimeout()}).DialContext,
					ResponseHeaderTimeout: rb.getRequestTimeout(),
				}
			}

			defaultCheckRedirectFunc = http.Client{}.CheckRedirect
		})

		tr := defaultTransport

		if cp := rb.CustomPool; cp != nil {
			if cp.Transport == nil {
				tr = &http.Transport{
					MaxIdleConnsPerHost:   rb.CustomPool.MaxIdleConnsPerHost,
					DialContext:           (&net.Dialer{Timeout: rb.getConnectionTimeout()}).DialContext,
					ResponseHeaderTimeout: rb.getRequestTimeout(),
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
					ctr.DialContext = (&net.Dialer{Timeout: rb.getConnectionTimeout()}).DialContext
					ctr.ResponseHeaderTimeout = rb.getRequestTimeout()
					tr = ctr
				} else {
					// If custom transport is not http.Transport, timeouts will not be overwritten.
					tr = cp.Transport
				}
			}
		}

		rb.Client = &http.Client{
			Transport: tr,
		}

		if rb.OAuth != nil {
			log.Debug("Using OAuth2 client")
			ctx = context.WithValue(ctx, oauth2.HTTPClient, rb.Client)
			rb.Client = rb.OAuth.Client(ctx)
		}
	})

	if !rb.FollowRedirect {
		rb.Client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return errors.New("avoided redirect attempt")
		}
	} else {
		rb.Client.CheckRedirect = defaultCheckRedirectFunc
	}

	if rb.Name == "" {
		log.Debugf("No name provided for request builder, using hostname")
		if hostname, found := os.LookupEnv("HOSTNAME"); found {
			rb.Name = hostname
		}

		rb.Name = "unknown"
	}

	return rb.Client
}

func (rb *RequestBuilder) getRequestTimeout() time.Duration {
	switch {
	case rb.DisableTimeout:
		return 0
	case rb.Timeout > 0:
		return rb.Timeout
	default:
		return DefaultTimeout
	}
}

func (rb *RequestBuilder) getConnectionTimeout() time.Duration {
	switch {
	case rb.DisableTimeout:
		return 0
	case rb.ConnectTimeout > 0:
		return rb.ConnectTimeout
	default:
		return DefaultConnectTimeout
	}
}

func (rb *RequestBuilder) setParams(req *http.Request, cacheResp *Response, cacheURL string) {
	// Default headers
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cache-Control", "no-cache")

	// If mockup
	if *mockUpEnv {
		req.Header.Set("X-Original-Url", cacheURL)
	}

	// Basic Auth
	if rb.BasicAuth != nil {
		req.SetBasicAuth(rb.BasicAuth.UserName, rb.BasicAuth.Password)
	}

	// User Agent
	req.Header.Set("User-Agent", func() string {
		if rb.UserAgent != "" {
			return rb.UserAgent
		}
		return "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient"
	}())

	// Encoding
	var cType string

	switch rb.ContentType {
	case JSON:
		cType = "json"
	case XML:
		cType = "xml"
	case FORM:
		cType = "x-www-form-urlencoded"
	}

	if cType != "" {
		req.Header.Set("Accept", "application/"+cType)

		if matchVerbs(req.Method, contentVerbs) {
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

	// @TODO: apineiro Replace by optional params when there is a lot of traffic
	rb.mtx.Lock()
	defer rb.mtx.Unlock()
	// Custom Headers
	if rb.Headers != nil {
		for key, values := range map[string][]string(rb.Headers) {
			if req.Header[key] != nil {
				req.Header.Del(key)
			}
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
}

func matchVerbs(s string, sarray [3]string) bool {
	for i := 0; i < len(sarray); i++ {
		if sarray[i] == s {
			return true
		}
	}

	return false
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
