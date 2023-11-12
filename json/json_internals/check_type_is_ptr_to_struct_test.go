package util_test

import (
	"testing"

	"fileserver/pkg/util"
	json "fileserver/pkg/util/json/json_internals"
)

func Test_Check_type_is_ptr_to_struct(t *testing.T) {
	t.Parallel()

	user_struct := UserStruct{} //nolint:exhaustruct // we don't need to initialize it here
	err := json.Check_type_is_ptr_to_struct(user_struct)
	util.Assert_error_equals(t, err, `Error: argument is not a pointer; is util_test.UserStruct`, 1)

	num := 2
	err = json.Check_type_is_ptr_to_struct(num)
	util.Assert_error_equals(t, err, `Error: argument is not a pointer; is int`, 1)

	err = json.Check_type_is_ptr_to_struct(&num)
	util.Assert_error_equals(t, err, `Error: argument is not a pointer to struct; is *int`, 1)

	err = json.Check_type_is_ptr_to_struct(&user_struct)
	util.Assert_no_error(t, err, 1)
}
