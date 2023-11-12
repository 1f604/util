// Unit tests for logging helper functions
package util_test

import (
	"runtime"
	"testing"
	"time"

	. "fileserver/pkg/util"
	util "fileserver/pkg/util/logging/logging_internals"
)

// hooray for generics!!!
func assert_Try_parse_log_file_line_returns[RESULT_or_ERROR RESULT | ERROR](t *testing.T, line string, expected_result RESULT_or_ERROR) {
	t.Helper()

	_, _, line_number, _ := runtime.Caller(1) // lol you have to use this to get the line number lmao
	result, err := util.Try_parse_log_file_line(line)
	// simple workaround: use any
	switch v := any(expected_result).(type) {
	case RESULT:
		ts, _ := v.S.(time.Time) // we know it is time.Time
		Assert_result_equals_time(t, result, err, ts, line_number)
	case ERROR:
		Assert_error_equals(t, err, v.S, 2)
	}
}

func Test_log_file_parser(t *testing.T) {
	t.Parallel()
	/*
		Being able to see each test case on a single line is so fucking good,
		because it allows me to see AT A GLANCE exactly what test cases I have and exactly what behavior I expect in each case.
	*/
	assert_Try_parse_log_file_line_returns(t, "", Error("No space found in line."))
	assert_Try_parse_log_file_line_returns(t, "asdfg", Error("No space found in line."))
	assert_Try_parse_log_file_line_returns(t, "24786543_hello", Error("No space found in line."))
	assert_Try_parse_log_file_line_returns(t, "24786543_hello\nasdf", Error("No space found in line."))

	assert_Try_parse_log_file_line_returns(t, "kjljg fdssd", Error("strconv.ParseInt: parsing \"kjljg\": invalid syntax"))
	assert_Try_parse_log_file_line_returns(t, "12345 fdssd", Error("Time represented is before Jan 2020."))
	assert_Try_parse_log_file_line_returns(t, "1369766400000000 fdssd", Error("Time represented is before Jan 2020."))

	assert_Try_parse_log_file_line_returns(t, "1669766400000000 fdssd", Result(time.Date(2023, 0, 0, 0, 0, 0, 0, time.UTC)))

	assert_Try_parse_log_file_line_returns(t, "1234567890123456789 fdssd", Error("Time represented is after the year 20,000."))

	assert_Try_parse_log_file_line_returns(t, "12345678901234567890123456789 fdssd", Error("strconv.ParseInt: parsing \"12345678901234567890123456789\": value out of range"))
}
