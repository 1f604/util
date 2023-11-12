// Defends against directory traversal attacks
package util

import (
	"fmt"
	"regexp"
	"strings"
)

func PosixValidatedURLDirPath(dirpath_str string) (*PosixValidatedURLDirPath_t, error) {
	// Special check: make sure it contains no dots
	if strings.Contains(dirpath_str, ".") {
		return nil, fmt.Errorf("Error: URL directory path \"%s\" contains dot.", dirpath_str)
	}
	// Regex validation
	// Check that it starts and ends with a slash, and contains no illegal characters
	alnum_char := `[0-9a-zA-Z]`
	alnum_seq := `[0-9a-zA-Z]+`
	alnum_or_sym_seq := `[0-9a-zA-Z_-]+` // don't need to escape dash if it's last character inside square brackets
	one_dir_name_followed_by_slash := `(` + alnum_seq + `|` + alnum_char + alnum_or_sym_seq + alnum_char + `)/`
	regex_str := `^/(` + one_dir_name_followed_by_slash + `)*$`

	// the above regex in one string: `^/(([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_-]+[0-9a-zA-Z])/)*$``
	validationRegex := regexp.MustCompile(regex_str)

	// check regex
	if !validationRegex.MatchString(dirpath_str) {
		return nil, fmt.Errorf("Error: URL directory path \"%s\" does not match URL directory path validation regex %s", dirpath_str, regex_str)
	}

	return &PosixValidatedURLDirPath_t{URLPrefix: &dirpath_str}, nil
}
