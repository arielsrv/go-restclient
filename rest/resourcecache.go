package rest

import (
	"container/list"
	"sync"
	"time"
)

// ResourceCache, is an LRU-TTL Cache, that caches Responses base on headers
// It uses 3 goroutines -> one for LRU, and the other two for TTL.

// The cache itself.
var resourceCache *resourceTTLLruMap

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
var MaxCacheSize = 1 * GB

// Current Cache Size.
var cacheSize int64

type lruOperation int

const (
	move lruOperation = iota
	push
	del
	last
)

type lruMsg struct {
	resp      *Response
	operation lruOperation
}

type resourceTTLLruMap struct {
	cache    map[string]*Response
	skipList *skipList    // skiplist for TTL
	lruList  *list.List   // List for LRU
	lruChan  chan *lruMsg // Channel for LRU messages
	ttlChan  chan bool    // Channel for TTL messages
	popChan  chan string
	rwMtx    sync.RWMutex // Read Write Locking Mutex
}

func init() {
	resourceCache = &resourceTTLLruMap{
		cache:    make(map[string]*Response),
		skipList: newSkipList(),
		lruList:  list.New(),
		lruChan:  make(chan *lruMsg, 10000),
		ttlChan:  make(chan bool, 1000),
		popChan:  make(chan string),
		rwMtx:    sync.RWMutex{},
	}

	go resourceCache.lruOperations()
	go resourceCache.ttl()
}

func (r *resourceTTLLruMap) lruOperations() {
	for {
		msg := <-r.lruChan

		switch msg.operation {
		case move:
			r.lruList.MoveToFront(msg.resp.listElement)
		case push:
			msg.resp.listElement = r.lruList.PushFront(msg.resp.Request.URL.String())
		case del:
			r.lruList.Remove(msg.resp.listElement)
		case last:
			if value, ok := r.lruList.Back().Value.(string); ok {
				r.popChan <- value
			}
		}
	}
}

func (r *resourceTTLLruMap) get(key string) *Response {
	// Read lock only
	r.rwMtx.RLock()
	resp := r.cache[key]
	r.rwMtx.RUnlock()

	// If expired, remove it
	if resp != nil && resp.ttl != nil && time.Until(*resp.ttl) <= 0 {
		// Full lock
		r.rwMtx.Lock()
		defer r.rwMtx.Unlock()

		// JIC, get the freshest version
		resp = r.cache[key]

		// Check again with the lock
		if resp != nil && resp.ttl != nil && time.Until(*resp.ttl) <= 0 {
			r.remove(key, resp)
			return nil // return. Do not send the move message
		}
	}

	if resp != nil {
		// Buffered msg to LruList
		// Move forward
		r.lruChan <- &lruMsg{
			operation: move,
			resp:      resp,
		}
	}

	return resp
}

// Set if key not exist.
func (r *resourceTTLLruMap) setNX(key string, value *Response) {
	// Full Lock
	r.rwMtx.Lock()
	defer r.rwMtx.Unlock()

	v := r.cache[key]

	if v == nil {
		r.cache[key] = value

		// PushFront in LruList
		r.lruChan <- &lruMsg{
			operation: push,
			resp:      value,
		}

		// Set ttl if necessary
		if value.ttl != nil {
			value.skipListElement = r.skipList.insert(key, *value.ttl)
			r.ttlChan <- true
		}

		// Add Response Size to Cache
		// Not necessary to use atomic
		cacheSize += value.size()

		for i := 0; ByteSize(cacheSize) >= MaxCacheSize && i < 10; i++ {
			r.lruChan <- &lruMsg{
				nil,
				last,
			}

			k := <-r.popChan
			response := r.cache[k]

			r.remove(k, response)
		}
	}
}

func (r *resourceTTLLruMap) remove(key string, resp *Response) {
	delete(r.cache, key)                    // Delete from map
	r.skipList.remove(resp.skipListElement) // Delete from skipList
	r.lruChan <- &lruMsg{                   // Delete from LruList
		operation: del,
		resp:      resp,
	}

	// Delete bytes cache
	// Not need for atomic
	cacheSize -= resp.size()
}

func (r *resourceTTLLruMap) ttl() {
	// Function to send a message when the timer expires
	backToFuture := func() {
		r.ttlChan <- true
	}

	// A timer.
	future := time.AfterFunc(24*time.Hour, backToFuture)

	for {
		<-r.ttlChan

		// Full Lock
		r.rwMtx.Lock()

		now := time.Now()

		// Traverse the skiplist which is ordered by ttl.
		// We do this by looping at level 0
		for node := r.skipList.head.next[0]; node != nil; node = node.next[0] {
			timeLeft := node.ttl.Sub(now)

			// If we still have time, check the timer and break
			if timeLeft > 0 {
				if !future.Reset(timeLeft) {
					future = time.AfterFunc(timeLeft, backToFuture)
				}

				break
			}

			// Remove from cache if time's up
			r.remove(node.key, r.cache[node.key])
		}

		r.rwMtx.Unlock()
	}
}
