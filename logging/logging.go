// We ALWAYS start every log message with the UnixMicro timestamp followed by a space followed by the UTC timestamp in RFC3339
// This is so that this package can parse the log file later and name the new log file appropriately
// VERY IMPORTANT NOTE: The log file max size should be large enough that each log file can contain at least two log file entries, otherwise it will cause an error.

package util

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1f604/util"
	logging_internals "github.com/1f604/util/logging/logging_internals"
	web_types "github.com/1f604/util/web_types"
)

func create_logging_dir_if_not_exists(logging_dir string) {
	err := os.MkdirAll(logging_dir, 0o755) // The execute bit on a directory allows you to access items that are inside the directory
	util.Check_err(err)
}

type RotateWriter struct {
	lock                    sync.Mutex
	maxfilesize_bytes       int64 // maximum allowed size of a log file in bytes
	DirectorySizeLimitBytes int64
	LogFilePath             string
	logfiledir              string
	logfileprefix           string
	fp                      *os.File
}

// Make a new RotateWriter. Return nil if error occurs during setup.
func NewRotateWriter(logfilepath string, loggingdir string, logfileprefix string, maxfilesize_bytes int64, directorymaxsize_bytes int64) *RotateWriter {
	// Try to create the log file
	fp, err := os.OpenFile(logfilepath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o644) // open in append mode, create if not already exist
	if err != nil {
		panic("ERROR: FAILED TO OPEN/CREATE NEW LOG FILE!!!")
	}

	w := &RotateWriter{lock: sync.Mutex{}, LogFilePath: logfilepath, logfiledir: loggingdir, logfileprefix: logfileprefix, maxfilesize_bytes: maxfilesize_bytes,
		DirectorySizeLimitBytes: directorymaxsize_bytes, fp: fp}
	return w
}

const (
	log_msg_timestamp_readable      = "2006-Jan-02T15:04:05.000Z07:00" // You can't use JAN, you have to use Jan. It doesn't recognize capitalized letter months. Lame.
	log_filename_timestamp_readable = "2006-Jan-02T15:04:05Z07:00"
)

// Write satisfies the io.Writer interface.
func (w *RotateWriter) Write(dangerous_string []byte) (int, error) {
	// first, get rid of the newline at the end
	// This means it doesn't matter if you use log.Print or log.Println since we remove any newline at the end of the argument
	if len(dangerous_string) == 0 {
		// don't do anything with it
		return 0, errors.New("Error: Cannot write empty string")
	}
	if dangerous_string[len(dangerous_string)-1] == '\n' {
		dangerous_string = dangerous_string[:len(dangerous_string)-1]
	}
	// next, sanitize the output by escaping all non-ASCII and non-printable symbols (including newline) as well as backslash
	encoded_string := logging_internals.Encode_log_msg(dangerous_string)

	// calls to write are serialized so only one thread is in write at any given moment
	// prefix the log message with the current timestamp
	cur_time := time.Now().UTC()
	// First add the UnixMicro timestamp
	prefix_array := append([]byte(strconv.FormatInt(cur_time.UnixMicro(), 10)), byte(' '))
	// Then add the RFC3339 timestamp in UTC - with slight modification
	prefix_array = cur_time.AppendFormat(prefix_array, log_msg_timestamp_readable)

	prefix_array = append(prefix_array, []byte(" (UTC) ")...)
	// Finally add the original message
	prefix_array = append(prefix_array, encoded_string...)
	// Now add the newline at the end
	prefix_array = append(prefix_array, '\n')
	// check if the log file size has exceeded the limit to decide whether to rotate the log
	filesize := util.Get_file_size(w.fp)
	if filesize+int64(len(prefix_array)) >= w.maxfilesize_bytes { // we need to rotate the log
		w.Rotate()
	}

	fmt.Print("log msg: ", string(prefix_array)) // this is safe because the user input has already been encoded at this point.
	// now write the line to the log file
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.fp.Write(prefix_array)
}

func (w *RotateWriter) get_new_filename() string {
	/*
	   Algorithm:
	   	1. Parse out first and last timestamp of file
	   	2. Then create a filename with first and last date as the filename
	*/
	var newfilename string
	var first_timestamp time.Time
	var last_timestamp time.Time

	failed_getting_timestamps := false
	firstline, err := logging_internals.Get_first_line_from_file(w.fp)
	if err != nil {
		fmt.Println("FAILED TO GET FIRST LINE FROM FILE", err) // can't print to log here because we are already in the logging function
		panic(err)
		failed_getting_timestamps = true
	}
	lastline, err := logging_internals.Get_last_nonempty_line_from_file(w.fp)
	if err != nil {
		fmt.Println("FAILED TO GET LAST LINE FROM FILE", err)
		failed_getting_timestamps = true
	}
	if !failed_getting_timestamps {
		// try to parse the timestamps
		first_timestamp, err = logging_internals.Try_parse_log_file_line(string(firstline))
		if err != nil {
			fmt.Println("FAILED TO PARSE FIRST TIMESTAMP", err)
			failed_getting_timestamps = true
		}
		last_timestamp, err = logging_internals.Try_parse_log_file_line(string(lastline))
		if err != nil {
			fmt.Println("FAILED TO PARSE LAST TIMESTAMP", err)
			failed_getting_timestamps = true
		}
	}

	// use proper path combination
	curtime := time.Now().UTC()
	logfilepath := filepath.Join(w.logfiledir, util.Int64_to_string(curtime.Unix())+"$$"+w.logfileprefix)

	if failed_getting_timestamps {
		fmt.Println("An error occurred while fetching first and last timestamps, defaulting to current timestamp instead")
		newfilename = logfilepath + "." + curtime.Format(time.RFC3339)
	} else {
		fmt.Println("Using first and last timestamp from log file")
		newfilename = logfilepath + "." + first_timestamp.Format(log_filename_timestamp_readable) + "$-$" + last_timestamp.Format(log_filename_timestamp_readable)
	}
	return newfilename
}

