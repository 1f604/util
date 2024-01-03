package util

import (
	"fmt"
	"sync"
)

// keys are strings
type ConcurrentPermanentMap struct {
	mut sync.Mutex
	m   MapWithPastesCount[*PermanentMapItem]
}

type PermanentMapItem struct {
	value         string
	itemValueType MapItemValueType // URL or paste
}

type CPMNonExistentKeyError struct{}

func (e CPMNonExistentKeyError) Error() string {
	return "ConcurrentPermanentMap: nonexistent key"
}

func (pmi *PermanentMapItem) MapItemToString() string {
	return fmt.Sprintf("PermanentMapItem{value:%#v}", pmi.value)
}

func (pmi *PermanentMapItem) GetValue() string {
	return pmi.value
}

func (pmi *PermanentMapItem) GetType() MapItemType {
	return MapItemType{
		IsTemporary: false,
		ValueType:   pmi.itemValueType,
	}
}

func (emi *PermanentMapItem) GetExpiryTime() int64 {
	return -1
}

func (cpm *ConcurrentPermanentMap) NumItems() int {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	return cpm.m.NumItems()
}

func (cpm *ConcurrentPermanentMap) NumPastes() int {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	return cpm.m.NumPastes()
}

// You can call this on nil receiver
func (*ConcurrentPermanentMap) BeginConstruction(stored_map_length int64, expiry_callback ExpiryCallback) ConcurrentMap {
	m := NewMapWithPastesCount[*PermanentMapItem](stored_map_length)
	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   m,
	}
}

// Caller must check that the key_str is not already in the map.
func (cpm *ConcurrentPermanentMap) ContinueConstruction(key_str string, value_str string, expiry_time int64, item_value_type MapItemValueType) {
	// just add it to the map
	err := cpm.m.InsertNew(string(key_str), &PermanentMapItem{
		value:         value_str,
		itemValueType: item_value_type,
	})
	Check_err(err)
}

func (cpm *ConcurrentPermanentMap) FinishConstruction() {} // Does nothing.

// Returns an error if the entry already exists, otherwise returns nil.
func (cpm *ConcurrentPermanentMap) Put_New_Entry(key string, value string, _ int64, item_value_type MapItemValueType) error {
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	// if entry already exists, return an error
	err := cpm.m.InsertNew(key, &PermanentMapItem{
		value:         value,
		itemValueType: item_value_type,
	})
	return err
}

func (cpm *ConcurrentPermanentMap) Get_Entry(key string) (MapItem, error) {
	// 1. acquire read lock
	cpm.mut.Lock()
	defer cpm.mut.Unlock()

	// 2. check if it's in the map
	item, err := cpm.m.GetKey(key)
	// If the key doesn't exist
	if err != nil {
		// return error saying key doesn't exist
		return nil, CPMNonExistentKeyError{}
	}

	return item, nil
}

func NewEmptyConcurrentPermanentMap() *ConcurrentPermanentMap {
	return &ConcurrentPermanentMap{
		mut: sync.Mutex{},
		m:   NewMapWithPastesCount[*PermanentMapItem](0),
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
