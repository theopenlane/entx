package entx

import (
	"encoding/json"
	"encoding/json/jsontext"
	"io"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalRawMessage provides a graphql.Marshaler for jsontext.Value
func MarshalRawMessage(t jsontext.Value) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		s, _ := t.MarshalJSON()
		io.Writer.Write(w, s) // nolint:errcheck
	})
}

// UnmarshalRawMessage provides a graphql.Unmarshaler for jsontext.Value
func UnmarshalRawMessage(v interface{}) (jsontext.Value, error) {
	switch j := v.(type) {
	case []byte:
		return jsontext.Value(j), nil
	case map[string]interface{}:
		js, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return jsontext.Value(js), nil
	default:
		js, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		return jsontext.Value(js), nil
	}
}
