package rest

import (
	"github.com/dgraph-io/ristretto/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// ResourceCache, is an LRU-TTL Cache, that caches Responses base on headers
// It uses 3 goroutines -> one for LRU, and the other two for TTL.

// The cache itself.
var resourceCache *ristretto.Cache[string, *Response]

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

// MaxCacheSize is the Maximum Byte Size to be hold by the ResourceCache
// Default is 1 GigaByte
// Type: rest.ByteSize.
var MaxCacheSize = int64(1 * GB)

func init() {
	cache, err := ristretto.NewCache(&ristretto.Config[string, *Response]{
		NumCounters: 1e7,          // number of keys to track frequency of (10M).
		MaxCost:     MaxCacheSize, // maximum cost of cache (1GB).
		BufferItems: 64,           // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}

	metrics.Collector.Prometheus().RecordValueFunc("ristretto_cache_ratio", func() float64 {
		return cache.Metrics.Ratio()
	})

	metrics.Collector.Prometheus().RecordValueFunc("ristretto_cache_hits", func() float64 {
		return float64(cache.Metrics.Hits())
	})

	metrics.Collector.Prometheus().RecordValueFunc("ristretto_cache_misses", func() float64 {
		return float64(cache.Metrics.Misses())
	})

	resourceCache = cache
}
