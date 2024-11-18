package rest

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/url"
	"strings"

	"github.com/pkg/errors"
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
	marshallers = map[ContentType]MediaMarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
		FORM: formMedia,
	}
	unmarshallers = map[ContentType]MediaUnmarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
	}
)

type NamedMedia interface {
	Name() string
}

type MediaMarshaler interface {
	NamedMedia
	Marshal(body any) (io.Reader, error)
}

type MediaUnmarshaler interface {
	NamedMedia
	Unmarshal(data []byte, v any) error
}

type JSONMedia struct {
	ContentType string
}

func (r JSONMedia) Marshal(body any) (io.Reader, error) {
	b, err := json.Marshal(body)
	return bytes.NewBuffer(b), err
}

func (r JSONMedia) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (r JSONMedia) Name() string {
	return r.ContentType
}

type XMLMedia struct {
	ContentType string
}

func (r XMLMedia) Marshal(body any) (io.Reader, error) {
	b, err := xml.Marshal(body)
	return bytes.NewBuffer(b), err
}

func (r XMLMedia) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

func (r XMLMedia) Name() string {
	return r.ContentType
}

type FormMedia struct {
	ContentType string
}

func (r FormMedia) Marshal(body any) (io.Reader, error) {
	b, ok := body.(url.Values)
	if !ok {
		return nil, errors.New("body must be of type url.Values or map[string]interface{}")
	}

	return strings.NewReader(b.Encode()), nil
}

func (r FormMedia) Name() string {
	return r.ContentType
}
