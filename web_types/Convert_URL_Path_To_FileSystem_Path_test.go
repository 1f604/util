package util_test

import (
	util "fileserver/pkg/util"
	json "fileserver/pkg/util/json"
	web_types "fileserver/pkg/util/web_types"
	"fmt"
	"reflect"
	"runtime"
	"testing"
)

type RESULT struct {
	string
}

type ERROR_INVALID struct{}
type ERROR_NO_MATCH struct{}

func run_test5[T RESULT | ERROR_NO_MATCH | ERROR_INVALID](t *testing.T, prefixes_map web_types.URLPrefixToFileSystemDirectoryMap, s string, expected_result T) {
	t.Helper()
	_, _, line_number, _ := runtime.Caller(1) // lol you have to use this to get the line number lmao

	result, err := web_types.PosixValidatedFullURLPath(s)

	if _, ok := any(expected_result).(ERROR_INVALID); ok {
		if err != nil {
			return
		} else {
			t.Fatal("Line number:", line_number, "Test failed. Expected error but did not get any.")
		}
	}

	if err != nil {
		t.Fatal("Line number:", line_number, "Test failed. Got URL Path validation error but did not expect any.")
	}

	fs_path_ptr, err := web_types.Convert_URLPath_To_Full_FileSystem_Path(*result, prefixes_map)

	switch v := any(expected_result).(type) {
	case RESULT:
		util.Assert_no_error(t, err, 2)
		util.Assert_result_equals_bytes(t, []byte(*fs_path_ptr.FileSystemPath), err, v.string, 2)
	case ERROR_NO_MATCH:
		expected_err := fmt.Sprintf("Error: %s prefix not in URL-to-dir map", s)
		util.Assert_error_equals(t, err, expected_err, 2)
	}

	// fmt.Println("input.URLPath:", input.URLPath)
	// err := web_types.Posix_filename_validator(input)
}

type TestConfiguration struct {
	ServeDir            *string                                      `json:"serve_dir"`
	LoggingDir          *string                                      `json:"logging_dir"`
	LogFileName         *string                                      `json:"log_file_name"`
	LogFilePrefix       *string                                      `json:"log_file_prefix"`
	LogFileMaxSizeBytes *int64                                       `json:"log_file_max_size_bytes"`
	Port                *uint64                                      `json:"port"`
	UrlPrefixes         *web_types.URLPrefixToFileSystemDirectoryMap `json:"url_prefix_to_directory_map"`
}

