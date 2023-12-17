// This struct is intended to provide persistence to the ConcurrentExpiringMap
// It allows the server to load from disk to recreate the ConcurrentExpiringMap
// Bucket (log) files are named "bucket_expires_before-18400.log" where the last number is a unix timestamp

// This file provides 2 structs: a "base" LogStructuredStorage for storing permanent data
// As well as an "expiring" LogStructuredStorage where you can remove expired entries from old log files and rewrite them into new log files
// The reason this works is because it's okay to see the same entry multiple times since we'll just ignore it when we see the same entry again
// We ignore entries that expire earlier than the current entry we have, and we overwrite the current entry as soon as we see another entry that has a later expiration time
package util

import (
	"crypto/md5"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
)

func LBSES_Get_bucket_filename(timestamp int64) string {
	return "bucket_expires_before-" + Int64_to_string(timestamp) + ".log"
}

var g_bucket_name_pattern = `^bucket_expires_before-([0-9]+)\.log$`
var g_bucket_name_regex = regexp.MustCompile(g_bucket_name_pattern)

func LBSES_Parse_bucket_filename_to_timestamp(filename string) (int64, error) {
	// use capture groups

	// caps is a slice of strings, where caps[0] matches the whole match
	// caps[1] == "202" etc
	matches := g_bucket_name_regex.FindStringSubmatch(filename)
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

func Compute_String_Checksum(str string) string {
	// TODO: FInish this
	return ""
}

// IMPORTANT: This function DOES NOT close the file handle!!!
func Write_Entry_To_File(key string, value string, timestamp int64, file_handle *os.File) error {
	// Generate the bytes to write to the file
	// validate key first
	for _, c := range key {
		if c == '\n' || c == '\t' || c == '\x1e' {
			return errors.New("Error: key contains newline or tab or x1e:" + string("c"))
		}
	}
	// validate value
	for _, c := range value {
		if c == '\n' || c == '\t' || c == '\x1e' {
			return errors.New("Error: value contains newline or tab or x1e:" + string("c"))
		}
	}
	// we use md5 to detect corruption - 16 bytes is enough.
	str_to_sum := key + string("\t") + value + string("\t") + Int64_to_string(timestamp)
	hash_bytes := md5.Sum([]byte(str_to_sum))
	hash_base64 := b64.StdEncoding.EncodeToString(hash_bytes[:])
	// convert hash to printable string
	string_to_write := str_to_sum + "\x1e" + hash_base64 + string("\n")
	if _, err := file_handle.WriteString(string_to_write); err != nil {
		log.Fatal(err)
		panic(err)
	}
	return nil
}
