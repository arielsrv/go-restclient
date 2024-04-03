package rest

import (
	"container/list"
	"encoding/xml"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/pkg/errors"

	"github.com/goccy/go-json"
)

// Response ...
type Response struct {
	*http.Response
	Err             error
	byteBody        []byte
	listElement     *list.Element
	skipListElement *skipListNode
	ttl             *time.Time
	lastModified    *time.Time
	etag            string
	revalidate      bool
	cacheHit        atomic.Value
}

func (r *Response) size() int64 {
	size := int64(unsafe.Sizeof(*r))

	size += int64(len(r.byteBody))
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
	return r.byteBody
}

// FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]interface{}`.
func (r *Response) FillUp(fill interface{}) error {
	ctypeJSON := "application/json"
	ctypeXML := "application/xml"

	ctype := strings.ToLower(r.Header.Get("Content-Type"))

	for i := 0; i < 2; i++ {
		switch {
		case strings.Contains(ctype, ctypeJSON):
			return json.Unmarshal(r.byteBody, fill)
		case strings.Contains(ctype, ctypeXML):
			return xml.Unmarshal(r.byteBody, fill)
		case i == 0:
			ctype = http.DetectContentType(r.byteBody)
		}
	}

	return errors.New("response format neither JSON nor XML")
}

// TypedFillUp FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]interface{}`.
func TypedFillUp[TResult any](r *Response) (*TResult, error) {
	target := new(TResult)
	err := r.FillUp(&target)
	if err != nil {
		return nil, err
	}
	return target, nil
}

// CacheHit shows if a response was get from the cache.
func (r *Response) CacheHit() bool {
	if hit, ok := r.cacheHit.Load().(bool); hit && ok {
		return true
	}
	return false
}

// Debug let any request/response to be dumped, showing how the request/response
// went through the wire, only if debug mode is *on* on RequestBuilder.
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
