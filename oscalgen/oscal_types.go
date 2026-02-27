package oscalgen

import oscal113 "github.com/theopenlane/oscalot/pkg/types/oscal-1-1-3"

// OSCALTypesVersion is the pinned OSCAL schema version from the external types package
const OSCALTypesVersion = oscal113.Version

// OscalModels aliases the external OSCAL models root container
type OscalModels = oscal113.OscalModels

// ComponentDefinition aliases the OSCAL component definition model
type ComponentDefinition = oscal113.ComponentDefinition

// SystemSecurityPlan aliases the OSCAL system security plan model
type SystemSecurityPlan = oscal113.SystemSecurityPlan

// PlanOfActionAndMilestones aliases the OSCAL POA&M model
type PlanOfActionAndMilestones = oscal113.PlanOfActionAndMilestones

// Metadata aliases OSCAL document metadata
type Metadata = oscal113.Metadata

// BackMatter aliases OSCAL back matter
type BackMatter = oscal113.BackMatter

// ControlImplementation aliases OSCAL control implementation for SSP
type ControlImplementation = oscal113.ControlImplementation

// ImplementedRequirement aliases OSCAL implemented requirement structures
type ImplementedRequirement = oscal113.ImplementedRequirement

// ByComponent aliases OSCAL by-component implementation structures
type ByComponent = oscal113.ByComponent

// Statement aliases OSCAL statement implementation structures
type Statement = oscal113.Statement

// oscalSchemaInfo stores internal schema-level OSCAL mapping details before template rendering
type oscalSchemaInfo struct {
	// name is the ent schema name
	name string
	// models are normalized OSCAL model identifiers for this schema
	models []string
	// assembly is the semantic OSCAL assembly mapping for this schema
	assembly string
	// fields are internal field-level OSCAL mappings
	fields []oscalFieldInfo
	// relationships are internal edge-level OSCAL mappings
	relationships []oscalRelationshipInfo
}

// oscalFieldInfo stores internal field-level OSCAL mapping details
type oscalFieldInfo struct {
	// name is the ent field name
	name string
	// role is the semantic OSCAL role for the field
	role string
	// models are optional OSCAL model constraints for the field
	models []string
	// identityAnchor indicates the field is an OSCAL identity anchor
	identityAnchor bool
}

// oscalRelationshipInfo stores internal edge-level OSCAL mapping details
type oscalRelationshipInfo struct {
	// name is the ent edge name
	name string
	// role is the semantic OSCAL relationship role
	role string
	// models are optional OSCAL model constraints for the relationship
	models []string
}

// oscalTemplateData is the top-level template input for generated OSCAL registry files
type oscalTemplateData struct {
	// Package is the output package name for the generated file
	Package string
	// Schemas contains template-ready schema mappings
	Schemas []oscalTemplateSchema
}

// oscalTemplateSchema is the template-ready schema mapping representation
type oscalTemplateSchema struct {
	// Name is the schema name key
	Name string
	// Models are OSCAL model identifiers supported by the schema
	Models []string
	// Assembly is the schema's semantic OSCAL assembly
	Assembly string
	// Fields contains template-ready field mappings
	Fields []oscalTemplateField
	// Relationships contains template-ready relationship mappings
	Relationships []oscalTemplateRelationship
}

// oscalTemplateField is the template-ready field mapping representation
type oscalTemplateField struct {
	// Name is the field name key
	Name string
	// Role is the semantic OSCAL field role
	Role string
	// Models are optional OSCAL model constraints for the field
	Models []string
	// IdentityAnchor indicates the field is an OSCAL identity anchor
	IdentityAnchor bool
}

// oscalTemplateRelationship is the template-ready relationship mapping representation
type oscalTemplateRelationship struct {
	// Name is the relationship name key
	Name string
	// Role is the semantic OSCAL relationship role
	Role string
	// Models are optional OSCAL model constraints for the relationship
	Models []string
}
