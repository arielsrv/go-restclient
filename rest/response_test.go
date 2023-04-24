package rest_test

import (
	"gitlab.com/iskaypetcom/digital/tools/dev/go-restclient/rest"
	"net/http"
	"strings"
	"testing"
)

func TestResponseBytesAndString(t *testing.T) {
	resp := rest.Get(server.URL + "/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	if string(resp.Bytes()) != resp.String() {
		t.Fatal("Bytes() and String() are not equal")
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

func TestGetTypedFillUpJSON(t *testing.T) {
	resp := rb.Get("/user")

	if resp.StatusCode != http.StatusOK {
		t.Fatal("Status != OK (200)")
	}

	result, err := rest.TypedFillUp[[]User](resp)
	if err != nil {
		t.Fatal("Json fill up failed. Error: " + err.Error())
	}

	for _, v := range *result {
		if v.Name == "Alice" {
			return
		}
	}

	t.Fatal("Couldn't found Alice")
}

func TestGetFillUpXML(t *testing.T) {
	var u []User

	var rbXML = rest.RequestBuilder{
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
