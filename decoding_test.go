package kubeapply

import "testing"

var decodingCases = []struct {
	name string
	out  Output
	want string
}{
	{
		name: "empty string",
		out:  Output(`hi`),
		want: `"hi"`,
	},
	{
		name: "string",
		out:  Output(`{invalid`),
		want: `"{invalid"`,
	},
	{
		name: "structure",
		out:  Output(`{"json": true}`),
		want: `{"json":true}`,
	},
}

func TestDecoding(t *testing.T) {
	for _, tt := range decodingCases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.out.MarshalJSON()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Expected Output(%v).Marshal() = (%v, %v), got (%v, %v) instead",
					tt.out, tt.want, nil, got, err)
			}
		})
	}
}
