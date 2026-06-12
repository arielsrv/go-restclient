package rest

import (
	"weak"

	"github.com/dgraph-io/ristretto/v2"
)

// registerMetrics is a placeholder for cache metrics registration.
// It is intentionally empty: metrics integration is opt-in and will be
// wired by consumers (e.g. Prometheus collectors) via the cache instance.
func registerMetrics(_ *ristretto.Cache[string, weak.Pointer[Response]]) {
	// Intentionally left blank — see doc comment above.
}
