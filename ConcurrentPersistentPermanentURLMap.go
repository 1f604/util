// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentPersistentPermanentURLMapFromDisk(expiration_check)

package util

import "time"

type ConcurrentPersistentPermanentURLMap struct {
	slice_map                     map[int]*RandomBag64
	urlmap                        *ConcurrentPermanentMap
	b53m                          *Base53IDManager
	lsps                          *LogStructuredPermanentStorage
	Generate_strings_up_to        int
	Extra_keeparound_seconds_ram  int64
	Extra_keeparound_seconds_disk int64
}

func (manager *ConcurrentPersistentPermanentURLMap) GetEntry(short_url string) (MapItem, error) {
	val, err := GetEntryCommon(manager.urlmap, short_url)
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func (manager *ConcurrentPersistentPermanentURLMap) PutEntry(requested_length int, long_url string) (string, error) {
	cur_unix_timestamp := time.Now().Unix()
	val, err := PutEntry_Common(requested_length, long_url, cur_unix_timestamp, manager.Generate_strings_up_to, manager.slice_map, manager.urlmap, manager.b53m, manager.lsps)
	return val, err
}

type CPPUMParams struct {
	Bucket_directory_path_absolute  string
	B53m                            *Base53IDManager
	Size_file_rounded_growth_amount int64
	Create_size_file_if_not_exists  bool
	Generate_strings_up_to          int
}

// This is the one you want to use in production
func CreateConcurrentPersistentPermanentURLMapFromDisk(cepum_params *CEPUMParams) *ConcurrentPersistentPermanentURLMap { //nolint:gocognit // yeah it's complicated
	slice_storage := make(map[int]*RandomBag64, 7)
	lsps := NewLogStructuredPermanentStorage(cepum_params.Bucket_interval, cepum_params.Bucket_directory_path_absolute)
	var nil_map_ptr *ConcurrentPermanentMap = nil

	// Now load from each file into the map
	concurrent_map := LoadStoredRecordsFromDisk(cepum_params, nil, lsps, nil, slice_storage, nil_map_ptr)

	manager := ConcurrentPersistentPermanentURLMap{
		slice_map:                     slice_storage,
		urlmap:                        concurrent_map.(*ConcurrentPermanentMap),
		b53m:                          cepum_params.B53m,
		lsps:                          lsps,
		Extra_keeparound_seconds_ram:  cepum_params.Extra_keeparound_seconds_ram,
		Extra_keeparound_seconds_disk: cepum_params.Extra_keeparound_seconds_disk,
		Generate_strings_up_to:        cepum_params.Generate_strings_up_to,
	}

	return &manager
}
