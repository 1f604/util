package util

import (
	"errors"
	"strings"
	"time"

	"github.com/1f604/util"
)

func Try_parse_log_file_line(line string) (time.Time, error) {
	// try to parse the timestamps
	first_unix_timestamp_str, _, found := strings.Cut(line, " ")
	errortime := time.Time{}
	if !found {
		// if there is no space
		return errortime, errors.New("No space found in line.")
	} else {
		// try to parse it
		num, err := util.String_to_int64(first_unix_timestamp_str)
		if err != nil {
			// if parse failed
			return errortime, err
		}
		// if parse succeeded, check if it is a unix timestamp
		// check it's between the year 2020 and 20000
		timestamp := time.UnixMicro(num).UTC()
		if timestamp.Before(time.Date(2020, 0, 0, 0, 0, 0, 0, time.UTC)) {
			return errortime, errors.New("Time represented is before Jan 2020.")
		}
		if timestamp.After(time.Date(20000, 0, 0, 0, 0, 0, 0, time.UTC)) {
			return errortime, errors.New("Time represented is after the year 20,000.")
		}
		return timestamp, nil
	}
}
