package rest

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// The default dfltTransport used by all RequestBuilders
// that haven't set up a CustomPool.
var dfltTransport http.RoundTripper

// Sync once to set default client and dfltTransport to default Request Builder.
var transportOnce sync.Once

var (
	// DefaultTimeout is the default timeout for all clients.
	DefaultTimeout = 500 * time.Millisecond
	// DefaultConnectTimeout is the time it takes to make a connection
	// Type: time.Duration.
	DefaultConnectTimeout = 1500 * time.Millisecond
)

type AsyncHTTPClient interface {
	AsyncGet(url string, headers ...http.Header) <-chan *Response
	AsyncGetWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	AsyncPost(url string, body any, headers ...http.Header) <-chan *Response
	AsyncPostWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	AsyncPutWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	AsyncPut(url string, body any, headers ...http.Header) <-chan *Response
	AsyncPatch(url string, body any, headers ...http.Header) <-chan *Response
	AsyncPatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	AsyncDelete(url string, headers ...http.Header) <-chan *Response
	AsyncDeleteWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	AsyncHead(url string, headers ...http.Header) <-chan *Response
	AsyncHeadWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	AsyncOptions(url string, headers ...http.Header) <-chan *Response
	AsyncOptionsWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
}

type HTTPClient interface {
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
}

type HTTPExporter interface {
	RawClient(ctx context.Context) *http.Client
	Do(*http.Request) (*http.Response, error)
}

// Client  is the baseline for creating requests
// There's a Default Builder that you may use for simple requests
// Client si thread-safe, and you should store it for later re-used.
type Client struct {
	// Create a CustomPool if you don't want to share the *dfltTransport*, with others
	// RESTClient
	CustomPool *CustomPool

	// Set Basic Auth for this RESTClient
	BasicAuth *BasicAuth

	// Public for custom fine-tuning
	Client *http.Client

	// OAuth Credentials
	OAuth *OAuth

	// Default headers for all requests
	DefaultHeaders http.Header

	// Headers to be included in all requests
	defaultHeaders sync.Map

	// Base URL to be used for each Request. The final URL will be BaseURL + URL.
	BaseURL string

	// Set a specific User Agent for this RESTClient
	UserAgent string

	// Public for metrics
	Name string

	// Complete request time out.
	Timeout time.Duration

	// ConnectTimeout, it bounds the time spent obtaining a successful connection
	ConnectTimeout time.Duration

	// ContentType
	ContentType ContentType

	// clientMtxOnce protects the http.Client
	clientMtxOnce sync.Once

	// Enable internal caching of Responses
	EnableCache bool

	// Disable timeout and default timeout = no timeout
	DisableTimeout bool

	// Set the http client to follow a redirect if we get a 3xx response
	FollowRedirect bool

	// Enable gzip compression for incoming and outgoing requests
	EnableGzip bool

	// Enable tracing
	EnableTrace bool
}

type AuthStyle int

const (
	// AuthStyleAutoDetect means to auto-detect which authentication
	// style the provider wants by trying both ways and caching
	// the successful way for the future.
	AuthStyleAutoDetect AuthStyle = iota

	// AuthStyleInParams sends the "client_id" and "client_secret"
	// in the POST body as application/x-www-form-urlencoded parameters.
	AuthStyleInParams

	// AuthStyleInHeader sends the client_id and client_password
	// using HTTP Basic Authorization. This is an optional style
	// described in the OAuth2 RFC 6749 section 2.3.1.
	AuthStyleInHeader
)

type OAuth struct {
	EndpointParams url.Values
	ClientID       string
	ClientSecret   string
	TokenURL       string
	Scopes         []string
	AuthStyle      AuthStyle
}

// CustomPool defines a separate internal *dfltTransport* and connection pooling.
type CustomPool struct {
	// Public for custom fine-tuning
	Transport http.RoundTripper
	Proxy     string

	MaxIdleConnsPerHost int
}

// BasicAuth gives the possibility to set Username and Password for a given
// RESTClient. Basic Auth is used by some APIs.
type BasicAuth struct {
	Username string
	Password string
}

// Get issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) Get(url string, headers ...http.Header) *Response {
	return r.GetWithContext(context.Background(), url, headers...)
}

