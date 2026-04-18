package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/arielsrv/go-restclient/rest"
)

// This example demonstrates ETag-based caching with go-restclient.
//
// Scenario A – 304 Not Modified (content unchanged):
//
//	Request 1 → server returns 200 + ETag "v1" → stored in cache (revalidate=true)
//	Request 2 → client sends If-None-Match: "v1" → server returns 304
//	           → client returns cached response, Cached()=true
//
// Scenario B – 200 after revalidation (content changed):
//
//	Request 1 → server returns 200 + ETag "v1" → stored in cache
//	Request 2 → client sends If-None-Match: "v1" → server returns 200 + ETag "v2"
//	           → cache is force-updated with new response (fix for setNX bug)
//	Request 3 → client sends If-None-Match: "v2" → server returns 304
//	           → client returns updated cached response, Cached()=true
func main() {
	rest.StartMockupServer()
	defer rest.StopMockupServer()

	ctx := context.Background()

	// ── Scenario A: content never changes ──────────────────────────────────────
	fmt.Println("=== Scenario A: content unchanged (304 revalidation) ===")

	const urlA = "http://example.com/resource/a"

	must(rest.AddMockups(&rest.Mock{
		URL:          urlA,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"version":1}`,
		RespHeaders:  http.Header{"Etag": {"v1"}, "Content-Type": {"application/json"}},
	}))

	clientA := &rest.Client{EnableCache: true, Timeout: 5 * time.Second}

	a1 := clientA.GetWithContext(ctx, urlA)
	checkResp(a1)
	fmt.Printf("  Request 1 → status=%d  cached=%t  etag=%q  body=%s\n",
		a1.StatusCode, a1.Cached(), a1.Header.Get(rest.ETagHeader), a1.String())

	rest.FlushMockups()
	must(rest.AddMockups(&rest.Mock{
		URL:          urlA,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusNotModified,
	}))

	a2 := clientA.GetWithContext(ctx, urlA)
	checkResp(a2)
	fmt.Printf("  Request 2 → status=%d  cached=%t  etag=%q  body=%s\n",
		a2.StatusCode, a2.Cached(), a2.Header.Get(rest.ETagHeader), a2.String())

	// ── Scenario B: content changes on second request ──────────────────────────
	fmt.Println("=== Scenario B: content changed (cache must be updated) ===")

	rest.FlushMockups()

	const urlB = "http://example.com/resource/b"

	must(rest.AddMockups(&rest.Mock{
		URL:          urlB,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"version":1}`,
		RespHeaders:  http.Header{"Etag": {"v1"}, "Content-Type": {"application/json"}},
	}))

	clientB := &rest.Client{EnableCache: true, Timeout: 5 * time.Second}

	b1 := clientB.GetWithContext(ctx, urlB)
	checkResp(b1)
	fmt.Printf("  Request 1 → status=%d  cached=%t  etag=%q  body=%s\n",
		b1.StatusCode, b1.Cached(), b1.Header.Get(rest.ETagHeader), b1.String())

	// Simulate content change: server returns 200 with a new ETag
	rest.FlushMockups()
	must(rest.AddMockups(&rest.Mock{
		URL:          urlB,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusOK,
		RespBody:     `{"version":2}`,
		RespHeaders:  http.Header{"Etag": {"v2"}, "Content-Type": {"application/json"}},
	}))

	b2 := clientB.GetWithContext(ctx, urlB)
	checkResp(b2)
	fmt.Printf("  Request 2 → status=%d  cached=%t  etag=%q  body=%s\n",
		b2.StatusCode, b2.Cached(), b2.Header.Get(rest.ETagHeader), b2.String())

	// Cache now has "v2". Next revalidation should return 304 and serve updated body.
	rest.FlushMockups()
	must(rest.AddMockups(&rest.Mock{
		URL:          urlB,
		HTTPMethod:   http.MethodGet,
		RespHTTPCode: http.StatusNotModified,
	}))

	b3 := clientB.GetWithContext(ctx, urlB)
	checkResp(b3)
	fmt.Printf("  Request 3 → status=%d  cached=%t  etag=%q  body=%s\n",
		b3.StatusCode, b3.Cached(), b3.Header.Get(rest.ETagHeader), b3.String())

	if b3.String() != `{"version":2}` {
		fmt.Println("  FAIL: cache was NOT updated after content change (stale body returned)")
		os.Exit(1)
	}
	fmt.Println("  OK: cache correctly updated after content change.")
}

func must(err error) {
	if err != nil {
		fmt.Printf("setup error: %v\n", err)
		os.Exit(1)
	}
}

func checkResp(r *rest.Response) {
	if r == nil || r.Err != nil {
		fmt.Printf("request error: %v\n", r.Err)
		os.Exit(1)
	}
}
