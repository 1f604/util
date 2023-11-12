// Defends against directory traversal attacks
package util

import (
	"fmt"
	"strings"
)

// returns nil if no match, or a key inside the map if there is a match.
func matchLongestURLPrefixMap(url_prefixes_map URLPrefixToFileSystemDirectoryMap,
	urlpath PosixValidatedFullURLPath_t) *PosixValidatedURLDirPath_t {
	var best_prefix *PosixValidatedURLDirPath_t = nil
	for prefix := range *url_prefixes_map.Map {
		if strings.HasPrefix(urlpath.URLPath, *prefix.URLPrefix) && best_prefix.Length() < prefix.Length() {
			best_prefix = prefix
		}
	}
	return best_prefix
}

// returns error if there is no match
// swaps the longest matching prefix for the mapped file system directory path and keep the rest of the string the same.
// This brute force implementation is faster than Trie when number of prefixes is less than 200
// And is roughly the same speed as Trie when number of prefixes is 200-300.
// See my Github repo where I do the benchmarks: https://github.com/1f604/longest_prefix_go
func Convert_URLPath_To_Full_FileSystem_Path(posix_validated_urlpath PosixValidatedFullURLPath_t,
	url_to_dir_map URLPrefixToFileSystemDirectoryMap) (*FullFileSystemPath_t, error) {
	// check prefix exists
	// find the matching prefix
	best_prefix := matchLongestURLPrefixMap(url_to_dir_map, posix_validated_urlpath)
	// no match
	if best_prefix == nil {
		return nil, fmt.Errorf("Error: %s prefix not in URL-to-dir map", posix_validated_urlpath.URLPath)
	}
	// match, then swap out the prefix
	best_prefix_str := best_prefix.URLPrefix
	posix_validated_urlpath_str := posix_validated_urlpath.URLPath
	if posix_validated_urlpath_str[:len(*best_prefix_str)] != *best_prefix_str {
		// this should never happen, just panic when it happens.
		panic_str := fmt.Sprintf("Unexpected error: returned prefix \"%s\" is not a prefix of the input string \"%s\"", *best_prefix_str, posix_validated_urlpath_str)
		panic(panic_str)
		// return nil, fmt.Errorf("Unexpected error: ")
	}
	matched_fs := (*url_to_dir_map.Map)[best_prefix]
	matched_fs_str := matched_fs.FileSystemPath

	converted_path := *matched_fs_str + posix_validated_urlpath_str[len(*best_prefix_str):]

	// sanity check
	converted_prefix := converted_path[:len(*matched_fs_str)]
	_, ok := (*url_to_dir_map.ReverseMap)[converted_prefix]
	if !ok {
		panic("This should never happen.")
	}

	return FullFileSystemPath(converted_path), nil
}
