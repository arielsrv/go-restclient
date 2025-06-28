package rest_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/require"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func TestGet_Raw(t *testing.T) {
	resp := rest.Get(server.URL + "/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	if !strings.EqualFold(resp.Raw(), resp.String()) {
		t.Fatal("Debug() failed!")
	}
}

func TestDebug(t *testing.T) {
	resp := rest.Get(server.URL + "/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	if !strings.Contains(resp.Debug(), resp.String()) {
		t.Fatal("Debug() failed!")
	}
}

func TestGetFillUpJSON(t *testing.T) {
	var u []User

	resp := rb.Get("/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	err := resp.FillUp(&u)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	for _, v := range users {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestGetFillUpJSON_IsOk(t *testing.T) {
	var u []User

	resp := rb.Get("/user")

	if !resp.IsOk() {
		t.Fatal("Status != OK (200)")
	}

	err := resp.FillUp(&u)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	err = resp.VerifyIsOkOrError()
	if err != nil {
		t.Fatal("Error in VerifyIsOkOrError: " + err.Error())
	}

	for _, v := range users {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestGetTypedFillUpJSON(t *testing.T) {
	resp := rb.Get("/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	result, err := rest.Deserialize[[]User](resp)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	for _, v := range result {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestGetTypedGenericUnmarshalJSON(t *testing.T) {
	resp := rb.Get("/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	result, err := rest.Deserialize[[]User](resp)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	for _, v := range result {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestGetFillUpXML(t *testing.T) {
	var u []User

	rbXML := rest.Client{
		BaseURL:     server.URL,
		ContentType: rest.XML,
	}

	resp := rbXML.Get("/xml/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	err := resp.FillUp(&u)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	for _, v := range users {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestResponse_Unmarshal_Error(t *testing.T) {
	response := &rest.Response{
		Response: &http.Response{
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		},
	}

	var user User
	require.Error(t, response.FillUp(user))
	require.Error(t, response.VerifyIsOkOrError())
	require.False(t, response.IsOk())
}

func TestResponse_Unmarshal_Error_Utf(t *testing.T) {
	response := &rest.Response{
		Response: &http.Response{
			Header: map[string][]string{
				"Content-Type": {"application/json; utf-8"},
			},
		},
	}

	var user User
	require.Error(t, response.FillUp(user))
	require.Error(t, response.VerifyIsOkOrError())
	require.False(t, response.IsOk())
}

func TestResponse_Unmarshal_Error_Type(t *testing.T) {
	response := &rest.Response{
		Response: &http.Response{
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		},
	}

	var user User
	require.Error(t, response.FillUp(&user))
	user, err := rest.Deserialize[User](response)
	require.Error(t, err)
}

func TestResponse_Unmarshal_Nil(t *testing.T) {
	user, err := rest.Deserialize[*User](nil)
	require.Error(t, err)
	require.Nil(t, user)
}

func TestResponse_Unmarshal_Nil_List(t *testing.T) {
	user, err := rest.Deserialize[[]User](nil)
	require.Error(t, err)
	require.Nil(t, user)
}

func TestFillUp_Err(t *testing.T) {
	var user User
	resp := new(rest.Response)
	resp.Response = &http.Response{}
	resp.Header = map[string][]string{"Content-Type": {"application/invalid"}}
	err := resp.FillUp(&user)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported content type: application/invalid")
}

func TestFillUp_TypedErr(t *testing.T) {
	resp := new(rest.Response)
	resp.Response = &http.Response{}
	resp.Header = map[string][]string{"Content-Type": {"application/invalid"}}
	user, err := rest.Deserialize[string](resp)
	require.Error(t, err)
	require.Empty(t, user)
	require.Contains(t, err.Error(), "unsupported content type: application/invalid")
}

func TestFillUp_Detection(t *testing.T) {
	resp := new(rest.Response)
	resp.Response = &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"text": "plain"}`)),
	}
	user, err := rest.Deserialize[string](resp)
	require.Error(t, err)
	require.Empty(t, user)
	require.Contains(t, err.Error(), "unsupported content type: text/plain")
}

func TestClient_GetWithContext_ConcurrentResponses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer srv.Close()

	client := &rest.Client{
		BaseURL:        srv.URL,
		Timeout:        time.Duration(100) * time.Millisecond,
		ConnectTimeout: time.Duration(100) * time.Millisecond,
		EnableTrace:    true,
		CustomPool: &rest.CustomPool{
			Transport: &http.Transport{
				MaxIdleConns:        2,
				MaxConnsPerHost:     2,
				MaxIdleConnsPerHost: 2,
			},
		},
	}

	const n = 1000

	errs := make(chan error, n)

	group, ctx := errgroup.WithContext(t.Context())
	for range n {
		group.Go(func() error {
			resp := client.GetWithContext(ctx, "/", http.Header{
				"X-Test": {"123"},
			})
			if resp.Err != nil {
				errs <- fmt.Errorf("request error: %w", resp.Err)
				return resp.Err
			}
			if resp.String() != `{"ok":true}` {
				err := fmt.Errorf("unexpected response: %q", resp.String())
				errs <- err
				return err
			}
			return nil
		})
	}
	require.NoError(t, group.Wait())
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}
