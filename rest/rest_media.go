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

// MIME type constants for various content types used in HTTP requests and responses.
const (
	// MIMETextXML is the MIME type for XML content in text format.
	MIMETextXML = "text/xml"

	// MIMETextPlain is the MIME type for plain text content.
	MIMETextPlain = "text/plain"

	// MIMEApplicationXML is the MIME type for XML content in application format.
	MIMEApplicationXML = "application/xml"

	// MIMEApplicationJSON is the MIME type for JSON content.
	MIMEApplicationJSON = "application/json"

	// MIMEApplicationProblemJSON is the MIME type for RFC7807 problem details in JSON format.
	MIMEApplicationProblemJSON = "application/problem+json"

	// MIMEApplicationProblemXML is the MIME type for RFC7807 problem details in XML format.
	MIMEApplicationProblemXML = "application/problem+xml"

	// MIMEApplicationForm is the MIME type for form-urlencoded content.
	MIMEApplicationForm = "application/x-www-form-urlencoded"
)

// HTTP header constants for content negotiation.
const (
	// CanonicalContentTypeHeader is the canonical name of the Content-Type header.
	CanonicalContentTypeHeader = "Content-Type"

	// CanonicalAcceptHeader is the canonical name of the Accept header.
	CanonicalAcceptHeader = "Accept"
)

// ContentType represents the content type for the body of HTTP verbs like
// POST, PUT, and PATCH. It's used to determine how to marshal and unmarshal
// request and response bodies.
type ContentType int

// ContentType constants for supported content types.
const (
	// JSON represents a JSON content type.
	JSON ContentType = iota

	// XML represents an XML content type.
	XML

	// FORM represents a form-urlencoded content type.
	FORM
)

// Media instances for each supported content type.
var (
	jsonMedia = &JSONMedia{
		ContentType: MIMEApplicationJSON,
	}
	xmlMedia = &XMLMedia{
		ContentType: MIMEApplicationXML,
	}
	formMedia = &FormMedia{
		ContentType: MIMEApplicationForm,
	}
)

// Maps of content types to their respective marshalers and unmarshalers.
var (
	// contentMarshalers maps ContentType to MediaMarshaler for request body serialization.
	contentMarshalers = map[ContentType]MediaMarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
		FORM: formMedia,
	}

	// readMarshalers maps ContentType to MediaUnmarshaler for response body deserialization.
	readMarshalers = map[ContentType]MediaUnmarshaler{
		JSON: jsonMedia,
		XML:  xmlMedia,
	}
)

// Media is an interface for types that can provide default HTTP headers
// for content negotiation.
type Media interface {
	// DefaultHeaders returns the default HTTP headers for this media type.
	DefaultHeaders() http.Header
}

// MediaMarshaler is an interface for types that can marshal data into a specific
// content type format.
type MediaMarshaler interface {
	Media
	// Marshal converts the given body into the appropriate format and returns
	// an io.Reader containing the marshaled data.
	Marshal(body any) (io.Reader, error)
}

// MediaUnmarshaler is an interface for types that can unmarshal data from a specific
// content type format.
type MediaUnmarshaler interface {
	Media
	// Unmarshal parses the data and stores the result in the value pointed to by v.
	Unmarshal(data []byte, v any) error
}

// JSONMedia implements the Media, MediaMarshaler, and MediaUnmarshaler interfaces
// for JSON content type.
type JSONMedia struct {
	// ContentType is the MIME type for this media, typically "application/json".
	ContentType string
}

// Marshal converts the given body into JSON format and returns an io.Reader
// containing the marshaled data.
func (r JSONMedia) Marshal(body any) (io.Reader, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

// Unmarshal parses the JSON-encoded data and stores the result in the value
// pointed to by v.
func (r JSONMedia) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// DefaultHeaders returns the default HTTP headers for JSON content type.
// It sets the Content-Type header to the configured ContentType and
// the Accept header to accept JSON and JSON problem responses.
func (r JSONMedia) DefaultHeaders() http.Header {
	return http.Header{
		CanonicalContentTypeHeader: []string{
			r.ContentType,
		},
		CanonicalAcceptHeader: []string{
			MIMEApplicationJSON,
			MIMEApplicationProblemJSON,
		},
	}
}

// XMLMedia implements the Media, MediaMarshaler, and MediaUnmarshaler interfaces
// for XML content type.
type XMLMedia struct {
	// ContentType is the MIME type for this media, typically "application/xml".
	ContentType string
}

// Marshal converts the given body into XML format and returns an io.Reader
// containing the marshaled data.
func (r XMLMedia) Marshal(body any) (io.Reader, error) {
	b, err := xml.Marshal(body)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(b), nil
}

// Unmarshal parses the XML-encoded data and stores the result in the value
// pointed to by v.
func (r XMLMedia) Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

// DefaultHeaders returns the default HTTP headers for XML content type.
// It sets the Content-Type header to the configured ContentType and
// the Accept header to accept various XML formats.
func (r XMLMedia) DefaultHeaders() http.Header {
	return http.Header{
		CanonicalContentTypeHeader: []string{
			r.ContentType,
		},
		CanonicalAcceptHeader: []string{
			MIMEApplicationXML,
			MIMEApplicationProblemXML,
			MIMETextXML,
		},
	}
}

// FormMedia implements the Media and MediaMarshaler interfaces for
// form-urlencoded content type.
type FormMedia struct {
	// ContentType is the MIME type for this media, typically "application/x-www-form-urlencoded".
	ContentType string
}

// Marshal converts the given body into form-urlencoded format and returns an io.Reader
// containing the marshaled data. The body must be of type url.Values.
func (r FormMedia) Marshal(body any) (io.Reader, error) {
	if values, ok := body.(url.Values); ok {
		return strings.NewReader(values.Encode()), nil
	}

	return nil, errors.New("body must be of type url.Values")
}

// DefaultHeaders returns the default HTTP headers for form-urlencoded content type.
// It sets the Content-Type header to the configured ContentType and
// the Accept header to accept JSON, XML, and plain text responses.
func (r FormMedia) DefaultHeaders() http.Header {
	return http.Header{
		CanonicalContentTypeHeader: []string{
			r.ContentType,
		},
		CanonicalAcceptHeader: []string{
			MIMEApplicationJSON,
			MIMEApplicationXML,
			MIMETextPlain,
		},
	}
}
