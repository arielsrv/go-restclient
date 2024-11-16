package rest

import (
	"net/http"
	"sync"
)

var dfltClient = Client{
	rwMtx: sync.RWMutex{},
}

// Get issues a GET HTTP verb to the specified URL.
//
// In Restful, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
//
// Get uses the DefaultBuilder.
func Get(url string) *Response {
	return dfltClient.Get(url)
}

// Post issues a POST HTTP verb to the specified URL.
//
// In Restful, POST is used for "creating" a resource.
// Client should expect a response status code of 201(Created), 400(Bad Request),
// 404(Not Found), or 409(Conflict) if resource already exist.
//
// Body could be any of the form: string, []byte, struct & map.
//
// Post uses the DefaultBuilder.
func Post(url string, body any) *Response {
	return dfltClient.Post(url, body)
}

// Put issues a PUT HTTP verb to the specified URL.
//
// In Restful, PUT is used for "updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
//
// Put uses the DefaultBuilder.
func Put(url string, body any) *Response {
	return dfltClient.Put(url, body)
}

// Patch issues a PATCH HTTP verb to the specified URL
//
// In Restful, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request). 200(OK) could be also 204(No Content)
//
// Body could be any of the form: string, []byte, struct & map.
//
// Patch uses the DefaultBuilder.
func Patch(url string, body any) *Response {
	return dfltClient.Patch(url, body)
}

// Delete issues a DELETE HTTP verb to the specified URL
//
// In Restful, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200(OK), 404(Not Found),
// or 400(Bad Request).
//
// Delete uses the DefaultBuilder.
func Delete(url string) *Response {
	return dfltClient.Delete(url)
}

// Head issues a HEAD HTTP verb to the specified URL
//
// In Restful, HEAD is used to "read" a resource headers only.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
//
// Head uses the DefaultBuilder.
func Head(url string) *Response {
	return dfltClient.Head(url)
}

// Options issues a OPTIONS HTTP verb to the specified URL
//
// In Restful, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200(OK) if resource exists,
// 404(Not Found) if it doesn't, or 400(Bad Request).
//
// Options uses the DefaultBuilder.
func Options(url string) *Response {
	return dfltClient.Options(url)
}

// AsyncGet is the *asynchronous* option for GET.
// The go routine calling AsyncGet(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncGet uses the DefaultBuilder.
func AsyncGet(url string, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncGet(url, f, headers...)
}

// AsyncPost is the *asynchronous* option for POST.
// The go routine calling AsyncPost(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncPost uses the DefaultBuilder.
func AsyncPost(url string, body any, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncPost(url, body, f, headers...)
}

// AsyncPut is the *asynchronous* option for PUT.
// The go routine calling AsyncPut(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncPut uses the DefaultBuilder.
func AsyncPut(url string, body any, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncPut(url, body, f, headers...)
}

// AsyncPatch is the *asynchronous* option for PATCH.
// The go routine calling AsyncPatch(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncPatch uses the DefaultBuilder.
func AsyncPatch(url string, body any, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncPatch(url, body, f, headers...)
}

// AsyncDelete is the *asynchronous* option for DELETE.
// The go routine calling AsyncDelete(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncDelete uses the DefaultBuilder.
func AsyncDelete(url string, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncDelete(url, f, headers...)
}

// AsyncHead is the *asynchronous* option for HEAD.
// The go routine calling AsyncHead(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncHead uses the DefaultBuilder.
func AsyncHead(url string, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncHead(url, f, headers...)
}

// AsyncOptions is the *asynchronous* option for OPTIONS.
// The go routine calling AsyncOptions(), will not be blocked.
//
// Whenever the Response is ready, the *f* function will be called back.
//
// AsyncOptions uses the DefaultBuilder.
func AsyncOptions(url string, f func(*Response), headers ...http.Header) {
	dfltClient.AsyncOptions(url, f, headers...)
}
