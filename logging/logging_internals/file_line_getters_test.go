// Unit tests for logging helper functions
package util_test

import (
	"log"
	"os"
	"testing"

	"github.com/1f604/util"
	. "github.com/1f604/util"
	logging_internals "github.com/1f604/util/logging/logging_internals"
)

type (
	test_fn      func(*os.File) ([]byte, error)
	testlogic_fn func(*os.File)
)

func test_helper(testfunc testlogic_fn, contents string) {
	// first create the file
	path := "./testfile"
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0o666)
	util.Check_err(err) // panic if file already exists
	// If we successfully created the file, then make sure to delete it afterwards
	defer func() {
		if err = os.Remove(path); err != nil {
			log.Fatal(err)
			panic(err)
		}
	}()
	// Now we write the contents into the file
	_, err = f.WriteString(contents)
	util.Check_err(err)
	// Now we run the test function
	testfunc(f)
}

// hooray for generics!!!
func assert_func_returns[RESULT_or_ERROR RESULT | ERROR](t *testing.T, testfunc test_fn, filecontents string, expected_result RESULT_or_ERROR) {
	t.Helper()

	test_logic := func(f *os.File) {
		result, err := testfunc(f)
		// simple workaround: use any
		switch v := any(expected_result).(type) {
		case RESULT:
			s, _ := v.S.(string) // we know it is string
			Assert_result_equals_bytes(t, result, err, s, 4)
		case ERROR:
			Assert_error_equals(t, err, v.S, 4)
		}
	}
	test_helper(test_logic, filecontents)
}

var (
	func_get_first_line_from_file test_fn = logging_internals.Get_first_line_from_file
	func_get__last_line_from_file test_fn = logging_internals.Get_last_nonempty_line_from_file
)

func Test_All(t *testing.T) {
	t.Parallel()
	/*
		Being able to see each test case on a single line is so fucking good,
		because it allows me to see AT A GLANCE exactly what test cases I have and exactly what behavior I expect in each case.
	*/
	assert_func_returns(t, func_get_first_line_from_file, "", Error("File is empty."))
	assert_func_returns(t, func_get__last_line_from_file, "", Error("File is empty."))

	assert_func_returns(t, func_get_first_line_from_file, "hello world!", Error("File contains no newlines."))
	assert_func_returns(t, func_get__last_line_from_file, "hello world!", Error("No newline in file or file only contains newlines or only one line."))

	assert_func_returns(t, func_get_first_line_from_file, "\nhello world!", Error("First line of file is empty."))
	assert_func_returns(t, func_get__last_line_from_file, "\nhello world!", Result("hello world!"))
	assert_func_returns(t, func_get_first_line_from_file, "\n\n\nhello world!", Error("First line of file is empty."))
	assert_func_returns(t, func_get__last_line_from_file, "\n\n\nhello world!", Result("hello world!"))

	assert_func_returns(t, func_get_first_line_from_file, "hello world!\n", Result("hello world!"))
	assert_func_returns(t, func_get__last_line_from_file, "hello world!\n", Error("No newline in file or file only contains newlines or only one line."))
	assert_func_returns(t, func_get_first_line_from_file, "hello world!\n\n\n", Result("hello world!"))
	assert_func_returns(t, func_get__last_line_from_file, "hello world!\n\n\n", Error("No newline in file or file only contains newlines or only one line."))

	assert_func_returns(t, func_get_first_line_from_file, "\nhello world!\n", Error("First line of file is empty."))
	assert_func_returns(t, func_get__last_line_from_file, "\nhello world!\n", Result("hello world!"))
	assert_func_returns(t, func_get_first_line_from_file, "\n\n\nhello world!\n\n\n", Error("First line of file is empty."))
	assert_func_returns(t, func_get__last_line_from_file, "\n\n\nhello world!\n\n\n", Result("hello world!"))

	assert_func_returns(t, func_get_first_line_from_file, "hello\nhello world!", Result("hello"))
	assert_func_returns(t, func_get__last_line_from_file, "hello\nhello world!", Result("hello world!"))

	assert_func_returns(t, func_get_first_line_from_file, "hello\nhello world!\n", Result("hello"))
	assert_func_returns(t, func_get__last_line_from_file, "hello\nhello world!\n", Result("hello world!"))
}