// This struct is copied from LogFileDeleter.go
type FileEntry struct {
	FilePath    string
	FileInfo    fs.FileInfo
	TimeCreated time.Time
}

// This function is copied from LogFileDeleter.go
func try_get_timestamp_from_filename(filename string) time.Time {
	firstpart, _, found := strings.Cut(filename, "$$")
	if !found {
		panic("File does not contain $$: " + filename)
	}
	result, err := util.String_to_int64(firstpart)
	if err != nil {
		panic("First part of file name is not a number")
	}
	if result < 1699737340 { //nolint: gomnd // see msg
		panic("File unix timestamp is in the past")
	}
	if result > 569724873339 { //nolint: gomnd // see msg
		panic("File unix timestamp is beyond the year 20,000")
	}
	return time.Unix(result, 0)
}

// This function is copied from LogFileDeleter.go
// DO NOT CALL log.Print from this function - it will result in infinite recursion thus causing the program to get stuck for no apparent reason.
func (w *RotateWriter) delete_Excess_Files() {
	// First, get a list of files in the directory
	entries, err := os.ReadDir(w.logfiledir)
	if err != nil {
		panic(err)
	}

	file_entries := []FileEntry{}
	var total_directory_size int64 = 0

	current_log_filename := filepath.Base(w.LogFilePath)
	for _, e := range entries {
		if e.IsDir() { // ignore directories
			continue
		}
		if e.Name() == current_log_filename {
			continue
		}
		time_file_created := try_get_timestamp_from_filename(e.Name())
		file_info, err1 := e.Info()
		util.Check_err(err1)

		total_directory_size += file_info.Size()

		file_entries = append(file_entries, FileEntry{FilePath: filepath.Join(w.logfiledir, e.Name()), FileInfo: file_info, TimeCreated: time_file_created})

		// fmt.Println(e.Name())
	}
	// fmt.Println("total_directory_size:", total_directory_size)
	sort.Slice(file_entries, func(i, j int) bool {
		return file_entries[i].TimeCreated.Before(file_entries[j].TimeCreated)
	})
	// the file_entries is sorted from earliest to latest, so first entry is the oldest file so we start deleting from there
	// while total size is greater than desired, delete

	for i := 0; total_directory_size > w.DirectorySizeLimitBytes; i++ {
		// try delete
		file_entry := file_entries[i]
		size_of_file_deleted := file_entry.FileInfo.Size()
		err1 := os.Remove(file_entry.FilePath)
		util.Check_err(err1)
		total_directory_size -= size_of_file_deleted
		// fmt.Println("Log file deleted:", file_entry.FilePath)
		// fmt.Println("total_directory_size:", total_directory_size)
	}

	// fmt.Println(file_entries)
}

// Perform the actual act of creating a new file and closing the existing file.
// Panics if anything goes wrong.
func (w *RotateWriter) Rotate() {
	w.lock.Lock()
	defer w.lock.Unlock()
	/*
	   Algorithm:
	   	1. Parse out first and last timestamp of file
	   	2. Then create a filename with first and last date as the filename
	   	3. Rename the current file to that
	*/
	// Attempt to parse first and last timestamp in file
	newfilename := w.get_new_filename() // we need to call this before closing the file

	// Close existing file
	err := w.fp.Close()
	w.fp = nil
	if err != nil {
		panic("ERROR: FAILED TO CLOSE LOG FILE!!!")
	}

	// Rename current log file.
	err = os.Rename(w.LogFilePath, newfilename)
	if err != nil {
		panic("ERROR: FAILED TO RENAME LOG FILE!!!")
	}

	// Create a new log file.
	w.fp, err = os.Create(w.LogFilePath)
	if err != nil {
		panic("ERROR: FAILED TO CREATE NEW LOG FILE!!!")
	}

	// Delete excess files.
	w.delete_Excess_Files()
}

func Set_up_logging_panic_on_err(logging_dir string, filename string, logfileprefix string, maxlogfilesize_bytes int64, directorymaxsize_bytes int64) {
	create_logging_dir_if_not_exists(logging_dir)
	// check log file name
	err := web_types.Posix_filename_validator(filename)
	util.Check_err(err)
	// check log file prefix
	if strings.Contains(logfileprefix, ".") {
		panic("Error: Log file prefix cannot contain dot.")
	}
	err = web_types.Posix_filename_validator(logfileprefix)
	util.Check_err(err)

	filename = filepath.Base(filename)
	// use proper path combination
	logfilepath := filepath.Join(logging_dir, filename)
	logger := NewRotateWriter(logfilepath, logging_dir, logfileprefix, maxlogfilesize_bytes, directorymaxsize_bytes)
	log.SetFlags(log.Llongfile) // tell the logger to only log the file name where the log.print function is called, we'll add in the date manually.

	log.SetOutput(logger)
}
