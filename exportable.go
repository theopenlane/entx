package entx

import "encoding/json"

// Exportable annotation marks a schema as exportable.
// This annotation can be used to indicate that a schema supports
// export functionality and should be included in export validation.
//
// Usage:
//
//	func (MySchema) Annotations() []schema.Annotation {
//		return []schema.Annotation{
//			entx.Exportable{},
//		}
//	}
type Exportable struct{}

// Name returns the name of the Exportable annotation.
func (Exportable) Name() string {
	return "Exportable"
}

// Decode unmarshalls the Exportable annotation
func (a *Exportable) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}
