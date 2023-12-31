// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentExpiringPersistentURLMapFromDisk(expiration_check)

package util

import (
	"log"
	"os"
	"sync"
	"time"
)

type ConcurrentExpiringPersistentURLMap struct {
	mut                           sync.Mutex
	slice_storage                 map[int]*RandomBag64
	map_storage                   *ConcurrentExpiringMap
	b53m                          *Base53IDManager
	lbses                         *LogBucketStructuredExpiringStorage
	ebs                           *ExpiringBucketStorage
	generate_strings_up_to        int
	extra_keeparound_seconds_ram  int64
	extra_keeparound_seconds_disk int64
	map_size_persister            *MapSizeFileManager
	xattr_params                  *XattrParams
}

type MapItem2 struct {
	key         string
	value       string
	expiry_time int64
}

// Define a slice of Person structs
type People []MapItem2

// Implement the Len method required by sort.Interface
func (p People) Len() int {
	return len(p)
}

// Implement the Less method required by sort.Interface
func (p People) Less(i, j int) bool {
	return p[i].expiry_time > p[j].expiry_time
}

// Implement the Swap method required by sort.Interface
func (p People) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

/*
	func (manager *ConcurrentExpiringPersistentURLMap) PrintInternalState() {
		manager.mut.Lock()
		defer manager.mut.Unlock()

		log.Println(" ============ Printing CCPUM internal state ===========")
		log.Println("Printing slice_storage:")
		for k, v := range manager.slice_storage {
			log.Println("k,v:", k, *v)
		}
		log.Println(manager.map_storage)
		values := People{}
		for k, v := range manager.map_storage.m {
			values = append(values, MapItem2{k, v.value, v.expiry_time_unix})
		}
		sort.Sort(values)
		for k, v := range values {
			fmt.Println("kv:", k, v)
		}
		fmt.Println("time now:", time.Now().Unix())

		log.Println(" ------------------------------------------------------")
	}
*/
func (manager *ConcurrentExpiringPersistentURLMap) NumItems() int { //nolint:ireturn // is ok
	manager.mut.Lock()
	defer manager.mut.Unlock()

	return manager.map_storage.NumItems()
}

func (manager *ConcurrentExpiringPersistentURLMap) NumPastes() int { //nolint:ireturn // is ok
	// No need for lock here.
	return manager.map_storage.NumPastes()
}

