// Package jsonx is a minimal JSON helper package standing in for the equivalent
// core-more package (github.com/theopenlane/core/pkg/jsonx), providing just the surface
// entityops-generated code calls
package jsonx

import "encoding/json"

// Decode unmarshals raw JSON into T
func Decode[T any](raw json.RawMessage) (T, error) {
	var out T
	if len(raw) == 0 {
		return out, nil
	}

	err := json.Unmarshal(raw, &out)

	return out, err
}

// ToRawMap marshals value and decodes it into a map of raw per-field JSON messages
func ToRawMap(value any) (map[string]json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// SetObjectKey sets key to value on the JSON object in base, returning the updated document.
// A nil base starts a new object. changed reports whether the key's value actually changed
func SetObjectKey(base json.RawMessage, key string, value any) (json.RawMessage, bool, error) {
	doc := map[string]json.RawMessage{}

	if len(base) > 0 {
		if err := json.Unmarshal(base, &doc); err != nil {
			return nil, false, err
		}
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, false, err
	}

	changed := string(doc[key]) != string(encoded)
	doc[key] = encoded

	out, err := json.Marshal(doc)
	if err != nil {
		return nil, false, err
	}

	return out, changed, nil
}
