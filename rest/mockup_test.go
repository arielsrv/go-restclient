package rest_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/arielsrv/go-restclient/rest"
	"github.com/stretchr/testify/require"
)

func TestMockup(t *testing.T) {
	defer rest.StopMockupServer()
	rest.StartMockupServer()

	myURL := "http://mytest.com/foo?val1=1&val2=2#fragment"

	myHeaders := make(http.Header)
	myHeaders.Add("Hello", "world")

	mock := rest.Mock{
		URL:          myURL,
		HTTPMethod:   http.MethodGet,
		ReqHeaders:   myHeaders,
		RespHTTPCode: http.StatusOK,
		RespBody:     "foo",
	}

	err := rest.AddMockups(&mock)
	require.NoError(t, err)

	v := rest.Get(myURL)
	if v.String() != "foo" {
		t.Fatal("Mockup Fail!")
	}
}

func TestMockup_RequestErr(t *testing.T) {
	defer rest.StopMockupServer()
	rest.StartMockupServer()

	myURL := "http://mytest.com/foo?val1=1&val2=2#fragment"

	myHeaders := make(http.Header)
	myHeaders.Add("Hello", "world")

	mock := rest.Mock{
		URL:          myURL,
		HTTPMethod:   http.MethodGet,
		ReqHeaders:   myHeaders,
		RespHTTPCode: http.StatusOK,
		RespBody:     "foo",
	}

	err := rest.AddMockups(&mock)
	require.NoError(t, err)

	v := rest.Get(":/invalid/url")
	if v.String() != "" {
		t.Fatal("Mockup Error Should Return Error!")
	}
}

func TestMockup_MockErr(t *testing.T) {
	defer rest.StopMockupServer()
	rest.StartMockupServer()

	myURL := ":/invalid/url"

	myHeaders := make(http.Header)
	myHeaders.Add("Hello", "world")

	mock := rest.Mock{
		URL:          myURL,
		Timeout:      time.Duration(100) * time.Millisecond,
		HTTPMethod:   http.MethodGet,
		ReqHeaders:   myHeaders,
		RespHTTPCode: http.StatusOK,
		RespBody:     "foo",
	}

	err := rest.AddMockups(&mock)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing protocol scheme")
}

func TestFlushMockups(t *testing.T) {
	defer rest.StopMockupServer()
	rest.StartMockupServer()

	myURL := "http://mytest.com/foo?val1=1&val2=2#fragment"

	myHeaders := make(http.Header)
	myHeaders.Add("Hello", "world")

	mock := rest.Mock{
		URL:          myURL,
		HTTPMethod:   http.MethodGet,
		ReqHeaders:   myHeaders,
		RespHTTPCode: http.StatusOK,
		RespBody:     "foo",
	}

	err := rest.AddMockups(&mock)
	require.NoError(t, err)

	rest.FlushMockups()

	// Verify that the mockup is removed
	v := rest.Get(myURL)
	if v.String() != rest.ErrMockNotFound.Error() {
		t.Fatal("Mockup Should Be Removed!")
	}
}
