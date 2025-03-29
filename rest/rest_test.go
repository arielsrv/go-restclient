package rest_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func TestAuthConstants(t *testing.T) {
	assert.Equal(t, rest.AuthStyleAutoDetect, rest.AuthStyle(0))
	assert.Equal(t, rest.AuthStyleInParams, rest.AuthStyle(1))
	assert.Equal(t, rest.AuthStyleInHeader, rest.AuthStyle(2))
}

func TestGet(t *testing.T) {
	resp := rest.Get(server.URL + "/user")
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestGet_GZip(t *testing.T) {
	client := &rest.Client{
		EnableGzip: true,
	}

	resp := client.Get(server.URL + "/gzip/user")
	if resp.Err != nil {
		t.Fatal("Error:", resp.Err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
	t.Log(resp.String())
	if !strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		t.Fatal("Content-Encoding header not found")
	}
}

func TestGet_LastModifiedErr(t *testing.T) {
	client := &rest.Client{
		EnableGzip: true,
	}

	resp := client.Get(server.URL + "/cache/lastmodified/user/err")
	if resp.Err != nil {
		t.Fatal("Error:", resp.Err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestGet_GZip_Err(t *testing.T) {
	client := &rest.Client{
		EnableGzip: true,
	}

	resp := client.Get(server.URL + "/gzip/user/err")
	if resp.Err == nil {
		t.Fatal("Error should not be nil")
	}
}

func TestGet_GZip_Headers(t *testing.T) {
	client := &rest.Client{
		EnableGzip: true,
	}

	headers := make(http.Header)
	headers.Add("Accept-Encoding", "gzip")

	resp := client.Get(server.URL+"/gzip/user", headers)
	if resp.Err != nil {
		t.Fatal("Error:", resp.Err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
	t.Log(resp.String())
	if !strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		t.Fatal("Content-Encoding header not found")
	}
}

func TestGet_NoCache(t *testing.T) {
	resp := rest.Get(server.URL + "/user")
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
	if resp.Cached() {
		t.Fatal("resp.Cached() == true")
	}
}

func TestSlowGet(t *testing.T) {
	var f [100]*rest.Response

	for i := range f {
		f[i] = rb.Get("/slow/user")

		if f[i].StatusCode != http.StatusOK {
			t.Fatal("f Status != OK (200)")
		}
	}
}

func TestHead(t *testing.T) {
	resp := rest.Head(server.URL + "/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestPost(t *testing.T) {
	resp := rest.Post(server.URL+"/user", &User{Name: "Maria"})

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("Status != OK (201)")
	}
}

func TestPost_XMLErr(t *testing.T) {
	client := rest.Client{
		ContentType: rest.XML,
	}
	resp := client.Post(server.URL+"/user", map[string]interface{}{})

	if resp.Err == nil {
		t.Fatal("Error should not be nil")
	}

	if !strings.Contains(resp.Err.Error(), "unsupported type") {
		t.Fatal("Error should contain 'invalid'")
	}
}

func TestPost_Err(t *testing.T) {
	resp := rest.Post(server.URL+"/user", make(chan any))
	if resp.Err == nil {
		t.Fatal("Error should not be nil")
	}
}

func TestPostXML(t *testing.T) {
	rbXML := rest.Client{
		BaseURL:     server.URL,
		ContentType: rest.XML,
	}

	resp := rbXML.Post("/xml/user", &User{Name: "Maria"})

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("Status != OK (201)")
	}
}

func TestPostForm(t *testing.T) {
	fbForm := rest.Client{
		BaseURL:     server.URL,
		ContentType: rest.FORM,
	}

	resp := fbForm.Post("/form/user", url.Values{"name": []string{"Maria"}})

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("Status != OK (201)")
	}
}

func TestPostForm_Err(t *testing.T) {
	fbForm := rest.Client{
		BaseURL:     server.URL,
		ContentType: rest.FORM,
	}

	resp := fbForm.Post("/form/user", &User{Name: "John Doe"})

	if resp.Err == nil {
		t.Fatal("Error should not be nil")
	}
}

func TestPut(t *testing.T) {
	resp := rest.Put(server.URL+"/user/3", &User{Name: "Pichucha"})

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func TestPatch(t *testing.T) {
	resp := rest.Patch(server.URL+"/user/3", &User{Name: "Pichucha"})

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func TestDelete(t *testing.T) {
	resp := rest.Delete(server.URL + "/user/4")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func TestOptions(t *testing.T) {
	resp := rest.Options(server.URL + "/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func TestHeaders(t *testing.T) {
	h := make(http.Header)
	h.Add("X-Params-Test", "test")

	client := rest.Client{
		BaseURL: server.URL,
		DefaultHeaders: map[string][]string{
			"X-Default-Test": {"test"},
		},
	}

	r := client.Get("/header", h)

	if r.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestSetName(t *testing.T) {
	t.Setenv("HOSTNAME", "localhost")
	builder := rest.Client{
		BaseURL: server.URL,
	}

	h := make(http.Header)
	h.Add("X-Test", "test")

	r := builder.Get("/header", h)

	if r.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestWrongURL(t *testing.T) {
	r := rest.Get("foo")
	if r.Err == nil {
		t.Fatal("Wrong URL should get an error")
	}
}

/*Increase percentage of net.go coverage. */
func TestRequestWithProxyAndFollowRedirect(t *testing.T) {
	host := "saraza"

	client := new(rest.Client)
	client.ContentType = rest.JSON
	client.DisableTimeout = true
	client.CustomPool = &rest.CustomPool{
		MaxIdleConnsPerHost: 100,
		Proxy:               fmt.Sprintf("http://%s", host),
	}
	client.FollowRedirect = true

	response := client.Get(server.URL + "/user")
	expected := fmt.Sprintf("Get \"%s/user\": proxyconnect tcp: dial tcp: lookup %s", server.URL, host)

	if !strings.Contains(response.Err.Error(), expected) {
		t.Fatalf("Expected %v Error, Got %v as Response", expected, response.Err.Error())
	}
}

func TestRequestWithProxyAndFollowRedirect_Trace(t *testing.T) {
	host := "saraza"
	customPool := rest.CustomPool{
		MaxIdleConnsPerHost: 100,
		Proxy:               fmt.Sprintf("http://%s", host),
	}

	restClient := new(rest.Client)
	restClient.ContentType = rest.JSON
	restClient.DisableTimeout = true
	restClient.CustomPool = &customPool
	restClient.FollowRedirect = true
	restClient.EnableTrace = true

	response := restClient.Get(server.URL + "/user")
	expected := fmt.Sprintf("Get \"%s/user\": proxyconnect tcp: dial tcp: lookup %s", server.URL, host)

	if !strings.Contains(response.Err.Error(), expected) {
		t.Fatalf("Expected %v Error, Got %v as Response", expected, response.Err.Error())
	}
}

func TestRequestSendingClientMetrics(t *testing.T) {
	restClient := new(rest.Client)

	response := restClient.Get(server.URL + "/user")

	if response.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestResponseExceedsConnectTimeout(t *testing.T) {
	restClient := rest.Client{CustomPool: &rest.CustomPool{}}
	restClient.ConnectTimeout = 1 * time.Nanosecond
	restClient.Timeout = 35 * time.Millisecond
	restClient.ContentType = rest.JSON

	scuResponse := restClient.Get(server.URL + "/cache/slow/user")

	scuResponseErrIsTimeoutExceeded := func() bool {
		expected := "dial tcp"
		if scuResponse.Err != nil {
			return strings.Contains(scuResponse.Err.Error(), expected)
		}
		return false
	}

	if !scuResponseErrIsTimeoutExceeded() {
		t.Errorf("Timeouts configuration should get an error when connect")
	}
}

func TestResponseExceedsRequestTimeout(t *testing.T) {
	restClient := rest.Client{CustomPool: &rest.CustomPool{Transport: &http.Transport{}}}
	restClient.ConnectTimeout = 10 * time.Millisecond
	restClient.Timeout = 1 * time.Millisecond
	restClient.ContentType = rest.JSON

	response := restClient.Get(server.URL + "/slow/user")

	suResponseErrIsTimeoutExceeded := func() bool {
		expected := "timeout awaiting response headers"
		if response.Err != nil {
			return strings.Contains(response.Err.Error(), expected)
		}
		return false
	}

	if !suResponseErrIsTimeoutExceeded() {
		t.Fatalf("Timeouts configuration should get an error after connect")
	}

	require.Error(t, response.VerifyIsOkOrError())
}

func TestResponse_InvalidContentType(t *testing.T) {
	restClient := rest.Client{CustomPool: &rest.CustomPool{Transport: &http.Transport{}}}
	restClient.ConnectTimeout = 35 * time.Millisecond
	restClient.Timeout = 9 * time.Millisecond
	restClient.ContentType = 4

	resp := restClient.Post(server.URL+"/users", map[string]interface{}{})
	require.Error(t, resp.VerifyIsOkOrError())
}

func TestResponseExceedsRequestOAuth(t *testing.T) {
	restClient := rest.Client{
		CustomPool: &rest.CustomPool{Transport: &http.Transport{}},
		OAuth: &rest.OAuth{
			ClientID:     "a11d0149-687e-452e-9c94-783d489d4f72",
			ClientSecret: "client_secret",
			TokenURL:     server.URL + "/auth/token",
			AuthStyle:    rest.AuthStyleInHeader,
		},
	}
	restClient.ConnectTimeout = 1000 * time.Millisecond
	restClient.Timeout = 2000 * time.Millisecond
	restClient.ContentType = rest.JSON

	suResponse := restClient.Get(server.URL + "/auth")

	assert.NotNil(t, suResponse)
	require.NoError(t, suResponse.Err)
}

func TestClient_AsyncGet(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncGet(server.URL + "/user")
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestClient_AsyncHead(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncHead(server.URL + "/user")
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestClient_AsyncOptions(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncOptions(server.URL + "/user")
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestClient_AsyncDelete(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncDelete(server.URL + "/user")
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func Test_AsyncPost(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncPost(server.URL+"/user", &User{Name: "Maria"})
	resp := <-rChan
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("Status != OK (201)")
	}
}

func Test_AsyncPut(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncPut(server.URL+"/user/3", &User{Name: "Pichucha"})
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func Test_AsyncPatch(t *testing.T) {
	client := rest.Client{}
	rChan := client.AsyncPatch(server.URL+"/user/3", &User{Name: "Pichucha"})
	resp := <-rChan
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200")
	}
}

func Test_Basic(t *testing.T) {
	client := rest.Client{
		BasicAuth: &rest.BasicAuth{
			Username: "Juan",
			Password: "Doe",
		},
		UserAgent: "elber",
	}

	resp := client.Get(server.URL + "/")
	require.NoError(t, resp.Err)
}
