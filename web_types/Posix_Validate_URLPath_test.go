package util_test

import (
	"fileserver/pkg/util"
	"fmt"
	"path/filepath"
	"testing"

	web_types "fileserver/pkg/util/web_types"
)

func run_test2[T SUCCESS | ERROR_DIR | ERROR_FILENAME | ERROR_DOT](t *testing.T, input string, expected_result T) {
	t.Helper()

	dirpart, filenamepart := filepath.Split(input)

	// fmt.Println("input.URLPath:", input.URLPath)
	_, err := web_types.PosixValidatedFullURLPath(input)
	switch any(expected_result).(type) {
	case SUCCESS: // do nothing
		util.Assert_no_error(t, err, 2)
	case ERROR_DIR:
		expected_err := fmt.Sprintf("Error: URL directory path \"%s\" does not match URL directory path validation regex ^/(([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_-]+[0-9a-zA-Z])/)*$", dirpart)
		util.Assert_error_equals(t, err, expected_err, 2)
	case ERROR_FILENAME:
		expected_err := fmt.Sprintf("Error: filename \"%s\" does not match filename validation regex ^([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_ ,'.-]+[0-9a-zA-Z])$", filenamepart)
		util.Assert_error_equals(t, err, expected_err, 2)
	case ERROR_DOT:
		error_dot := fmt.Sprintf("Error: URL directory path \"%s\" contains dot.", dirpart)
		util.Assert_error_equals(t, err, error_dot, 2)
	}
}

func Test_Posix_Validate_url_path(t *testing.T) {
	t.Parallel()

	// These are just sanity checks. Check the other tests for the actual full testing.

	// control
	run_test2(t, "", ERROR_FILENAME{})
	run_test2(t, "a", ERROR_DIR{})
	run_test2(t, "/", ERROR_FILENAME{})
	run_test2(t, "//", ERROR_FILENAME{})
	run_test2(t, "/a/", ERROR_FILENAME{})
	run_test2(t, "/milk.txt", SUCCESS{})
	run_test2(t, "/../milk", ERROR_DOT{})
	run_test2(t, "../milk", ERROR_DOT{})
	run_test2(t, "./milk", ERROR_DOT{})
	run_test2(t, "/a/milk.txt", SUCCESS{})
	run_test2(t, "/a//milk.txt", ERROR_DIR{})
	run_test2(t, "/eggs/asdf_/milk.txt", ERROR_DIR{})
	run_test2(t, "/eggs/_asdf/milk.txt", ERROR_DIR{})

	// test URL path starts with no slashes
	run_test2(t, "eggs/milk.txt", ERROR_DIR{})

	// test URL path starts with two slashes
	run_test2(t, "//eggs/milk.txt", ERROR_DIR{})

	// test directory traversal
	run_test2(t, "/meow/a..a/milk.txt", ERROR_DOT{})
	run_test2(t, "/meow/a../milk.txt", ERROR_DOT{})
	run_test2(t, "/meow/..a/milk.txt", ERROR_DOT{})
	run_test2(t, "/meow/.../milk.txt", ERROR_DOT{})
	run_test2(t, "/meow/../milk.txt", ERROR_DOT{})
	run_test2(t, "/meow/./milk.txt", ERROR_DOT{})
	run_test2(t, "/meow//milk.txt", ERROR_DIR{})

	// test directory leading slash
	run_test2(t, "/meow/milk.txt", SUCCESS{})
	run_test2(t, "/cool/cat/milk.txt", SUCCESS{})
	run_test2(t, "cool/cat/milk.txt", ERROR_DIR{})

	// test filename starts with symbols
	run_test2(t, "meow/cat/.milk.txt", ERROR_FILENAME{})
	run_test2(t, "meow/cat/-milk.txt", ERROR_FILENAME{})
	run_test2(t, "meow/cat/_milk.txt", ERROR_FILENAME{})
}
