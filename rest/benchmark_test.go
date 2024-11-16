package rest_test

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-restclient/rest"
	"net/http"
	"strconv"
	"testing"
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

func BenchmarkResty_Get(b *testing.B) {
	client := resty.New()
	for i := 0; i < b.N; i++ {
		resp, err := client.R().Get(fmt.Sprintf("https://gorest.co.in/public/v2/users"))
		if err != nil {
			log.Info("f[" + strconv.Itoa(i) + "] Error")
			continue
		}
		if resp.StatusCode() != http.StatusOK {
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
