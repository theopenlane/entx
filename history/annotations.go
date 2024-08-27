package enthistory

import (
	"encoding/json"
)

const (
	ValueTypeInt ValueType = iota
	ValueTypeString
)

type ValueType uint

func (ValueType) ValueType() string {
	return "ValueType"
}

const (
	annotationName = "History"
)

// Annotations of the history extension
type Annotations struct {
	Exclude   bool `json:"exclude,omitempty"`   // Will exclude history tracking for this schema
	IsHistory bool `json:"isHistory,omitempty"` // DO NOT APPLY TO ANYTHING EXCEPT HISTORY SCHEMAS
}

// Name of the annotation
func (Annotations) Name() string {
	return annotationName
}

// jsonUnmarshalAnnotations unmarshals the annotations from the schema
// this is useful when you have a map[string]any and want to get the fields
// from the annotation
func jsonUnmarshalAnnotations(data any) (a Annotations, err error) {
	var out []byte

	out, err = json.Marshal(data)
	if err != nil {
		return
	}

	if err = json.Unmarshal(out, &a); err != nil {
		return
	}

	return
}
