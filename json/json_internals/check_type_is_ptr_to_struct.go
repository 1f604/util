package util

import (
	"fmt"
	"reflect"
)

func Check_type_is_ptr_to_struct(struct_ptr interface{}) error {
	v := reflect.ValueOf(struct_ptr)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("Error: argument is not a pointer; is %T", struct_ptr)
	}
	v = v.Elem() // dereference the pointer
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("Error: argument is not a pointer to struct; is %T", struct_ptr)
	}
	return nil
}
