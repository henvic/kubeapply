package kubeapply

import (
	"encoding/json"
)

// Output can be used to encode an outgoing output value to a JSON structure.
type Output string

// MarshalJSON returns a JSON string encoding of v.
func (o Output) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(json.RawMessage(o))

	if err != nil {
		return json.Marshal(string(o))
	}

	return b, nil
}