func (manager *ConcurrentExpiringPersistentURLMap) GetEntry(short_url string) (MapItem, error) { //nolint:ireturn // is ok
	manager.mut.Lock()
	defer manager.mut.Unlock()

	val, err := GetEntryCommon(manager.map_storage, short_url)
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func (manager *ConcurrentExpiringPersistentURLMap) PutEntry(requested_length int, long_url string, expiry_time int64, value_type MapItemValueType) (string, error) {
	manager.mut.Lock()
	defer manager.mut.Unlock()

	val, err := PutEntry_Common(requested_length, long_url, value_type, expiry_time, manager.generate_strings_up_to, manager.slice_storage,
		manager.map_storage, manager.b53m, manager.lbses, manager.ebs, manager.map_size_persister, manager.xattr_params)
	return val, err
}

type CEPUMParams struct {
	Expiry_check_interval_seconds_ram    int
	Expiry_check_interval_seconds_disk   int
	Extra_keeparound_seconds_ram         int64
	Extra_keeparound_seconds_disk        int64
	Bucket_interval                      int64
	Bucket_directory_path_absolute       string
	Paste_bucket_directory_path_absolute string
	Size_file_path_absolute              string
	B53m                                 *Base53IDManager
	Size_file_rounded_multiple           int64
	Generate_strings_up_to               int
	Xattr_params                         *XattrParams
}

// This is the one you want to use in production
func CreateConcurrentExpiringPersistentURLMapFromDisk(cepum_params *CEPUMParams) *ConcurrentExpiringPersistentURLMap {
	if !(cepum_params.Extra_keeparound_seconds_disk > (cepum_params.Extra_keeparound_seconds_ram+5)*2) {
		log.Fatal("Extra keep around seconds disk must be much greater than ram!")
		panic("Invalid config")
	}

	cur_unix_timestamp := time.Now().Unix()
	Entry_should_be_deleted_fn := func(expiry_time int64) bool {
		return expiry_time < cur_unix_timestamp
	}
	slice_storage := make(map[int]*RandomBag64)
	expiry_callback := _internal_get_cem_expiry_callback(&slice_storage, cepum_params.Generate_strings_up_to) // this won't get called until much later so it's okay...

	lbses := NewLogBucketStructuredExpiringStorage(cepum_params.Bucket_interval, cepum_params.Bucket_directory_path_absolute)
	ebs := NewExpiringBucketStorage(cepum_params.Paste_bucket_directory_path_absolute)
	// delete expired log files on startup
	lbses.DeleteExpiredLogFiles(cepum_params.Extra_keeparound_seconds_disk)

	var nil_map_ptr *ConcurrentExpiringMap = nil

	// Now load from each file into the map
	params := LSRFD_Params{
		B53m:                        cepum_params.B53m,
		Log_directory_path_absolute: cepum_params.Bucket_directory_path_absolute,
		Entry_should_be_deleted_fn:  Entry_should_be_deleted_fn,
		Lss:                         lbses,
		Expiry_callback:             expiry_callback,
		Slice_storage:               slice_storage,
		Nil_ptr:                     nil_map_ptr,
		Size_file_rounded_multiple:  cepum_params.Size_file_rounded_multiple,
		Generate_strings_up_to:      cepum_params.Generate_strings_up_to,
		Size_file_path_absolute:     cepum_params.Size_file_path_absolute,
	}

	concurrent_map, map_size_persister := LoadStoredRecordsFromDisk(&params)

	manager := ConcurrentExpiringPersistentURLMap{ //nolint:forcetypeassert // just let it crash.
		mut:                           sync.Mutex{},
		slice_storage:                 slice_storage,
		map_storage:                   concurrent_map.(*ConcurrentExpiringMap),
		b53m:                          cepum_params.B53m,
		lbses:                         lbses,
		ebs:                           ebs,
		extra_keeparound_seconds_ram:  cepum_params.Extra_keeparound_seconds_ram,
		extra_keeparound_seconds_disk: cepum_params.Extra_keeparound_seconds_disk,
		generate_strings_up_to:        cepum_params.Generate_strings_up_to,
		map_size_persister:            map_size_persister,
		xattr_params:                  cepum_params.Xattr_params,
	}

	// It is very important to ensure that these functions run ONLY AFTER the LoadStoredRecordsFromDisk has finished.
	// This is because we need to load in the expired entries and delete the associated paste files on startup.
	//TODO: REmove this line
	//time.Sleep(60 * time.Second)
	go RunFuncEveryXSeconds(manager.RemoveAllExpiredURLsFromDisk, cepum_params.Expiry_check_interval_seconds_disk)
	go RunFuncEveryXSeconds(manager.RemoveAllExpiredURLsFromRAM, cepum_params.Expiry_check_interval_seconds_ram)
	return &manager
}

// Removed expired URLs from map in RAM every x seconds
func (manager *ConcurrentExpiringPersistentURLMap) RemoveAllExpiredURLsFromRAM() {
	// Don't need lock here because cem has lock
	manager.map_storage.Remove_All_Expired(manager.extra_keeparound_seconds_ram)
}

// Removed expired URLs from disk every x seconds
func (manager *ConcurrentExpiringPersistentURLMap) RemoveAllExpiredURLsFromDisk() {
	// Don't need lock here because lbses has lock
	manager.lbses.DeleteExpiredLogFiles(manager.extra_keeparound_seconds_disk)
}

// This callback puts the expired short URL ID back into the internal slice so that it can be reused
// It also deletes the associated file on disk if any
func _internal_get_cem_expiry_callback(slice_storage *map[int]*RandomBag64, generate_strings_up_to int) ExpiryCallback {
	return func(url_str string, map_item MapItem) {
		// check length of URL string
		length := len(url_str)
		if length <= generate_strings_up_to {
			// convert string back to uint64
			uint_num := Convert_str_to_uint64(url_str)
			(*slice_storage)[length].Push(uint_num)
		}
		// delete the associated file on disk
		if map_item.GetType().ValueType == TYPE_MAP_ITEM_PASTE {
			absfilepath := map_item.GetValue()
			err := os.Remove(absfilepath)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
		}
	}
}
