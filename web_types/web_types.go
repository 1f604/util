// This file contains types that are useful for webservers

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"

	json_internals "fileserver/pkg/util/json/json_internals"
)

type PosixValidatedFullURLPath_t struct {
	URLPath string
}

type FullFileSystemPath_t struct {
	FileSystemPath *string
}

func FullFileSystemPath(s string) *FullFileSystemPath_t {
	return &FullFileSystemPath_t{FileSystemPath: &s}
}

type PosixValidatedURLDirPath_t struct {
	URLPrefix *string
}

type UrlPrefixToFSDirMap_internal_t map[*PosixValidatedURLDirPath_t]*FullFileSystemPath_t

type URLPrefixToFileSystemDirectoryMap struct {
	Map        *UrlPrefixToFSDirMap_internal_t
	ReverseMap *map[string]bool // used as a set of all the file system directories defined by the user
}

type URLRedirectMap struct { // maps source to destination URL...
	Map *map[string]string
}

func (the_map *URLRedirectMap) UnmarshalJSON(data []byte) error {
	pairs := new([]*URLRedirectPair) // new can never return a nil ptr

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields() // This line is super critical. It makes the decoder error on JSON fields that aren't in the user struct.
	err := decoder.Decode(pairs)
	if err != nil {
		return err
	}
	// Check pairs for nil pointers
	err = json_internals.Recursively_Check_for_nil_pointers(pairs, "pairs")
	if err != nil {
		return err
	}

	m := make(map[string]string)
	the_map.Map = &m
	for _, pair := range *pairs {
		(*the_map.Map)[*pair.SourceURL] = *pair.DestinationURL
	}

	return nil
}

func (the_map *URLPrefixToFileSystemDirectoryMap) UnmarshalJSON(data []byte) error {
	pairs := new([]*URLPrefixFileSystemDirectoryPair) // new can never return a nil ptr

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields() // This line is super critical. It makes the decoder error on JSON fields that aren't in the user struct.
	err := decoder.Decode(pairs)
	if err != nil {
		return err
	}
	// Check pairs for nil pointers
	err = json_internals.Recursively_Check_for_nil_pointers(pairs, "pairs")
	if err != nil {
		return err
	}

	m := make(UrlPrefixToFSDirMap_internal_t)
	the_map.Map = &m
	rm := make(map[string]bool)
	the_map.ReverseMap = &rm
	for _, pair := range *pairs {
		if pair == nil || pair.URLPrefix == nil || pair.FileSystemPath == nil {
			// should never happen since we check it.
			panic("This should never happen.")
		}
		(*the_map.Map)[pair.URLPrefix] = pair.FileSystemPath
		(*the_map.ReverseMap)[*pair.FileSystemPath.FileSystemPath] = true
	}

	return nil
}

type URLPrefixFileSystemDirectoryPair struct {
	URLPrefix      *PosixValidatedURLDirPath_t `json:"url_prefix"`
	FileSystemPath *FullFileSystemPath_t       `json:"file_system_directory_path"`
}

type URLRedirectPair struct {
	SourceURL      *string `json:"source_path"`
	DestinationURL *string `json:"destination_path"`
}

// should we panic here? Ideally we should panic in this kind of situation.
func (filesystempath *FullFileSystemPath_t) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	// Standardize it by ensuring the file system path ends with a slash
	// Check length is greater than 0
	if len(str) < 1 {
		return errors.New("File system directory path must contain at least 1 character: " + str)
	}
	// Check last character is a slash
	if str[len(str)-1] != '/' {
		return errors.New("File system directory path must end with slash: " + str)
	}

	filesystempath.FileSystemPath = &str
	return nil
}

// should we panic here? Ideally we should panic in this kind of situation.
func (urlprefix *PosixValidatedURLDirPath_t) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	// Standardize it by ensuring the URL prefix has a leading and trailing slash
	// Check length is greater than 0
	if len(str) < 1 {
		return errors.New("URL Prefix must contain at least 1 character.")
	}

	// Check first character is a slash
	if str[0] != '/' {
		return errors.New("URL Prefix must start with slash: " + str)
	}
	// Check last character is a slash
	if str[len(str)-1] != '/' {
		return errors.New("URL Prefix must end with slash: " + str)
	}

	// Is valid POSIX URL dir
	result, err := PosixValidatedURLDirPath(str)
	if err != nil {
		return err
	}

	urlprefix.URLPrefix = result.URLPrefix
	return nil
}

// In Go, it is valid to call a method on a nil pointer!!!
func (p *PosixValidatedURLDirPath_t) Length() int {
	if p == nil {
		return 0
	} else {
		return len(*p.URLPrefix)
	}
}

type TimeDurationSeconds struct {
	Duration *time.Duration
}

func (timeduration *TimeDurationSeconds) UnmarshalJSON(bytes []byte) error {
	var num float64
	err := json.Unmarshal(bytes, &num)
	if err != nil {
		return err
	}
	temp := time.Duration(num * float64(time.Second))
	timeduration.Duration = &temp
	return nil
}
