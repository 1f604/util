// Buckets are directories that contain files
// The name of a bucket is the expiry time (unix) of that bucket
// The idea is that when a bucket expires it should be deleted

// It provides an API that has 3 methods:
// 1. InsertFile(contents, expiry_time) returns file path
// 2. GetFile(file_path) returns contents of file
// 3. Delete expired buckets

package util

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PermanentBucketStorage struct {
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
func NewPermanentBucketStorage(bucket_directory_path_absolute string) *PermanentBucketStorage {
	// check if bucket directory exists
	// create it if it doesn't exist.
	err := os.MkdirAll(bucket_directory_path_absolute, os.ModePerm)
	if err != nil {
		log.Fatal("Fatal error: Could not create directory:", err)
		panic(err)
	}

	return &PermanentBucketStorage{
		mut:                            sync.Mutex{},
		bucket_directory_path_absolute: bucket_directory_path_absolute,
	}
}

func (pbs *PermanentBucketStorage) InsertFile(file_contents []byte, _ int64) string {
	pbs.mut.Lock()
	defer pbs.mut.Unlock()

	// Now generate a new filename that doesn't already exist
	// Just generate a random 8-character string, should be good enough
	var absfilepath string
	cur_timestamp := time.Now().Unix()
	for count := 0; count < 10; count++ {
		absfilepath = filepath.Join(pbs.bucket_directory_path_absolute, GetPasteFileName_Common("created_at_", file_contents, cur_timestamp))
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
			return absfilepath
		}
		log.Println("Unexpected collision occurred!!!", absfilepath)
	}
	log.Fatal("Tried 10 times to write new file, all failed. Is the disk full?")
	panic("This should never happen.")
}
