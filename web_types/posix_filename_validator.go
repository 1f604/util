package util

import (
	"fmt"
	"regexp"
)

// Checks that the file name only contains numbers, letters, underscore, dot, and dash.
func Posix_filename_validator(filename string) error {
	alnum_char := `[0-9a-zA-Z]`
	alnum_seq := `[0-9a-zA-Z]+`
	alnum_or_sym_seq := `[0-9a-zA-Z_ ,'.-]+` // don't need to escape dash (-) if it's last character inside square brackets
	regex_str := `^(` + alnum_seq + `|` + alnum_char + alnum_or_sym_seq + alnum_char + `)$`

	// check regex
	validationRegex := regexp.MustCompile(regex_str)
	if !validationRegex.MatchString(filename) {
		return fmt.Errorf("Error: filename \"%s\" does not match filename validation regex %s", filename, regex_str)
	}

	return nil
}
