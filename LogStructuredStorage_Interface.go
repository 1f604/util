// This struct is intended to provide persistence to the ConcurrentExpiringMap
// It allows the server to load from disk to recreate the ConcurrentExpiringMap
// Bucket (log) files are named "bucket_expires_before-18400.log" where the last number is a unix timestamp

// This file provides 2 structs: a "base" LogStructuredStorage for storing permanent data
// As well as an "expiring" LogStructuredStorage where you can remove expired entries from old log files and rewrite them into new log files
// The reason this works is because it's okay to see the same entry multiple times since we'll just ignore it when we see the same entry again
// We ignore entries that expire earlier than the current entry we have, and we overwrite the current entry as soon as we see another entry that has a later expiration time
package util

import (
	"fmt"
)

type LogStructuredStorage interface {
	ValidateLogFilename(filename string) error
}

// Can be called with nil receiver.
func (*LogBucketStructuredExpiringStorage) ValidateLogFilename(filename string) error {
	timestamp, err := LBSES_Parse_bucket_filename_to_timestamp(filename)
	if err != nil {
		return err
	}
	err = Validate_Timestamp_Common(timestamp)
	if err != nil {
		return err
	}
	return nil
}

// Can be called with nil receiver.
func (*LogStructuredPermanentStorage) ValidateLogFilename(filename string) error {
	number, err := LSPS_Parse_log_filename_to_number(filename)
	if err != nil {
		return err
	}
	if number < 0 {
		return fmt.Errorf("Error: File number %#v is less than zero", number)
	}
	return nil
}
