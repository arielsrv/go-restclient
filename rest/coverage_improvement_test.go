package rest

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCoverageImprovement_Response(t *testing.T) {
	t.Run("IsOk_NilResponse", func(t *testing.T) {
		var r *Response
		if r.IsOk() {
			t.Error("IsOk() should return false for nil Response")
		}
	})

	t.Run("IsOk_NilHTTPResponse", func(t *testing.T) {
		r := &Response{Response: nil}
		if r.IsOk() {
			t.Error("IsOk() should return false for nil http.Response")
		}
	})

	t.Run("VerifyIsOkOrError_NilResponse", func(t *testing.T) {
		var r *Response
		err := r.VerifyIsOkOrError()
		if err == nil || err.Error() != "response is nil" {
			t.Errorf("expected 'response is nil' error, got %v", err)
		}
	})

	t.Run("VerifyIsOkOrError_WithErr", func(t *testing.T) {
		expectedErr := errors.New("some error")
		r := &Response{Err: expectedErr}
		err := r.VerifyIsOkOrError()
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("VerifyIsOkOrError_NilHTTPResponse", func(t *testing.T) {
		r := &Response{Response: nil}
		err := r.VerifyIsOkOrError()
		if err == nil || err.Error() != "http.Response is nil" {
			t.Errorf("expected 'http.Response is nil' error, got %v", err)
		}
	})

	t.Run("VerifyIsOkOrError_NotOk", func(t *testing.T) {
		r := &Response{
			Response: &http.Response{StatusCode: http.StatusBadRequest},
			bytes:    []byte("error body"),
		}
		err := r.VerifyIsOkOrError()
		if err == nil || !strings.Contains(err.Error(), "status code 400") {
			t.Errorf("expected error with status code 400, got %v", err)
		}
	})

	t.Run("Debug_NilResponse", func(t *testing.T) {
		var r *Response
		debug := r.Debug()
		if debug != "Response is nil" {
			t.Errorf("expected 'Response is nil', got %q", debug)
		}
	})

	t.Run("Debug_NilRequestAndResponse", func(t *testing.T) {
		r := &Response{
			Response: nil,
		}
		debug := r.Debug()
		if !strings.Contains(debug, "Request is nil") || !strings.Contains(debug, "Response is nil") {
			t.Errorf("expected both Request and Response to be nil, got %q", debug)
		}
	})

	t.Run("Debug_WithRequestAndResponse", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		r := &Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    req,
			},
			bytes: []byte("response body"),
		}
		debug := r.Debug()
		if !strings.Contains(debug, "GET / HTTP/1.1") || !strings.Contains(debug, "response body") {
			t.Errorf("debug output should contain request and response info, got %q", debug)
		}
	})

	t.Run("FillUp_NilResponse", func(t *testing.T) {
		var r *Response
		var data any
		err := r.FillUp(&data)
		if err == nil || err.Error() != "response is nil" {
			t.Errorf("expected 'response is nil' error, got %v", err)
		}
	})

	t.Run("FillUp_WithErr", func(t *testing.T) {
		expectedErr := errors.New("some error")
		r := &Response{Err: expectedErr}
		var data any
		err := r.FillUp(&data)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	})

	t.Run("FillUp_NilHTTPResponse", func(t *testing.T) {
		r := &Response{Response: nil}
		var data any
		err := r.FillUp(&data)
		if err == nil || err.Error() != "http.Response is nil" {
			t.Errorf("expected 'http.Response is nil' error, got %v", err)
		}
	})

	t.Run("FillUp_InvalidContentType", func(t *testing.T) {
		r := &Response{
			Response: &http.Response{
				Header: http.Header{
					CanonicalContentTypeHeader: []string{"invalid; content/type"},
				},
			},
		}
		var data any
		err := r.FillUp(&data)
		if err == nil || !strings.Contains(err.Error(), "invalid content type") {
			t.Errorf("expected 'invalid content type' error, got %v", err)
		}
	})
}

func TestCoverageImprovement_Net(t *testing.T) {
	t.Run("setProblem", func(t *testing.T) {
		// Test when Content-Type contains "problem" but unmarshal fails
		r := &Response{
			Response: &http.Response{
				Header: http.Header{
					CanonicalContentTypeHeader: []string{"application/problem+json"},
				},
			},
			bytes: []byte("invalid json"),
		}
		setProblem(r)
		// It should return early without panic
		if r.Problem != nil {
			t.Errorf("Problem should be nil (unmarshal failed), got %+v", r.Problem)
		}
	})

	t.Run("checkMockup_InvalidURL", func(t *testing.T) {
		// checkMockup is now inlined in newRequest; test the same invalid-URL path
		// by making a real request through the client while mock is active.
		StartMockupServer()
		defer StopMockupServer()

		c := &Client{}
		resp := c.Get("http://[::1]:invalid/path")
		if resp.Err == nil {
			t.Error("expected error for invalid URL")
		}
	})
}

func TestCoverageImprovement_RegisterMetrics(t *testing.T) {
	// Just call it to cover the empty function
	registerMetrics(nil)
}

func TestCoverageImprovement_Net_Additional(t *testing.T) {
	t.Run("setRespReader_GzipError", func(t *testing.T) {
		c := &Client{}
		req, _ := http.NewRequest(http.MethodGet, "http://example.com", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		resp := &http.Response{
			Header: http.Header{
				"Content-Encoding": []string{"gzip"},
			},
			Body: io.NopCloser(strings.NewReader("not a gzip stream")),
		}
		_, err := c.setRespReader(req, resp)
		if err == nil {
			t.Error("expected error for invalid gzip stream in setRespReader")
		}
	})
}
