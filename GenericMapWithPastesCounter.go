package util

// Users of this map are expected to access it with a mutex.
type MapWithPastesCount[T MapItem] interface {
	InsertNew(key string, value T) error
	GetKey(key string) (T, error)
	DeleteKey(key string)
	NumPastes() int
	NumItems() int
}

type MapWithPastesCount_impl[T MapItem] struct {
	m            map[string]T
	pastes_count int
}

func NewMapWithPastesCount[T MapItem](size int64) MapWithPastesCount[T] {
	return &MapWithPastesCount_impl[T]{
		m:            make(map[string]T, size),
		pastes_count: 0,
	}
}

func (mwpc *MapWithPastesCount_impl[T]) InsertNew(key string, value T) error {
	_, ok := mwpc.m[key]
	if ok {
		return KeyAlreadyExistsError{}
	}

	mwpc.m[key] = value
	if value.GetType().ValueType == TYPE_MAP_ITEM_PASTE {
		mwpc.pastes_count++
	}
	return nil
}

func (mwpc *MapWithPastesCount_impl[T]) GetKey(key string) (T, error) {
	val, ok := mwpc.m[key]
	if ok {
		return val, nil
	} else {
		var zero_value T
		return zero_value, CPMNonExistentKeyError{}
	}
}

func (mwpc *MapWithPastesCount_impl[T]) DeleteKey(key string) {
	// check if key is already in map
	val, ok := mwpc.m[key]
	if !ok {
		return
	}

	if val.GetType().ValueType == TYPE_MAP_ITEM_PASTE {
		mwpc.pastes_count--
	}

	delete(mwpc.m, key)
}

func (mwpc *MapWithPastesCount_impl[T]) NumItems() int {
	return len(mwpc.m)
}

func (mwpc *MapWithPastesCount_impl[T]) NumPastes() int {
	return mwpc.pastes_count
}
