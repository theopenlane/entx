package oscalgen

import "encoding/json"

// OSCALModelAnnotationName is the annotation name for schema-level OSCAL mapping
const OSCALModelAnnotationName = "OSCALModel"

// OSCALFieldAnnotationName is the annotation name for field-level OSCAL mapping
const OSCALFieldAnnotationName = "OSCALField"

// OSCALRelationshipAnnotationName is the annotation name for edge-level OSCAL mapping
const OSCALRelationshipAnnotationName = "OSCALRelationship"

// OSCALModelType identifies a supported OSCAL document model
type OSCALModelType string

const (
	// OSCALModelComponentDefinition maps to OSCAL Component Definition
	OSCALModelComponentDefinition OSCALModelType = "component-definition"
	// OSCALModelSSP maps to OSCAL System Security Plan
	OSCALModelSSP OSCALModelType = "ssp"
	// OSCALModelPOAM maps to OSCAL Plan of Action and Milestones
	OSCALModelPOAM OSCALModelType = "poam"
)

// OSCALFieldRole captures semantic field intent for OSCAL mapping
type OSCALFieldRole string

const (
	// OSCALFieldRoleTitle marks a field as representing a title in OSCAL documents
	OSCALFieldRoleTitle OSCALFieldRole = "title"
	// OSCALFieldRoleDescription marks a field as representing a description in OSCAL documents
	OSCALFieldRoleDescription OSCALFieldRole = "description"
	// OSCALFieldRoleSystemName marks a field as representing a system name in OSCAL documents
	OSCALFieldRoleSystemName OSCALFieldRole = "system-name"
	// OSCALFieldRoleInventoryItemIdentifier marks a field as representing an inventory item identifier in OSCAL documents
	OSCALFieldRoleInventoryItemIdentifier OSCALFieldRole = "inventory-item-identifier"
	// OSCALFieldRoleImplementationDetails marks a field as representing implementation details in OSCAL documents
	OSCALFieldRoleImplementationDetails OSCALFieldRole = "implementation-details"
	// OSCALFieldRoleImplementationStatus marks a field as representing implementation status in OSCAL documents
	OSCALFieldRoleImplementationStatus OSCALFieldRole = "implementation-status"
	// OSCALFieldRoleResponsibleRole marks a field as representing a responsible role in OSCAL documents
	OSCALFieldRoleResponsibleRole OSCALFieldRole = "responsible-role"
	// OSCALFieldRoleControlID marks a field as representing a control ID in OSCAL documents
	OSCALFieldRoleControlID OSCALFieldRole = "control-id"
	// OSCALFieldRoleStatementID marks a field as representing a statement ID in OSCAL documents
	OSCALFieldRoleStatementID OSCALFieldRole = "statement-id"
	// OSCALFieldRoleUUID marks a field as representing a UUID in OSCAL documents
	OSCALFieldRoleUUID OSCALFieldRole = "uuid"
)

// OSCALRelationshipRole captures semantic edge intent for OSCAL mapping
type OSCALRelationshipRole string

const (
	// OSCALRelationshipRoleComponentContains marks a relationship as representing a component containing another component
	OSCALRelationshipRoleComponentContains OSCALRelationshipRole = "component-contains"
	// OSCALRelationshipRoleSatisfiesControl marks a relationship as representing satisfaction of a control
	OSCALRelationshipRoleSatisfiesControl OSCALRelationshipRole = "satisfies-control"
	// OSCALRelationshipRoleImplementedByComponent marks a relationship as representing implementation by a component
	OSCALRelationshipRoleImplementedByComponent OSCALRelationshipRole = "implemented-by-component"
	// OSCALRelationshipRoleLinksToControlID marks a relationship as linking to a control ID
	OSCALRelationshipRoleLinksToControlID OSCALRelationshipRole = "links-to-control-id"
	// OSCALRelationshipRoleLinksToStatementID marks a relationship as linking to a statement ID
	OSCALRelationshipRoleLinksToStatementID OSCALRelationshipRole = "links-to-statement-id"
)

// OSCALModel is a schema-level annotation for OSCAL model participation
type OSCALModel struct {
	// Models identifies which OSCAL models this schema participates in
	Models []OSCALModelType
	// Assembly captures the target semantic assembly (not a JSON path)
	Assembly string
}

// OSCALModelOption configures OSCALModel annotation options
type OSCALModelOption func(*OSCALModel)

// NewOSCALModel creates a new schema-level OSCAL annotation
func NewOSCALModel(opts ...OSCALModelOption) OSCALModel {
	model := OSCALModel{}

	for _, opt := range opts {
		opt(&model)
	}

	model.Models = normalizeOSCALModels(model.Models)

	return model
}

// WithOSCALModels sets target OSCAL models for a schema
func WithOSCALModels(models ...OSCALModelType) OSCALModelOption {
	return func(m *OSCALModel) {
		m.Models = normalizeOSCALModels(append(m.Models, models...))
	}
}

