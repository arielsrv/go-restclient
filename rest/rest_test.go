package rest_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func TestGet(t *testing.T) {
	resp := rest.Get(server.URL + "/user")
	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestSlowGet(t *testing.T) {
	var f [100]*rest.Response

	for i := range f {
		f[i] = rb.Get("/slow/user")

		if f[i].Response.StatusCode != http.StatusOK {
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

func TestPostXML(t *testing.T) {
	rbXML := rest.RequestBuilder{
		BaseURL:     server.URL,
		ContentType: rest.XML,
	}

	resp := rbXML.Post("/xml/user", &User{Name: "Maria"})

	if resp.StatusCode != http.StatusCreated {
		t.Fatal("Status != OK (201)")
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

func TestAsyncGet(t *testing.T) {
	rest.AsyncGet(server.URL+"/user", func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncHead(t *testing.T) {
	rest.AsyncHead(server.URL+"/user", func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncPost(t *testing.T) {
	rest.AsyncPost(server.URL+"/user", &User{Name: "Maria"}, func(r *rest.Response) {
		if r.StatusCode != http.StatusCreated {
			t.Fatal("Status != OK (201)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncPut(t *testing.T) {
	rest.AsyncPut(server.URL+"/user/3", &User{Name: "Pichucha"}, func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncPatch(t *testing.T) {
	rest.AsyncPatch(server.URL+"/user/3", &User{Name: "Pichucha"}, func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncDelete(t *testing.T) {
	rest.AsyncDelete(server.URL+"/user/4", func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestAsyncOptions(t *testing.T) {
	rest.AsyncOptions(server.URL+"/user", func(r *rest.Response) {
		if r.StatusCode != http.StatusOK {
			t.Fatal("Status != OK (200)")
		}
	})

	time.Sleep(50 * time.Millisecond)
}

func TestHeaders(t *testing.T) {
	h := make(http.Header)
	h.Add("X-Test", "test")

	builder := rest.RequestBuilder{
		BaseURL: server.URL,
		Headers: h,
	}

	r := builder.Get("/header")

	if r.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestSetHeader(t *testing.T) {
	builder := rest.RequestBuilder{
		BaseURL: server.URL,
	}

	builder.Headers = make(http.Header)
	builder.SetHeader("X-Test", "test")

	r := builder.Get("/header")

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
	customPool := rest.CustomPool{
		MaxIdleConnsPerHost: 100,
		Proxy:               fmt.Sprintf("http://%s", host),
	}

	restClient := new(rest.RequestBuilder)
	restClient.ContentType = rest.JSON
	restClient.DisableTimeout = true
	restClient.CustomPool = &customPool
	restClient.FollowRedirect = true

	response := restClient.Get(server.URL + "/user")
	expected := fmt.Sprintf("Get \"%s/user\": proxyconnect tcp: dial tcp: lookup %s", server.URL, host)

	if !strings.Contains(response.Err.Error(), expected) {
		t.Fatalf("Expected %v Error, Got %v as Response", expected, response.Err.Error())
	}
}

func TestRequestSendingClientMetrics(t *testing.T) {
	restClient := new(rest.RequestBuilder)

	response := restClient.Get(server.URL + "/user")

	if response.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}
}

func TestResponseExceedsConnectTimeout(t *testing.T) {
	restClient := rest.RequestBuilder{CustomPool: &rest.CustomPool{}}
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
	restClient := rest.RequestBuilder{CustomPool: &rest.CustomPool{Transport: &http.Transport{}}}
	restClient.ConnectTimeout = 35 * time.Millisecond
	restClient.Timeout = 9 * time.Millisecond
	restClient.ContentType = rest.JSON

	suResponse := restClient.Get(server.URL + "/slow/user")

	suResponseErrIsTimeoutExceeded := func() bool {
		expected := "timeout awaiting response headers"
		if suResponse.Err != nil {
			return strings.Contains(suResponse.Err.Error(), expected)
		}
		return false
	}

	if !suResponseErrIsTimeoutExceeded() {
		t.Fatalf("Timeouts configuration should get an error after connect")
	}
}

func TestResponseExceedsRequestOAuth(t *testing.T) {
	restClient := rest.RequestBuilder{
		CustomPool: &rest.CustomPool{Transport: &http.Transport{}},
		OAuth: &clientcredentials.Config{
			ClientID:     "a11d0149-687e-452e-9c94-783d489d4f72",
			ClientSecret: "client_secret",
			TokenURL:     server.URL + "/auth/token",
			AuthStyle:    oauth2.AuthStyleInHeader,
		},
	}
	restClient.ConnectTimeout = 1000 * time.Millisecond
	restClient.Timeout = 2000 * time.Millisecond
	restClient.ContentType = rest.JSON

	suResponse := restClient.Get(server.URL + "/auth")

	assert.NotNil(t, suResponse)
	assert.NoError(t, suResponse.Err)
}
