package decoding

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// FlagValue can be used to decode an incoming flag value.
type FlagValue string

// MarshalJSON returns a JSON string encoding of v.
func (f FlagValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(f))
}

// UnmarshalJSON parses the JSON-encoded flag value.
func (f *FlagValue) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)

	if err == nil {
		*f = FlagValue(s)
		return nil
	}

	return f.unmarshalJSONSubtypes(data, err)
}

func (f *FlagValue) unmarshalJSONSubtypes(data []byte, ie error) error {
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlagValue(fmt.Sprintf("%v", n))
		return nil
	}

	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*f = FlagValue(fmt.Sprintf("%v", b))
		return nil
	}

	if ec, ok := ie.(*json.UnmarshalTypeError); ok {
		ec.Type = reflect.TypeOf(f)
		return ec
	}

	return ie
}
