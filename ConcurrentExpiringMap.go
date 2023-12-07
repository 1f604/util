// Map with expiring entries
// For slightly better performance, replace map[string]string with map[int64]string. See https://www.komu.engineer/blogs/01/go-gc-maps
// Memory usage can be more than double what you actually store in it.
// Based on my own testing, storing 10 million 128 byte URLs will take around 3.6GB of RAM, so each 128 byte URL took around 360 bytes of RAM.
// Entries can only be inserted, they cannot be updated or deleted before they expire.
// Uses sync.Mutex to protect concurrent access. Adding, getting, and removing entries require obtaining the mutex first.
// TODO: Benchmark switching to use a RWMutex or a sync.Map for improved performance.
// I tested sync.Map, it apparently has no reserve feature? Bulk load is slow - 7.8 seconds.
// Heap-based implementation for performance and simplicity
// Benchmarks show that Remove_All_Expired takes 3 seconds to remove 10 million expired entries
// Benchmarks show that NewConcurrentExpiringMapFromSlice takes 3.5 seconds to load 10 million entries
// No requirement for entries to have same TTL duration
// No support for updating expiry time - though this functionality can be added later if necessary.
// Example use cases:
// 1. Expiring short URLs - short URL -> long URL map
// 2. Expiring pastebins - short URL -> file path map
// 3. Expiring tokens - token -> expiry time map
// Map will return an error for expired entries
package util

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

type ExpiringHeapItem struct {
	// Yes, sometimes it is more space-efficient to store only a pointer to the map key rather than the key itself, but here keys are expected to be small, around 6-7 bytes on average.
	key              interface{}
	expiry_time_unix int64 // When the item expires. This is used as the priority. Doesn't have to be unix time.
}

type ExpiringMapItem struct {
	value interface{} // The actual value of the item; arbitrary.
	// Yes, expiry_time_unix is duplicated but it's only 8 bytes, using a pointer here wouldn't gain much.
	expiry_time_unix int64 // When the item expires. This is used as the priority. Doesn't have to be unix time.
}

type ExpiryCallback func(interface{})

// keys are strings
type ConcurrentExpiringMap struct {
	mut             sync.Mutex
	m               map[interface{}]ExpiringMapItem
	hq              ExpiringHeapQueue
	expiry_callback ExpiryCallback
}

func NewEmptyConcurrentExpiringMap(expiry_callback ExpiryCallback) *ConcurrentExpiringMap {
	m := make(map[interface{}]ExpiringMapItem, 0)
	hq := make(ExpiringHeapQueue, 0)
	// heap.Init(&hq) // No need to initialize an empty heap.

	return &ConcurrentExpiringMap{
		mut:             sync.Mutex{},
		m:               m,
		hq:              hq,
		expiry_callback: expiry_callback,
	}
}

func (cem *ConcurrentExpiringMap) Put_New_Entry(key interface{}, value interface{}, expiry_time int64) error {
	cem.mut.Lock()
	defer cem.mut.Unlock()
	// First check if it's already in the map
	_, ok := cem.m[key]
	// If the key already exists
	if ok {
		// return error saying key already exists
		return KeyAlreadyExistsError{}
	}

	// first add it to the map
	map_item := ExpiringMapItem{
		value:            value,
		expiry_time_unix: expiry_time,
	}
	cem.m[key] = map_item

	// then add it to the heap
	heap_item := ExpiringHeapItem{
		key:              key,
		expiry_time_unix: expiry_time,
	}
	heap.Push(&cem.hq, &heap_item)
	// fmt.Println("added:", item)
	// fmt.Println("New map:", cem.m)
	// fmt.Printf("New heap: %+v\n", cem.hq)

	return nil
}

type CEMItem struct {
	Key              interface{}
	Value            interface{}
	Expiry_time_unix int64
}

