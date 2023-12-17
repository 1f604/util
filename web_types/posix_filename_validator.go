package util

import (
	"fmt"
	"regexp"
)

var g_posix_alnum_char = `[0-9a-zA-Z]`
var g_posix_alnum_seq = `[0-9a-zA-Z]+`
var g_posix_alnum_or_sym_seq = `[0-9a-zA-Z_ ,'.-]+` // don't need to escape dash (-) if it's last character inside square brackets
var g_posix_regex_str = `^(` + g_posix_alnum_seq + `|` + g_posix_alnum_char + g_posix_alnum_seq + g_posix_alnum_char + `)$`
var g_validationRegex = regexp.MustCompile(g_posix_regex_str)

// Checks that the file name only contains numbers, letters, underscore, dot, and dash.
func Posix_filename_validator(filename string) error {
	// check regex
	if !g_validationRegex.MatchString(filename) {
		return fmt.Errorf("Error: filename \"%s\" does not match filename validation regex %s", filename, g_validationRegex)
	}

	return nil
}