// WithOSCALAssembly sets the schema's OSCAL semantic assembly
func WithOSCALAssembly(assembly string) OSCALModelOption {
	return func(m *OSCALModel) {
		m.Assembly = assembly
	}
}

// Name returns the annotation name
func (OSCALModel) Name() string {
	return OSCALModelAnnotationName
}

// normalizeModels normalizes model constraints for OSCALModel
func (a *OSCALModel) normalizeModels() {
	a.Models = normalizeOSCALModels(a.Models)
}

// Decode unmarshals the OSCALModel annotation
func (a *OSCALModel) Decode(annotation any) error {
	return decodeOSCALAnnotation(annotation, a)
}

// OSCALField is a field-level annotation for semantic OSCAL mapping roles
type OSCALField struct {
	// Role identifies the field's semantic role for OSCAL document builders
	Role OSCALFieldRole
	// Models optionally scopes this field role to specific OSCAL models
	Models []OSCALModelType
	// IdentityAnchor marks fields that should be treated as stable identity anchors
	IdentityAnchor bool
}

// OSCALFieldOption configures OSCALField annotation options
type OSCALFieldOption func(*OSCALField)

// NewOSCALField creates a new field-level OSCAL annotation
func NewOSCALField(role OSCALFieldRole, opts ...OSCALFieldOption) OSCALField {
	field := OSCALField{
		Role: role,
	}

	for _, opt := range opts {
		opt(&field)
	}

	field.Models = normalizeOSCALModels(field.Models)

	return field
}

// WithOSCALFieldModels scopes a field role to specific OSCAL models
func WithOSCALFieldModels(models ...OSCALModelType) OSCALFieldOption {
	return func(f *OSCALField) {
		f.Models = normalizeOSCALModels(append(f.Models, models...))
	}
}

// WithOSCALIdentityAnchor marks a field as an OSCAL identity anchor
func WithOSCALIdentityAnchor() OSCALFieldOption {
	return func(f *OSCALField) {
		f.IdentityAnchor = true
	}
}

// Name returns the annotation name
func (OSCALField) Name() string {
	return OSCALFieldAnnotationName
}

// normalizeModels normalizes model constraints for OSCALField
func (a *OSCALField) normalizeModels() {
	a.Models = normalizeOSCALModels(a.Models)
}

// Decode unmarshals the OSCALField annotation
func (a *OSCALField) Decode(annotation any) error {
	return decodeOSCALAnnotation(annotation, a)
}

// OSCALRelationship is an edge-level annotation for OSCAL semantics
type OSCALRelationship struct {
	// Role identifies semantic relationship intent for OSCAL builders
	Role OSCALRelationshipRole
	// Models optionally scopes this relationship role to specific OSCAL models
	Models []OSCALModelType
}

// OSCALRelationshipOption configures OSCALRelationship annotation options
type OSCALRelationshipOption func(*OSCALRelationship)

// NewOSCALRelationship creates a new edge-level OSCAL annotation
func NewOSCALRelationship(role OSCALRelationshipRole, opts ...OSCALRelationshipOption) OSCALRelationship {
	rel := OSCALRelationship{
		Role: role,
	}

	for _, opt := range opts {
		opt(&rel)
	}

	rel.Models = normalizeOSCALModels(rel.Models)

	return rel
}

// WithOSCALRelationshipModels scopes a relationship role to specific OSCAL models
func WithOSCALRelationshipModels(models ...OSCALModelType) OSCALRelationshipOption {
	return func(r *OSCALRelationship) {
		r.Models = normalizeOSCALModels(append(r.Models, models...))
	}
}

// Name returns the annotation name
func (OSCALRelationship) Name() string {
	return OSCALRelationshipAnnotationName
}

// normalizeModels normalizes model constraints for OSCALRelationship
func (a *OSCALRelationship) normalizeModels() {
	a.Models = normalizeOSCALModels(a.Models)
}

// Decode unmarshals the OSCALRelationship annotation
func (a *OSCALRelationship) Decode(annotation any) error {
	return decodeOSCALAnnotation(annotation, a)
}

// oscalModelsNormalizer defines the normalization hook required by decodeOSCALAnnotation
type oscalModelsNormalizer interface {
	normalizeModels()
}

// decodeOSCALAnnotation decodes a raw Ent annotation into a typed OSCAL annotation and normalizes model constraints
func decodeOSCALAnnotation[T oscalModelsNormalizer](annotation any, target T) error {
	buf, err := json.Marshal(annotation)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, target); err != nil {
		return err
	}

	target.normalizeModels()

	return nil
}

// normalizeOSCALModels removes empty and duplicate model values while preserving order
func normalizeOSCALModels(models []OSCALModelType) []OSCALModelType {
	seen := make(map[OSCALModelType]struct{}, len(models))
	out := make([]OSCALModelType, 0, len(models))

	for _, model := range models {
		if model == "" {
			continue
		}

		if _, ok := seen[model]; ok {
			continue
		}

		seen[model] = struct{}{}
		out = append(out, model)
	}

	return out
}
