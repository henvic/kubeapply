package decoding

import (
	"errors"
	"fmt"
	"testing"
)

var flagValueEncodingCases = []struct {
	name    string
	in      string
	decoded FlagValue
	err     error
}{
	{
		name:    "empty string",
		in:      `""`,
		decoded: FlagValue(""),
		err:     nil,
	},
	{
		name:    "string",
		in:      `"common"`,
		decoded: FlagValue("common"),
		err:     nil,
	},
	{
		name:    "boolean value 'false'",
		in:      `false`,
		decoded: FlagValue("false"),
		err:     nil,
	},
	{
		name:    "boolean value 'true'",
		in:      `true`,
		decoded: FlagValue("true"),
		err:     nil,
	},
	{
		name:    "number value 0",
		in:      `0`,
		decoded: FlagValue("0"),
		err:     nil,
	},
	{
		name:    "number value 123",
		in:      `123`,
		decoded: FlagValue("123"),
		err:     nil,
	},
	{
		name:    "number value 3.14159265",
		in:      `3.14159265`,
		decoded: FlagValue("3.14159265"),
		err:     nil,
	},
	{
		name:    "object error",
		in:      "{}",
		decoded: FlagValue(""),
		err:     errors.New("json: cannot unmarshal object into Go value of type *decoding.FlagValue"),
	},
	{
		name:    "unexpected end of JSON input",
		in:      "",
		decoded: FlagValue(""),
		err:     errors.New(`unexpected end of JSON input`),
	},
}

func TestFlagValueEncoding(t *testing.T) {
	for _, tt := range flagValueEncodingCases {
		t.Run(tt.name, func(t *testing.T) {
			var v FlagValue
			var err = v.UnmarshalJSON([]byte(tt.in))

			if v != tt.decoded || fmt.Sprint(tt.err) != fmt.Sprint(err) {
				t.Errorf("Expected FlagValue.Unmarshal(%v) = (%v, %v), got (%v, %v) instead",
					tt.in, tt.decoded, tt.err, v, err)
			}
		})
	}
}

func TestFlagValueMarshal(t *testing.T) {
	var v = FlagValue(`hello`)

	var b, err = v.MarshalJSON()

	if string(b) != `"hello"` && err != nil {
		t.Errorf("Unexepcted FlagValue.Marshal value")
	}

}