// batched mode for fast loading from disk
// Takes around 3.5s to load 10 million items, 300ms for loading 1 million items
func NewConcurrentExpiringMapFromSlice(expiry_callback ExpiryCallback, kv_pairs []CEMItem) *ConcurrentExpiringMap {
	m := make(map[interface{}]ExpiringMapItem, len(kv_pairs))
	hq := make(ExpiringHeapQueue, 0, len(kv_pairs))

	for _, cem_item := range kv_pairs {
		key := cem_item.Key
		value := cem_item.Value
		expiry_time := cem_item.Expiry_time_unix

		// first add it to the map
		map_item := ExpiringMapItem{
			value:            value,
			expiry_time_unix: expiry_time,
		}
		m[key] = map_item

		// then add it to the heap
		heap_item := ExpiringHeapItem{
			key:              key,
			expiry_time_unix: expiry_time,
		}
		hq.Push(&heap_item)
	}
	// Now initialize the heap
	heap.Init(&hq)

	// fmt.Println("added:", item)
	// fmt.Println("New map:", cem.m)
	// fmt.Printf("New heap: %+v\n", cem.hq)

	return &ConcurrentExpiringMap{
		mut:             sync.Mutex{},
		m:               m,
		hq:              hq,
		expiry_callback: expiry_callback,
	}
}

// keep links around for extra_keeparound_seconds just to tell people that the link has expired
// this function will remove 10 million entries in 3 seconds
func (cem *ConcurrentExpiringMap) Remove_All_Expired(extra_keeparound_seconds int64) {
	cem.mut.Lock()
	defer cem.mut.Unlock()

	cur_time := time.Now().Unix()
	// pop root from hq until root is no longer expired or the thing is empty
	for len(cem.hq) > 0 && cem.hq[0].expiry_time_unix+extra_keeparound_seconds <= cur_time {
		// first remove from heap
		heap_item := heap.Pop(&cem.hq)
		item, ok := heap_item.(*ExpiringHeapItem)
		if !ok {
			panic("Expected ExpiringHeapItem, got something else. This should never happen.")
		}
		// now call the callback with the removed item key
		if cem.expiry_callback != nil {
			cem.expiry_callback(item.key)
		}
		// then remove from map
		delete(cem.m, item.key)
		// fmt.Println("removed:", item)
		// fmt.Println("New map:", cem.m)
		// fmt.Printf("New heap: %+v\n", cem.hq)
	}
	// fmt.Println("Done expiring items.")
}

type NonExistentKeyError struct{}

func (e NonExistentKeyError) Error() string {
	return "ConcurrentExpiringMap: nonexistent key"
}

type KeyExpiredError struct{}

func (e KeyExpiredError) Error() string {
	return "ConcurrentExpiringMap: key expired"
}

type KeyAlreadyExistsError struct{}

func (e KeyAlreadyExistsError) Error() string {
	return "ConcurrentExpiringMap: key already exists"
}

func (cem *ConcurrentExpiringMap) NumItems() int {
	return len(cem.m)
}

func (cem *ConcurrentExpiringMap) Get_Entry(key interface{}) (interface{}, error) {
	// 1. acquire read lock
	cem.mut.Lock()
	defer cem.mut.Unlock()

	// 2. check if it's in the map
	map_item, ok := cem.m[key]
	// If the key doesn't exist
	if !ok {
		// return error saying key already exists
		return nil, NonExistentKeyError{}
	}

	// 3. check if it's expired
	if map_item.expiry_time_unix <= time.Now().Unix() {
		return nil, KeyExpiredError{}
	}

	// 4. if it's not expired, then return it
	return map_item.value, nil
}

/*
type TTLMap [K any, V any] struct {
    Data []T
	queue chan[K]
	Map map[K]V
}*/

// This example demonstrates a priority queue built using the heap interface.

func (p ExpiringMapItem) String() string {
	return fmt.Sprintf("ExpiringMapItem{value:%v, expiry_time:%d}", p.value, p.expiry_time_unix)
}

func (p ExpiringHeapItem) String() string {
	return fmt.Sprintf("ExpiringHeapItem{key:%v, expiry_time:%d}", p.key, p.expiry_time_unix)
}

// ============= All this stuff is just to implement the interface required by heap ===================
type ExpiringHeapQueue []*ExpiringHeapItem

func (pq ExpiringHeapQueue) Len() int { return len(pq) }

func (pq ExpiringHeapQueue) Less(i, j int) bool { // root is the element with smallest expiry date
	return pq[i].expiry_time_unix < pq[j].expiry_time_unix
}

func (pq ExpiringHeapQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *ExpiringHeapQueue) Push(x any) {
	item := x.(*ExpiringHeapItem)
	*pq = append(*pq, item)
}

func (pq *ExpiringHeapQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// ====================================================================================================
