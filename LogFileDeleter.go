// Deletes the oldest files until the size of directory is back within limit
package util

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type LogFileDeleter struct {
	AbsoluteDirectoryPath   string
	CurrentLogFileName      string
	DirectorySizeLimitBytes int64
}

func NewLogFileDeleter(directory_path string, size_limit int64, log_file_name string) *LogFileDeleter {
	return &LogFileDeleter{
		AbsoluteDirectoryPath:   directory_path,
		DirectorySizeLimitBytes: size_limit,
		CurrentLogFileName:      log_file_name,
	}
}

func (lfd *LogFileDeleter) RunThread(time_interval_secs int) {
	for range time.Tick(time.Second * time.Duration(time_interval_secs)) {
		lfd.Delete_Excess_Files()
	}
}

func try_get_timestamp_from_filename(filename string) time.Time {
	firstpart, _, found := strings.Cut(filename, "$$")
	if !found {
		panic("File does not contain $$")
	}
	result, err := String_to_int64(firstpart)
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

type FileEntry struct {
	FilePath    string
	FileInfo    fs.FileInfo
	TimeCreated time.Time
}

func (lfd *LogFileDeleter) Delete_Excess_Files() {
	// First, get a list of files in the directory
	entries, err := os.ReadDir(lfd.AbsoluteDirectoryPath)
	if err != nil {
		panic(err)
	}

	file_entries := []FileEntry{}
	var total_directory_size int64 = 0

	for _, e := range entries {
		if e.IsDir() { // ignore directories
			continue
		}
		if e.Name() == lfd.CurrentLogFileName {
			continue
		}
		time_file_created := try_get_timestamp_from_filename(e.Name())
		file_info, err1 := e.Info()
		Check_err(err1)

		total_directory_size += file_info.Size()

		file_entries = append(file_entries, FileEntry{FilePath: filepath.Join(lfd.AbsoluteDirectoryPath, e.Name()), FileInfo: file_info, TimeCreated: time_file_created})

		fmt.Println(e.Name())
	}
	// fmt.Println("total_directory_size:", total_directory_size)
	sort.Slice(file_entries, func(i, j int) bool {
		return file_entries[i].TimeCreated.Before(file_entries[j].TimeCreated)
	})
	// the file_entries is sorted from earliest to latest, so first entry is the oldest file so we start deleting from there
	// while total size is greater than desired, delete

	for i := 0; total_directory_size > lfd.DirectorySizeLimitBytes; i++ {
		// try delete
		file_entry := file_entries[i]
		size_of_file_deleted := file_entry.FileInfo.Size()
		err1 := os.Remove(file_entry.FilePath)
		Check_err(err1)
		total_directory_size -= size_of_file_deleted
		log.Println("Log file deleted:", file_entry.FilePath)
		// fmt.Println("total_directory_size:", total_directory_size)
	}

	// fmt.Println(file_entries)
}

/* Manual Test Logs:
1699744929$$log.2023-11-11T23:22:09Z
1699744976$$log.2023-11-11T23:22:56Z
total_directory_size: 1358
Log file deleted: /opt/1f604_fileserver/logs/1699744301$$log.2023-11-11T23:11:41Z
total_directory_size: 1193
Log file deleted: /opt/1f604_fileserver/logs/1699744302$$log.2023-11-11T23:11:42Z
total_directory_size: 1028
Log file deleted: /opt/1f604_fileserver/logs/1699744303$$log.2023-11-11T23:11:43Z
total_directory_size: 863
[{/opt/1f604_fileserver/logs/1699744301$$log.2023-11-11T23:11:41Z
*/
