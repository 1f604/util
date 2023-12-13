// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentExpiringPersistentURLMapFromDisk(expiration_check)

package util

import (
	"errors"
	"log"
	"time"
)

type ConcurrentExpiringPersistentURLMap struct {
	slice_storage                 map[int]*RandomBag64
	map_storage                   *ConcurrentExpiringMap
	b53m                          *Base53IDManager
	lbses                         *LogBucketStructuredExpiringStorage
	Generate_strings_up_to        int
	Extra_keeparound_seconds_ram  int64
	Extra_keeparound_seconds_disk int64
}

func (manager *ConcurrentExpiringPersistentURLMap) GetEntry(short_url string) (MapItem, error) { //nolint:ireturn // is ok
	val, err := GetEntryCommon(manager.map_storage, short_url)
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func (manager *ConcurrentExpiringPersistentURLMap) PutEntry(requested_length int, long_url string, expiry_time int64) (string, error) {
	if requested_length < 2 { //nolint:gomnd // 2 is not magic here. BASE53 can only go down to 2 characters because it uses one character for the checksum
		return "", errors.New("Requested length is too small.")
	}
	// check the expiry time
	cur_time_unix := time.Now().Unix()
	if expiry_time < (cur_time_unix + 5) { //nolint:gomnd // 5 seconds is a good time...
		return "", errors.New("Entry is already expired.")
	}

	// if length is <= 5, grab it from one of the slices
	if requested_length <= manager.Generate_strings_up_to { //nolint:nestif // yeah it's complicated
		randombag, ok := manager.slice_storage[requested_length]
		if !ok {
			log.Fatal("Failed to index slice_storage. This should never happen.")
			panic("Failed to index slice_storage. This should never happen.")
		}
		item, err := randombag.PopRandom()
		if err != nil {
			// This should be a common scenario.
			// We haven't modified anything at this point, so it's fine to return error here.
			return "", errors.New("No short URLs left")
		}
		// At this point, the item has been removed from the slice, so add it to the map.
		// Add item to the map
		id_str := Convert_uint64_to_str(item, requested_length)
		err = manager.map_storage.Put_New_Entry(id_str, long_url, expiry_time)
		if err != nil { // Only possible error is if entry already exists, which it should never do since we got it from the slice.
			log.Fatal("Put_New_Entry failed. This should never happen. Error:", err)
			panic("Put_New_Entry failed. This should never happen. Error:" + err.Error())
		}
		return id_str, nil
	} else { // Otherwise randomly generate it and see if it already exists
		id, err := manager.b53m.B53_generate_random_Base53ID(requested_length)
		if err != nil {
			log.Fatal("Failed to generate new random ID. This should never happen. Error:", err)
			panic(err)
		}
		// try 100 times, trying again when it fails due to already existing in the map
		// probability of failing 100 times in a row should be astronomically small
		for i := 0; i < 100; i++ {
			id_str := id.GetCombinedString()
			err = manager.map_storage.Put_New_Entry(id_str, long_url, expiry_time)
			if err == nil {
				// Successfully put it into the map. Now write it to disk too
				// It's okay if this is slow since it's just a write. Most operations are going to be reads.
				err = manager.lbses.AppendNewEntry(id_str, long_url, expiry_time)
				if err != nil {
					// It should never fail.
					log.Fatal("AppendNewEntry failed:", err)
					panic(err)
				}
				return id_str, nil
			}
			if i > 3 { //nolint:gomnd // 3 is a good number.
				log.Println("Unexpected event: got duplicate ID", i, "times in a row. ID is:", id_str)
			}
		}
		log.Fatal("Failed to generate new random string 100 times, this should never happen")
		panic("Failed to generate new random string 100 times, this should never happen")
	}
}

// This is just an example function, you wouldn't use it in production.
/*
func _newEmptyConcurrentExpiringPersistentURLMap(expiry_check_interval_seconds int, extra_keeparound_seconds_ram int, extra_keeparound_seconds_disk int,
	bucket_interval int64, bucket_directory_path_absolute string, b53m *Base53IDManager) *ConcurrentExpiringPersistentURLMap {

	slice_storage := make(map[int]*RandomBag64)
	for n := 2; n < 5; n++ {
		log.Println("Generating all Base 53 IDs of length", n)
		slice, err := b53m.B53_generate_all_Base53IDs_int64_optimized(n)
		if err != nil {
			log.Fatal("B53_generate_all_Base53IDs_int64_optimized failed", err)
			panic("B53_generate_all_Base53IDs_int64_optimized failed: " + err.Error())
		}
		slice_storage[n] = CreateRandomBagFromSlice(slice)
	}
	manager := ConcurrentExpiringPersistentURLMap{
		slice_storage: slice_storage,
		map_storage:    *NewEmptyConcurrentExpiringMap(nil),
		b53m:      b53m,
		lbses:     NewLogBucketStructuredExpiringStorage(bucket_interval, bucket_directory_path_absolute),
	}

	go manager._internal_expire_URLs_from_RAM(expiry_check_interval_seconds, extra_keeparound_seconds_ram)
	go manager.RemoveAllExpiredURLsFromDisk(expiry_check_interval_seconds, extra_keeparound_seconds_disk)
	return &manager
}
*/

type CEPUMParams struct {
	Expiry_check_interval_seconds_ram  int
	Expiry_check_interval_seconds_disk int
	Extra_keeparound_seconds_ram       int64
	Extra_keeparound_seconds_disk      int64
	Bucket_interval                    int64
	Bucket_directory_path_absolute     string
	Size_file_path_absolute            string
	B53m                               *Base53IDManager
	Size_file_rounded_multiple         int64
	Create_size_file_if_not_exists     bool
	Generate_strings_up_to             int
}

// This is the one you want to use in production
func CreateConcurrentExpiringPersistentURLMapFromDisk(cepum_params *CEPUMParams) *ConcurrentExpiringPersistentURLMap {
	cur_unix_timestamp := time.Now().Unix()
	entry_should_be_ignored_fn := func(expiry_time int64) bool {
		return expiry_time < cur_unix_timestamp
	}
	slice_storage := make(map[int]*RandomBag64)
	expiry_callback := _internal_get_cem_expiry_callback(&slice_storage, cepum_params.Generate_strings_up_to) // this won't get called until much later so it's okay...

	// Now load from each file into the map
	lbses := NewLogBucketStructuredExpiringStorage(cepum_params.Bucket_interval, cepum_params.Bucket_directory_path_absolute)

	var nil_map_ptr *ConcurrentExpiringMap = nil
	concurrent_map := LoadStoredRecordsFromDisk(cepum_params, entry_should_be_ignored_fn, lbses, expiry_callback, slice_storage, nil_map_ptr)

	manager := ConcurrentExpiringPersistentURLMap{ //nolint:forcetypeassert // just let it crash.
		slice_storage:                 slice_storage,
		map_storage:                   concurrent_map.(*ConcurrentExpiringMap),
		b53m:                          cepum_params.B53m,
		lbses:                         lbses,
		Extra_keeparound_seconds_ram:  cepum_params.Extra_keeparound_seconds_ram,
		Extra_keeparound_seconds_disk: cepum_params.Extra_keeparound_seconds_disk,
		Generate_strings_up_to:        cepum_params.Generate_strings_up_to,
	}

	go RunFuncEveryXSeconds(manager.RemoveAllExpiredURLsFromDisk, cepum_params.Expiry_check_interval_seconds_disk)
	go RunFuncEveryXSeconds(manager.RemoveAllExpiredURLsFromRAM, cepum_params.Expiry_check_interval_seconds_ram)
	return &manager
}

// Removed expired URLs from map in RAM every x seconds
func (manager *ConcurrentExpiringPersistentURLMap) RemoveAllExpiredURLsFromRAM() {
	manager.map_storage.Remove_All_Expired(manager.Extra_keeparound_seconds_ram)
}

// Removed expired URLs from disk every x seconds
func (manager *ConcurrentExpiringPersistentURLMap) RemoveAllExpiredURLsFromDisk() {
	manager.lbses.DeleteExpiredLogFiles(manager.Extra_keeparound_seconds_disk)
}

// This callback puts the expired short URL ID back into the internal slice so that it can be reused
func _internal_get_cem_expiry_callback(slice_storage *map[int]*RandomBag64, generate_strings_up_to int) func(string) {
	return func(url_str string) {
		// check length of URL string
		length := len(url_str)
		if length <= generate_strings_up_to {
			// convert string back to uint64
			uint_num := Convert_str_to_uint64(url_str)
			(*slice_storage)[length].Push(uint_num)
		}
	}
}
