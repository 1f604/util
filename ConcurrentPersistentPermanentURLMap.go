// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentPersistentPermanentURLMapFromDisk(expiration_check)

package util

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type ConcurrentPersistentPermanentURLMap struct {
	slice_map                     map[int]*RandomBag64
	urlmap                        *ConcurrentPermanentMap
	b53m                          *Base53IDManager
	lsps                          *LogStructuredPermanentStorage
	Generate_strings_up_to        int
	Extra_keeparound_seconds_ram  int64
	Extra_keeparound_seconds_disk int64
}

func (manager *ConcurrentPersistentPermanentURLMap) GetEntry(short_url string) (string, error) {
	val, err := GetEntryCommon(manager.urlmap, short_url)
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func (manager *ConcurrentPersistentPermanentURLMap) PutEntry(requested_length int, long_url string) (string, error) {
	if requested_length < 2 { //nolint:gomnd // 2 is not magic here. BASE53 can only go down to 2 characters because it uses one character for the checksum
		return "", errors.New("Requested length is too small.")
	}

	// if length is <= 5, grab it from one of the slices
	if requested_length <= manager.Generate_strings_up_to { //nolint:nestif // yeah it's complicated
		randombag, ok := manager.slice_map[requested_length]
		if !ok {
			log.Fatal("Failed to index slice_map. This should never happen.")
			panic("Failed to index slice_map. This should never happen.")
		}
		item, err := randombag.PopRandom()
		if err != nil {
			// This should be a common scenario.
			// We haven't modified anything at this point, so it's fine to return error here.
			return "", errors.New("No short URLs left")
		}
		// At this point, the item has been removed from the slice, so add it to the map.
		// Add item to the RAM map
		id_str := Convert_uint64_to_str(item, requested_length)
		err = manager.urlmap.Put_New_Entry(id_str, long_url)
		if err != nil { // Only possible error is if entry already exists, which it should never do since we got it from the slice.
			log.Fatal("Put_New_Entry failed. This should never happen. Error:", err)
			panic("Put_New_Entry failed. This should never happen. Error:" + err.Error())
		}
		// Add item to the disk storage
		// record the generation time
		generation_time_unix := time.Now().Unix()
		manager.lsps.AppendNewEntry(id_str, long_url, generation_time_unix)

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
			err = manager.urlmap.Put_New_Entry(id_str, long_url, expiry_time)
			if err == nil {
				// Successfully put it into the map. Now write it to disk too
				// It's okay if this is slow since it's just a write. Most operations are going to be reads.
				err = manager.lsps.AppendNewEntry(id_str, long_url, expiry_time)
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

type CPPUMParams struct {
	Bucket_directory_path_absolute  string
	B53m                            *Base53IDManager
	Size_file_rounded_growth_amount int64
	Create_size_file_if_not_exists  bool
	Generate_strings_up_to          int
}

// This is the one you want to use in production
func CreateConcurrentPersistentPermanentURLMapFromDisk(cepum_params *CEPUMParams) *ConcurrentPersistentPermanentURLMap { //nolint:gocognit // yeah it's complicated
	// First, list all the files in the directory
	entries, err := os.ReadDir(cepum_params.Bucket_directory_path_absolute)
	if err != nil {
		log.Fatal("Failed to open bucket directory:", cepum_params.Bucket_directory_path_absolute, "error:", err)
		panic(err)
	}
	// Now for each file, try to parse the file's filename and load it into the map if it's not expired
	files_to_be_loaded_from := make([]string, 0, len(entries))
	cur_unix_timestamp := time.Now().Unix()
	year_20000 := time.Date(20000, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	year_2023 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	for _, entry := range entries {
		if entry.IsDir() { // ignore directories
			continue
		}
		// if you can't parse it, raise an error
		expiry_timestamp_unix, err1 := LBSES_Parse_bucket_filename_to_timestamp(entry.Name())
		if err1 != nil {
			log.Fatal("Failed to parse name of file in bucket directory:", entry.Name(), "got error:", err1)
			panic(err1)
		}
		// if it's expired, then delete it with grace period
		absolute_file_path := filepath.Join(cepum_params.Bucket_directory_path_absolute, entry.Name())
		if (expiry_timestamp_unix + cepum_params.Extra_keeparound_seconds_disk) < cur_unix_timestamp {
			if err = os.Remove(absolute_file_path); err != nil {
				log.Fatal("Could not remove expired log file:", err)
				panic(err)
			}
		} else { // otherwise, add it to the list of files to be loaded from
			files_to_be_loaded_from = append(files_to_be_loaded_from, absolute_file_path)
		}
	}
	// First, create the size file if it doesn't exist
	sizefilepath := Get_LBSES_Size_File_Absolute_Path(cepum_params.Bucket_directory_path_absolute)
	if cepum_params.Create_size_file_if_not_exists { //nolint:nestif // yeah I know it's complicated
		// Check if it exists using os.stat
		_, err = os.Stat(sizefilepath)
		if err != nil {
			// if it doesn't exist then create it
			f, err := os.OpenFile(sizefilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:govet // ignore err shadow
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
			// set it to the size_growth_amount to begin with.
			_, err = f.WriteString(Int64_to_string(cepum_params.Size_file_rounded_growth_amount))
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
			// close the file!
			err = f.Close()
			Check_err(err)
		}
	}
	// Now load from each file into the map
	lsps := NewLogStructuredPermanentStorage(cepum_params.Bucket_interval, cepum_params.Bucket_directory_path_absolute, cepum_params.Size_file_rounded_growth_amount)
	// Load size of map from file
	stored_map_length, err := Get_current_rounded_size(lsps.size_file_path, lsps.size_growth_amount)
	if err != nil {
		log.Fatal("Get_current_rounded_size error: ", err)
		panic(err)
	}
	// Create the map and slice efficiently using the loaded rounded size. It's okay if it's too small, since these will grow automatically.
	m := make(map[string]interface{}, stored_map_length)

	for _, absolute_filepath := range files_to_be_loaded_from {
		f, err := os.Open(absolute_filepath) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("Failed to open bucket log file:", absolute_filepath, "err:", err)
			panic(err)
		}

		// Now scan the input from the file
		br := bufio.NewReader(f)
		for {
			key_str, err := br.ReadBytes('\t') //nolint:govet // ignore err shadow
			// check if error is EOF
			if errors.Is(err, io.EOF) {
				// make sure we're not waiting for more input
				// If ReadBytes encounters an error before finding a delimiter,
				// it returns the data read before the error and the error itself (often io.EOF).
				if len(key_str) != 0 {
					log.Fatal("File ", absolute_filepath, " does not end with newline, indicating some kind of corruption")
					panic("File doesn't end with newline.")
				}
				break
			}
			if err != nil {
				log.Fatal("Unexpected non-EOF error")
				panic(err)
			}
			value_str, err := br.ReadBytes('\t')
			if err != nil {
				panic(err)
			}
			expiry_str, err := br.ReadBytes('\n')
			if err != nil {
				panic(err)
			}
			// convert expiry_str to expiry_time_unix
			expiry_time, err := String_to_int64(string(expiry_str))
			if err != nil {
				log.Fatal("Could not convert expiry time to int64", err)
				panic(err)
			}
			switch {
			case expiry_time < cur_unix_timestamp:
				// already expired, so ignore it
				continue // go to next loop iteration
			case expiry_time < year_2023:
				// should never happen
				log.Fatal("Expiry time less than year 2023", expiry_time)
				panic(fmt.Sprintf("Expiry time %d greater than year 2023", expiry_time))
			case expiry_time > year_20000:
				// expiry_time too big
				// fatal error
				log.Fatal("Expiry time greater than year 20000", expiry_time)
				panic("Expiry time greater than year 20000")
			}
			// So now we know the entry in the file is not expired.
			// But what if there is already an entry in the map???
			val, ok := m[string(key_str)]
			if ok { // This implies that we've already seen a non-expired entry for that URL ID, which should never happen
				log.Fatal("Multiple non-expired entries found in log files for same key string: ", val, " expiry time: ", val.expiry_time_unix, " value: ", val.value, " key_str: ", string(key_str))
				panic("Multiple non-expired entries found in log files for same URL ID")
			}

			// first add it to the map
			m[string(key_str)] = value_str
		}
	}

	// fmt.Println("added:", item)
	// fmt.Println("New map:", cem.m)
	// fmt.Printf("New heap: %+v\n", cem.hq)
	should_be_added_fn := func(keystr string) bool {
		_, ok := m[keystr]
		return !ok
	}

	slice_map := make(map[int]*RandomBag64)
	for n := 2; n <= cepum_params.Generate_strings_up_to; n++ {
		log.Println("Generating all Base 53 IDs of length", n)
		slice, err := cepum_params.B53m.B53_generate_all_Base53IDs_int64_optimized(n, should_be_added_fn) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("B53_generate_all_Base53IDs_int64_optimized failed", err)
			panic("B53_generate_all_Base53IDs_int64_optimized failed: " + err.Error())
		}
		slice_map[n] = CreateRandomBagFromSlice(slice)
	}

	manager := ConcurrentPersistentPermanentURLMap{
		slice_map:                     slice_map,
		urlmap:                        NewEmptyConcurrentPermanentMap(),
		b53m:                          cepum_params.B53m,
		lsps:                          lsps,
		Extra_keeparound_seconds_ram:  cepum_params.Extra_keeparound_seconds_ram,
		Extra_keeparound_seconds_disk: cepum_params.Extra_keeparound_seconds_disk,
		Generate_strings_up_to:        cepum_params.Generate_strings_up_to,
	}

	return &manager
}
