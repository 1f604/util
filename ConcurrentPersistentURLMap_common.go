// API:
// 1. PutEntry(long_url, expiry_date) -> (short_url, err)
// 2. GetEntry(short_url) -> (long_url, err)
// 3. CreateConcurrentExpiringPersistentURLMapFromDisk(expiration_check)

package util

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type MapItem interface {
	MapItemToString() string
}

type ConcurrentMap interface {
	Get_Entry(string) (MapItem, error)
	BeginConstruction(int64, ExpiryCallback) ConcurrentMap
	ContinueConstruction(string, string, int64)
	FinishConstruction()
}

type URLMap interface {
	Put_New_Entry(string, interface{}, int64) error
}

type LogStorage interface {
	AppendNewEntry(string, string, int64) error
}

func GetEntryCommon(cm ConcurrentMap, short_url string) (MapItem, error) {
	// Literally just pass it directly to the map. Reads should never hit disk.
	val, err := cm.Get_Entry(short_url)
	if err != nil {
		return nil, err
	}
	return val, err //nolint:forcetypeassert // just let it panic.
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func PutEntry_Common(requested_length int, long_url string, timestamp int64, Generate_strings_up_to int,
	slice_storage map[int]*RandomBag64, urlmap URLMap, b53m *Base53IDManager, log_storage LogStorage) (string, error) {
	if requested_length < 2 { //nolint:gomnd // 2 is not magic here. BASE53 can only go down to 2 characters because it uses one character for the checksum
		return "", errors.New("Requested length is too small.")
	}
	// if length is <= 5, grab it from one of the slices
	if requested_length <= Generate_strings_up_to { //nolint:nestif // yeah it's complicated
		randombag, ok := slice_storage[requested_length]
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
		err = urlmap.Put_New_Entry(id_str, long_url, timestamp)
		if err != nil { // Only possible error is if entry already exists, which it should never do since we got it from the slice.
			log.Fatal("Put_New_Entry failed. This should never happen. Error:", err)
			panic("Put_New_Entry failed. This should never happen. Error:" + err.Error())
		}
		return id_str, nil
	} else { // Otherwise randomly generate it and see if it already exists
		id, err := b53m.B53_generate_random_Base53ID(requested_length)
		if err != nil {
			log.Fatal("Failed to generate new random ID. This should never happen. Error:", err)
			panic(err)
		}
		// try 100 times, trying again when it fails due to already existing in the map
		// probability of failing 100 times in a row should be astronomically small
		for i := 0; i < 100; i++ {
			id_str := id.GetCombinedString()
			err = urlmap.Put_New_Entry(id_str, long_url, timestamp)
			if err == nil {
				// Successfully put it into the map. Now write it to disk too
				// It's okay if this is slow since it's just a write. Most operations are going to be reads.
				err = log_storage.AppendNewEntry(id_str, long_url, timestamp)
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

// This is the one you want to use in production
func LoadStoredRecordsFromDisk(cepum_params *CEPUMParams,
	entry_should_be_ignored_fn func(int64) bool, add_to_map_fn func(string, string, int64), heap_init_fn func(), lss LogStructuredStorage, expiry_callback ExpiryCallback,
	nil_map_ptr ConcurrentMap) (map[int]*RandomBag64, ConcurrentMap, LogStructuredStorage) { //nolint:gocognit // yeah it's complicated
	// First, list all the files in the directory
	entries, err := os.ReadDir(cepum_params.Bucket_directory_path_absolute)
	if err != nil {
		log.Fatal("Failed to open bucket directory:", cepum_params.Bucket_directory_path_absolute, "error:", err)
		panic(err)
	}
	// Now for each file, try to parse the file's filename and load it into the map if it's not expired
	files_to_be_loaded_from := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { // ignore directories
			continue
		}
		// validate file name
		err := lss.ValidateLogFilename(entry.Name())
		if err != nil {
			log.Fatal("Failed to parse name of file in log directory:", entry.Name(), "got error:", err)
			panic(err)
		}

		// add it to the list of files to be loaded from
		absolute_file_path := filepath.Join(cepum_params.Bucket_directory_path_absolute, entry.Name())
		files_to_be_loaded_from = append(files_to_be_loaded_from, absolute_file_path)
	}

	map_size_persister := NewMapSizeFileManager(cepum_params.Size_file_path_absolute, cepum_params.Size_file_rounded_multiple)
	// Load size of map from file
	stored_map_length := map_size_persister.current_rounded_size

	// Create the map and slice efficiently using the loaded rounded size. It's okay if it's too small, since these will grow automatically.
	concurrent_map := nil_map_ptr.BeginConstruction(stored_map_length, expiry_callback)

	// Now load from each file into the map
	lbses := NewLogBucketStructuredExpiringStorage(cepum_params.Bucket_interval, cepum_params.Bucket_directory_path_absolute)

	for _, absolute_filepath := range files_to_be_loaded_from {
		f, err := os.Open(absolute_filepath) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("Failed to open bucket log file:", absolute_filepath, "err:", err)
			panic(err)
		}

		// Now scan the input from the file
		br := bufio.NewReader(f)
		for {
			str_without_hash, err := br.ReadBytes('\r') // TODO: Remove the trailing \r otherwise it will fail haha
			Check_err(err)
			// check if error is EOF
			if errors.Is(err, io.EOF) {
				// make sure we're not waiting for more input
				// If ReadBytes encounters an error before finding a delimiter,
				// it returns the data read before the error and the error itself (often io.EOF).
				if len(str_without_hash) != 0 {
					log.Fatal("File ", absolute_filepath, " does not end with newline, indicating some kind of corruption")
					panic("File doesn't end with newline.")
				}
				break
			}
			if err != nil {
				log.Fatal("Unexpected non-EOF error")
				panic(err)
			}

			md5_base64, err := br.ReadBytes('\n') // TODO: Remove the trailing \n otherwise it will fail haha
			Check_err(err)
			parts := strings.Split(string(str_without_hash), "\t")
			if len(parts) != 3 {
				log.Fatal("Expected 3 parts (key, value, timestamp), got", len(parts))
				panic("Got unexpected number of parts")
			}
			key_str := parts[0]
			value_str := parts[1]
			timestamp_str := parts[2]

			// Check URL ID
			_, err = cepum_params.B53m.NewBase53ID(key_str[:len(key_str)-1], key_str[len(key_str)-1], false)
			if err != nil {
				log.Fatal("Invalid URL ID:", key_str, "Error:", err)
				panic(err)
			}

			// Check md5_base64
			md5_bytes, err := base64.StdEncoding.DecodeString(string(md5_base64))
			if err != nil {
				log.Fatal("Could not decode base64-encoded md5", err)
				panic(err)
			}
			// Now recompute the md5 and check it against the stored value
			recomputed_md5 := md5.Sum(str_without_hash)
			if !bytes.Equal(recomputed_md5[:], md5_bytes) {
				log.Fatalf("md5 does not match. Stored: %s Recomputed: %s", hex.EncodeToString(md5_bytes), hex.EncodeToString(recomputed_md5[:]))
				panic("md5 does not match.")
			}

			// convert timestamp_str to timestamp_unix
			timestamp_unix, err := String_to_int64(string(timestamp_str))
			if err != nil {
				log.Fatal("Could not convert timestamp_str to int64", err)
				panic(err)
			}

			err = Validate_Timestamp_Common(timestamp_unix)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}

			if entry_should_be_ignored_fn != nil {
				ignore_entry := entry_should_be_ignored_fn(timestamp_unix)
				if ignore_entry {
					continue
				}
			}

			// So now we know the entry in the file is not expired.
			// But what if there is already an entry in the map???
			val, err := concurrent_map.Get_Entry(string(key_str)) // if map already contains item, err will be nil
			if err == nil {                                       // This implies that we've already seen a non-expired entry for that URL ID, which should never happen
				log.Fatal("Multiple non-expired entries found in log files for same key string: ", val.MapItemToString(), " key_str: ", string(key_str))
				panic("Multiple non-expired entries found in log files for same URL ID")
			}

			// Insert it into map (and push it into heap for ConcurrentExpiringMap)
			concurrent_map.ContinueConstruction(string(key_str), string(value_str), timestamp_unix)
		}
	}
	// Call heap.Init() for ConcurrentExpiringMap
	concurrent_map.FinishConstruction()

	should_be_added_fn := func(keystr string) bool { // Only add to slice if it's not in the map
		_, err := concurrent_map.Get_Entry(keystr)
		if err != nil && !errors.Is(err, NonExistentKeyError{}) {
			log.Fatal("Unexpected error from Get_Entry", err)
			panic(err)
		}
		return err != nil
	}
	slice_storage := make(map[int]*RandomBag64)
	for n := 2; n <= cepum_params.Generate_strings_up_to; n++ {
		log.Println("Generating all Base 53 IDs of length", n)
		slice, err := cepum_params.B53m.B53_generate_all_Base53IDs_int64_optimized(n, should_be_added_fn) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("B53_generate_all_Base53IDs_int64_optimized failed", err)
			panic("B53_generate_all_Base53IDs_int64_optimized failed: " + err.Error())
		}
		slice_storage[n] = CreateRandomBagFromSlice(slice)
	}

	return slice_storage, concurrent_map, lbses
}
