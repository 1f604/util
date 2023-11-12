package util

import (
	"fmt"
	"io"
	"os"

	"github.com/1f604/util"
	json_internals "github.com/1f604/util/json/json_internals"
)

// constructor functions take no argument and return a struct pointer.
// Unfortunately in Go it is impossible to specify an argument must be a struct.
// So we have to use reflection to check it at runtime. SO FUCKING LAME.
type ctor_fn[T any] func() *T

func Parse_json_file_into_struct[T any](json_file_location string, struct_constructor ctor_fn[T]) (*T, error) {
	// read the JSON file
	file, err := os.Open(json_file_location)
	if err != nil {
		fmt.Println("Error: ", err, "   Please copy the config file ./config.json to", json_file_location)
		os.Exit(1)
	}
	data, err := io.ReadAll(file)
	util.Check_err(err)
	return Parse_json_data_into_struct(data, struct_constructor)
}

func Parse_json_data_into_struct[T any](data []byte, struct_constructor ctor_fn[T]) (*T, error) {
	// create the struct
	struct_ptr := struct_constructor()
	// check that struct_ptr is a pointer to a struct
	err := json_internals.Check_type_is_ptr_to_struct(struct_ptr)
	if err != nil {
		return nil, err
	}

	/*
		We want to ensure that the JSON file contains exactly the fields that are in the user struct, no more and no less.
		Step 1. we want to check for "no more" i.e. it doesn't contain any fields that are not in the user struct.
		Step 2. we want to check for "no less" i.e. it contains all of the fields that are in the user struct.

		By performing these 2 steps we ensure that the JSON file contains exactly the fields that are in the user struct.
	*/
	// Step 1. Unmarshal the JSON into user struct and error if it contains any fields that are not in the user struct.
	err = json_internals.Json_unmarshall_error_on_extra_fields(data, struct_ptr)
	if err != nil {
		return nil, err
	}
	// Step 2. Check the user struct to see if it contains any nil pointers
	// We have to use reflect here. See https://stackoverflow.com/questions/19633763/unmarshaling-json-in-go-required-field
	// This has to be recursive so that it works for nested structs
	err = json_internals.Recursively_Check_for_nil_pointers(struct_ptr, "struct_ptr")
	if err != nil {
		return nil, err
	}
	// And that's it. Now we can be sure that the json contains exactly the fields in the user struct, no more and no less.
	return struct_ptr, nil

	/*
	 Every field must be a pointer in order for us to check if it was initialized as "nil" or not
	 If it wasn't initialized then it will be nil, otherwise it won't
	 This is one of the stupid annoying things I hate about Go
	 See https://stackoverflow.com/questions/19633763/unmarshaling-json-in-go-required-field
	 Example:
	   type Configuration struct {
	   	ServeDir       *string `json:"serve_dir"`
	   	LoggingDir     *string `json:"logging_dir"`
	   	LogFileName    *string `json:"log_file_name"`
	   	LogFileMaxSizeBytes *int64  `json:"log_file_max_size_bytes"`
	   	Port           *uint64  `json:"port"`
	   }
	*/
	// Need to pass non-nil pointer to here
	/*
		You can't have comments in JSON so I put comments here instead.
		You also can't have line breaks in JSON strings so you have to write everything in one line.
		JSON doesn't even allow trailing commas, so annoying.
		I really hate JSON but I hate XML even more. I'm only using JSON because Go has a JSON parser built in.
		I could write my own parser but it would be more lines of code for no real benefit.

		Another thing I really hate about Go is that the jsonunmarshal function has no option to allow you to declare
		certain fields in a struct as "required". You can make it error on extraneous fields but you can't tell it to
		error on missing fields. I had to implement it myself using reflection. So fucking bad.
	*/
}

func Parse_json_config_file_panic_on_err[T any](config_file_location string, struct_constructor ctor_fn[T], out_ptr **T) {
	result, err := Parse_json_file_into_struct(config_file_location, struct_constructor)
	util.Check_err(err)
	*out_ptr = result
}
