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
	cache, err := ristretto.NewCache(&ristretto.Config[string, *Response]{
		NumCounters: int64(NumCounters), // number of keys to track frequency of (10M).
		MaxCost:     MaxCacheSize,       // maximum cost of cache (1GB).
		BufferItems: int64(BufferItems), // number of keys per Get buffer.
		Metrics:     true,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create go-restclient cache: %w", err))
	}

	recordMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		Cache: cache,
	}
}

// setNX sets a new value to the cache, if the key does not exist (like Redis SETNX).
func (r *resourceTTLLfuMap) setNX(key string, value *Response) {
	if _, hit := r.Get(key); hit {
		return
	}

	cost := value.size()
	if value.ttl != nil {
		resourceCache.SetWithTTL(key, value, cost, time.Until(*value.ttl))
		return
	}

	resourceCache.Set(key, value, cost)
}

// recordMetrics records the cache's metrics to Prometheus.
func recordMetrics(cache *ristretto.Cache[string, *Response]) {
	// config
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "num_counters"), NumCounters)
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "max_cost_bytes"), float64(cache.MaxCost()))
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "buffer_items"), float64(BufferItems))

	// metrics
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "ratio"), func() float64 { return cache.Metrics.Ratio() })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "hits"), func() float64 { return float64(cache.Metrics.Hits()) })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "misses"), func() float64 { return float64(cache.Metrics.Misses()) })

	// cache metrics
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "keys_added"), func() float64 { return float64(cache.Metrics.KeysAdded()) })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "keys_evicted"), func() float64 { return float64(cache.Metrics.KeysEvicted()) })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "keys_updated"), func() float64 { return float64(cache.Metrics.KeysUpdated()) })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "cost_added_bytes"), func() float64 { return float64(cache.Metrics.CostAdded()) })
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "cost_evicted_bytes"), func() float64 { return float64(cache.Metrics.CostEvicted()) })
}
