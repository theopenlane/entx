package entityops

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"

	"github.com/theopenlane/entx"
)

// integrationSystemFieldNames is the set of system-managed field names excluded from provider mapping
var integrationSystemFieldNames = map[string]struct{}{
	"id":                        {},
	"owner_id":                  {},
	"organization_id":           {},
	"org_id":                    {},
	"created_at":                {},
	"updated_at":                {},
	"created_by":                {},
	"updated_by":                {},
	"updated_by_impersonator":   {},
	"deleted_at":                {},
	"deleted_by":                {},
	workflowEligibleMarkerField: {},
}

// integrationFieldMeta carries the per-field integration mapping metadata folded onto EntityField
type integrationFieldMeta struct {
	// InputKey is the GraphQL create-input field name (lowerCamel)
	InputKey string
	// InputGoField is the exported Go struct field name for the input key on ent create inputs
	InputGoField string
	// LookupKey reports whether the field is the ingest upsert lookup column for its schema
	LookupKey bool
	// SystemControlled excludes the field from provider mappings
	SystemControlled bool
	// Volatile excludes the field from triggering an ingest change
	Volatile bool
}

// integrationSchemaMeta carries the schema-level integration mapping metadata folded onto EntitySchema
type integrationSchemaMeta struct {
	// Mapped reports whether the schema has at least one integration mapping field
	Mapped bool
	// LookupAlternatives are the schema's annotation-declared composite ingest lookup keys, if any
	LookupAlternatives [][]string
	// InstanceScoped reports whether same-key records from different source instances are distinct rows
	InstanceScoped bool
}

// collectIntegrationMapping returns the per-field integration mapping metadata (keyed by ent field
// name) and the schema-level mapping metadata for one schema. It mirrors the field eligibility,
// include/exclude, and runtime-default rules of the standalone integration mapping generator so the
// metadata can be folded onto the unified entityops field catalog
func collectIntegrationMapping(schema *load.Schema) (map[string]integrationFieldMeta, integrationSchemaMeta, error) {
	meta := map[string]integrationFieldMeta{}

	var schemaMeta integrationSchemaMeta

	if schema == nil {
		return meta, schemaMeta, nil
	}

	schemaAnt := integrationSchemaAnnotation(schema)
	stockPersist := schemaAnt != nil && schemaAnt.StockPersist
	includeSet := map[string]struct{}{}
	excludeSet := map[string]struct{}{}
	hasInclude := false

	if schemaAnt != nil {
		for _, name := range schemaAnt.Include {
			includeSet[name] = struct{}{}
		}

		hasInclude = len(includeSet) > 0

		for _, name := range schemaAnt.Exclude {
			excludeSet[name] = struct{}{}
		}

		schemaMeta.LookupAlternatives = schemaAnt.LookupAlternatives
		schemaMeta.InstanceScoped = schemaAnt.InstanceScoped
	}

	for _, field := range schema.Fields {
		if !integrationFieldEligible(field) {
			continue
		}

		ant := integrationFieldAnnotation(field)

		if !integrationFieldIncluded(field.Name, includeSet, excludeSet, hasInclude, stockPersist, ant) {
			continue
		}

		if schemaAnt == nil && ant == nil {
			continue
		}

		key := ""
		if ant != nil {
			key = ant.Key
		}

		if key == "" {
			// snake_case so the ingest payload key matches the create-input's snake_case json tag,
			// letting it unmarshal directly without a camel->snake re-key map
			key = field.Name
		}

		goField := templates.ToGo(key)

		meta[field.Name] = integrationFieldMeta{
			InputKey:         key,
			InputGoField:     goField,
			LookupKey:        ant != nil && ant.LookupKey,
			SystemControlled: isIntegrationSystemField(field.Name) || (ant != nil && ant.SystemControlled),
			Volatile:         ant != nil && ant.Volatile,
		}
	}

	for _, field := range meta {
		if !field.SystemControlled {
			schemaMeta.Mapped = true
			break
		}
	}

	return meta, schemaMeta, nil
}

// integrationFieldIncluded reports whether a field should be collected given the schema's
// include/exclude/system-field rules. The include list takes full precedence: when present, only
// listed fields are included. When no include list is set, excluded fields and system-managed fields
// are skipped unless the schema uses stock persistence and the field carries an explicit annotation
func integrationFieldIncluded(fieldName string, includeSet, excludeSet map[string]struct{}, hasInclude, stockPersist bool, ant *entx.IntegrationMappingFieldAnnotation) bool {
	if hasInclude {
		_, ok := includeSet[fieldName]

		return ok
	}

	if _, ok := excludeSet[fieldName]; ok {
		return false
	}

	if isIntegrationSystemField(fieldName) {
		return stockPersist && ant != nil
	}

	return true
}

// integrationFieldEligible reports whether a field is eligible for integration mapping, excluding
// sensitive fields and fields skipped from the GraphQL type or both mutation inputs
func integrationFieldEligible(field *load.Field) bool {
	if field.Sensitive {
		return false
	}

	ant, ok := entx.GetAnnotation[*entgql.Annotation](field)
	if !ok {
		return true
	}

	switch {
	case ant.Skip.Is(entgql.SkipType):
		return false
	case ant.Skip.Is(entgql.SkipAll):
		return false
	case ant.Skip.Is(entgql.SkipMutationCreateInput) && ant.Skip.Is(entgql.SkipMutationUpdateInput):
		return false
	default:
		return true
	}
}

// integrationFieldAnnotation retrieves the IntegrationMappingFieldAnnotation from a field
func integrationFieldAnnotation(field *load.Field) *entx.IntegrationMappingFieldAnnotation {
	ant, ok := entx.GetAnnotation[*entx.IntegrationMappingFieldAnnotation](field)
	if !ok {
		return nil
	}

	return ant
}

// integrationSchemaAnnotation retrieves the IntegrationMappingSchemaAnnotation from a schema
func integrationSchemaAnnotation(schema *load.Schema) *entx.IntegrationMappingSchemaAnnotation {
	ant, ok := entx.GetAnnotation[*entx.IntegrationMappingSchemaAnnotation](schema)
	if !ok {
		return nil
	}

	return ant
}

// isIntegrationSystemField reports whether a field name is system-managed and excluded from mapping by default
func isIntegrationSystemField(name string) bool {
	_, ok := integrationSystemFieldNames[name]

	return ok
}
