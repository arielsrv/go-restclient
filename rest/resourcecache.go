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
		lowLevelCache: cache,
	}
}

func (r *resourceTTLLfuMap) get(key string) (*Response, bool) {
	now := time.Now()
	defer metrics.Collector.Prometheus().RecordExecutionTime(fmt.Sprintf(prefix, "get_execution_time_seconds"), time.Since(now))

	if value, hit := r.lowLevelCache.Get(key); hit {
		return value, true
	}

	return nil, false
}

// setNX sets a new value to the cache, if the key does not exist (like Redis SETNX).
func (r *resourceTTLLfuMap) setNX(key string, value *Response) {
	if _, hit := r.lowLevelCache.Get(key); hit {
		return
	}

	now := time.Now()
	defer metrics.Collector.Prometheus().RecordExecutionTime(fmt.Sprintf(prefix, "set_execution_time_seconds"), time.Since(now))

	cost := value.size()
	if value.ttl != nil {
		r.lowLevelCache.SetWithTTL(key, value, cost, time.Until(*value.ttl))
		return
	}

	r.lowLevelCache.Set(key, value, cost)
}

// recordMetrics records the cache's metrics to Prometheus.
func recordMetrics(cache *ristretto.Cache[string, *Response]) {
	// config
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "num_counters"), NumCounters)
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "max_cost_bytes"), float64(cache.MaxCost()))
	metrics.Collector.Prometheus().RecordValue(fmt.Sprintf(prefix, "buffer_items"), float64(BufferItems))

	// metrics
	metrics.Collector.Prometheus().RecordValueFunc(fmt.Sprintf(prefix, "ratio"), func() float64 { return cache.Metrics.Ratio() })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "hits_total"), func() float64 { return float64(cache.Metrics.Hits()) })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "misses_total"), func() float64 { return float64(cache.Metrics.Misses()) })

	// cache metrics
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "keys_added_total"), func() float64 { return float64(cache.Metrics.KeysAdded()) })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "keys_evicted_total"), func() float64 { return float64(cache.Metrics.KeysEvicted()) })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "keys_updated_total"), func() float64 { return float64(cache.Metrics.KeysUpdated()) })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "cost_added_bytes_total"), func() float64 { return float64(cache.Metrics.CostAdded()) })
	metrics.Collector.Prometheus().IncrementCounterFunc(fmt.Sprintf(prefix, "cost_evicted_bytes_total"), func() float64 { return float64(cache.Metrics.CostEvicted()) })
}
