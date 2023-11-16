package rest_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
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
