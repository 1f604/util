// Defends against directory traversal attacks
package util

import (
	"fmt"
	"regexp"
	"strings"
)

// Regex validation
// Check that it starts and ends with a slash, and contains no illegal characters
var posix_url_dir_alnum_char = `[0-9a-zA-Z]`
var posix_url_dir_alnum_seq = `[0-9a-zA-Z]+`
var posix_url_dir_alnum_or_sym_seq = `[0-9a-zA-Z_-]+` // don't need to escape dash if it's last character inside square brackets
var posix_url_dir_one_dir_name_followed_by_slash = `(` + posix_url_dir_alnum_seq + `|` + posix_url_dir_alnum_char + posix_url_dir_alnum_or_sym_seq + posix_url_dir_alnum_char + `)/`
var posix_url_dir_regex_str = `^/(` + posix_url_dir_one_dir_name_followed_by_slash + `)*$`

// the above regex in one string: `^/(([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_-]+[0-9a-zA-Z])/)*$“
var posix_url_dir_validationRegex = regexp.MustCompile(posix_url_dir_regex_str)

func PosixValidatedURLDirPath(dirpath_str string) (*PosixValidatedURLDirPath_t, error) {
	// Special check: make sure it contains no dots
	if strings.Contains(dirpath_str, ".") {
		return nil, fmt.Errorf("Error: URL directory path \"%s\" contains dot.", dirpath_str)
	}

	// check regex
	if !posix_url_dir_validationRegex.MatchString(dirpath_str) {
		return nil, fmt.Errorf("Error: URL directory path \"%s\" does not match URL directory path validation regex %s", dirpath_str, posix_url_dir_validationRegex)
	}

	return &PosixValidatedURLDirPath_t{URLPrefix: &dirpath_str}, nil
}
