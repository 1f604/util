package util

import (
	"errors"
	"io"
	"os"

	"fileserver/pkg/util"
)

// This function returns the last non-empty line in the file
// It will return an error if there's only one line in the file, or if the file is empty, or if the file only contains newlines
// Make sure every read is preceded by a seek.
func Get_last_nonempty_line_from_file(f *os.File) ([]byte, error) {
	// Note that here I am finding the offset of the newline first, and then reading the file again from that offset to EOF
	// It may be faster to just build a string from the characters and then reverse that string at the end
	// Might be worth benchmarking to see which approach is faster
	// read from back to front
	fi, err := f.Stat()
	util.Check_err(err)
	filesize := fi.Size()
	if filesize == 0 {
		return nil, errors.New("File is empty.")
	}

	b := make([]byte, 1)
	var pos int64 = filesize - 1
	success := false
	seen_non_newline := false
	prev_newline_pos := filesize
	for ; pos >= 0; pos-- {
		_, err = f.Seek(pos, 0)
		util.Check_err(err)
		_, err = f.Read(b)
		util.Check_err(err)
		if b[0] == '\n' {
			if seen_non_newline { // we only break at newline characters
				success = true
				break
			}
			prev_newline_pos = pos
		} else {
			seen_non_newline = true
		}
	}
	if !success {
		return nil, errors.New("No newline in file or file only contains newlines or only one line.")
	}
	lastline := make([]byte, prev_newline_pos-pos-1)
	_, err = f.Seek(pos+1, 0)
	util.Check_err(err)
	_, err = f.Read(lastline)
	util.Check_err(err)
	return lastline, nil
}

// This function returns the first line in the file
// It will return an error if the file is empty or contains no newlines
// Make sure every read is preceded by a seek.
func Get_first_line_from_file(f *os.File) ([]byte, error) {
	fi, err := f.Stat()
	util.Check_err(err)
	if fi.Size() == 0 {
		return nil, errors.New("File is empty.")
	}

	b := make([]byte, 1)
	success := false
	var pos int64 = 0
	_, err = f.Seek(0, 0)
	util.Check_err(err)
	for {
		_, err = f.Read(b)
		if errors.Is(err, io.EOF) { // don't panic if it's just EOF
			break
		}
		util.Check_err(err)
		if b[0] == '\n' {
			success = true
			break
		}
		pos++
	}
	if !success {
		return nil, errors.New("File contains no newlines.")
	}
	if pos == 0 {
		return nil, errors.New("First line of file is empty.")
	}
	firstline := make([]byte, pos)
	_, err = f.Seek(0, 0)
	util.Check_err(err)
	_, err = f.Read(firstline)
	util.Check_err(err)
	return firstline, nil
}
