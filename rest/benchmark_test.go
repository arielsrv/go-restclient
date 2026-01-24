package rest_test

import (
	"net/http"
	"testing"

	"github.com/arielsrv/go-restclient/rest"
)

func BenchmarkGet(b *testing.B) {
	client := &rest.Client{}

	for range b.N {
		resp := client.Get("https://gorest.co.in/public/v2/users")
		if resp.Err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	for range b.N {
		resp := rb.Get("/cache/user")

		if resp.StatusCode != http.StatusOK {
		}
	}
}

func BenchmarkSlowGet(b *testing.B) {
	for range b.N {
		resp := rb.Get("/slow/user")

		if resp.StatusCode != http.StatusOK {
		}
	}
}
