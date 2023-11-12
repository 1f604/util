package util_test

import (
	"testing"

	json "github.com/1f604/util/json"

	"github.com/1f604/util"
)

type T1 struct {
	A string  `json:"a"`
	B *string `json:"b"`
	C int     `json:"c"`
}

type T2 struct {
	A *string `json:"a"`
	B *string `json:"b"`
	C int     `json:"c"`
}

type T3 struct {
	A *string `json:"a"`
	B *string `json:"b"`
	C *int    `json:"c"`
}

type UserStruct struct {
	A *string `json:"a"`
	B *int    `json:"b"`
	C *uint   `json:"c"`
}

func Test_Parse_json_data_check_argument_is_ptr_to_struct(t *testing.T) {
	t.Parallel()

	buildInt := func() *int {
		x := 2
		return &x
	}

	data := []byte(`{"a":"hi", "b":1, "c":1}`)

	_, err := json.Parse_json_data_into_struct(data, buildInt)
	util.Assert_error_equals(t, err, `Error: argument is not a pointer to struct; is *int`, 1)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_no_error(t, err, 1)
}

func Test_Parse_json_data_check_struct_fields_are_all_pointers(t *testing.T) {
	t.Parallel()
	data := []byte(`{"a":"hi", "b":"hey", "c":1}`)
	_, err := json.Parse_json_data_into_struct(data, util.BuildStruct[T1])
	util.Assert_error_equals(t, err, "Error: Field A (string) is not a pointer", 1)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[T2])
	util.Assert_error_equals(t, err, "Error: Field C (int) is not a pointer", 1)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[T3])
	util.Assert_no_error(t, err, 1)
}

func Test_Parse_json_data_check_json_data_values_are_correct_types(t *testing.T) {
	t.Parallel()

	data := []byte(`{"a":1, "b":1, "c":1}`)
	_, err := json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `json: cannot unmarshal number into Go struct field UserStruct.a of type string`, 1)

	data = []byte(`{"a":"1", "b":"1", "c":1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `json: cannot unmarshal string into Go struct field UserStruct.b of type int`, 1)

	data = []byte(`{"a":"1", "b":1, "c":-1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `json: cannot unmarshal number -1 into Go struct field UserStruct.c of type uint`, 1)

	data = []byte(`{"a":"1", "b":1, "c":1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_no_error(t, err, 1)

	// nested data
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": 23,
			"file_system_directory": "./example/test456"
			},
			{
			"url_prefix": "/test123",
			"file_system_directory": "./example/test456"
			}
		]
	}
	`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `json: cannot unmarshal number into Go struct field CustomStruct.url_prefix_to_directory_map.url_prefix of type string`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": 11,
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {

				"url_prefix":"",
				"file_system_directory": "./example/test"
			}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `json: cannot unmarshal number into Go struct field CustomStruct.url_prefix_to_directory_map.url_prefix of type string`, 1)
}

func Test_Parse_json_data_check_for_missing_fields_in_json_data(t *testing.T) { //nolint:funlen // it's ok.
	t.Parallel()

	data := []byte(`{"b":1, "c":1}`)
	_, err := json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `Error: Field A (type *string) is a nil pointer`, 1)

	data = []byte(`{"a":"1", "c":1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `Error: Field B (type *int) is a nil pointer`, 1)

	data = []byte(`{"a":"1", "b":1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `Error: Field C (type *uint) is a nil pointer`, 1)

	data = []byte(`{"a":"1", "b":1, "c":1}`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_no_error(t, err, 1)

	// nested data
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test123",
			"file_system_directory": "./example/test456"
			}
		]
	}
	`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	//	fmt.Println(((*(test_struct.UrlPrefixes))[0].URLPrefix))
	//	fmt.Println(((*(test_struct.UrlPrefixes))[0].FileSystemDirectory))
	//	fmt.Println(((*(test_struct.UrlPrefixes))[1].URLPrefix))
	//	fmt.Println(((*(test_struct.UrlPrefixes))[1].FileSystemDirectory))
	//	fmt.Println(*((*(test_struct.UrlPrefixes))[0].URLPrefix))
	//	fmt.Println(*((*(test_struct.UrlPrefixes))[0].FileSystemDirectory))
	//	fmt.Println(*((*(test_struct.UrlPrefixes))[1].URLPrefix))
	//	fmt.Println(*((*(test_struct.UrlPrefixes))[1].FileSystemDirectory))
	util.Assert_error_equals(t, err, `Error: Field URLPrefix (type *string) is a nil pointer`, 1)

	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/test",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test123"
			}
		]
	}
	`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `Error: Field FileSystemDirectory (type *string) is a nil pointer`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		]
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `Error: Field MapOfStructs (type *map[string]util_test.CustomStruct) is a nil pointer`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `Error: Field NestedStruct (type *util_test.CustomStruct) is a nil pointer`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{},
		"nested_struct": {
			"url_prefix": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `Error: Field FileSystemDirectory (type *string) is a nil pointer`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_no_error(t, err, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": ""
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, "json: cannot unmarshal string into Go struct field TestConfiguration.map_of_structs of type util_test.CustomStruct", 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, "Error: Map map[string]util_test.CustomStruct contains unexpected value of kind: struct type: util_test.CustomStruct", 1)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration2])
	util.Assert_error_equals(t, err, "Error: Field URLPrefix (type *string) is a nil pointer", 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {"url_prefix":""}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration2])
	util.Assert_error_equals(t, err, "Error: Field FileSystemDirectory (type *string) is a nil pointer", 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {

				"url_prefix":"",
				"file_system_directory": "./example/test"
			}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration2])
	util.Assert_no_error(t, err, 1)
}

