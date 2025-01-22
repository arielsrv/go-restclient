package rest

import (
	"fmt"

	"github.com/dgraph-io/ristretto/v2"
	"gitlab.com/iskaypetcom/digital/sre/tools/dev/go-metrics-collector/metrics"
)

// CounterFunc increments a counter.
type CounterFunc func() uint64

// RecordValueFunc records a value to Prometheus.
type RecordValueFunc func() float64

// registerMetrics records the cache's metrics to Prometheus.
func registerMetrics(cache *ristretto.Cache[string, *Response]) {
	// config
	recordValue("num_counters", NumCounters)
	recordValue("max_cost_bytes", float64(cache.MaxCost()))
	recordValue("buffer_items", float64(BufferItems))

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
func incrementCounterFunc(metricName string, counterFunc CounterFunc) {
	metrics.Collector.Prometheus().IncrementCounterFunc(buildMetricName(metricName), func() float64 {
		return float64(counterFunc())
	})
}

// recordValue records a Prometheus gauge.
func recordValue(metricName string, value float64) {
	metrics.Collector.Prometheus().RecordValue(buildMetricName(metricName), value)
}

// recordValueFunc records a Prometheus gauge using a function.
func recordValueFunc(metricName string, recordValueFunc RecordValueFunc) {
	metrics.Collector.Prometheus().RecordValueFunc(buildMetricName(metricName), metrics.RecordValueFunc(recordValueFunc))
}
