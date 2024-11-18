package rest_test

import (
	"net/http"
	"strconv"
	"testing"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
)

func BenchmarkGet(b *testing.B) {
	client := &rest.Client{}

	for i := 0; i < b.N; i++ {
		resp := client.Get("https://gorest.co.in/public/v2/users")
		if resp.Err != nil {
			log.Info("f[" + strconv.Itoa(i) + "] Error")
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Info("f[" + strconv.Itoa(i) + "] Status != OK (200)")
		}
	}
}

func BenchmarkCacheGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := rb.Get("/cache/user")

		if resp.StatusCode != http.StatusOK {
			log.Info("f[" + strconv.Itoa(i) + "] Status != OK (200)")
		}
	}
}

func BenchmarkSlowGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := rb.Get("/slow/user")

		if resp.StatusCode != http.StatusOK {
			log.Info("f[" + strconv.Itoa(i) + "] Status != OK (200)")
		}
	}
}
