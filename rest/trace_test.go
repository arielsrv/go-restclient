package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	"github.com/arielsrv/go-restclient/rest"
)

// newTracerHarness installs a fresh in-memory tracer provider and W3C propagator
// for the duration of a single test, returning the recorder used to inspect spans
// and a cleanup func that restores the previous global state.
func newTracerHarness(t *testing.T) (*tracetest.SpanRecorder, func()) {
	t.Helper()

	rec := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec))

	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return rec, func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	}
}

// newTracedServer spins up an httptest server invoking handler and returns
// a rest.Client wired to it with EnableTrace=enableTrace.
func newTracedServer(t *testing.T, enableTrace bool, handler http.HandlerFunc) (*rest.Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := &rest.Client{
		BaseURL:     srv.URL,
		ContentType: rest.JSON,
		EnableTrace: enableTrace,
	}
	return client, srv.Close
}

// findSpanByName returns the first span whose Name equals name, or nil.
func findSpanByName(spans []sdktrace.ReadOnlySpan, name string) sdktrace.ReadOnlySpan {
	for _, s := range spans {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

// statusOKHandler is a tiny [http.HandlerFunc] that always responds 200.
func statusOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// TestEnableTrace_EmitsHTTPSpan verifies that with EnableTrace=true otelhttp wraps
// the transport and emits a span for every outbound request.
func TestEnableTrace_EmitsHTTPSpan(t *testing.T) {
	rec, cleanup := newTracerHarness(t)
	defer cleanup()

	client, closeSrv := newTracedServer(t, true, statusOKHandler)
	defer closeSrv()

	resp := client.GetWithContext(context.Background(), "/ping")
	require.NoError(t, resp.Err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	spans := rec.Ended()
	require.NotEmpty(t, spans, "expected at least one span when EnableTrace=true")

	// The otelhttp transport span should exist and carry the configured
	// span-name formatter output ("METHOD /path").
	httpSpan := findSpanByName(spans, "GET /ping")
	require.NotNil(t, httpSpan, "expected an HTTP span named with METHOD + path, got: %v", spanNames(spans))
	assert.Equal(t, "GET /ping", httpSpan.Name())
}

// TestEnableTrace_HttptraceSubSpansAreChildren verifies the fix: httptrace sub-spans
// (DNS, connect, ...) must be children of the otelhttp HTTP span, not orphans.
func TestEnableTrace_HttptraceSubSpansAreChildren(t *testing.T) {
	rec, cleanup := newTracerHarness(t)
	defer cleanup()

	client, closeSrv := newTracedServer(t, true, statusOKHandler)
	defer closeSrv()

	resp := client.GetWithContext(context.Background(), "/x")
	require.NoError(t, resp.Err)

	spans := rec.Ended()
	require.NotEmpty(t, spans)

	httpSpan := findSpanByName(spans, "GET /x")
	require.NotNil(t, httpSpan, "missing HTTP span; got %v", spanNames(spans))

	httpSpanID := httpSpan.SpanContext().SpanID()
	traceID := httpSpan.SpanContext().TraceID()

	// At least one of the other recorded spans must be an httptrace sub-span
	// (its name comes from otelhttptrace, e.g. "http.getconn", "http.dns",
	// "http.connect", "http.receive", ...). The bug we fixed caused those
	// spans to have a zero parent and a different trace id.
	var subSpanFound bool
	for _, s := range spans {
		if s.SpanContext().SpanID() == httpSpanID {
			continue
		}
		if !strings.HasPrefix(s.Name(), "http.") {
			continue
		}
		subSpanFound = true
		assert.Equal(t, traceID, s.SpanContext().TraceID(),
			"httptrace sub-span %q must share the trace id of the HTTP span", s.Name())
		assert.True(t, s.Parent().IsValid(),
			"httptrace sub-span %q must have a valid parent (regression: was orphaned)", s.Name())
	}
	assert.True(t, subSpanFound,
		"expected at least one httptrace sub-span (http.*) as a child of the HTTP span; got %v",
		spanNames(spans))
}

// TestEnableTrace_PropagatesParentContext verifies that an outer (caller) span is
// propagated correctly: the HTTP span must be a child of the outer span and the
// W3C traceparent header must be injected into the outgoing request.
func TestEnableTrace_PropagatesParentContext(t *testing.T) {
	rec, cleanup := newTracerHarness(t)
	defer cleanup()

	var gotTraceparent string
	client, closeSrv := newTracedServer(t, true, func(w http.ResponseWriter, r *http.Request) {
		gotTraceparent = r.Header.Get("Traceparent")
		w.WriteHeader(http.StatusOK)
	})
	defer closeSrv()

	tracer := otel.Tracer("test")
	ctx, parentSpan := tracer.Start(context.Background(), "caller")
	resp := client.GetWithContext(ctx, "/p")
	parentSpan.End()

	require.NoError(t, resp.Err)

	// The outgoing request must carry a W3C traceparent so downstream services
	// can continue the trace.
	require.NotEmpty(t, gotTraceparent, "traceparent header must be propagated")

	spans := rec.Ended()
	httpSpan := findSpanByName(spans, "GET /p")
	require.NotNil(t, httpSpan, "missing HTTP span; got %v", spanNames(spans))

	assert.Equal(t, parentSpan.SpanContext().TraceID(), httpSpan.SpanContext().TraceID(),
		"HTTP span must share trace id with the caller span")
	assert.Equal(t, parentSpan.SpanContext().SpanID(), httpSpan.Parent().SpanID(),
		"HTTP span must be a direct child of the caller span")
}

// TestDisableTrace_NoSpansEmitted ensures the instrumentation is fully bypassed
// when EnableTrace is false (no accidental performance/leak hit on the hot path).
func TestDisableTrace_NoSpansEmitted(t *testing.T) {
	rec, cleanup := newTracerHarness(t)
	defer cleanup()

	client, closeSrv := newTracedServer(t, false, func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Traceparent"),
			"traceparent must not be injected when EnableTrace=false")
		w.WriteHeader(http.StatusOK)
	})
	defer closeSrv()

	resp := client.Get("/nope")
	require.NoError(t, resp.Err)

	assert.Empty(t, rec.Ended(), "no spans should be recorded when EnableTrace=false")
}

// TestEnableTrace_SpanRecordsHTTPError ensures non-2xx responses are surfaced
// on the span (status / attributes) by otelhttp.
func TestEnableTrace_SpanRecordsHTTPError(t *testing.T) {
	rec, cleanup := newTracerHarness(t)
	defer cleanup()

	client, closeSrv := newTracedServer(t, true, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer closeSrv()

	resp := client.GetWithContext(context.Background(), "/boom")
	require.NoError(t, resp.Err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	httpSpan := findSpanByName(rec.Ended(), "GET /boom")
	require.NotNil(t, httpSpan)

	// otelhttp records http.response.status_code (semconv) as an int attribute.
	var sawStatus bool
	for _, attr := range httpSpan.Attributes() {
		key := string(attr.Key)
		if key == "http.response.status_code" || key == "http.status_code" {
			sawStatus = true
			assert.EqualValues(t, http.StatusInternalServerError, attr.Value.AsInt64())
		}
	}
	assert.True(t, sawStatus, "HTTP span must record the response status code")
}

func spanNames(spans []sdktrace.ReadOnlySpan) []string {
	out := make([]string, 0, len(spans))
	for _, s := range spans {
		out = append(out, s.Name())
	}
	return out
}

// Compile-time assertion that the SpanContext type used above matches the one
// from go.opentelemetry.io/otel/trace.
var _ trace.SpanContext = (trace.SpanContext{})
