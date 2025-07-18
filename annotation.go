package entx

import "encoding/json"

// CascadeAnnotationName is a name for our cascading delete annotation
var CascadeAnnotationName = "OPENLANE_CASCADE"

// CascadeThroughAnnotationName is a name for our cascading through edge delete annotation
var CascadeThroughAnnotationName = "OPENLANE_CASCADE_THROUGH"

// SchemaGenAnnotationName is a name for our graphql schema generation annotation
var SchemaGenAnnotationName = "OPENLANE_SCHEMAGEN"

// QueryGenAnnotationName is a name for our graphql query generation annotation
var QueryGenAnnotationName = "OPENLANE_QUERYGEN"

// SearchFieldAnnotationName is a name for the search field annotation
var SearchFieldAnnotationName = "OPENLANE_SEARCH"

// FeatureAnnotationName is the annotation name used for module feature tags
var FeatureAnnotationName = "OPENLANE_FEATURE"

// FeatureVisibilityAnnotationName is the annotation name used to flag schema visibility
var FeatureVisibilityAnnotationName = "OPENLANE_SCHEMA_VISIBILITY"

// CascadeAnnotation is an annotation used to indicate that an edge should be cascaded
type CascadeAnnotation struct {
	Field string
}

// CascadeThroughAnnotation is an annotation used to indicate that an edge should be cascaded through
type CascadeThroughAnnotation struct {
	Schemas []ThroughCleanup
}

// ThroughCleanup is a struct used to indicate the field and through edge to cascade through
type ThroughCleanup struct {
	Field   string
	Through string
}

// SchemaGenAnnotation is an annotation used to indicate that schema generation should be skipped for this type
// When Skip is true, the search schema generation is always skipped
// SkipSearch allow for schemas to be be opt out of search schema generation
type SchemaGenAnnotation struct {
	// Skip indicates that the schema generation should be skipped for this type
	Skip bool
	// SkipSearch indicates that the schema should not be searchable
	// Schemas are also not searchable if not fields are marked as searchable
	SkipSearch bool
}

// QueryGenAnnotation is an annotation used to indicate that query generation should be skipped for this type
type QueryGenAnnotation struct {
	Skip bool
}

// SearchFieldAnnotation is an annotation used to indicate that the field should be searchable
type SearchFieldAnnotation struct {
	// Searchable indicates that the field should be searchable
	Searchable bool
	// ExcludeAdmin indicates that the field will be excluded from the admin search which includes all fields by default
	ExcludeAdmin bool
	// JSONPath is the path to the field in the JSON object
	JSONPath string
	// JSONDotPath is the path to the field in the JSON object using dot notation
	JSONDotPath string
}

// FeatureVisibilityAnnotation is an annotation used to flag schema visibility
type FeatureVisibilityAnnotation struct {
	Visibility Visibility
}

// FeatureAnnotation is an annotation used to specify modules that enable a schema
type FeatureAnnotation struct {
	// Features contains the list of feature identifiers
	Features []FeatureModule
	// Visibility marks the associated feature modules visibility
	Visibility Visibility `json:",omitempty"`
}

// Name returns the name of the CascadeAnnotation
func (a CascadeAnnotation) Name() string {
	return CascadeAnnotationName
}

// Name returns the name of the CascadeThroughAnnotation
func (a CascadeThroughAnnotation) Name() string {
	return CascadeThroughAnnotationName
}

// Name returns the name of the SchemaGenAnnotation
func (a SchemaGenAnnotation) Name() string {
	return SchemaGenAnnotationName
}

// Name returns the name of the QueryGenAnnotation
func (a QueryGenAnnotation) Name() string {
	return QueryGenAnnotationName
}

// Name returns the name of the SearchFieldAnnotation
func (a SearchFieldAnnotation) Name() string {
	return SearchFieldAnnotationName
}

// Name returns the name of the FeatureAnnotation
func (a FeatureAnnotation) Name() string {
	return FeatureAnnotationName
}

// Name returns the name of the FeatureVisibilityAnnotation
func (a FeatureVisibilityAnnotation) Name() string {
	return FeatureVisibilityAnnotationName
}

// CascadeAnnotationField sets the field name of the edge containing the ID of a record from the current schema
func CascadeAnnotationField(fieldname string) *CascadeAnnotation {
	return &CascadeAnnotation{
		Field: fieldname,
	}
}

// CascadeThroughAnnotationField sets the field name of the edge containing the ID of a record from the current schema
func CascadeThroughAnnotationField(schemas []ThroughCleanup) *CascadeThroughAnnotation {
	return &CascadeThroughAnnotation{
		Schemas: schemas,
	}
}

// SchemaGenSkip sets whether the schema generation should be skipped for this type
func SchemaGenSkip(skip bool) *SchemaGenAnnotation {
	return &SchemaGenAnnotation{
		Skip: skip,
	}
}

// SchemaSearchable sets if the schema should be searchable and generated in the search schema template
func SchemaSearchable(s bool) *SchemaGenAnnotation {
	return &SchemaGenAnnotation{
		SkipSearch: !s,
	}
}

// QueryGenSkip sets whether the query generation should be skipped for this type
func QueryGenSkip(skip bool) *QueryGenAnnotation {
	return &QueryGenAnnotation{
		Skip: skip,
	}
}

// FeatureVisibility returns a FeatureVisibilityAnnotation with the provided visibility.
func FeatureVisibility(v Visibility) *FeatureVisibilityAnnotation {
	return &FeatureVisibilityAnnotation{Visibility: v}
}

// FieldJSONPathSearchable returns a new SearchFieldAnnotation with the searchable flag set and the JSONPath set
func FieldJSONPathSearchable(path string) *SearchFieldAnnotation {
	return &SearchFieldAnnotation{
		JSONPath:   path,
		Searchable: true,
	}
}

// FieldJSONDotPathSearchable returns a new SearchFieldAnnotation with the searchable flag set and the JSONDotPath set
func FieldJSONDotPathSearchable(path string) *SearchFieldAnnotation {
	return &SearchFieldAnnotation{
		JSONDotPath: path,
		Searchable:  true,
	}
}

// FieldSearchable returns a new SearchFieldAnnotation with the searchable flag set
func FieldSearchable() *SearchFieldAnnotation {
	return &SearchFieldAnnotation{
		Searchable: true,
	}
}

// FieldAdminSearchable returns a new SearchFieldAnnotation with the exclude admin searchable flag set
func FieldAdminSearchable(s bool) *SearchFieldAnnotation {
	return &SearchFieldAnnotation{
		ExcludeAdmin: !s,
	}
}

// Features returns a new FeatureAnnotation for the provided module identifiers
func Features(features ...FeatureModule) *FeatureAnnotation {
	return &FeatureAnnotation{Features: features}
}

// FeaturesWithVisibility returns a new FeatureAnnotation with a visibility flag
func FeaturesWithVisibility(v Visibility, features ...FeatureModule) *FeatureAnnotation {
	return &FeatureAnnotation{Features: features, Visibility: v}
}

// Decode unmarshalls the CascadeAnnotation
func (a *CascadeAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the CascadeThroughAnnotation
func (a *CascadeThroughAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the SchemaGenAnnotation
func (a *SchemaGenAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the QueryGenAnnotation
func (a *QueryGenAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the SearchFieldAnnotation
func (a *SearchFieldAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the FeatureAnnotation
func (a *FeatureAnnotation) Decode(annotation any) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the FeatureVisibilityAnnotation
func (a *FeatureVisibilityAnnotation) Decode(annotation any) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}
