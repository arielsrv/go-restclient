package rest

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// resourceTTLLfuMap, is an LRU-TTL Cache, that caches Responses base on headers
// The cache itself.
var (
	resourceCache *resourceTTLLfuMap
	prefix        = "go_restclient_cache_%s"
)

type resourceTTLLfuMap struct {
	lowLevelCache *ristretto.Cache[string, *Response]
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
	MaxCacheSize       = int64(1 * GB)
	NumCounters  int64 = 1e7
	BufferItems  int64 = 64
)

func init() {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, *Response]{
		NumCounters: NumCounters,  // number of keys to track frequency of (10M).
		MaxCost:     MaxCacheSize, // maximum cost of cache (1GB).
		BufferItems: BufferItems,  // number of keys per Get buffer.
		Metrics:     true,
	})

	setupMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		lowLevelCache: cache,
	}
}

func (r *resourceTTLLfuMap) get(url string) (*Response, bool) {
	if value, hit := r.lowLevelCache.Get(url); hit {
		return value, true
	}

	return nil, false
}

// setNX sets a new value to the cache, if the key does not exist (like Redis SETNX).
func (r *resourceTTLLfuMap) setNX(url string, response *Response) {
	if _, hit := r.lowLevelCache.Get(url); hit {
		return
	}

	cost := response.size()
	if ttl := response.ttl; ttl != nil {
		r.lowLevelCache.SetWithTTL(url, response, cost, time.Until(*ttl))
		return
	}

	r.lowLevelCache.Set(url, response, cost)
}

// setupMetrics records the cache's metrics to Prometheus.
func setupMetrics(cache *ristretto.Cache[string, *Response]) {
	buildName := func(key string) string {
		return fmt.Sprintf(prefix, key)
	}

	// config
	metrics.Collector.Prometheus().RecordValue(buildName("num_counters"), float64(NumCounters))
	metrics.Collector.Prometheus().RecordValue(buildName("max_cost_bytes"), float64(cache.MaxCost()))
	metrics.Collector.Prometheus().RecordValue(buildName("buffer_items"), float64(BufferItems))

	// ratio
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "ratio"), cache.Metrics.Ratio)

	incrementCounter := func(part string, metricFunc func() uint64) {
		metrics.Collector.Prometheus().IncrementCounterFunc(buildName(part), func() float64 {
			return float64(metricFunc())
		})
	}

	// counters
	incrementCounter("hits_total", cache.Metrics.Hits)
	incrementCounter("misses_total", cache.Metrics.Misses)
	incrementCounter("keys_added_total", cache.Metrics.KeysAdded)
	incrementCounter("keys_evicted_total", cache.Metrics.KeysEvicted)
	incrementCounter("keys_updated_total", cache.Metrics.KeysUpdated)
	incrementCounter("cost_added_bytes_total", cache.Metrics.CostAdded)
	incrementCounter("cost_evicted_bytes_total", cache.Metrics.CostEvicted)
	incrementCounter("gets_kept_total", cache.Metrics.GetsKept)
	incrementCounter("gets_dropped_total", cache.Metrics.GetsDropped)
	incrementCounter("sets_dropped_total", cache.Metrics.SetsDropped)
	incrementCounter("sets_rejected_total", cache.Metrics.SetsRejected)
}