// correctness check.
func Test_Convert_URL_Path_To_File_System_Path(t *testing.T) { //nolint:funlen, maintidx // it's fine
	t.Parallel()

	data := []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/hel/",
			"file_system_directory_path": "/root/"
			},
			{
			"url_prefix": "/hello/",
			"file_system_directory_path": "./example/yo/"
			},
			{
			"url_prefix": "/hello/world/",
			"file_system_directory_path": "wassup/"
			}
		]
	}
	`)
	result, err := json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_no_error(t, err, 1)
	// finish setup

	prefixes_map := *result.UrlPrefixes

	// the actual tests
	// invalid URL path
	run_test5(t, prefixes_map, "", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a.", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".a", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a.a", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "///", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".///", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "///.", ERROR_INVALID{})

	run_test5(t, prefixes_map, "..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "..//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "././", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./.", ERROR_INVALID{})

	run_test5(t, prefixes_map, "...", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/...", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../.", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//...", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "..//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "...//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "././.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./../", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".././", ERROR_INVALID{})

	run_test5(t, prefixes_map, "/.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})

	// valid URL path but no matching prefix
	run_test5(t, prefixes_map, "/a", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/a/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/help", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/help/hi", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/helloworld", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/hell/world", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/hello1/world", ERROR_NO_MATCH{})

	// invalid URL path
	run_test5(t, prefixes_map, "", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})

	// valid URL path but no matching prefix
	run_test5(t, prefixes_map, "/helloworld", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/hell/world", ERROR_NO_MATCH{})
	run_test5(t, prefixes_map, "/hello1/world", ERROR_NO_MATCH{})

	// match
	run_test5(t, prefixes_map, "/hel/hi", RESULT{"/root/hi"})
	run_test5(t, prefixes_map, "/hello/worl", RESULT{"./example/yo/worl"})
	run_test5(t, prefixes_map, "/hello/world", RESULT{"./example/yo/world"})
	run_test5(t, prefixes_map, "/hello/worlda", RESULT{"./example/yo/worlda"})
	run_test5(t, prefixes_map, "/hello/world/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/hello/world/a", RESULT{"wassup/a"})
	run_test5(t, prefixes_map, "/hello/world/hi", RESULT{"wassup/hi"})

	// add root handler and try again
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/",
			"file_system_directory_path": "woot/"
			},
			{
			"url_prefix": "/hel/",
			"file_system_directory_path": "/root/"
			},
			{
			"url_prefix": "/hello/",
			"file_system_directory_path": "./example/yo/"
			},
			{
			"url_prefix": "/hello/world/",
			"file_system_directory_path": "wassup/"
			}
		]
	}
	`)
	result, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_no_error(t, err, 1)
	prefixes_map = *result.UrlPrefixes

	// invalid URL path
	run_test5(t, prefixes_map, "", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a.", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".a", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a.a", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "///", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".///", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "///.", ERROR_INVALID{})

	run_test5(t, prefixes_map, "..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "..//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "././", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./.", ERROR_INVALID{})

	run_test5(t, prefixes_map, "...", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/...", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../.", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//...", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".//..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "..//.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "...//", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/.../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "././.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./../", ERROR_INVALID{})
	run_test5(t, prefixes_map, ".././", ERROR_INVALID{})

	run_test5(t, prefixes_map, "/.", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/..", ERROR_INVALID{})
	run_test5(t, prefixes_map, "./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/./", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/../", ERROR_INVALID{})
	run_test5(t, prefixes_map, "a/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "//", ERROR_INVALID{})

	// match
	run_test5(t, prefixes_map, "/a", RESULT{"woot/a"})
	run_test5(t, prefixes_map, "/a/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/hel", RESULT{"woot/hel"})
	run_test5(t, prefixes_map, "/help", RESULT{"woot/help"})
	run_test5(t, prefixes_map, "/hel/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/hel/a", RESULT{"/root/a"})
	run_test5(t, prefixes_map, "/hello", RESULT{"woot/hello"})
	run_test5(t, prefixes_map, "/hello/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/hello/worl", RESULT{"./example/yo/worl"})
	run_test5(t, prefixes_map, "/hello/worlda", RESULT{"./example/yo/worlda"})
	run_test5(t, prefixes_map, "/hello/world", RESULT{"./example/yo/world"})
	run_test5(t, prefixes_map, "/helloworld", RESULT{"woot/helloworld"})
	run_test5(t, prefixes_map, "/hell/world", RESULT{"woot/hell/world"})
	run_test5(t, prefixes_map, "/hello1/world", RESULT{"woot/hello1/world"})

	run_test5(t, prefixes_map, "/hello/world/", ERROR_INVALID{})
	run_test5(t, prefixes_map, "/hello/world/a", RESULT{"wassup/a"})

	// test longest match takes precedence
	run_test5(t, prefixes_map, "/hello1/world1/hi", RESULT{"woot/hello1/world1/hi"})
	run_test5(t, prefixes_map, "/hello1/world/hi", RESULT{"woot/hello1/world/hi"})
	run_test5(t, prefixes_map, "/hello/world1/hi", RESULT{"./example/yo/world1/hi"})
	run_test5(t, prefixes_map, "/hello/world/hi", RESULT{"wassup/hi"})

	// Test again with different order in JSON

	// add root handler and try again
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/",
			"file_system_directory_path": "woot/"
			},
			{
			"url_prefix": "/hel/",
			"file_system_directory_path": "/root/"
			},
			{
			"url_prefix": "/hello/world/",
			"file_system_directory_path": "wassup/"
			},
			{
			"url_prefix": "/hello/",
			"file_system_directory_path": "./example/yo/"
			}
		]
	}
	`)
	result, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_no_error(t, err, 1)
	prefixes_map = *result.UrlPrefixes

	// test longest match takes precedence
	run_test5(t, prefixes_map, "/hello1/world1/hi", RESULT{"woot/hello1/world1/hi"})
	run_test5(t, prefixes_map, "/hello1/world/hi", RESULT{"woot/hello1/world/hi"})
	run_test5(t, prefixes_map, "/hello/world1/hi", RESULT{"./example/yo/world1/hi"})
	run_test5(t, prefixes_map, "/hello/world/hi", RESULT{"wassup/hi"})

	// check ReverseMap is as expected
	/*
		"file_system_directory_path": "woot/"
		"file_system_directory_path": "/root/"
		"file_system_directory_path": "./example/yo/"
		"file_system_directory_path": "wassup/"
	*/
	expected_map := map[string]bool{
		"woot/":         true,
		"/root/":        true,
		"./example/yo/": true,
		"wassup/":       true,
	}
	eq := reflect.DeepEqual((*prefixes_map.ReverseMap), expected_map)
	if !eq {
		t.Fatal("Maps not equal.")
	}
}
