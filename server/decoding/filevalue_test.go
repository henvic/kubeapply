package decoding

import (
	"errors"
	"fmt"
	"testing"
)

var fileValueEncodingCases = []struct {
	name    string
	in      string
	decoded FileValue
	err     error
}{
	{
		name:    "empty string",
		in:      `""`,
		decoded: FileValue(""),
		err:     nil,
	},
	{
		name:    "string",
		in:      `"common"`,
		decoded: FileValue("common"),
		err:     nil,
	},
	{
		name:    "empty object",
		in:      "{}",
		decoded: FileValue("{}"),
		err:     nil,
	},
	{
		name: "number",
		in:   "3",
		err:  errors.New("json: cannot unmarshal number into Go value of type *decoding.FileValue"),
	},
	{
		name:    "object",
		in:      `{"foo": "bar"}`,
		decoded: FileValue(`{"foo": "bar"}`),
		err:     nil,
	},
	{
		name:    "unexpected end of JSON input",
		in:      "",
		decoded: FileValue(""),
		err:     errors.New(`unexpected end of JSON input`),
	},
}

func TestFileValueEncoding(t *testing.T) {
	for _, tt := range fileValueEncodingCases {
		t.Run(tt.name, func(t *testing.T) {
			var v FileValue
			var err = v.UnmarshalJSON([]byte(tt.in))

			if string(v) != string(tt.decoded) || fmt.Sprint(tt.err) != fmt.Sprint(err) {
				t.Errorf("Expected FileValue.Unmarshal(%v) = (%v, %v), got (%v, %v) instead",
					tt.in, tt.decoded, tt.err, v, err)
			}
		})
	}
}
