package rest

import (
	"time"
	"weak"

	"github.com/dgraph-io/ristretto/v2"
)

// Key is an interface for types that can be used as cache keys.
// It supports various primitive types that can be used as keys.
type Key interface {
	uint64 | string | []byte | byte | int | int32 | uint32 | int64
}

// Cache is an interface for cache implementations.
// It provides methods for getting, setting, and setting with TTL values.
type Cache[K Key, V any] interface {
	// Get retrieves a value from the cache by its key.
	// Returns the value and a boolean indicating whether the key was found.
	Get(key K) (V, bool)

	// Set adds a value to the cache with the specified key and cost.
	// The cost can be the size of the value in bytes or any other metric.
	// Returns true if the value was added successfully.
	Set(key K, value V, cost int64) bool

	// SetWithTTL adds a value to the cache with the specified key, cost, and time-to-live.
	// Returns true if the value was added successfully.
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
}

// resourceTTLLfuMap is an LRU-TTL Cache that caches Responses based on headers.
// It implements a least-frequently-used eviction policy with time-to-live expiration.
var resourceCache *resourceTTLLfuMap

// resourceTTLLfuMap is the internal implementation of the response cache.
// It uses a weak pointer cache to avoid memory leaks when responses are no longer needed.
type resourceTTLLfuMap struct {
	lowLevelCache Cache[string, weak.Pointer[Response]]
}

// ByteSize is a helper type for configuring cache sizes in bytes.
// It provides constants for common size units (KB, MB, GB).
type ByteSize int64

const (
	_ = iota

	// KB represents a kilobyte (1024 bytes).
	KB ByteSize = 1 << (10 * iota)

	// MB represents a megabyte (1024 kilobytes).
	MB

	// GB represents a gigabyte (1024 megabytes).
	GB
)

// MaxCacheSize is the maximum total size of the cache in bytes.
// Default is 256 MB.
var MaxCacheSize = 256 * MB

var (
	// NumCounters is the number of keys to track frequency of (100K).
	NumCounters = 1e5

	// BufferItems is the number of keys per Get buffer.
	BufferItems = 64
)

// init initializes the resourceTTLLfuMap with a Ristretto cache.
// It configures the cache with the specified MaxCacheSize, NumCounters, and BufferItems,
// and enables a metrics collection.
func init() {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, weak.Pointer[Response]]{
		MaxCost:     int64(MaxCacheSize), // maximum cost of cache (256Mb by default)
		NumCounters: int64(NumCounters),  // number of keys to track frequency of (100K)
		BufferItems: int64(BufferItems),  // number of keys per Get buffer
		Metrics:     true,                // enable metrics collection
	})

	registerMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		lowLevelCache: cache,
	}
}

// get retrieves a Response from the cache, if it exists.
// It returns the cached Response and a boolean indicating whether the key was found.
// If the weak pointer's value is nil (garbage collected), it returns false.
func (r *resourceTTLLfuMap) get(url string) (*Response, bool) {
	if weakPtr, hit := r.lowLevelCache.Get(url); hit && weakPtr.Value() != nil {
		return weakPtr.Value(), hit
	}

	return nil, false
}

// setNX sets a new value to the cache, if the key does not exist (like Redis SETNX).
// If the key already exists, it does nothing.
// If the response has a TTL, it sets the value with the TTL.
// Otherwise, it sets the value without a TTL.
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
