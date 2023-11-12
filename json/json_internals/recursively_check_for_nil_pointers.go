// Checks that all the fields in a struct are initialized
package util // gaah this is soo annoying I want to name it util/user_struct but I can't!!

import (
	"fmt"
	"reflect"
)

// Recursively_Check_for_nil_pointers is a function that checks the 'thing_to_be_checked' and its nested elements for nil pointers.
// It has to be recursive so that it checks inside nested data structures.
// Note that it does not check maps and slices for nil values, since we don't really care about that
// If a slice is empty, that's okay and allowed. Likewise if a map is empty, that's okay too. We want users to be able to do that.
// What we don't want are empty fields in structs, whether these structs are inside other structs, or slices, or maps.
//
// The function uses reflection to inspect the structure of the data.
// It performs the following steps:
// 1. Checks whether the argument is a pointer. If not, it returns an error.
// 2. Checks whether the argument is a nil pointer. If so, it returns an error.
// 3. Dereferences the pointer to find the actual data type.
// 4. Calls itself on each nested element.
//
// Guarantees:
// - If the data structure contains a nil pointer, then the function is guaranteed to return an error.
// - Therefore, if it doesn't throw an error then we can be sure that the data structure doesn't contain a nil pointer.
// - It does not modify the input data structure.
//
// Parameters:
// - thing_to_be_checked: The thing which the function checks for nil pointers.
//
// Returns:
// - An error if a nil pointer or unsupported data type is found within the data structure, indicating the type of error.
// - nil if no nil pointers or unsupported data types are found within the data structure.
func Recursively_Check_for_nil_pointers(thing_to_be_checked interface{}, name_of_the_thing string) error { //nolint: gocognit // yes this is a complicated function.
	// fmt.Println("=============================")
	// fmt.Printf("Function called Recursively_Check_for_nil_pointers: Type of thing to be checked: %T\n", thing_to_be_checked)
	// fmt.Println(": Thing type: ", reflect.TypeOf(thing_to_be_checked))

	// First check if the thing is a pointer
	thingVal := reflect.ValueOf(thing_to_be_checked)
	// fmt.Println(": Thing kind: ", thingVal.Kind())

	// fmt.Println(": ThingVal.type(): ", thingVal.Type())
	if thingVal.Kind() != reflect.Ptr {
		return fmt.Errorf("Error: %s (%T) is not a pointer", name_of_the_thing, thing_to_be_checked)
	}
	// Second check if it's a nil pointer
	// This is important because dereferencing a nil pointer with Elem() results in reflect.Invalid
	if thingVal.IsNil() {
		return fmt.Errorf("Error: %s (type %T) is a nil pointer", name_of_the_thing, thing_to_be_checked)
	}

	// Now dereference it and see what type it actually is
	dereferencedVal := thingVal.Elem() // dereference the pointer
	// fmt.Println(": dereferencedVal.Kind(): ", dereferencedVal.Kind())
	switch dereferencedVal.Kind() { // find all kinds here: https://pkg.go.dev/reflect#Kind
	// These cases are exhaustive, as checked by the linter (exhaustive).
	// error on unexpected types
	case reflect.Invalid, reflect.Chan, reflect.Func, reflect.Interface, reflect.Pointer, reflect.UnsafePointer:
		return fmt.Errorf("Error: Unsupported data type: Type of %s: %s Kind of *%s: %s\n", name_of_the_thing, thingVal.Type(), name_of_the_thing, dereferencedVal.Kind())
		// fmt.Println("Field", fieldType.Name, fieldVal.Type(), fieldVal.Interface(), "is a pointer!")
	case reflect.Array, reflect.Slice: // if it's an array or slice, check every element
		// fmt.Printf("thing_to_be_checked is an array or slice: %T %s\n", thing_to_be_checked, dereferencedVal.Kind())
		// Allow slices of strings and other primitive types.
		for i := 0; i < dereferencedVal.Len(); i++ {
			childVal := dereferencedVal.Index(i)
			switch childVal.Kind() { //nolint:exhaustive // it's ok...
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
				reflect.Complex64, reflect.Complex128, reflect.String: // do nothing for non-container primitive types.
			case reflect.Ptr:
				err := Recursively_Check_for_nil_pointers(childVal.Interface(), "[slice element has no name]")
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Error: Slice %s contains unexpected value of kind: %s", dereferencedVal.Type(), childVal.Kind())
			}
		}
	case reflect.Map: // if it's a map, check every key and value
		// no need to check type of key and value
		// keyType := dereferencedVal.Type().Key()
		// valueType := dereferencedVal.Type().Elem()
		for _, keyVal := range dereferencedVal.MapKeys() {
			valueVal := dereferencedVal.MapIndex(keyVal)
			// if key is a primitive type, don't check it.
			keyType := dereferencedVal.Type().Key()
			valueType := dereferencedVal.Type().Elem()
			switch keyType.Kind() { //nolint:exhaustive // it's ok...
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
				reflect.Complex64, reflect.Complex128, reflect.String: // do nothing for non-container primitive types.
			case reflect.Ptr:
				err := Recursively_Check_for_nil_pointers(keyVal.Interface(), "[map key has no name]")
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unexpected map key type: %s", keyType.Kind())
			}

			// if value is a primitive type, don't check it.
			switch valueVal.Kind() { //nolint:exhaustive // it's ok...
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
				reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
				reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
				reflect.Complex64, reflect.Complex128, reflect.String: // do nothing for non-container primitive types.
			case reflect.Ptr:
				err := Recursively_Check_for_nil_pointers(valueVal.Interface(), "[map value has no name]")
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("Error: Map %s contains unexpected value of kind: %s type: %s", dereferencedVal.Type(), valueType.Kind(), valueType.String())
			}
		}
	case reflect.Struct: // if it's a struct, check every field
		for i := 0; i < dereferencedVal.NumField(); i++ {
			childVal := dereferencedVal.Field(i)
			childType := dereferencedVal.Type().Field(i)
			err := Recursively_Check_for_nil_pointers(childVal.Interface(), "Field "+childType.Name)
			if err != nil {
				return err
			}
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.String: // do nothing for non-container primitive types.
	}

	return nil
}
