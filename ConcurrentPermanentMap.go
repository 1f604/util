package util

import (
	"errors"
	"fmt"
	"sync"
)

// keys are strings
type ConcurrentPermanentMap struct {
	mut sync.Mutex
	m   map[string]PermanentMapItem
}

type PermanentMapItem struct {
	value string
}

func (pmi *PermanentMapItem) MapItemToString() string {
	return fmt.Sprintf("PermanentMapItem{value:%#v}", pmi.value)
}

func (cpm *ConcurrentPermanentMap) NumItems() int {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	return len(cpm.m)
}

// You can call this on nil receiver
func (*ConcurrentPermanentMap) BeginConstruction(stored_map_length int64, expiry_callback ExpiryCallback) *ConcurrentPermanentMap {
	m := make(map[string]PermanentMapItem, stored_map_length)
	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   m,
	}
}

// Caller must check that the key_str is not already in the map.
func (cpm *ConcurrentPermanentMap) ContinueConstruction(key_str string, value_str string, expiry_time int64) {
	// just add it to the map
	cpm.m[string(key_str)] = PermanentMapItem{value_str}
}

func (cpm *ConcurrentPermanentMap) FinishConstruction() {} // Does nothing.

// Returns an error if the entry already exists, otherwise returns nil.
func (cpm *ConcurrentPermanentMap) Put_New_Entry(key string, value string) error {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	// if entry already exists, return an error
	_, ok := cpm.m[key]
	if ok {
		return errors.New("Entry already exists!!")
	}

	cpm.m[key] = PermanentMapItem{value}
	return nil
}

func (cpm *ConcurrentPermanentMap) Get_Entry(key string) (MapItem, error) {
	// 1. acquire read lock
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	// 2. check if it's in the map
	map_item, ok := cpm.m[key]
	// If the key doesn't exist
	if !ok {
		// return error saying key doesn't exist
		return nil, NonExistentKeyError{}
	}

	return &map_item, nil
}

func NewEmptyConcurrentPermanentMap() *ConcurrentPermanentMap {
	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   make(map[string]PermanentMapItem),
	}
}

/*
type CPMItem struct {
	Key   string
	Value string
}

func NewConcurrentPermanentMapFromSlice(kv_pairs []CPMItem) *ConcurrentPermanentMap {
	m := make(map[string]string, len(kv_pairs))

	for _, cpm_item := range kv_pairs {
		m[cpm_item.Key] = cpm_item.Value
	}

	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   m,
	}
}

type TTLMap [K any, V any] struct {
    Data []T
	queue chan[K]
	Map map[K]V
}*/
