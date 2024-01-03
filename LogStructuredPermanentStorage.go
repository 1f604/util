// Very simple append-only log. The only operation is AppendNewEntry.

// It provides an API that has 3 methods:
// 1. Update map size rounded
// 2. Append new entry to log file
// 3. Delete expired log files
// As well as an "expiring" LogStructuredStorage where you can remove expired entries from old log files and rewrite them into new log files
// The reason this works is because it's okay to see the same entry multiple times since we'll just ignore it when we see the same entry again
// We ignore entries that expire earlier than the current entry we have, and we overwrite the current entry as soon as we see another entry that has a later expiration time
package util

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

type LogStructuredPermanentStorage struct {
	directory_lock              sync.Mutex
	log_file_max_size           int64
	log_directory_path_absolute string
	current_log_filepath        string
	current_log_file_handle     *os.File
}

// Works just like the log rotation library - once log file reaches the max size, create a new log file
// Except we don't need any clever naming scheme, just an increasing number will do, since we're going to read in every file on startup anyway
// The increasing number naming scheme is actually good for cloud backups since we can just send the highest numbered file every time
func NewLogStructuredPermanentStorage(log_file_max_size int64, log_directory_path_absolute string) *LogStructuredPermanentStorage {
	// check if log directory exists
	_, err := os.Stat(log_directory_path_absolute)
	if err != nil {
		log.Fatal("Fatal error: Could not stat log directory:", err)
		panic(err)
	}
	// list all the files in the directory and find the file with the highest numbered name
	// the file names should be "1.log", "2.log", "3.log" and so on
	entries, err := os.ReadDir(log_directory_path_absolute)
	if err != nil {
		log.Fatal("Failed to open log_directory_path_absolute:", log_directory_path_absolute, "error:", err)
		panic(err)
	}
	// Now try to parse each file's filename
	// Find the name of the file with the biggest number
	var biggest_numbered_filename string
	var biggest_seen_number int64 = 0
	for _, entry := range entries {
		if entry.IsDir() { // ignore directories
			continue
		}
		// if you can't parse it, raise an error
		number, err := LSPS_Parse_log_filename_to_number(entry.Name())
		if err != nil {
			log.Fatal("Failed to parse name of file in bucket directory:", entry.Name(), "got error:", err)
			panic(err)
		}
		if number > biggest_seen_number {
			biggest_seen_number = number
			biggest_numbered_filename = entry.Name()
		}
	}
	// If there are no entries, then create 0.log
	if len(biggest_numbered_filename) == 0 {
		biggest_numbered_filename = "0.log"
	}

	fmt.Println("biggest_numbered_filename:", biggest_numbered_filename)
	current_log_filepath_absolute := filepath.Join(log_directory_path_absolute, biggest_numbered_filename)
	// We should keep track of the file size too, so that we rotate it when we get to max size
	fh, err := os.OpenFile(current_log_filepath_absolute, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644) // open in append mode, create if not already exist
	if err != nil {
		log.Fatal(err)
		panic("ERROR: FAILED TO OPEN/CREATE NEW LOG FILE!!!")
	}

	return &LogStructuredPermanentStorage{
		directory_lock:              sync.Mutex{},
		log_file_max_size:           log_file_max_size,
		log_directory_path_absolute: log_directory_path_absolute,
		current_log_filepath:        current_log_filepath_absolute,
		current_log_file_handle:     fh,
	}
}

// Adds a new entry to the log file
//
// Also important: Make sure the input does not contain carriage return or newline.
func (lsps *LogStructuredPermanentStorage) AppendNewEntry(key string, value string, value_type MapItemValueType, generation_time_unix int64) error {
	lsps.directory_lock.Lock()
	defer lsps.directory_lock.Unlock()
	// Write to the log file unless the log file size is too big, in which case we create a new log file and write to that one
	// Get the current log file size
	file_size := Get_file_size(lsps.current_log_file_handle)
	if file_size > lsps.log_file_max_size { // Rotate the log file
		// Create new log file and point to that instead
		dir_part, cur_log_filename := filepath.Split(lsps.current_log_filepath)
		file_number, err := LSPS_Parse_log_filename_to_number(cur_log_filename)
		Check_err(err)
		file_number++
		// Remember to close the existing handle before creating a new one
		err = lsps.current_log_file_handle.Close()
		Check_err(err)
		// Check if new file already exists, if so panic
		new_file_name := Int64_to_string(file_number) + ".log"
		new_file_path := filepath.Join(dir_part, new_file_name)
		_, err = os.Stat(new_file_path)
		if !errors.Is(err, os.ErrNotExist) { // if it exists, then panic
			log.Fatal("This shouldn't happen. Log file ", cur_log_filename, " already exists.")
			panic("This shouldn't happen. Log file already exists.")
		}
		// Otherwise create the file and point the pointers to it
		fh, err := os.OpenFile(new_file_path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644) // open in append mode, create if not already exist
		Check_err(err)
		lsps.current_log_filepath = new_file_path
		lsps.current_log_file_handle = fh
	}
	return Write_Entry_To_File(key, value, value_type, generation_time_unix, lsps.current_log_file_handle)
}

var g_lsps_log_name_pattern = `^([0-9]+)\.log$`
var g_lsps_log_name_regex = regexp.MustCompile(g_lsps_log_name_pattern)

func LSPS_Parse_log_filename_to_number(filename string) (int64, error) {
	// use capture groups

	// caps is a slice of strings, where caps[0] matches the whole match
	// caps[1] == "202" etc
	matches := g_lsps_log_name_regex.FindStringSubmatch(filename)
	if matches == nil {
		return -1, errors.New("Failed to parse file name")
	}
	if len(matches) != 2 {
		fmt.Println("matches:", matches)
		return -1, errors.New("Expected exactly 2 matches")
	}
	match := matches[1]
	num, err := String_to_int64(match)
	return num, err
}
