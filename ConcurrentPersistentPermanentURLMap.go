// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentPersistentPermanentURLMapFromDisk(expiration_check)

package util

import (
	"log"
	"sync"
	"time"
)

type ConcurrentPersistentPermanentURLMap struct {
	mut                    sync.Mutex
	slice_map              map[int]*RandomBag64
	urlmap                 *ConcurrentPermanentMap
	b53m                   *Base53IDManager
	lsps                   *LogStructuredPermanentStorage
	pbs                    *PermanentBucketStorage
	generate_strings_up_to int
	map_size_persister     *MapSizeFileManager
}

func (manager *ConcurrentPersistentPermanentURLMap) PrintInternalState() {
	manager.mut.Lock()
	defer manager.mut.Unlock()

	log.Println(" ============ Printing CPPUM internal state ===========")
	log.Println("Printing slice_maps:")
	for k, v := range manager.slice_map {
		log.Println("k,v:", k, *v)
	}
	log.Println(manager.urlmap)
	log.Println(" ------------------------------------------------------")
}

func (manager *ConcurrentPersistentPermanentURLMap) NumItems() int { //nolint:ireturn // is ok
	manager.mut.Lock()
	defer manager.mut.Unlock()

	return manager.urlmap.NumItems()
}

func (manager *ConcurrentPersistentPermanentURLMap) GetEntry(short_url string) (MapItem, error) { //nolint:ireturn //this is ok
	manager.mut.Lock()
	defer manager.mut.Unlock()

	val, err := GetEntryCommon(manager.urlmap, short_url)
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func (manager *ConcurrentPersistentPermanentURLMap) PutEntry(requested_length int, long_url string, _ int64, value_type MapItemValueType) (string, error) {
	manager.mut.Lock()
	defer manager.mut.Unlock()

	cur_unix_timestamp := time.Now().Unix()

	val, err := PutEntry_Common(requested_length, long_url, value_type, cur_unix_timestamp, manager.generate_strings_up_to, manager.slice_map, manager.urlmap, manager.b53m, manager.lsps, manager.pbs, manager.map_size_persister)
	return val, err
}

type CPPUMParams struct {
	Log_directory_path_absolute    string
	Bucket_directory_path_absolute string
	B53m                           *Base53IDManager
	Generate_strings_up_to         int
	Log_file_max_size_bytes        int64
	Size_file_rounded_multiple     int64
	Size_file_path_absolute        string
}

// This is the one you want to use in production
func CreateConcurrentPersistentPermanentURLMapFromDisk(cppum_params *CPPUMParams) *ConcurrentPersistentPermanentURLMap {
	slice_storage := make(map[int]*RandomBag64)
	lsps := NewLogStructuredPermanentStorage(cppum_params.Log_file_max_size_bytes, cppum_params.Log_directory_path_absolute)
	pbs := NewPermanentBucketStorage(cppum_params.Bucket_directory_path_absolute)
	var nil_map_ptr *ConcurrentPermanentMap = nil

	// Now load from each file into the map
	params := LSRFD_Params{
		B53m:                        cppum_params.B53m,
		Log_directory_path_absolute: cppum_params.Log_directory_path_absolute,
		Entry_should_be_ignored_fn:  nil,
		Lss:                         lsps,
		Expiry_callback:             nil,
		Slice_storage:               slice_storage,
		Nil_ptr:                     nil_map_ptr,
		Size_file_rounded_multiple:  cppum_params.Size_file_rounded_multiple,
		Generate_strings_up_to:      cppum_params.Generate_strings_up_to,
		Size_file_path_absolute:     cppum_params.Size_file_path_absolute,
	}

	concurrent_map, map_size_persister := LoadStoredRecordsFromDisk(&params)

	manager := ConcurrentPersistentPermanentURLMap{ //nolint:forcetypeassert // it's okay. Just let it crash.
		mut:                    sync.Mutex{},
		slice_map:              slice_storage,
		urlmap:                 concurrent_map.(*ConcurrentPermanentMap),
		b53m:                   cppum_params.B53m,
		lsps:                   lsps,
		pbs:                    pbs,
		generate_strings_up_to: cppum_params.Generate_strings_up_to,
		map_size_persister:     map_size_persister,
	}

	return &manager
}
