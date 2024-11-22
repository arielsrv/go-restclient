package rest

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

// The default transport used by all RequestBuilders
// that haven't set up a CustomPool.
var dfltTransport http.RoundTripper

// Sync once to set default client and transport to default Request Builder.
var dfltTransportOnce sync.Once

var (
	// DefaultTimeout is the default timeout for all clients.
	DefaultTimeout = 500 * time.Millisecond
	// DefaultConnectTimeout is the time it takes to make a connection
	// Type: time.Duration.
	DefaultConnectTimeout = 1500 * time.Millisecond
	// DefaultMaxIdleConnsPerHost is the default maximum idle connections to have
	// per Host for all clients, that use *any* RESTClient that don't set
	// a CustomPool.
	DefaultMaxIdleConnsPerHost = 2
)

type AsyncHTTPClient interface {
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

type AsyncChanHTTPClient interface {
	ChanGet(url string, headers ...http.Header) <-chan *Response
	ChanGetWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	ChanPost(url string, body any, headers ...http.Header) <-chan *Response
	ChanPostWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	ChanPut(url string, body any, headers ...http.Header) <-chan *Response
	ChanPutWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	ChanPatch(url string, body any, headers ...http.Header) <-chan *Response
	ChanPatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response
	ChanDelete(url string, headers ...http.Header) <-chan *Response
	ChanDeleteWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	ChanHead(url string, headers ...http.Header) <-chan *Response
	ChanHeadWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
	ChanOptions(url string, headers ...http.Header) <-chan *Response
	ChanOptionsWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response
}

type HTTPClient interface {
	AsyncHTTPClient
	AsyncChanHTTPClient

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

// Client  is the baseline for creating requests
// There's a Default Builder that you may use for simple requests
// Client si thread-safe, and you should store it for later re-used.
type Client struct {
	// Create a CustomPool if you don't want to share the *transport*, with others
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

	// Connection timeout, it bounds the time spent obtaining a successful connection
	ConnectTimeout time.Duration

	// ContentType
	ContentType ContentType

	// Trace logs HTTP requests and responses
	rwMtx sync.RWMutex

	// clientMtx protects the clientMtxOnce
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

type Option func(*Client)

// NewClient creates a new Client with the given options.
func NewClient(opts ...Option) *Client {
	client := &Client{
		Timeout:        30 * time.Second, // default timeout
		ConnectTimeout: 10 * time.Second, // default connection timeout
		ContentType:    JSON,
		FollowRedirect: true,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// WithBaseURL sets the BaseURL of the Client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithName sets the name of the Client.
func WithName(name string) Option {
	return func(c *Client) {
		c.Name = name
	}
}

// WithUserAgent sets the UserAgent of the Client.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithTimeout sets the timeout of the Client.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

// WithConnectTimeout sets the connection timeout of the Client.
func WithConnectTimeout(connectTimeout time.Duration) Option {
	return func(c *Client) {
		c.ConnectTimeout = connectTimeout
	}
}

// WithBasicAuth sets the BasicAuth credentials for the Client.
func WithBasicAuth(basicAuth *BasicAuth) Option {
	return func(c *Client) {
		c.BasicAuth = basicAuth
	}
}

// WithCustomPool sets the custom pool for the Client.
func WithCustomPool(customPool *CustomPool) Option {
	return func(c *Client) {
		c.CustomPool = customPool
	}
}

// WithOAuth sets the OAuth credentials for the Client.
func WithOAuth(oauth *OAuth) Option {
	return func(c *Client) {
		c.OAuth = oauth
	}
}

// WithDisableCache disables caching in the Client.
func WithDisableCache() Option {
	return func(c *Client) {
		c.DisableCache = true
	}
}

// WithContentType sets the Content-Type of the Client.
func WithContentType(contentType ContentType) Option {
	return func(c *Client) {
		c.ContentType = contentType
	}
}

// WithFollowRedirect enables or disables following redirects.
func WithFollowRedirect(enabled bool) Option {
	return func(c *Client) {
		c.FollowRedirect = enabled
	}
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
	*clientcredentials.Config
	EndpointParams url.Values
	ClientID       string
	ClientSecret   string
	TokenURL       string
	Scopes         []string
	AuthStyle      AuthStyle
}

// CustomPool defines a separate internal *transport* and connection pooling.
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
// Client should expect a response status code of 200(OK) if resource exists,
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

// AsyncGet is the *asynchronous* option for GET.
// The go routine calling AsyncGet(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncGet(url string, f func(*Response), headers ...http.Header) {
	r.AsyncGetWithContext(context.Background(), url, f, headers...)
}

// AsyncGetWithContext is the *asynchronous* option for GET.
// The go routine calling AsyncGet(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncGetWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.GetWithContext(ctx, url, headers...), f)
}

// AsyncPost is the *asynchronous* option for POST.
// The go routine calling AsyncPost(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPost(url string, body any, f func(*Response), headers ...http.Header) {
	r.AsyncPostWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPostWithContext is the *asynchronous* option for POST.
// The go routine calling AsyncPost(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPostWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.PostWithContext(ctx, url, body, headers...), f)
}

// AsyncPut is the *asynchronous* option for PUT.
// The go routine calling AsyncPut(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPut(url string, body any, f func(*Response), headers ...http.Header) {
	r.AsyncPutWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPutWithContext is the *asynchronous* option for PUT.
// The go routine calling AsyncPut(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPutWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.PutWithContext(ctx, url, body, headers...), f)
}

// AsyncPatch is the *asynchronous* option for PATCH.
// The go routine calling AsyncPatch(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPatch(url string, body any, f func(*Response), headers ...http.Header) {
	r.AsyncPatchWithContext(context.Background(), url, body, f, headers...)
}

