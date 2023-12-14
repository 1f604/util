// The purpose of this file is for storing the map size on disk.
// The purpose of the map size file is to make it faster to load entries from disk into map
// By storing the size of the map in the file, the next time on program on startup creates the map, the make() function can be given the correct size
// It doesn't matter if it's slightly too big or small since it won't affect the performance much.
// Better to be too big than too small to avoid resizing which is costly
package util

import (
	"errors"
	"log"
	"os"
	"sync"
)

type MapSizeFileManager struct {
	mut                     sync.Mutex
	size_multiple           int64
	current_rounded_size    int64
	size_file_path_absolute string
}

func NewMapSizeFileManager(size_file_path_absolute string, size_multiple int64) *MapSizeFileManager {
	// Get current rounded size
	// Try to open the file
	// First, create the size file if it doesn't exist
	// Check if it exists using os.stat
	_, err := os.Stat(size_file_path_absolute)
	if err != nil {
		// if it doesn't exist then create it
		log.Println("Size file doesn't exist, creating it...")
		f, err := os.OpenFile(size_file_path_absolute, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) //nolint:govet // ignore err shadow
		Check_err(err)
		// set it to the size_growth_amount to begin with.
		_, err = f.WriteString(Int64_to_string(size_multiple))
		Check_err(err)
		// close the file!
		err = f.Close()
		Check_err(err)
	}
	// Now get the current rounded size
	current_rounded_size, err := _internal_get_current_rounded_size(size_file_path_absolute, size_multiple)
	if err != nil {
		log.Fatal("Get_current_rounded_size error: ", err)
		panic(err)
	}

	return &MapSizeFileManager{
		size_multiple:           size_multiple,
		size_file_path_absolute: size_file_path_absolute,
		current_rounded_size:    current_rounded_size,
	}
}

func _internal_get_current_rounded_size(size_file_path string, size_multiple int64) (int64, error) {
	// try to open it
	buf, err := os.ReadFile(size_file_path)
	if err != nil {
		return -1, err
	}
	// try to convert it into int64
	num, err := String_to_int64(string(buf))
	if err != nil {
		return -1, err
	}
	// Check that it's a multiple of size_growth_amount
	if num < size_multiple || num%size_multiple != 0 {
		return -1, errors.New("Contents of file is not a positive multiple of size_growth_amount")
	}
	// all is fine, so return the number
	return num, nil
}

// rounds to the nearest size
//
// You can call this function as many times as you like, since it only updates file if rounded size has changed.
func (msfm *MapSizeFileManager) UpdateMapSizeRounded(actual_size int64) {
	msfm.mut.Lock()
	defer msfm.mut.Unlock()

	number_to_round := actual_size + msfm.size_multiple/2 //nolint:gomnd // 2 is a good number
	// Now round it to the nearest size
	rounded_size := ((number_to_round / msfm.size_multiple) + 1) * msfm.size_multiple

	if rounded_size > msfm.current_rounded_size {
		msfm.current_rounded_size = rounded_size
	} else { // do nothing
		return
	}
	// If updated current_rounded_size, write it into file.
	err := os.WriteFile(msfm.size_file_path_absolute, []byte(Int64_to_string(msfm.current_rounded_size)), 0644)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
