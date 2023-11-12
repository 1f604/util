package util_test

import (
	"fmt"
	"testing"

	"github.com/1f604/util"
	web_types "github.com/1f604/util/web_types"
)

func run_test4[T SUCCESS | ERROR](t *testing.T, input string, expected_result T) {
	t.Helper()

	expected_err := fmt.Sprintf("Error: filename \"%s\" does not match filename validation regex ^([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_ ,'.-]+[0-9a-zA-Z])$", input)

	// fmt.Println("input.URLPath:", input.URLPath)
	err := web_types.Posix_filename_validator(input)
	switch any(expected_result).(type) {
	case SUCCESS: // do nothing
		util.Assert_no_error(t, err, 2)
	case ERROR:
		util.Assert_error_equals(t, err, expected_err, 2)
	}
}

func Test_filename_validator(t *testing.T) { //nolint:funlen // is okay
	t.Parallel()

	// example usage
	run_test4(t, "SoMe_file_123.hahaha-okay_la", SUCCESS{})
	run_test4(t, "/hello", ERROR{})
	run_test4(t, "hello/", ERROR{})
	run_test4(t, "hello/world", ERROR{})
	run_test4(t, "hello world", SUCCESS{})
	run_test4(t, ".hello", ERROR{})
	run_test4(t, "hello.", ERROR{})
	run_test4(t, "_hello", ERROR{})
	run_test4(t, "hello_", ERROR{})

	// empty string
	run_test4(t, "", ERROR{})

	// single character
	run_test4(t, "a", SUCCESS{})
	run_test4(t, "0", SUCCESS{})
	run_test4(t, "/", ERROR{})
	run_test4(t, ".", ERROR{})
	run_test4(t, "-", ERROR{})
	run_test4(t, "_", ERROR{})

	// two characters
	// same
	run_test4(t, "aa", SUCCESS{})
	run_test4(t, "//", ERROR{})
	run_test4(t, "..", ERROR{})
	run_test4(t, "__", ERROR{})
	run_test4(t, "--", ERROR{})
	// different
	run_test4(t, "/a", ERROR{})
	run_test4(t, "a/", ERROR{})
	run_test4(t, "_a", ERROR{})
	run_test4(t, "a_", ERROR{})
	run_test4(t, ".a", ERROR{})
	run_test4(t, "a.", ERROR{})
	run_test4(t, "-a", ERROR{})
	run_test4(t, "a-", ERROR{})
	run_test4(t, "/.", ERROR{})
	run_test4(t, "./", ERROR{})

	// three characters - slash, dot, underscore, dash
	// no symbols
	run_test4(t, "aaa", SUCCESS{})
	// one symbol
	run_test4(t, "/aa", ERROR{})
	run_test4(t, "a/a", ERROR{})
	run_test4(t, "aa/", ERROR{})
	// two symbol
	run_test4(t, "a//", ERROR{})
	run_test4(t, "/a/", ERROR{})
	run_test4(t, "//a", ERROR{})
	// three symbol
	run_test4(t, "///", ERROR{})

	// one symbol
	run_test4(t, "-aa", ERROR{})
	run_test4(t, "a-a", SUCCESS{})
	run_test4(t, "aa-", ERROR{})
	// two symbol
	run_test4(t, "a--", ERROR{})
	run_test4(t, "-a-", ERROR{})
	run_test4(t, "--a", ERROR{})
	// three symbol
	run_test4(t, "---", ERROR{})

	// one symbol
	run_test4(t, ".aa", ERROR{})
	run_test4(t, "a.a", SUCCESS{})
	run_test4(t, "aa.", ERROR{})
	// two symbol
	run_test4(t, "a..", ERROR{})
	run_test4(t, ".a.", ERROR{})
	run_test4(t, "..a", ERROR{})
	// three symbol
	run_test4(t, "...", ERROR{})

	// one symbol
	run_test4(t, "_aa", ERROR{})
	run_test4(t, "a_a", SUCCESS{})
	run_test4(t, "aa_", ERROR{})
	// two symbol
	run_test4(t, "a__", ERROR{})
	run_test4(t, "_a_", ERROR{})
	run_test4(t, "__a", ERROR{})
	// three symbol
	run_test4(t, "___", ERROR{})

	run_test4(t, "helloworld!", ERROR{})
	run_test4(t, "-hello", ERROR{})
	run_test4(t, "hello-", ERROR{})
	run_test4(t, "SoMe_file_123.hahaha-okay_", ERROR{})
}
