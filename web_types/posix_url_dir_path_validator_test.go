package util_test

import (
	"fileserver/pkg/util"
	"fmt"
	"testing"

	web_types "fileserver/pkg/util/web_types"
)

func run_test3[T SUCCESS | ERROR | ERROR_DOT](t *testing.T, input string, expected_result T) {
	t.Helper()

	expected_err := fmt.Sprintf("Error: URL directory path \"%s\" does not match URL directory path validation regex ^/(([0-9a-zA-Z]+|[0-9a-zA-Z][0-9a-zA-Z_-]+[0-9a-zA-Z])/)*$", input)
	error_dot := fmt.Sprintf("Error: URL directory path \"%s\" contains dot.", input)

	// fmt.Println("input.URLPath:", input.URLPath)
	_, err := web_types.PosixValidatedURLDirPath(input)
	switch any(expected_result).(type) {
	case SUCCESS: // do nothing
		util.Assert_no_error(t, err, 2)
	case ERROR:
		util.Assert_error_equals(t, err, expected_err, 2)
	case ERROR_DOT:
		util.Assert_error_equals(t, err, error_dot, 2)
	}
}

func Test_Posix_validate_URL_dir_path(t *testing.T) { //nolint:funlen // it's ok, we just have lots of test cases.
	t.Parallel()

	// examples
	run_test3(t, "/", SUCCESS{})
	run_test3(t, "/a/", SUCCESS{})
	run_test3(t, "a", ERROR{})
	run_test3(t, "/a", ERROR{})
	run_test3(t, "a/", ERROR{})
	run_test3(t, "/../", ERROR_DOT{})
	run_test3(t, "//", ERROR{})

	// empty string
	run_test3(t, "", ERROR{})

	// single character
	run_test3(t, "0", ERROR{})
	run_test3(t, " ", ERROR{})
	run_test3(t, "_", ERROR{})
	run_test3(t, "-", ERROR{})
	run_test3(t, ".", ERROR_DOT{})
	// two characters
	run_test3(t, "./", ERROR_DOT{})
	run_test3(t, "/.", ERROR_DOT{})

	// not starting with slash
	run_test3(t, "a", ERROR{})
	run_test3(t, "a/", ERROR{})
	run_test3(t, "a/a", ERROR{})
	run_test3(t, "a/a/", ERROR{})
	run_test3(t, "/a/a/", SUCCESS{})

	// not ending with slash
	run_test3(t, "a", ERROR{})
	run_test3(t, "/a", ERROR{})
	run_test3(t, "a/a", ERROR{})
	run_test3(t, "/a/a", ERROR{})

	// double slash
	run_test3(t, "a//", ERROR{})
	run_test3(t, "//a", ERROR{})
	run_test3(t, "/a//", ERROR{})
	run_test3(t, "//a/", ERROR{})
	run_test3(t, "a//a", ERROR{})
	run_test3(t, "/a//a", ERROR{})
	run_test3(t, "a//a/", ERROR{})
	run_test3(t, "/a//a/", ERROR{})

	// illegal characters
	run_test3(t, "/./", ERROR_DOT{})
	run_test3(t, "/a.a/", ERROR_DOT{})
	run_test3(t, "/a..a/", ERROR_DOT{})
	run_test3(t, "/a/./", ERROR_DOT{})
	run_test3(t, "/a/ ./", ERROR_DOT{})
	run_test3(t, "/a /./", ERROR_DOT{})
	run_test3(t, "/a/.a/", ERROR_DOT{})
	run_test3(t, "/a/a./", ERROR_DOT{})
	run_test3(t, "/a/a.a/", ERROR_DOT{})
	run_test3(t, ".", ERROR_DOT{})
	run_test3(t, "a.", ERROR_DOT{})
	run_test3(t, ".a", ERROR_DOT{})
	run_test3(t, "/.", ERROR_DOT{})
	run_test3(t, "/a.", ERROR_DOT{})
	run_test3(t, "/.a", ERROR_DOT{})
	run_test3(t, "/a/.", ERROR_DOT{})
	run_test3(t, "/./a/", ERROR_DOT{})
	run_test3(t, "/a.a/.", ERROR_DOT{})
	run_test3(t, "/aa/a.a", ERROR_DOT{})

	run_test3(t, "/aa/aa/", SUCCESS{})
	run_test3(t, "_/aa/aa/", ERROR{})
	run_test3(t, "/_aa/aa/", ERROR{})
	run_test3(t, "/a_a/aa/", SUCCESS{})
	run_test3(t, "/aa_/aa/", ERROR{})
	run_test3(t, "/aa/_aa/", ERROR{})
	run_test3(t, "/aa/a_a/", SUCCESS{})
	run_test3(t, "/aa/aa_/", ERROR{})
	run_test3(t, "/aa/aa/_", ERROR{})

	run_test3(t, "./aa/aa/", ERROR_DOT{})
	run_test3(t, "/.aa/aa/", ERROR_DOT{})
	run_test3(t, "/a.a/aa/", ERROR_DOT{})
	run_test3(t, "/aa./aa/", ERROR_DOT{})
	run_test3(t, "/aa/.aa/", ERROR_DOT{})
	run_test3(t, "/aa/a.a/", ERROR_DOT{})
	run_test3(t, "/aa/aa./", ERROR_DOT{})
	run_test3(t, "/aa/aa/.", ERROR_DOT{})

	run_test3(t, " /aa/aa/", ERROR{})
	run_test3(t, "/ aa/aa/", ERROR{})
	run_test3(t, "/a a/aa/", ERROR{})
	run_test3(t, "/aa /aa/", ERROR{})
	run_test3(t, "/aa/ aa/", ERROR{})
	run_test3(t, "/aa/a a/", ERROR{})
	run_test3(t, "/aa/aa /", ERROR{})
	run_test3(t, "/aa/aa/ ", ERROR{})

	// typical sequences
	run_test3(t, "/home/", SUCCESS{})
	run_test3(t, "/eggs/milk.txt", ERROR_DOT{})
	run_test3(t, "eggs/milk.txt", ERROR_DOT{})
	run_test3(t, "/eggs/milk.txt/", ERROR_DOT{})
	run_test3(t, "/eggs/asdf_-/milk.txt", ERROR_DOT{})
	run_test3(t, "/eggs/asdf_-/milk.txt/", ERROR_DOT{})

	// test URL path starts with two slashes
	run_test3(t, "//eggs/milk.txt", ERROR_DOT{})

	// test directory traversal
	run_test3(t, "/meow/a/milk_txt", ERROR{})
	run_test3(t, "/meow/a/milk_txt/", SUCCESS{})
	run_test3(t, "/meow/a..a/milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/a../milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/..a/milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/.../milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/../milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/./milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow//milk_txt/", ERROR{})

	// test name starts with symbols
	run_test3(t, "/a/", SUCCESS{})
	run_test3(t, "/_/", ERROR{})
	run_test3(t, "/-/", ERROR{})
	run_test3(t, "/./", ERROR_DOT{})

	run_test3(t, "/a_a/", SUCCESS{})
	run_test3(t, "/a-a/", SUCCESS{})
	run_test3(t, "/a.a/", ERROR_DOT{})

	run_test3(t, "/a_/", ERROR{})
	run_test3(t, "/_a/", ERROR{})
	run_test3(t, "/a-/", ERROR{})
	run_test3(t, "/-a/", ERROR{})
	run_test3(t, "/a./", ERROR_DOT{})
	run_test3(t, "/.a/", ERROR_DOT{})

	run_test3(t, "/a_a/_", ERROR{})
	run_test3(t, "/a_a/_a", ERROR{})
	run_test3(t, "/a_a/_a_", ERROR{})
	run_test3(t, "/a_a/_/", ERROR{})
	run_test3(t, "/a_a/_a/", ERROR{})
	run_test3(t, "/a_a/_a_/", ERROR{})
	run_test3(t, "/a_a/a_a/", SUCCESS{})
	run_test3(t, "/a_a/a_a_/", ERROR{})
	run_test3(t, "/a_a/a_a_a/", SUCCESS{})
	run_test3(t, "/a_a/a_a__a/", SUCCESS{})

	run_test3(t, "/a-a/-", ERROR{})
	run_test3(t, "/a-a/-a", ERROR{})
	run_test3(t, "/a-a/-a-", ERROR{})
	run_test3(t, "/a-a/-/", ERROR{})
	run_test3(t, "/a-a/-a/", ERROR{})
	run_test3(t, "/a-a/-a-/", ERROR{})
	run_test3(t, "/a-a/a-a/", SUCCESS{})
	run_test3(t, "/a-a/a-a-/", ERROR{})
	run_test3(t, "/a-a/a-a-a/", SUCCESS{})
	run_test3(t, "/a-a/a-a--a/", SUCCESS{})

	run_test3(t, "/a.a/.", ERROR_DOT{})
	run_test3(t, "/a.a/.a", ERROR_DOT{})
	run_test3(t, "/a.a/.a.", ERROR_DOT{})
	run_test3(t, "/a.a/./", ERROR_DOT{})
	run_test3(t, "/a.a/.a/", ERROR_DOT{})
	run_test3(t, "/a.a/.a./", ERROR_DOT{})
	run_test3(t, "/a.a/a.a/", ERROR_DOT{})
	run_test3(t, "/a.a/a.a./", ERROR_DOT{})
	run_test3(t, "/a.a/a.a.a/", ERROR_DOT{})
	run_test3(t, "/a.a/a.a..a/", ERROR_DOT{})

	run_test3(t, "/meow/cat/_milk_txt/", ERROR{})
	run_test3(t, "/meow/cat/a_milk_txt/", SUCCESS{})
	run_test3(t, "/meow/cat/a.milk_txt/", ERROR_DOT{})
	run_test3(t, "/meow/cat/a-milk_txt/", SUCCESS{})
}
