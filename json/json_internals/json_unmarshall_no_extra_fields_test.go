package util_test

import (
	"testing"

	. "fileserver/pkg/util"
	util "fileserver/pkg/util/json/json_internals"
)

type UserStruct struct {
	A *string `json:"a"`
	B *int    `json:"b"`
	C *uint   `json:"c"`
}

func Test_Json_unmarshall_error_on_extra_fields(t *testing.T) {
	t.Parallel()

	// Check 1. Check it fails on extra field
	data := []byte(`{"a":"hi", "b":1, "d":1, "c":1}`)
	user_struct := UserStruct{} //nolint:exhaustruct // we don't need to initialize it here
	err := util.Json_unmarshall_error_on_extra_fields(data, &user_struct)
	Assert_error_equals(t, err, `json: unknown field "d"`, 1)

	// Check 2. Check it fails on trailing data
	data = []byte(`{"a":"hi", "b":1, "c":1} "a":"hi"`)
	user_struct = UserStruct{} //nolint:exhaustruct // we don't need to initialize it here
	err = util.Json_unmarshall_error_on_extra_fields(data, &user_struct)
	Assert_error_equals(t, err, `JSON ERROR: trailing data after top-level value`, 1)
}
