package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

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
)

var (
	jsonMedia = &JSONMedia{ContentType: "application/json"}
	xmlMedia  = &XMLMedia{ContentType: "application/xml"}
	formMedia = &FormMedia{ContentType: "application/x-www-form-urlencoded"}
)

var (
	mediaMarshalers = map[ContentType]MediaMarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
		FORM: formMedia,
	}
	mediaUnmarshalers = map[ContentType]MediaUnmarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
	}
)

type Media interface {
	DefaultHeaders() http.Header
}

type MediaMarshaler interface {
	Media
	Marshal(body any) (io.Reader, error)
}

type MediaUnmarshaler interface {
	Media
	Unmarshal(data []byte, v any) error
}

type JSONMedia struct {
	ContentType string
}

func (r JSONMedia) Marshal(body any) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

func (r JSONMedia) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (r JSONMedia) DefaultHeaders() http.Header {
	return http.Header{
		"Content-Type": []string{
			r.ContentType,
			"application/problem+json",
		},
		"Accept": []string{
			"application/json",
			"application/problem+json",
		},
	}
}

type XMLMedia struct {
	ContentType string
}

func (r XMLMedia) Marshal(body any) (io.Reader, error) {
	b, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

func (r XMLMedia) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

func (r XMLMedia) DefaultHeaders() http.Header {
	return http.Header{
		"Content-Type": []string{
			r.ContentType,
		},
		"Accept": []string{
			"application/xml",
			"application/problem+xml",
			"txt/xml",
		},
	}
}

type FormMedia struct {
	ContentType string
}

func (r FormMedia) Marshal(body any) (io.Reader, error) {
	if values, ok := body.(url.Values); ok {
		return strings.NewReader(values.Encode()), nil
	}

	return nil, errors.New("body must be of type url.Values")
}

func (r FormMedia) DefaultHeaders() http.Header {
	return http.Header{
		"Content-Type": []string{
			r.ContentType,
		},
		"Accept": []string{
			"application/json",
			"application/xml",
			"text/plain",
		},
	}
}
