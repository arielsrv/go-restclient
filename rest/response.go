package rest

import (
	"encoding/xml"
	"errors"
	"fmt"
	"maps"
	"mime"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
	"unsafe"
)

// Response ...
type Response struct {
	*http.Response
	Err          error
	Problem      *Problem
	ttl          *time.Time
	lastModified *time.Time
	etag         string
	bytes        []byte
	revalidate   bool
}

// Problem represents a rfc7807 json|xml API problem response. https://datatracker.ietf.org/doc/html/rfc7807#section-1
type Problem struct {
	XMLName  xml.Name `json:"-"                  xml:"problem,omitempty"`
	XMLNS    xml.Name `json:"-"                  xml:"xmlns,attr,omitempty"`
	Type     string   `json:"type,omitempty"     xml:"type,omitempty"`
	Title    string   `json:"title,omitempty"    xml:"title,omitempty"`
	Detail   string   `json:"detail,omitempty"   xml:"detail,omitempty"`
	Instance string   `json:"instance,omitempty" xml:"instance,omitempty"`
	Status   int      `json:"status,omitempty"   xml:"status,omitempty"`
}

// size returns the size of the Response in bytes.
func (r *Response) size() int64 {
	size := int64(unsafe.Sizeof(*r))

	size += int64(len(r.bytes))
	size += int64(unsafe.Sizeof(*r.Problem))
	size += int64(unsafe.Sizeof(*r.ttl))
	size += int64(unsafe.Sizeof(*r.lastModified))
	size += int64(len(r.etag))

	size += int64(len(r.Response.Proto))
	size += int64(len(r.Response.Status))

	return size
}

// String return the Response Body as a String.
func (r *Response) String() string {
	return string(r.bytes)
}

// FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]any`.
func (r *Response) FillUp(fill any) error {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(r.bytes)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	for unmarshaller := range maps.Values(mediaUnmarshaler) {
		values := unmarshaller.DefaultHeaders().Values("Content-Type")
		for i := range values {
			value := values[i]
			if mediaType == value {
				return unmarshaller.Unmarshal(r.bytes, fill)
			}
		}
	}

	return fmt.Errorf("unmarshal fail, unsupported content type: %s", contentType)
}

// TypedFillUp FillUp set the *fill* parameter with the corresponding JSON or XML response.
// fill could be `struct` or `map[string]any`.
// Deprecated: use Deserialize[T] instead.
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
