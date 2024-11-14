package rest_test

import (
	"net/http"
	"strconv"
	"testing"

	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger/log"
)

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := rb.Get("/user")

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
