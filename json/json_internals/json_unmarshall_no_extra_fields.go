// Checks that JSON file does not contain any fields that are not in the struct
package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

func Json_unmarshall_error_on_extra_fields(data []byte, struct_ptr interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields() // This line is super critical. It makes the decoder error on JSON fields that aren't in the user struct.
	err := decoder.Decode(struct_ptr)
	if err != nil {
		return err
	}
	/*
		Next, we need to check decoder.Token() because it will continue parsing the JSON even after hitting the close bracket of top level
		See https://github.com/golang/go/issues/36225

		If you have this JSON: {} "port":90
		decoder.Token() will return token=port, err: nil.
		So we have to check that decoder.Token() gives io.EOF error.
	*/
	_, err = decoder.Token()
	if !errors.Is(err, io.EOF) {
		return errors.New("JSON ERROR: trailing data after top-level value")
	}
	return nil
}