// AsyncPatchWithContext is the *asynchronous* option for PATCH.
// The go routine calling AsyncPatch(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncPatchWithContext(ctx context.Context, url string, body any, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.PatchWithContext(ctx, url, body, headers...), f)
}

// AsyncDelete is the *asynchronous* option for DELETE.
// The go routine calling AsyncDelete(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncDelete(url string, f func(*Response), headers ...http.Header) {
	r.AsyncDeleteWithContext(context.Background(), url, f, headers...)
}

// AsyncDeleteWithContext is the *asynchronous* option for DELETE.
// The go routine calling AsyncDelete(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncDeleteWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.DeleteWithContext(ctx, url, headers...), f)
}

// AsyncHead is the *asynchronous* option for HEAD.
// The go routine calling AsyncHead(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncHead(url string, f func(*Response), headers ...http.Header) {
	r.AsyncHeadWithContext(context.Background(), url, f, headers...)
}

// AsyncHeadWithContext is the *asynchronous* option for HEAD.
// The go routine calling AsyncHead(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncHeadWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.HeadWithContext(ctx, url, headers...), f)
}

// AsyncOptions is the *asynchronous* option for OPTIONS.
// The go routine calling AsyncOptions(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncOptions(url string, f func(*Response), headers ...http.Header) {
	r.AsyncOptionsWithContext(context.Background(), url, f, headers...)
}

// AsyncOptionsWithContext is the *asynchronous* option for OPTIONS.
// The go routine calling AsyncOptions(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
func (r *Client) AsyncOptionsWithContext(ctx context.Context, url string, f func(*Response), headers ...http.Header) {
	go doAsyncRequest(r.OptionsWithContext(ctx, url, headers...), f)
}

func doAsyncRequest(r *Response, f func(*Response)) {
	f(r)
}

// doChanAsync creates a new request and sends it asynchronously.
func (r *Client) doChanAsync(ctx context.Context, url string, verb string, body any, headers ...http.Header) <-chan *Response {
	rChan := make(chan *Response, 1)
	go func() {
		defer close(rChan)
		select {
		case <-ctx.Done():
			rChan <- &Response{Err: ctx.Err()}
		default:
			rChan <- r.newRequest(ctx, verb, url, body, headers...)
		}
	}()

	return rChan
}

// ChanGet sends an asynchronous GET request to the specified URL.
//
// Parameters:
//   - url: The URL to send the GET request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanGet(url string, headers ...http.Header) <-chan *Response {
	return r.ChanGetWithContext(context.Background(), url, headers...)
}

// ChanGetWithContext sends an asynchronous GET request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the GET request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanGetWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodGet, nil, headers...)
}

// ChanPost sends an asynchronous POST request to the specified URL.
//
// Parameters:
//   - url: The URL to send the POST request to.
//   - body: The body of the POST request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPost(url string, body any, headers ...http.Header) <-chan *Response {
	return r.ChanPostWithContext(context.Background(), url, body, headers...)
}

// ChanPostWithContext sends an asynchronous POST request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the POST request to.
//   - body: The body of the POST request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPostWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodPost, body, headers...)
}

// ChanPut sends an asynchronous PUT request to the specified URL.
//
// Parameters:
//   - url: The URL to send the PUT request to.
//   - body: The body of the PUT request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPut(url string, body any, headers ...http.Header) <-chan *Response {
	return r.ChanPutWithContext(context.Background(), url, body, headers...)
}

// ChanPutWithContext sends an asynchronous PUT request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the PUT request to.
//   - body: The body of the PUT request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPutWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodPut, body, headers...)
}

// ChanPatch sends an asynchronous PATCH request to the specified URL.
//
// Parameters:
//   - url: The URL to send the PATCH request to.
//   - body: The body of the PATCH request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPatch(url string, body any, headers ...http.Header) <-chan *Response {
	return r.ChanPatchWithContext(context.Background(), url, body, headers...)
}

// ChanPatchWithContext sends an asynchronous PATCH request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the PATCH request to.
//   - body: The body of the PATCH request.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanPatchWithContext(ctx context.Context, url string, body any, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodPatch, body, headers...)
}

// ChanDelete sends an asynchronous DELETE request to the specified URL.
//
// Parameters:
//   - url: The URL to send the DELETE request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanDelete(url string, headers ...http.Header) <-chan *Response {
	return r.ChanDeleteWithContext(context.Background(), url, headers...)
}

// ChanDeleteWithContext sends an asynchronous DELETE request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the DELETE request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanDeleteWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodDelete, nil, headers...)
}

// ChanHead sends an asynchronous HEAD request to the specified URL.
//
// Parameters:
//   - url: The URL to send the HEAD request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanHead(url string, headers ...http.Header) <-chan *Response {
	return r.ChanHeadWithContext(context.Background(), url, headers...)
}

// ChanHeadWithContext sends an asynchronous HEAD request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the HEAD request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanHeadWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodHead, nil, headers...)
}

// ChanOptions sends an asynchronous OPTIONS request to the specified URL.
//
// Parameters:
//   - url: The URL to send the OPTIONS request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanOptions(url string, headers ...http.Header) <-chan *Response {
	return r.ChanOptionsWithContext(context.Background(), url, headers...)
}

// ChanOptionsWithContext sends an asynchronous OPTIONS request to the specified URL with context.
//
// Parameters:
//   - ctx: The context for the request.
//   - url: The URL to send the OPTIONS request to.
//   - headers: Optional HTTP headers to include in the request.
//
// Returns:
//
//	A channel that will receive the Response when it's ready.
func (r *Client) ChanOptionsWithContext(ctx context.Context, url string, headers ...http.Header) <-chan *Response {
	return r.doChanAsync(ctx, url, http.MethodOptions, nil, headers...)
}
