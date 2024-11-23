package rest

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// resourceTTLLfuMap, is an LRU-TTL Cache, that caches Responses base on headers
// The cache itself.
var resourceCache *resourceTTLLfuMap

type resourceTTLLfuMap struct {
	*ristretto.Cache[string, *Response]
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
	// Default is 1 GigaByte
	// Type: rest.ByteSize.
	MaxCacheSize = int64(1 * GB)
	NumCounters  = 1e7
	BufferItems  = 64
)

func init() {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, *Response]{
		NumCounters: int64(NumCounters), // number of keys to track frequency of (10M).
		MaxCost:     MaxCacheSize,       // maximum cost of cache (1GB).
		BufferItems: int64(BufferItems), // number of keys per Get buffer.
	})

	recordMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		Cache: cache,
	}
}

func (r *resourceTTLLfuMap) setNX(key string, value *Response) {
	if _, found := r.Get(key); !found {
		if value.ttl != nil {
			resourceCache.SetWithTTL(key, value, value.size(), time.Until(*value.ttl))
		} else {
			resourceCache.Set(key, value, value.size())
		}
		resourceCache.Wait()
	}
}

// recordMetrics records the cache's metrics to Prometheus.
func recordMetrics(cache *ristretto.Cache[string, *Response]) {
	metrics.Collector.Prometheus().RecordValue("go_restclient_cache_size_numcounters", NumCounters)
	metrics.Collector.Prometheus().RecordValue("go_restclient_cache_size_maxcost", float64(cache.MaxCost()))
	metrics.Collector.Prometheus().RecordValue("go_restclient_cache_size_bufferitems", float64(BufferItems))
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_ratio", func() float64 { return cache.Metrics.Ratio() })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_hits", func() float64 { return float64(cache.Metrics.Hits()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_misses", func() float64 { return float64(cache.Metrics.Misses()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_keys_added", func() float64 { return float64(cache.Metrics.KeysAdded()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_keys_evicted", func() float64 { return float64(cache.Metrics.KeysEvicted()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_keys_updated", func() float64 { return float64(cache.Metrics.KeysUpdated()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_cost_added", func() float64 { return float64(cache.Metrics.CostAdded()) })
	metrics.Collector.Prometheus().RecordValueFunc("go_restclient_cache_cost_evicted", func() float64 { return float64(cache.Metrics.CostEvicted()) })
}
