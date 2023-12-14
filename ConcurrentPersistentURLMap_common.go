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
	GetValue() string
}

type ConcurrentMap interface {
	Get_Entry(string) (MapItem, error)
	BeginConstruction(int64, ExpiryCallback) ConcurrentMap
	ContinueConstruction(string, string, int64)
	FinishConstruction()
	NumItems() int
}

type URLMap interface {
	Put_New_Entry(string, string, int64) error
	NumItems() int
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
	return val, err
}

// Shorten long URL into short URL and return the short URL and store the entry both in map and on disk
func PutEntry_Common(requested_length int, long_url string, timestamp int64, generate_strings_up_to int,
	slice_storage map[int]*RandomBag64, urlmap URLMap, b53m *Base53IDManager, log_storage LogStorage, map_size_persister *MapSizeFileManager) (string, error) {
	if requested_length < 2 { //nolint:gomnd // 2 is not magic here. BASE53 can only go down to 2 characters because it uses one character for the checksum
		return "", errors.New("Requested length is too small.")
	}
	// if length is <= 5, grab it from one of the slices
	var result_str string
	if requested_length <= generate_strings_up_to { //nolint:nestif // yeah it's complicated
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
		result_str = Convert_uint64_to_str(item, requested_length)
		err = urlmap.Put_New_Entry(result_str, long_url, timestamp)
		if err != nil { // Only possible error is if entry already exists, which it should never do since we got it from the slice.
			log.Fatal("Put_New_Entry failed. This should never happen. Error:", err)
			panic("Put_New_Entry failed. This should never happen. Error:" + err.Error())
		}
		// Successfully put it into the map. Now write it to disk too
		goto added_item_to_map
	} else { // Otherwise randomly generate it and see if it already exists
		// try 100 times, trying again when it fails due to already existing in the map
		// probability of failing 100 times in a row should be astronomically small
		for i := 0; i < 100; i++ {
			id, err := b53m.B53_generate_random_Base53ID(requested_length)
			if err != nil {
				log.Fatal("Failed to generate new random ID. This should never happen. Error:", err)
				panic(err)
			}
			result_str = id.GetCombinedString()
			err = urlmap.Put_New_Entry(result_str, long_url, timestamp)
			if err == nil {
				// Successfully put it into the map. Now write it to disk too
				goto added_item_to_map
			}
			if i > 3 { //nolint:gomnd // 3 is a good number.
				log.Println("Unexpected event: got duplicate ID", i, "times in a row. ID is:", result_str)
			}
		}
		log.Fatal("Failed to generate new random string 100 times, this should never happen")
		panic("Failed to generate new random string 100 times, this should never happen")
	}
	// Successfully put it into the map. Now write it to disk too
added_item_to_map:
	// Update the size file if necessary
	// log.Print("urlmap.NumItems():", urlmap.NumItems())
	map_size_persister.UpdateMapSizeRounded(int64(urlmap.NumItems()))
	// It's okay if this is slow since it's just a write. Most operations are going to be reads.
	err := log_storage.AppendNewEntry(result_str, long_url, timestamp)
	// log.Println("calling log_storage.AppendNewEntry(result_str, long_url, timestamp)")
	if err != nil {
		// It should never fail.
		log.Fatal("AppendNewEntry failed:", err)
		panic(err)
	}
	return result_str, nil
}

type NonExistentKeyError interface {
	NonExistentKeyError() string
}

type LSRFD_Params struct {
	B53m                        *Base53IDManager
	Log_directory_path_absolute string
	Size_file_path_absolute     string
	Entry_should_be_ignored_fn  func(int64) bool
	Lss                         LogStructuredStorage
	Expiry_callback             ExpiryCallback
	Slice_storage               map[int]*RandomBag64
	Nil_ptr                     ConcurrentMap
	Size_file_rounded_multiple  int64
	Generate_strings_up_to      int
}

