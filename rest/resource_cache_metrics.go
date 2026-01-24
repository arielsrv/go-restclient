package rest

import (
	"weak"

	"github.com/dgraph-io/ristretto/v2"
)

// registerMetrics is a placeholder for cache metrics registration.
func registerMetrics(cache *ristretto.Cache[string, weak.Pointer[Response]]) {
}
