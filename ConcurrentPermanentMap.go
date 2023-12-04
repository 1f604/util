package util

import (
	"sync"
)

// keys are strings
type ConcurrentPermanentMap struct {
	mut sync.Mutex
	m   map[interface{}]interface{}
}

func (cpm *ConcurrentPermanentMap) NumItems() int {
	return len(cpm.m)
}

func (cpm *ConcurrentPermanentMap) Put_Entry(key interface{}, value interface{}) {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	cpm.m[key] = value
}

func (cpm *ConcurrentPermanentMap) Get_Entry(key interface{}) (interface{}, bool) {
	// 1. acquire lock
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	// 2. check if it's in the map
	map_item, ok := cpm.m[key]

	return map_item, ok
}

func NewEmptyConcurrentPermanentMap() *ConcurrentPermanentMap {
	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   make(map[interface{}]interface{}),
	}
}

type CPMItem struct {
	Key   interface{}
	Value interface{}
}

func NewConcurrentPermanentMapFromSlice(kv_pairs []CPMItem) *ConcurrentPermanentMap {
	m := make(map[interface{}]interface{}, len(kv_pairs))

	for _, cpm_item := range kv_pairs {
		m[cpm_item.Key] = cpm_item.Value
	}

	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   m,
	}
}

/*
type TTLMap [K any, V any] struct {
    Data []T
	queue chan[K]
	Map map[K]V
}*/
