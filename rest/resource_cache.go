package rest

import (
	"time"
	"weak"

	"github.com/dgraph-io/ristretto/v2"
)

type Key interface {
	uint64 | string | []byte | byte | int | int32 | uint32 | int64
}

// Cache is an interface for cache implementations.
type Cache[K Key, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V, cost int64) bool
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
}

// resourceTTLLfuMap, is an LRU-TTL Cache, that caches Responses base on headers
// The cache itself.
var resourceCache *resourceTTLLfuMap

type resourceTTLLfuMap struct {
	lowLevelCache Cache[string, weak.Pointer[Response]]
}

// ByteSize is a helper for configuring MaxCacheSize.
type ByteSize int64

const (
	_ = iota

	// KB = KiloBytes.
	KB ByteSize = 1 << (10 * iota)

	// MB = MegaBytes.
	MB

	// GB = GigaBytes.
	GB
)

var (
	// MaxCacheSize is the Maximum Byte Size to be hold by the resourceTTLLfuMap
	// Default is 128Mb
	// Type: rest.ByteSize.
	MaxCacheSize = 1 * GB
	NumCounters  = 1e5
	BufferItems  = 64
)

// init initializes the resourceTTLLfuMap with a Ristretto cache.
func init() {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, weak.Pointer[Response]]{
		MaxCost:     int64(MaxCacheSize), // maximum cost of cache (128Mb).
		NumCounters: int64(NumCounters),  // number of keys to track frequency of (100K).
		BufferItems: int64(BufferItems),  // number of keys per Get buffer.
		Metrics:     true,
	})

	registerMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		lowLevelCache: cache,
	}
}

// get retrieves a Response from the cache, if it exists.
func (r *resourceTTLLfuMap) get(url string) (*Response, bool) {
	if weakPtr, hit := r.lowLevelCache.Get(url); hit && weakPtr.Value() != nil {
		return weakPtr.Value(), hit
	}

	return nil, false
}

// setNX sets a new value to the cache, if the key does not exist (like Redis SETNX).
func (r *resourceTTLLfuMap) setNX(url string, response *Response) {
	if _, hit := r.get(url); hit {
		return
	}
	cost := response.size()
	if ttl := response.ttl; ttl != nil {
		r.lowLevelCache.SetWithTTL(url, weak.Make(response), cost, time.Until(*ttl))
		return
	}
	r.lowLevelCache.Set(url, weak.Make(response), cost)
}
