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
type Exportable struct {
	// OrgOwned indicates if the schema is owned by an organization and will automatically add the ownerID filter during export validation.
	OrgOwned bool
	// HasSystemOwned indicates if the schema has systemOwned field and will automatically add the systemOwned=false filter during export validation.
	HasSystemOwned bool
}

// Options for the Exportable annotation.
type ExportableOption func(*Exportable)

// NewExportable creates a new Exportable annotation with the given options.
func NewExportable(opts ...ExportableOption) Exportable {
	e := Exportable{}

	for _, opt := range opts {
		opt(&e)
	}

	return e
}

// WithOrgOwned is an option for the Exportable annotation
// that indicates the schema is owned by an organization.
func WithOrgOwned() ExportableOption {
	return func(e *Exportable) {
		e.OrgOwned = true
	}
}

// WithSystemOwned is an option for the Exportable annotation
// that indicates the schema is owned by the system.
func WithSystemOwned() ExportableOption {
	return func(e *Exportable) {
		e.HasSystemOwned = true
	}
}

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
