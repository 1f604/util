// Buckets are directories that contain files
// The name of a bucket is the expiry time (unix) of that bucket
// The idea is that when a bucket expires it should be deleted

// It provides an API that has 3 methods:
// 1. InsertFile(contents, expiry_time) returns file path
// 2. GetFile(file_path) returns contents of file
// 3. Delete expired buckets

package util

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"golang.org/x/sys/unix"
)

type ExpiringBucketStorage struct {
	mut                            sync.Mutex
	bucket_directory_path_absolute string
}

// the bucket interval is the all-important parameter that determines the number of buckets and when buckets will be deleted
// the bucket interval is in Unix time (seconds).
// it means that all entries between two time points will go into one bucket
// when that bucket expires, it will be deleted
// example: if bucket interval is 100, then all timestamps from 0 to 100 will go into one bucket, all timestamps 100 to 200 will go into next bucket and so on
// bucketing is done simply by the / (round-to-zero division) operation.
// expiry time will be divided by the bucket interval and placed into appropriate bucket (log file)
// e.g. if bucket interval is 100, then all timestamps from 86400 to 86500 will go into bucket 865
// e.g. if bucket interval is 200, then all timestamps from 1200 to 1400 will go into bucket 7, all timestamps from 1400 to 1600 will go to bucket 8 and so on.
// e.g. if bucket interval is 200, then bucket 200 holds all timestamps 0-199, bucket 400 holds all timestamps 200-399, bucket 600 holds 400-599, and so on.
// bucket files are named "expires_before_18400" where the last number is a unix timestamp
func NewExpiringBucketStorage(bucket_interval int64, bucket_directory_path_absolute string) *ExpiringBucketStorage {
	// check if bucket directory exists
	// create it if it doesn't exist.
	err := os.MkdirAll(bucket_directory_path_absolute, os.ModePerm)
	if err != nil {
		log.Fatal("Fatal error: Could not create directory:", err)
		panic(err)
	}

	return &ExpiringBucketStorage{
		mut:                            sync.Mutex{},
		bucket_directory_path_absolute: bucket_directory_path_absolute,
	}
}

func GetPasteFileName_Common(prefix string, file_contents []byte, timestamp int64) string {
	// we use sha1 to detect corruption because it's fast - 16 bytes is enough.
	hash_bytes := sha1.Sum(file_contents)
	// convert hash to printable string
	hex_sha1 := hex.EncodeToString(hash_bytes[:])[:8]
	rand_string := Crypto_Rand_Alnum_String(8) //nolint:gomnd // 8 characters is more than we need but birthday paradox means that collisions are more likely than they seem...

	return prefix + Int64_to_string(timestamp) + "_sha1_" + hex_sha1 + "_rand_" + rand_string
}

// Adds a new entry to the log file
//
// Also important: Make sure the input does not contain carriage return or newline.
func (ebs *ExpiringBucketStorage) InsertFile(file_contents []byte, expiry_time int64, xattr_params *XattrParams) string {
	ebs.mut.Lock()
	defer ebs.mut.Unlock()
	// Don't check expiry time. Just put it.
	// Compute md5 hash of file
	// we use md5 to detect corruption - 16 bytes is enough.

	// Now generate a new filename that doesn't already exist
	// Just generate a random 8-character string, should be good enough
	var absfilepath string
	for count := 0; count < 10; count++ {
		absfilepath = filepath.Join(ebs.bucket_directory_path_absolute, GetPasteFileName_Common("expires_at_", file_contents, expiry_time))
		// Check if file already exists
		f, err := os.OpenFile(absfilepath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		// If file already exists try again
		if err == nil {
			_, err := f.Write(file_contents)
			if err != nil {
				log.Fatal(err)
				panic(err)
			}
			if err = f.Close(); err != nil {
				log.Fatal(err)
				panic(err)
			}
			if xattr_params.SetXattr {
				err = unix.Setxattr(absfilepath, xattr_params.XattrName, []byte(xattr_params.Xattrvalue), 0)
				if err != nil {
					log.Fatal(err)
					panic(err)
				}
			}
			return absfilepath
		}
		log.Println("Unexpected collision occurred!!!", absfilepath)
	}
	log.Fatal("Tried 10 times to write new file, all failed. Is the disk full?")
	panic("This should never happen.")
}

var bucket_dir_name_pattern = `^([0-9]+)$`
var bucket_dir_name_regex = regexp.MustCompile(bucket_dir_name_pattern)

func parse_bucket_filename_to_timestamp(filename string) (int64, error) {
	// use capture groups

	// caps is a slice of strings, where caps[0] matches the whole match
	// caps[1] == "202" etc
	matches := bucket_dir_name_regex.FindStringSubmatch(filename)
	if matches == nil {
		return -1, errors.New("Failed to parse bucket")
	}
	if len(matches) != 2 {
		fmt.Println("matches:", matches)
		return -1, errors.New("Expected exactly 2 matches")
	}
	match := matches[1]
	num, err := String_to_int64(match)
	return num, err
}

// Delete expired buckets (directories)
// extra_keeparound_seconds_disk defines how long to keep around buckets after they expired
// Redundant: You should never need to use this function if you're using CEPUM
/* It wont'work now anyway.
func (ebs *ExpiringBucketStorage) DeleteExpiredBuckets() {
	ebs.mut.Lock()
	defer ebs.mut.Unlock()
	// First, list all of the directories and then delete the ones that are expired.
	// First, list all the dirs in the directory
	entries, err := os.ReadDir(ebs.bucket_directory_path_absolute)
	if err != nil {
		panic(err)
	}

	cur_timestamp := time.Now().Unix()
	for _, e := range entries {
		if !e.IsDir() { // ignore files
			continue
		}
		// if you can't parse it, raise an error
		expiry_timestamp_unix, err1 := parse_bucket_filename_to_timestamp(e.Name())
		if err1 != nil {
			log.Fatal("Failed to parse name of bucket file:", e.Name(), "got error:", err)
			panic(err1)
		}
		// if it's expired, then delete it
		// add grace period
		if (expiry_timestamp_unix + ebs.extra_keeparound_seconds_disk) < cur_timestamp {
			absdirpath := filepath.Join(ebs.bucket_directory_path_absolute, e.Name())
			log.Println("Deleting dir ", absdirpath)
			log.Println("Current time:", time.Now().Unix())
			if err = os.RemoveAll(absdirpath); err != nil {
				log.Fatal(err)
				panic(err)
			}
		}
	}
}
*/