func Test_Parse_json_data_check_for_extraneous_fields_in_json_data(t *testing.T) {
	t.Parallel()

	data := []byte(`{"a":"hi", "b":1, "d":1, "c":1}`)
	_, err := json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `json: unknown field "d"`, 1)

	// in nested data
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/test",
			"url_prefix1": "/test",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		]
	}
	`)
	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration])
	util.Assert_error_equals(t, err, `json: unknown field "url_prefix1"`, 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {

				"url_prefix":"",
				"url_prefix23": "",
				"file_system_directory": "./example/test"
			}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration2])
	util.Assert_error_equals(t, err, "json: unknown field \"url_prefix23\"", 1)

	// control
	data = []byte(`{
		"port": 80,
		"serve_dir": "./example/",
		"logging_dir": "./logs/",
		"log_file_name": "current.log",
		"log_file_prefix": "log",
		"log_file_max_size_bytes": 5000,
		"url_prefix_to_directory_map": [
			{
			"url_prefix": "/files",
			"file_system_directory": "./example"
			},
			{
			"url_prefix": "/test",
			"file_system_directory": "./example/test"
			}
		],
		"map_of_structs":{
			"a": {

				"url_prefix":"",
				"file_system_directory": "./example/test"
			}
		},
		"nested_struct": {
			"url_prefix": "",
			"file_system_directory": ""
		}
	}
	`)

	_, err = json.Parse_json_data_into_struct(data, util.BuildStruct[TestConfiguration2])
	util.Assert_no_error(t, err, 1)
}

func Test_Parse_json_data_check_for_trailing_data_in_json_file(t *testing.T) {
	t.Parallel()

	data := []byte(`{"a":"hi", "b":1, "c":1} "a":"hi"`)
	_, err := json.Parse_json_data_into_struct(data, util.BuildStruct[UserStruct])
	util.Assert_error_equals(t, err, `JSON ERROR: trailing data after top-level value`, 1)
}

type CustomStruct struct {
	URLPrefix           *string `json:"url_prefix"`
	FileSystemDirectory *string `json:"file_system_directory"`
}

// JSON map keys must be strings.
type TestConfiguration struct {
	ServeDir            *string                  `json:"serve_dir"`
	LoggingDir          *string                  `json:"logging_dir"`
	LogFileName         *string                  `json:"log_file_name"`
	LogFilePrefix       *string                  `json:"log_file_prefix"`
	LogFileMaxSizeBytes *int64                   `json:"log_file_max_size_bytes"`
	Port                *uint64                  `json:"port"`
	SliceOfStructs      *[]*CustomStruct         `json:"url_prefix_to_directory_map"`
	MapOfStructs        *map[string]CustomStruct `json:"map_of_structs"`
	NestedStruct        *CustomStruct            `json:"nested_struct"`
}

type TestConfiguration2 struct {
	ServeDir            *string                   `json:"serve_dir"`
	LoggingDir          *string                   `json:"logging_dir"`
	LogFileName         *string                   `json:"log_file_name"`
	LogFilePrefix       *string                   `json:"log_file_prefix"`
	LogFileMaxSizeBytes *int64                    `json:"log_file_max_size_bytes"`
	Port                *uint64                   `json:"port"`
	SliceOfStructs      *[]*CustomStruct          `json:"url_prefix_to_directory_map"`
	MapOfStructs        *map[string]*CustomStruct `json:"map_of_structs"`
	NestedStruct        *CustomStruct             `json:"nested_struct"`
}
