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
var transportMtxOnce sync.Once

var (
	// DefaultTimeout is the default timeout for all clients.
	DefaultTimeout = 500 * time.Millisecond
	// DefaultConnectTimeout is the time it takes to make a connection
	// Type: time.Duration.
	DefaultConnectTimeout = 1500 * time.Millisecond
)

// AsyncHTTPClient defines the interface for making asynchronous HTTP requests.
// It returns a channel that will eventually receive the *Response.
type AsyncHTTPClient interface {
	// AsyncGet issues a GET HTTP verb to the specified URL asynchronously.
	AsyncGet(url string, headers ...http.Header) <-chan *Response
	// AsyncGetWithContext issues a GET HTTP verb with context to the specified URL asynchronously.
	AsyncGetWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	// AsyncPost issues a POST HTTP verb to the specified URL asynchronously.
	AsyncPost(url string, body any, headers ...http.Header) <-chan *Response
	// AsyncPostWithContext issues a POST HTTP verb with context to the specified URL asynchronously.
	AsyncPostWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	// AsyncPutWithContext issues a PUT HTTP verb with context to the specified URL asynchronously.
	AsyncPutWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	// AsyncPut issues a PUT HTTP verb to the specified URL asynchronously.
	AsyncPut(url string, body any, headers ...http.Header) <-chan *Response
	// AsyncPatch issues a PATCH HTTP verb to the specified URL asynchronously.
	AsyncPatch(url string, body any, headers ...http.Header) <-chan *Response
	// AsyncPatchWithContext issues a PATCH HTTP verb with context to the specified URL asynchronously.
	AsyncPatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	// AsyncDelete issues a DELETE HTTP verb to the specified URL asynchronously.
	AsyncDelete(url string, headers ...http.Header) <-chan *Response
	// AsyncDeleteWithContext issues a DELETE HTTP verb with context to the specified URL asynchronously.
	AsyncDeleteWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	// AsyncHead issues a HEAD HTTP verb to the specified URL asynchronously.
	AsyncHead(url string, headers ...http.Header) <-chan *Response
	// AsyncHeadWithContext issues a HEAD HTTP verb with context to the specified URL asynchronously.
	AsyncHeadWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	// AsyncOptions issues an OPTIONS HTTP verb to the specified URL asynchronously.
	AsyncOptions(url string, headers ...http.Header) <-chan *Response
	// AsyncOptionsWithContext issues an OPTIONS HTTP verb with context to the specified URL asynchronously.
	AsyncOptionsWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
}

// HTTPClient defines the interface for making synchronous HTTP requests.
type HTTPClient interface {
	// Get issues a GET HTTP verb to the specified URL.
	Get(url string, headers ...http.Header) *Response
	// GetWithContext issues a GET HTTP verb with context to the specified URL.
	GetWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	// Post issues a POST HTTP verb to the specified URL.
	Post(url string, body any, headers ...http.Header) *Response
	// PostWithContext issues a POST HTTP verb with context to the specified URL.
	PostWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	// PutWithContext issues a PUT HTTP verb with context to the specified URL.
	PutWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	// Put issues a PUT HTTP verb to the specified URL.
	Put(url string, body any, headers ...http.Header) *Response
	// Patch issues a PATCH HTTP verb to the specified URL.
	Patch(url string, body any, headers ...http.Header) *Response
	// PatchWithContext issues a PATCH HTTP verb with context to the specified URL.
	PatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) *Response
	// Delete issues a DELETE HTTP verb to the specified URL.
	Delete(url string, headers ...http.Header) *Response
	// DeleteWithContext issues a DELETE HTTP verb with context to the specified URL.
	DeleteWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	// Head issues a HEAD HTTP verb to the specified URL.
	Head(url string, headers ...http.Header) *Response
	// HeadWithContext issues a HEAD HTTP verb with context to the specified URL.
	HeadWithContext(ctx context.Context, url string, headers ...http.Header) *Response
	// Options issues an OPTIONS HTTP verb to the specified URL.
	Options(url string, headers ...http.Header) *Response
	// OptionsWithContext issues an OPTIONS HTTP verb with context to the specified URL.
	OptionsWithContext(ctx context.Context, url string, headers ...http.Header) *Response
}

// HTTPExporter provides access to the underlying http.Client and the Do method.
type HTTPExporter interface {
	// RawClient returns the underlying http.Client.
	RawClient(ctx context.Context) *http.Client
	// Do executes a standard http.Request.
	Do(*http.Request) (*http.Response, error)
}

// Client is the main structure for making REST requests.
// It is thread-safe and should be reused.
// Use the package-level functions (Get, Post, etc.) for quick requests using a default client,
// or create a new Client instance for custom configuration.
type Client struct {
	// CustomPool defines a separate internal transport and connection pooling.
	// If nil, the default transport is used.
	CustomPool *CustomPool

	// BasicAuth sets the username and password for Basic Authentication.
	BasicAuth *BasicAuth

	// Client is the underlying http.Client. If not provided, one will be created.
	Client *http.Client

	// OAuth credentials for OAuth2 authentication.
	OAuth *OAuth

	// DefaultHeaders are headers included in every request.
	DefaultHeaders http.Header

	// defaultHeaders stores headers to be included in all requests (internal use).
	defaultHeaders sync.Map

	// BaseURL is the prefix for all request URLs. Final URL = BaseURL + path.
	BaseURL string

	// UserAgent is the User-Agent header value for all requests.
	UserAgent string

	// Name is a label for the client, used in metrics.
	Name string

	// Timeout is the maximum time for the entire request/response cycle.
	Timeout time.Duration

	// ConnectTimeout is the maximum time allowed to establish a connection.
	ConnectTimeout time.Duration

	// ContentType specifies the default media type (JSON, XML, Form).
	ContentType ContentType

	// clientMtx protects the http.Client creation.
	clientMtx     sync.Mutex
	clientMtxOnce sync.Once

	// EnableCache enables internal response caching.
	EnableCache bool

	// DisableTimeout disables any timeout for the requests.
	DisableTimeout bool

	// FollowRedirect enables following HTTP redirects (3xx).
	FollowRedirect bool

	// EnableGzip enables Gzip compression for requests and responses.
	EnableGzip bool

	// EnableTrace enables OpenTelemetry tracing.
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
	return r.newHTTPClient(ctx)
}

// Do execute a REST request.
func (r *Client) Do(req *http.Request) (*http.Response, error) {
	return r.RawClient(req.Context()).Do(req)
}
