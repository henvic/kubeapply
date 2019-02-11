package decoding

import (
	"encoding/json"
	"reflect"
)

// FileValue can be used to decode an incoming file value.
type FileValue []byte

// UnmarshalJSON parses the JSON-encoded flag value.
func (f *FileValue) UnmarshalJSON(data []byte) error {
	if err := f.unmarshalJSONData(data); err == nil {
		return nil
	}

	var s string
	err := json.Unmarshal(data, &s)
	*f = FileValue(s)

	if ec, ok := err.(*json.UnmarshalTypeError); ok {
		ec.Type = reflect.TypeOf(f)
		return ec
	}

	return err
}

func (f *FileValue) unmarshalJSONData(data []byte) error {
	var r map[string]json.RawMessage
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	*f = FileValue(data)
	return nil
}
