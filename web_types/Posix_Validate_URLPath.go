// Defends against directory traversal attacks
package util

import (
	"fmt"
	"path/filepath"
)

// Generic url path sanitizer: ensures only POSIX symbols in path and defends against directory traversal.
// You should call this function on a file path before passing the path to http.ServeFile.
func PosixValidatedFullURLPath(path_str string) (*PosixValidatedFullURLPath_t, error) {
	// performs simple POSIX validation of a URL path string, mostly just checks for illegal characters
	// Split the path into the directory and the filename
	dirpath, filename := filepath.Split(path_str)
	// Step 1. Check the file name is valid
	err := Posix_filename_validator(filename)
	if err != nil {
		return nil, err
	}
	// Step 2. Check the directory path is valid
	_, err = PosixValidatedURLDirPath(dirpath)
	if err != nil {
		return nil, err
	}
	// Step 3. Sanity check
	joined := filepath.Join(dirpath, filename)
	if joined != path_str {
		return nil, fmt.Errorf("Unexpected Error: input string %s does not equal joined string %s", path_str, joined)
	}
	// And that's it!
	return &PosixValidatedFullURLPath_t{URLPath: path_str}, nil
}
