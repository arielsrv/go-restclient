package rest

import (
	"container/list"
	"fmt"
	"maps"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pkg/errors"
)

// Response ...
type Response struct {
	Err      error
	cacheHit atomic.Value
	*http.Response
	listElement     *list.Element
	skipListElement *skipListNode
	ttl             *time.Time
	lastModified    *time.Time
	etag            string
	bytes           []byte
	revalidate      bool
}

func (r *Response) size() int64 {
	size := int64(unsafe.Sizeof(*r))

	size += int64(len(r.bytes))
	size += int64(unsafe.Sizeof(*r.listElement))
	size += int64(unsafe.Sizeof(*r.skipListElement))
	size += int64(unsafe.Sizeof(*r.ttl))
	size += int64(unsafe.Sizeof(*r.lastModified))
	size += int64(len(r.etag))

	size += int64(len(r.Response.Proto))
	size += int64(len(r.Response.Status))

	return size
}

// String return the Response Body as a String.
func (r *Response) String() string {
	return string(r.Bytes())
}

// Bytes return the Response Body as bytes.
func (r *Response) Bytes() []byte {
	return r.bytes
}

// FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]any`.
func (r *Response) FillUp(fill any) error {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(r.bytes)
	}

	for key := range maps.Keys(unmarshallers) {
		media := unmarshallers[key]
		if strings.Contains(contentType, media.Name()) {
			return media.Unmarshal(r.bytes, fill)
		}
	}

	return fmt.Errorf("unsupported content type: %s", contentType)
}

// TypedFillUp FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]any`.
func TypedFillUp[T any](r *Response) (*T, error) {
	result, err := Deserialize[T](r)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// Deserialize fills the provided pointer with the JSON or XML response.
func Deserialize[T any](r *Response) (T, error) {
	var zero T
	if r == nil {
		return zero, errors.New("response is nil")
	}

	var result T
	err := r.FillUp(&result)
	if err != nil {
		return zero, err
	}

	return result, nil
}

// CacheHit shows if a response was get from the cache.
func (r *Response) CacheHit() bool {
	if hit, ok := r.cacheHit.Load().(bool); hit && ok {
		return true
	}

	return false
}

// Debug let any request/response to be dumped, showing how the request/response
// went through the wire, only if debug mode is *on* on RESTClient.
func (r *Response) Debug() string {
	var strReq, strResp string

	if req, err := httputil.DumpRequest(r.Request, true); err != nil {
		strReq = err.Error()
	} else {
		strReq = string(req)
	}

	if resp, err := httputil.DumpResponse(r.Response, false); err != nil {
		strResp = err.Error()
	} else {
		strResp = string(resp)
	}

	const separator = "--------\n"

	dump := separator
	dump += "REQUEST\n"
	dump += separator
	dump += strReq
	dump += "\n" + separator
	dump += "RESPONSE\n"
	dump += separator
	dump += strResp
	dump += r.String() + "\n"

	return dump
}

func (r *Response) IsOk() bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusBadRequest
}

func (r *Response) VerifyIsOkOrError() error {
	if r.Err != nil {
		return r.Err
	}

	if !r.IsOk() {
		return fmt.Errorf("status code %d, body: %s", r.StatusCode, r.String())
	}

	return nil
}
