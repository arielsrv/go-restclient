package rest

var dfltClient = Client{}

// Get issues a GET HTTP verb to the specified URL using the default client.
//
// In RESTful design, GET is used for "reading" or retrieving a resource.
// Client should expect a response status code of 200 (OK) if the resource exists,
// 404 (Not Found) if it doesn't, or 400 (Bad Request).
func Get(url string) *Response {
	return dfltClient.Get(url)
}

// Post issues a POST HTTP verb to the specified URL using the default client.
//
// In RESTful design, POST is used for "creating" a resource.
// Client should expect a response status code of 201 (Created), 400 (Bad Request),
// 404 (Not Found), or 409 (Conflict) if the resource already exists.
//
// Body could be any of the following: string, []byte, struct or map.
func Post(url string, body any) *Response {
	return dfltClient.Post(url, body)
}

// Put issues a PUT HTTP verb to the specified URL using the default client.
//
// In RESTful design, PUT is used for "updating" a resource.
// Client should expect a response status code of 200 (OK), 404 (Not Found),
// or 400 (Bad Request). 200 (OK) could also be 204 (No Content).
//
// Body could be any of the following: string, []byte, struct or map.
func Put(url string, body any) *Response {
	return dfltClient.Put(url, body)
}

// Patch issues a PATCH HTTP verb to the specified URL using the default client.
//
// In RESTful design, PATCH is used for "partially updating" a resource.
// Client should expect a response status code of 200 (OK), 404 (Not Found),
// or 400 (Bad Request). 200 (OK) could also be 204 (No Content).
//
// Body could be any of the following: string, []byte, struct or map.
func Patch(url string, body any) *Response {
	return dfltClient.Patch(url, body)
}

// Delete issues a DELETE HTTP verb to the specified URL using the default client.
//
// In RESTful design, DELETE is used to "delete" a resource.
// Client should expect a response status code of 200 (OK), 404 (Not Found),
// or 400 (Bad Request).
func Delete(url string) *Response {
	return dfltClient.Delete(url)
}

// Head issues a HEAD HTTP verb to the specified URL using the default client.
//
// In RESTful design, HEAD is used to "read" resource headers only.
// Client should expect a response status code of 200 (OK) if the resource exists,
// 404 (Not Found) if it doesn't, or 400 (Bad Request).
func Head(url string) *Response {
	return dfltClient.Head(url)
}

// Options issues an OPTIONS HTTP verb to the specified URL using the default client.
//
// In RESTful design, OPTIONS is used to get information about the resource
// and supported HTTP verbs.
// Client should expect a response status code of 200 (OK) if the resource exists,
// 404 (Not Found) if it doesn't, or 400 (Bad Request).
func Options(url string) *Response {
	return dfltClient.Options(url)
}
