package rest

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

// The default transport used by all RequestBuilders
// that haven't set up a CustomPool.
var defaultTransport http.RoundTripper

// Sync once to set default client and transport to default Request Builder.
var dTransportMtxOnce sync.Once

// DefaultTimeout is the default timeout for all clients.
// DefaultConnectTimeout is the time it takes to make a connection
// Type: time.Duration.
var DefaultTimeout = 500 * time.Millisecond

var DefaultConnectTimeout = 1500 * time.Millisecond

// DefaultMaxIdleConnsPerHost is the default maximum idle connections to have
// per Host for all clients, that use *any* RequestBuilder that don't set
// a CustomPool.
var DefaultMaxIdleConnsPerHost = 2

// ContentType represents the Content Type for the Body of HTTP Verbs like
// POST, PUT, and PATCH.
type ContentType int

const (
	// JSON represents a JSON Content Type.
	JSON ContentType = iota

	// XML represents an XML Content Type.
	XML

	// FORM represents an FORM Content Type.
	FORM

	// BYTES represents a plain Content Type.
	BYTES
)

type IRequestBuilder interface {
	Get(url string, headers ...http.Header) *Response
	GetWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	Post(url string, body any, headers ...http.Header) *Response
	PostWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	PutWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	Put(url string, body any, headers ...http.Header) *Response
	Patch(url string, body any, headers ...http.Header) *Response
	PatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	Delete(url string, headers ...http.Header) *Response
	DeleteWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	Head(url string, headers ...http.Header) *Response
	HeadWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	Options(url string, headers ...http.Header) *Response
	OptionsWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	AsyncGet(url string, f func(*Response), headers ...http.Header)
	AsyncGetWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header)
	AsyncPost(url string, body any, f func(*Response), headers ...http.Header)
	AsyncPostWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header)
	AsyncPut(url string, body any, f func(*Response), headers ...http.Header)
	AsyncPutWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header)
	AsyncPatch(url string, body any, f func(*Response), headers ...http.Header)
	AsyncPatchWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header)
	AsyncDelete(url string, f func(*Response), headers ...http.Header)
	AsyncDeleteWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header)
	AsyncHead(url string, f func(*Response), headers ...http.Header)
	AsyncHeadWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header)
	AsyncOptions(url string, f func(*Response), headers ...http.Header)
	AsyncOptionsWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header)
}

// RequestBuilder is the baseline for creating requests
// There's a Default Builder that you may use for simple requests
// RequestBuilder si thread-safe, and you should store it for later re-used.
type RequestBuilder struct {
	// Create a CustomPool if you don't want to share the *transport*, with others
	// RequestBuilder
	CustomPool *CustomPool

	// Set Basic Auth for this RequestBuilder
	BasicAuth *BasicAuth

	// Public for custom fine-tuning
	Client *http.Client

	// OAuth Credentials
	OAuth *clientcredentials.Config

	// Base URL to be used for each Request. The final URL will be BaseURL + URL.
	BaseURL string

	// Set a specific User Agent for this RequestBuilder
	UserAgent string

	// Public for metrics
	Name string

	// Complete request time out.
	Timeout time.Duration

	// Connection timeout, it bounds the time spent obtaining a successful connection
	ConnectTimeout time.Duration

	// ContentType
	ContentType ContentType

	mtx sync.RWMutex

	clientMtxOnce sync.Once

	// Disable 	internal caching of Responses
	DisableCache bool

	// Disable timeout and default timeout = no timeout
	DisableTimeout bool

	// Set the http client to follow a redirect if we get a 3xx response
	FollowRedirect bool

	// Enable tracing
	EnableTrace bool
}

// CustomPool defines a separate internal *transport* and connection pooling.
type CustomPool struct {
	// Public for custom fine-tuning
	Transport http.RoundTripper
	Proxy     string

	MaxIdleConnsPerHost int
}

// BasicAuth gives the possibility to set UserName and Password for a given
// RequestBuilder. Basic Auth is used by some APIs.
type BasicAuth struct {
	UserName string
	Password string
}

// Get issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) Get(url string, headers ...http.Header) *Response {
	return rb.GetWithContext(context.Background(), url, headers...)
}

// GetWithContext issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) GetWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodGet, url, nil, headers...)
}

// Post issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) Post(url string, body any, headers ...http.Header) *Response {
	return rb.PostWithContext(context.Background(), url, body, headers...)
}

