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

// FeatureVisibilityAnnotationName is the annotation name used to flag schema visibility
var FeatureVisibilityAnnotationName = "OPENLANE_SCHEMA_VISIBILITY"

// WorkflowEligibleAnnotationName is the annotation name for workflow-eligible fields
var WorkflowEligibleAnnotationName = "OPENLANE_WORKFLOW_ELIGIBLE"

// WebhookPayloadFieldAnnotationName is the annotation name for fields to include in webhook payloads
var WebhookPayloadFieldAnnotationName = "OPENLANE_WEBHOOK_PAYLOAD_FIELD"

// WorkflowObjectConfigAnnotationName is the annotation name for workflow object configuration
var WorkflowObjectConfigAnnotationName = "OPENLANE_WORKFLOW_OBJECT_CONFIG"

// CSVReferenceAnnotationName is the annotation name for CSV reference field mappings
var CSVReferenceAnnotationName = "OPENLANE_CSV_REFERENCE"

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
	// Schemas are also not searchable if no fields are marked as searchable
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

// WorkflowEligibleAnnotation is an annotation used to indicate that a field can be modified via workflow proposed changes
type WorkflowEligibleAnnotation struct {
	// Eligible indicates that the field can be included in workflow definitions and modified via proposed changes
	Eligible bool
}

// WebhookPayloadFieldAnnotation is an annotation used to indicate that a field should be included in webhook payloads
type WebhookPayloadFieldAnnotation struct {
	// Include indicates that the field should be included in webhook payloads
	Include bool
}

// WorkflowObjectConfigAnnotation is an annotation used to configure workflow object loading behavior
type WorkflowObjectConfigAnnotation struct {
	// EagerLoadEdges lists edges that should be eagerly loaded when loading workflow objects
	EagerLoadEdges []string
}

// CSVReferenceAnnotation is an annotation used to map CSV columns to ID fields via lookups
// All lookups are automatically scoped to the organization context from the request.
type CSVReferenceAnnotation struct {
	// CSVColumn is the friendly CSV header name that users will see (e.g., AssignedToUserEmail)
	CSVColumn string
	// MatchField is the field on the target entity to match against (e.g., email, name, ref_code)
	MatchField string
	// TargetEntity optionally specifies the target entity type when it cannot be inferred from edges
	TargetEntity string
	// CreateIfMissing allows auto-creation of missing records during CSV import (e.g., platforms)
	CreateIfMissing bool
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

// Name returns the name of the WorkflowEligibleAnnotation
func (a WorkflowEligibleAnnotation) Name() string {
	return WorkflowEligibleAnnotationName
}

// Name returns the name of the WebhookPayloadFieldAnnotation
func (a WebhookPayloadFieldAnnotation) Name() string {
	return WebhookPayloadFieldAnnotationName
}

// Name returns the name of the WorkflowObjectConfigAnnotation
func (a WorkflowObjectConfigAnnotation) Name() string {
	return WorkflowObjectConfigAnnotationName
}

// Name returns the name of the CSVReferenceAnnotation
func (a CSVReferenceAnnotation) Name() string {
	return CSVReferenceAnnotationName
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

// FieldWorkflowEligible returns a new WorkflowEligibleAnnotation with the eligible flag set
func FieldWorkflowEligible() *WorkflowEligibleAnnotation {
	return &WorkflowEligibleAnnotation{
		Eligible: true,
	}
}

// FieldWebhookPayloadField returns a new WebhookPayloadFieldAnnotation with the include flag set
func FieldWebhookPayloadField() *WebhookPayloadFieldAnnotation {
	return &WebhookPayloadFieldAnnotation{
		Include: true,
	}
}

// WorkflowObjectConfig returns a new WorkflowObjectConfigAnnotation with the specified eager load edges
func WorkflowObjectConfig(eagerLoadEdges []string) *WorkflowObjectConfigAnnotation {
	return &WorkflowObjectConfigAnnotation{
		EagerLoadEdges: eagerLoadEdges,
	}
}

// CSVRefBuilder provides a fluent interface for building CSVReferenceAnnotation.
// All lookups are automatically scoped to the organization context from the request.
// The target entity is inferred from the edge associated with the annotated ID field.
// Use TargetEntity() only when no edge exists.
//
// Example - annotate a user_id field to be populated from CSV email column:
//
//	field.String("assigned_to_user_id").
//	    Annotations(
//	        entx.CSVRef().
//	            FromColumn("AssignedToEmail").
//	            MatchOn("email"),
//	    )
//
// Example - annotate a group_ids field:
//
//	field.Strings("blocked_group_ids").
//	    Annotations(
//	        entx.CSVRef().
//	            FromColumn("BlockedGroupNames").
//	            MatchOn("name"),
//	    )
//
// Example - specify target entity when no edge exists:
//
//	field.String("control_id").
//	    Annotations(
//	        entx.CSVRef().
//	            FromColumn("ControlRefCode").
//	            MatchOn("ref_code").
//	            TargetEntity("Control"),
//	    )
type CSVRefBuilder struct {
	annotation *CSVReferenceAnnotation
}

// CSVRef starts building a CSV reference annotation
// Place this annotation on ID fields that should be populated from friendly CSV values
func CSVRef() *CSVRefBuilder {
	return &CSVRefBuilder{
		annotation: &CSVReferenceAnnotation{},
	}
}

// FromColumn sets the CSV column header that users will see and provide values in
// For example, "AssignedToEmail" means the CSV will have a column with that header
func (b *CSVRefBuilder) FromColumn(csvColumn string) *CSVRefBuilder {
	b.annotation.CSVColumn = csvColumn

	return b
}

// MatchOn sets the field on the target entity to match against when resolving
// friendly CSV values to IDs. Common values: email, name, ref_code, slug.
func (b *CSVRefBuilder) MatchOn(field string) *CSVRefBuilder {
	b.annotation.MatchField = field

	return b
}

// TargetEntity explicitly sets the target entity type for lookups.
// Only needed when the target cannot be inferred from an edge definition.
func (b *CSVRefBuilder) TargetEntity(entity string) *CSVRefBuilder {
	b.annotation.TargetEntity = entity

	return b
}

// CreateIfMissing enables auto-creation of missing records during CSV import
// only for entities where auto-creation makes sense
func (b *CSVRefBuilder) CreateIfMissing() *CSVRefBuilder {
	b.annotation.CreateIfMissing = true

	return b
}

// Name returns the annotation name, implementing the ent Annotation interface.
func (b *CSVRefBuilder) Name() string {
	return b.annotation.Name()
}

// MarshalJSON serializes the builder as the underlying CSVReferenceAnnotation.
// This ensures ent stores the annotation in the expected format for decoding.
func (b *CSVRefBuilder) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.annotation)
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

// Decode unmarshalls the WorkflowEligibleAnnotation
func (a *WorkflowEligibleAnnotation) Decode(annotation interface{}) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the WebhookPayloadFieldAnnotation
func (a *WebhookPayloadFieldAnnotation) Decode(annotation any) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the WorkflowObjectConfigAnnotation
func (a *WorkflowObjectConfigAnnotation) Decode(annotation any) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}

// Decode unmarshalls the CSVReferenceAnnotation
func (a *CSVReferenceAnnotation) Decode(annotation any) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, a)
}
