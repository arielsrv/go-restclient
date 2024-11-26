package rest

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// Cache is an interface for cache implementations.
type Cache[K any, V any] interface {
	Set(key K, value V, cost int64) bool
	SetWithTTL(key K, value V, cost int64, ttl time.Duration) bool
	Get(key K) (V, bool)
}

// resourceTTLLfuMap, is an LRU-TTL Cache, that caches Responses base on headers
// The cache itself.
var resourceCache *resourceTTLLfuMap

type resourceTTLLfuMap struct {
	lowLevelCache Cache[string, *Response]
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

// MaxCacheSize is the Maximum Byte Size to be hold by the resourceTTLLfuMap
// Default is 1 GigaByte
// Type: rest.ByteSize.
var MaxCacheSize = 1 * GB

const (
	numCounters int64 = 1e7
	bufferItems int64 = 64
)

// init initializes the resourceTTLLfuMap with a Ristretto cache.
func init() {
	cache, _ := ristretto.NewCache(&ristretto.Config[string, *Response]{
		MaxCost:     int64(MaxCacheSize), // maximum cost of cache (1GB).
		NumCounters: numCounters,         // number of keys to track frequency of (10M).
		BufferItems: bufferItems,         // number of keys per Get buffer.
		Metrics:     true,
	})

	registerMetrics(cache)

	resourceCache = &resourceTTLLfuMap{
		lowLevelCache: cache,
	}
}

// get retrieves a Response from the cache, if it exists.
func (r *resourceTTLLfuMap) get(url string) (*Response, bool) {
	if value, hit := r.lowLevelCache.Get(url); hit {
		if value != nil {
			value.cached.Store(hit)
		}
		return value, hit
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
		r.lowLevelCache.SetWithTTL(url, response, cost, time.Until(*ttl))
		return
	}

	r.lowLevelCache.Set(url, response, cost)
}

// registerMetrics records the cache's metrics to Prometheus.
func registerMetrics(cache *ristretto.Cache[string, *Response]) {
	// config
	recordValue("num_counters", float64(numCounters))
	recordValue("max_cost_bytes", float64(cache.MaxCost()))
	recordValue("buffer_items", float64(bufferItems))

	// ratio
	recordValueFunc("ratio", cache.Metrics.Ratio)

	// counters
	incrementCounterFunc("hits_total", cache.Metrics.Hits)
	incrementCounterFunc("misses_total", cache.Metrics.Misses)
	incrementCounterFunc("keys_added_total", cache.Metrics.KeysAdded)
	incrementCounterFunc("keys_evicted_total", cache.Metrics.KeysEvicted)
	incrementCounterFunc("keys_updated_total", cache.Metrics.KeysUpdated)
	incrementCounterFunc("cost_added_bytes_total", cache.Metrics.CostAdded)
	incrementCounterFunc("cost_evicted_bytes_total", cache.Metrics.CostEvicted)
	incrementCounterFunc("gets_kept_total", cache.Metrics.GetsKept)
	incrementCounterFunc("gets_dropped_total", cache.Metrics.GetsDropped)
	incrementCounterFunc("sets_dropped_total", cache.Metrics.SetsDropped)
	incrementCounterFunc("sets_rejected_total", cache.Metrics.SetsRejected)
}

// buildMetricName constructs a Prometheus metric name.
func buildMetricName(suffix string) string {
	return fmt.Sprintf("__go_restclient_cache_%s", suffix)
}

// incrementCounterFunc increments a Prometheus counter.
func incrementCounterFunc(metricName string, metricFunc func() uint64) {
	metrics.Collector.Prometheus().IncrementCounterFunc(buildMetricName(metricName), func() float64 {
		return float64(metricFunc())
	})
}

// recordValue records a Prometheus gauge.
func recordValue(metricName string, value float64) {
	metrics.Collector.Prometheus().RecordValue(buildMetricName(metricName), value)
}

// recordValueFunc records a Prometheus gauge using a function.
func recordValueFunc(metricName string, metricFunc func() float64) {
	metrics.Collector.Prometheus().RecordValueFunc(buildMetricName(metricName), metricFunc)
}
