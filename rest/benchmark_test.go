package rest_test

import (
	"net/http"
	"testing"

	"github.com/arielsrv/go-restclient/rest"
)

func BenchmarkGet(b *testing.B) {
	client := &rest.Client{}

	for b.Loop() {
		resp := client.Get("https://gorest.co.in/public/v2/users")
		if resp.Err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	for b.Loop() {
		resp := rb.Get("/cache/user")

		if resp.StatusCode != http.StatusOK {
		}
	}
}

func BenchmarkSlowGet(b *testing.B) {
	for b.Loop() {
		resp := rb.Get("/slow/user")

		if resp.StatusCode != http.StatusOK {
		}
	}
}