// PostWithContext issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) PostWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodPost, url, body, headers...)
}

// Put issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) Put(url string, body any, headers ...http.Header) *Response {
	return rb.PutWithContext(context.Background(), url, body, headers...)
}

// PutWithContext issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) PutWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodPut, url, body, headers...)
}

// Patch issues a PATCH HTTP verb to the specified URL.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) Patch(url string, body any, headers ...http.Header) *Response {
	return rb.PatchWithContext(context.Background(), url, body, headers...)
}

// PatchWithContext issues a PATCH HTTP verb to the specified URL.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (rb *RequestBuilder) PatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodPatch, url, body, headers...)
}

// Delete issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (rb *RequestBuilder) Delete(url string, headers ...http.Header) *Response {
	return rb.DeleteWithContext(context.Background(), url, headers...)
}

// DeleteWithContext issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (rb *RequestBuilder) DeleteWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodDelete, url, nil, headers...)
}

// Head issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) Head(url string, headers ...http.Header) *Response {
	return rb.HeadWithContext(context.Background(), url, headers...)
}

// HeadWithContext issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) HeadWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodHead, url, nil, headers...)
}

// Options issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) Options(url string, headers ...http.Header) *Response {
	return rb.OptionsWithContext(context.Background(), url, headers...)
}

// OptionsWithContext issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (rb *RequestBuilder) OptionsWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return rb.doRequest(ctx, http.MethodOptions, url, nil, headers...)
}

// AsyncGet is the *asynchronous* option for GET.
// The go routine calling AsyncGet(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncGet(url string, f func(*Response), headers ...http.Header) {
	rb.AsyncGetWithContext(context.Background(), url, f, headers...)
}

// AsyncGetWithContext is the *asynchronous* option for GET.
// The go routine calling AsyncGet(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncGetWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.GetWithContext(ctx, url, headers...), f)
}

// AsyncPost is the *asynchronous* option for POST.
// The go routine calling AsyncPost(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPost(url string, body any, f func(*Response), headers ...http.Header) {
	rb.AsyncPostWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPostWithContext is the *asynchronous* option for POST.
// The go routine calling AsyncPost(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPostWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.PostWithContext(ctx, url, body, headers...), f)
}

// AsyncPut is the *asynchronous* option for PUT.
// The go routine calling AsyncPut(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPut(url string, body any, f func(*Response), headers ...http.Header) {
	rb.AsyncPutWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPutWithContext is the *asynchronous* option for PUT.
// The go routine calling AsyncPut(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPutWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.PutWithContext(ctx, url, body, headers...), f)
}

// AsyncPatch is the *asynchronous* option for PATCH.
// The go routine calling AsyncPatch(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPatch(url string, body any, f func(*Response), headers ...http.Header) {
	rb.AsyncPatchWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPatchWithContext is the *asynchronous* option for PATCH.
// The go routine calling AsyncPatch(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncPatchWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.PatchWithContext(ctx, url, body, headers...), f)
}

// AsyncDelete is the *asynchronous* option for DELETE.
// The go routine calling AsyncDelete(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncDelete(url string, f func(*Response), headers ...http.Header) {
	rb.AsyncDeleteWithContext(context.Background(), url, f, headers...)
}

// AsyncDeleteWithContext is the *asynchronous* option for DELETE.
// The go routine calling AsyncDelete(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncDeleteWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.DeleteWithContext(ctx, url, headers...), f)
}

// AsyncHead is the *asynchronous* option for HEAD.
// The go routine calling AsyncHead(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncHead(url string, f func(*Response), headers ...http.Header) {
	rb.AsyncHeadWithContext(context.Background(), url, f, headers...)
}

// AsyncHeadWithContext is the *asynchronous* option for HEAD.
// The go routine calling AsyncHead(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncHeadWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.HeadWithContext(ctx, url, headers...), f)
}

// AsyncOptions is the *asynchronous* option for OPTIONS.
// The go routine calling AsyncOptions(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncOptions(url string, f func(*Response), headers ...http.Header) {
	rb.AsyncOptionsWithContext(context.Background(), url, f, headers...)
}

// AsyncOptionsWithContext is the *asynchronous* option for OPTIONS.
// The go routine calling AsyncOptions(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (rb *RequestBuilder) AsyncOptionsWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(rb.OptionsWithContext(ctx, url, headers...), f)
}

func doAsyncRequest(r *Response, f func(*Response)) {
	f(r)
}
