package rest

import (
	"container/list"
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"unsafe"
)

type IConcurrent interface {
	Get(url string) *FutureResponse
	Post(url string, body any) *FutureResponse
	Patch(url string, body any) *FutureResponse
	Put(url string, body any) *FutureResponse
	Delete(url string) *FutureResponse
	Head(url string) *FutureResponse
	Options(url string) *FutureResponse

	GetWithContext(ctx context.Context, url string) *FutureResponse
	PostWithContext(ctx context.Context, url string, body any) *FutureResponse
	PatchWithContext(ctx context.Context, url string, body any) *FutureResponse
	PutWithContext(ctx context.Context, url string, body any) *FutureResponse
	DeleteWithContext(ctx context.Context, url string) *FutureResponse
	HeadWithContext(ctx context.Context, url string) *FutureResponse
	OptionsWithContext(ctx context.Context, url string) *FutureResponse
}

// FutureResponse represents a response to be completed after a ForkJoin
// operation is done.
//
// FutureResponse will never be nil, and has a Response function for getting the
// Response, that will be nil after the ForkJoin operation is completed.
type FutureResponse struct {
	p unsafe.Pointer
}

// Response gives you the Response of a Request,after the ForkJoin operation
// is completed.
//
// Response will be nil if the ForkJoin operation is not completed.
func (fr *FutureResponse) Response() *Response {
	return (*Response)(fr.p)
}

// Concurrent has methods for Get, Post, Put, Patch, Delete, Head & Options,
// with almost the same API as the synchronous methods.
// The difference is that these methods return a FutureResponse, which holds a pointer to
// Response. Response inside FutureResponse is nil until the request has finished.
//
//	rest.ForkJoin(func(c *rest.Concurrent){
//		futureA = c.Get("/url/1")
//		futureB = c.Get("/url/2")
//	})
//
// The difference is that Concurrent methods returns a FutureResponse, instead
// of a Response.
type Concurrent struct {
	reqBuilder *RequestBuilder
	list       list.List
	wg         sync.WaitGroup
}

// Get issues a GET HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) Get(url string) *FutureResponse {
	return c.GetWithContext(context.Background(), url)
}

// Post issues a POST HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) Post(url string, body any) *FutureResponse {
	return c.PostWithContext(context.Background(), url, body)
}

// Patch issues a PATCH HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) Patch(url string, body any) *FutureResponse {
	return c.PatchWithContext(context.Background(), url, body)
}

// Put issues a PUT HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) Put(url string, body any) *FutureResponse {
	return c.PutWithContext(context.Background(), url, body)
}

// Delete issues a DELETE HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (c *Concurrent) Delete(url string) *FutureResponse {
	return c.DeleteWithContext(context.Background(), url)
}

// Head issues a HEAD HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) Head(url string) *FutureResponse {
	return c.HeadWithContext(context.Background(), url)
}

// Options issues a OPTIONS HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) Options(url string) *FutureResponse {
	return c.OptionsWithContext(context.Background(), url)
}

// GetWithContext issues a GET HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) GetWithContext(ctx context.Context, url string) *FutureResponse {
	return c.doRequest(ctx, http.MethodGet, url, nil)
}

// PostWithContext issues a POST HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) PostWithContext(ctx context.Context, url string, body any) *FutureResponse {
	return c.doRequest(ctx, http.MethodPost, url, body)
}

// PatchWithContext issues a PATCH HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) PatchWithContext(ctx context.Context, url string, body any) *FutureResponse {
	return c.doRequest(ctx, http.MethodPatch, url, body)
}

// PutWithContext issues a PUT HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
func (c *Concurrent) PutWithContext(ctx context.Context, url string, body any) *FutureResponse {
	return c.doRequest(ctx, http.MethodPut, url, body)
}

// DeleteWithContext issues a DELETE HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
func (c *Concurrent) DeleteWithContext(ctx context.Context, url string) *FutureResponse {
	return c.doRequest(ctx, http.MethodDelete, url, nil)
}

// HeadWithContext issues a HEAD HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) HeadWithContext(ctx context.Context, url string) *FutureResponse {
	return c.doRequest(ctx, http.MethodHead, url, nil)
}

// OptionsWithContext issues a OPTIONS HTTP verb to the specified URL, concurrently with any other
// concurrent requests that may be called.
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
func (c *Concurrent) OptionsWithContext(ctx context.Context, url string) *FutureResponse {
	return c.doRequest(ctx, http.MethodOptions, url, nil)
}

func (c *Concurrent) doRequest(ctx context.Context, verb string, url string, reqBody any) *FutureResponse {
	fr := new(FutureResponse)

	future := func() {
		defer c.wg.Done()
		r := c.reqBuilder.doRequest(ctx, verb, url, reqBody)
		atomic.StorePointer(&fr.p, unsafe.Pointer(r))
	}
	c.list.PushBack(future)

	return fr
}