// GetWithContext issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) GetWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodGet, url, nil, headers...)
}

// Post issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) Post(url string, body any, headers ...http.Header) *Response {
	return r.PostWithContext(context.Background(), url, body, headers...)
}

// PostWithContext issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) PostWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodPost, url, body, headers...)
}

// Put issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) Put(url string, body any, headers ...http.Header) *Response {
	return r.PutWithContext(context.Background(), url, body, headers...)
}

// PutWithContext issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) PutWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodPut, url, body, headers...)
}

// Patch issues a PATCH HTTP verb to the specified URL.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) Patch(url string, body any, headers ...http.Header) *Response {
	return r.PatchWithContext(context.Background(), url, body, headers...)
}

// PatchWithContext issues a PATCH HTTP verb to the specified URL.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) PatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodPatch, url, body, headers...)
}

// Delete issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (r *Client) Delete(url string, headers ...http.Header) *Response {
	return r.DeleteWithContext(context.Background(), url, headers...)
}

// DeleteWithContext issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (r *Client) DeleteWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodDelete, url, nil, headers...)
}

// Head issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) Head(url string, headers ...http.Header) *Response {
	return r.HeadWithContext(context.Background(), url, headers...)
}

// HeadWithContext issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) HeadWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodHead, url, nil, headers...)
}

// Options issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) Options(url string, headers ...http.Header) *Response {
	return r.OptionsWithContext(context.Background(), url, headers...)
}

// OptionsWithContext issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) OptionsWithContext(ctx context.Context, url string, headers ...http.Header) *Response {
	return r.newRequest(ctx, http.MethodOptions, url, nil, headers...)
}

// AsyncGet issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncGet(url string, headers ...http.Header) <-chan *Response {
	return r.AsyncGetWithContext(context.Background(), url, headers...)
}

// AsyncGetWithContext issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncGetWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodGet, url, nil, headers...)
}

// AsyncPost issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPost(url string, body any, headers ...http.Header) <-chan *Response {
	return r.AsyncPostWithContext(context.Background(), url, body, headers...)
}

// AsyncPostWithContext issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPostWithContext(
	ctx context.Context,
	url string,
	body any,
	headers ...http.Header,
) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodPost, url, body, headers...)
}

// AsyncPut issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPut(url string, body any, headers ...http.Header) <-chan *Response {
	return r.AsyncPutWithContext(context.Background(), url, body, headers...)
}

// AsyncPutWithContext issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPutWithContext(
	ctx context.Context,
	url string,
	body any,
	headers ...http.Header,
) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodPut, url, body, headers...)
}

// AsyncPatch issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPatch(url string, body any, headers ...http.Header) <-chan *Response {
	return r.AsyncPatchWithContext(context.Background(), url, body, headers...)
}

// AsyncPatchWithContext issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (r *Client) AsyncPatchWithContext(
	ctx context.Context,
	url string,
	body any,
	headers ...http.Header,
) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodPatch, url, body, headers...)
}

// AsyncDelete issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (r *Client) AsyncDelete(url string, headers ...http.Header) <-chan *Response {
	return r.AsyncDeleteWithContext(context.Background(), url, headers...)
}

// AsyncDeleteWithContext issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (r *Client) AsyncDeleteWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodDelete, url, nil, headers...)
}

// AsyncHead issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncHead(url string, headers ...http.Header) <-chan *Response {
	return r.AsyncHeadWithContext(context.Background(), url, headers...)
}

// AsyncHeadWithContext issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncHeadWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodHead, url, nil, headers...)
}

// AsyncOptions issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncOptions(url string, headers ...http.Header) <-chan *Response {
	return r.AsyncOptionsWithContext(context.Background(), url, headers...)
}

// AsyncOptionsWithContext issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (r *Client) AsyncOptionsWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.asyncNewRequest(ctx, http.MethodOptions, url, nil, headers...)
}

// RawClient returns the underlying http.Client used by the RESTClient.
func (r *Client) RawClient(ctx context.Context) *http.Client {
	return r.onceHTTPClient(ctx)
}

// Do execute a REST request.
func (r *Client) Do(req *http.Request) (*http.Response, error) {
	return r.RawClient(req.Context()).Do(req)
}