// This is the one you want to use in production
func LoadStoredRecordsFromDisk(params *LSRFD_Params) (ConcurrentMap, *MapSizeFileManager) { //nolint:gocognit,ireturn // yeah, it is complicated...
	// First, list all the files in the directory
	entries, err := os.ReadDir(params.Log_directory_path_absolute)
	if err != nil {
		log.Fatal("Failed to open bucket directory:", params.Log_directory_path_absolute, "error:", err)
		panic(err)
	}
	// Now for each file, try to parse the file's filename and load it into the map if it's not expired
	files_to_be_loaded_from := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() { // ignore directories
			continue
		}
		// validate file name
		err = params.Lss.ValidateLogFilename(entry.Name())
		if err != nil {
			log.Fatal("Failed to parse name of file in log directory:", entry.Name(), "got error:", err)
			panic(err)
		}

		// add it to the list of files to be loaded from
		absolute_file_path := filepath.Join(params.Log_directory_path_absolute, entry.Name())
		files_to_be_loaded_from = append(files_to_be_loaded_from, absolute_file_path)
	}

	map_size_persister := NewMapSizeFileManager(params.Size_file_path_absolute, params.Size_file_rounded_multiple)
	// Load size of map from file
	stored_map_length := map_size_persister.current_rounded_size

	// Create the map and slice efficiently using the loaded rounded size. It's okay if it's too small, since these will grow automatically.
	concurrent_map := params.Nil_ptr.BeginConstruction(stored_map_length, params.Expiry_callback)

	for _, absolute_filepath := range files_to_be_loaded_from {
		f, err := os.Open(absolute_filepath) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("Failed to open bucket log file:", absolute_filepath, "err:", err)
			panic(err)
		}

		// Now scan the input from the file
		br := bufio.NewReader(f)
		for {
			str_without_hash, err := br.ReadBytes('\x1e')
			//			log.Println("str_without_hash, err:", str_without_hash, err)
			if len(str_without_hash) > 0 {
				str_without_hash = str_without_hash[:len(str_without_hash)-1] // Remove trailing \r
			}
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

			md5_base64, err := br.ReadBytes('\n')
			if err != nil {
				log.Println("Failed on file:", absolute_filepath)
			}
			Check_err(err)
			md5_base64 = md5_base64[:len(md5_base64)-1] // remove trailing newline
			parts := strings.Split(string(str_without_hash), "\t")
			if len(parts) != 3 { //nolint:gomnd // 3 is okay here...
				log.Fatal("Expected 3 parts (key, value, timestamp), got", len(parts))
				panic("Got unexpected number of parts")
			}
			key_str := parts[0]
			value_str := parts[1]
			timestamp_str := parts[2]

			// Check URL ID
			_, err = params.B53m.NewBase53ID(key_str[:len(key_str)-1], key_str[len(key_str)-1], false)
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
			recomputed_md5 := md5.Sum(str_without_hash) //nolint:gosec // md5 is fine here.
			if !bytes.Equal(recomputed_md5[:], md5_bytes) {
				log.Fatalf("md5 does not match. Stored: %s Recomputed: %s", hex.EncodeToString(md5_bytes), hex.EncodeToString(recomputed_md5[:]))
				panic("md5 does not match.")
			}

			// convert timestamp_str to timestamp_unix
			timestamp_unix, err := String_to_int64(timestamp_str)
			if err != nil {
				log.Fatal("Could not convert timestamp_str to int64", err)
				panic(err)
			}

			err = Validate_Timestamp_Common(timestamp_unix)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}

			if params.Entry_should_be_ignored_fn != nil {
				ignore_entry := params.Entry_should_be_ignored_fn(timestamp_unix)
				if ignore_entry {
					continue
				}
			}

			// So now we know the entry in the file is not expired.
			// But what if there is already an entry in the map???
			val, err := concurrent_map.Get_Entry(key_str) // if map already contains item, err will be nil
			if err == nil {                               // This implies that we've already seen a non-expired entry for that URL ID, which should never happen
				log.Fatal("Multiple non-expired entries found in log files for same key string: ", val.MapItemToString(), " key_str: ", key_str)
				panic("Multiple non-expired entries found in log files for same URL ID")
			}

			// Insert it into map (and push it into heap for ConcurrentExpiringMap)
			concurrent_map.ContinueConstruction(key_str, value_str, timestamp_unix)
		}
	}
	// Call heap.Init() for ConcurrentExpiringMap
	concurrent_map.FinishConstruction()

	should_be_added_fn := func(keystr string) bool { // Only add to slice if it's not in the map
		_, err := concurrent_map.Get_Entry(keystr) //nolint:govet // shadow is okay here.
		if err != nil {
			switch err.(type) { //nolint:errorlint // just let it fail
			case CPMNonExistentKeyError, CEMNonExistentKeyError:
				// okay, good
			default:
				log.Fatal("Unexpected error from Get_Entry", err)
				panic(err)
			}
		}
		return err != nil
	}
	for n := 2; n <= params.Generate_strings_up_to; n++ {
		log.Println("Generating all Base 53 IDs of length", n)
		slice, err := params.B53m.B53_generate_all_Base53IDs_int64_optimized(n, should_be_added_fn) //nolint:govet // ignore err shadow
		if err != nil {
			log.Fatal("B53_generate_all_Base53IDs_int64_optimized failed", err)
			panic("B53_generate_all_Base53IDs_int64_optimized failed: " + err.Error())
		}
		params.Slice_storage[n] = CreateRandomBagFromSlice(slice)
	}

	if !IsSameType(concurrent_map, params.Nil_ptr) {
		log.Fatalf("concurrent_map is of type %T while nil_ptr is of type %T", concurrent_map, params.Nil_ptr)
		panic("Not same type.")
	}
	map_size_persister.UpdateMapSizeRounded(int64(concurrent_map.NumItems()))
	return concurrent_map, map_size_persister
}
