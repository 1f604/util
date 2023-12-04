package util

import (
	"bytes"
	"runtime"
	"testing"
	"time"
)

type RESULT struct {
	S interface{}
}

type ERROR struct {
	S string
}

func Error(s string) ERROR {
	return ERROR{S: s}
}

func Result(s interface{}) RESULT {
	return RESULT{S: s}
}

func Assert_no_error(t *testing.T, err error, skip_level int) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(skip_level) // lol you have to use this to get the line number lmao

	if err != nil {
		t.Fatal("Line number:", line_number, "Test failed. Expected no error but an error occurred:", err)
	}
}

func Assert_error_equals(t *testing.T, err error, expected string, skip_level int) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(skip_level)

	if err == nil {
		t.Fatal("Line number:", line_number, "Test failed. Expected error but no error occurred.")
	}
	if err.Error() != expected {
		t.Fatal("Line number:", line_number, "Test failed. Expected error", expected, "got", err.Error())
	}
}

func Assert_result_equals_bytes(t *testing.T, actual []byte, err error, expected string, skip_level int) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(skip_level)

	if err != nil {
		t.Fatal("Line number:", line_number, "Fail. Expected no error, got", err.Error())
	}
	if !bytes.Equal(actual, []byte(expected)) {
		t.Fatal("Line number:", line_number, "Fail. Strings not equal. Expected", expected, "got", string(actual))
	}
}

func Assert_result_equals_time(t *testing.T, actual time.Time, err error, expected time.Time, line_number int) {
	t.Helper()

	if err != nil {
		t.Fatal("Line number:", line_number, "Fail. Expected no error, got", err.Error())
	}
	if !actual.Equal(expected) {
		t.Fatal("Line number:", line_number, "Fail. Times not equal. Expected:", expected, "got:", actual)
	}
}

func Assert_result_equals_bool(t *testing.T, actual bool, err error, expected bool, skip_level int) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(skip_level)

	if err != nil {
		t.Fatal("Line number:", line_number, "Fail. Expected no error, got", err.Error())
	}
	if actual != expected {
		t.Fatal("Line number:", line_number, "Fail. Booleans not equal. Expected:", expected, "got:", actual)
	}
}

func Assert_result_equals_interface(t *testing.T, actual interface{}, err error, expected interface{}, skip_level int) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(skip_level)

	if err != nil {
		t.Fatal("Line number:", line_number, "Fail. Expected no error, got", err.Error())
	}
	if actual != expected {
		t.Fatal("Line number:", line_number, "Fail. Interfaces not equal. Expected:", expected, "got:", actual)
	}
}
