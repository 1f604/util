// This struct is intended to provide persistence to the ConcurrentExpiringMap
// It allows the server to load from disk to recreate the ConcurrentExpiringMap
// Bucket (log) files are named "bucket_expires_before-18400.log" where the last number is a unix timestamp

// It provides an API that has 3 methods:
// 1. Update map size rounded
// 2. Append new entry to log file
// 3. Delete expired log files
// As well as an "expiring" LogStructuredStorage where you can remove expired entries from old log files and rewrite them into new log files
// The reason this works is because it's okay to see the same entry multiple times since we'll just ignore it when we see the same entry again
// We ignore entries that expire earlier than the current entry we have, and we overwrite the current entry as soon as we see another entry that has a later expiration time
package util

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogBucketStructuredExpiringStorage struct {
	directory_lock                 sync.Mutex
	bucket_interval                int64
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
func NewLogBucketStructuredExpiringStorage(bucket_interval int64, bucket_directory_path_absolute string) *LogBucketStructuredExpiringStorage {
	// check if bucket directory exists
	_, err := os.Stat(bucket_directory_path_absolute)
	if err != nil {
		log.Fatal("Fatal error: Could not stat bucket directory:", err)
		panic(err)
	}

	return &LogBucketStructuredExpiringStorage{
		directory_lock:                 sync.Mutex{},
		bucket_interval:                bucket_interval,
		bucket_directory_path_absolute: bucket_directory_path_absolute,
	}
}

// Adds a new entry to the log file
//
// Also important: Make sure the input does not contain carriage return or newline.
func (lbses *LogBucketStructuredExpiringStorage) AppendNewEntry(key string, value string, expiry_time int64) error {
	lbses.directory_lock.Lock()
	defer lbses.directory_lock.Unlock()
	// Don't check for expiry time
	// If entry is already expired then it will be written to an already expired log file, which will be removed at some point automatically.

	// Find the corresponding bucket number. This should always succeed
	corresponding_bucket_timestamp := ((expiry_time / lbses.bucket_interval) + 1) * lbses.bucket_interval
	bucket_path := filepath.Join(lbses.bucket_directory_path_absolute, LBSES_Get_bucket_filename(corresponding_bucket_timestamp))
	// Find the corresponding log file
	// If it doesn't exist, create it
	// If it does exist, then append to it
	f, err := os.OpenFile(bucket_path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	err = Write_Entry_To_File(key, value, expiry_time, f)
	if err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
		panic(err)
	}
	return nil
}

// Delete expired buckets (log files)
// extra_keeparound_seconds_disk defines how long to keep around log files after they expired
func (lbses *LogBucketStructuredExpiringStorage) DeleteExpiredLogFiles(extra_keeparound_seconds_disk int64) {
	lbses.directory_lock.Lock()
	defer lbses.directory_lock.Unlock()
	// First, list all the files in the directory
	entries, err := os.ReadDir(lbses.bucket_directory_path_absolute)
	if err != nil {
		panic(err)
	}

	cur_timestamp := time.Now().Unix()
	for _, e := range entries {
		if e.IsDir() { // ignore directories
			continue
		}
		// if you can't parse it, raise an error
		expiry_timestamp_unix, err1 := LBSES_Parse_bucket_filename_to_timestamp(e.Name())
		if err1 != nil {
			log.Fatal("Failed to parse name of bucket file:", e.Name(), "got error:", err)
			panic(err1)
		}
		// if it's expired, then delete it
		// add grace period
		if (expiry_timestamp_unix + extra_keeparound_seconds_disk) < cur_timestamp {
			if err = os.Remove(filepath.Join(lbses.bucket_directory_path_absolute, e.Name())); err != nil {
				log.Fatal(err)
				panic(err)
			}
		}
	}
}
