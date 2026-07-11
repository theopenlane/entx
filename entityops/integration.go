package entityops

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"

	"github.com/theopenlane/entx"
)

// integrationSystemFieldNames is the set of system-managed field names excluded from integration
// mapping unless they carry an explicit IntegrationMappingFieldAnnotation
var integrationSystemFieldNames = map[string]struct{}{
	"id":                        {},
	"owner_id":                  {},
	"organization_id":           {},
	"org_id":                    {},
	"created_at":                {},
	"updated_at":                {},
	"created_by":                {},
	"updated_by":                {},
	"deleted_at":                {},
	"deleted_by":                {},
	workflowEligibleMarkerField: {},
}

// EntityRuntimeDefault represents one integration-injected field used by stock ingest preparation
type EntityRuntimeDefault struct {
	// GoField is the Go struct field name on the ent create input that receives the injected value
	GoField string
	// Required reports whether the field is non-pointer on the input type, which determines whether
	// the zero value or nil is checked before injection
	Required bool
	// IntegrationField is the Go struct field name on *generated.Integration that sources the value
	IntegrationField string
}

// integrationFieldMeta carries the per-field integration mapping metadata folded onto EntityField
type integrationFieldMeta struct {
	// InputKey is the GraphQL create-input field name (lowerCamel)
	InputKey string
	// InputGoField is the exported Go struct field name for the input key on ent create inputs
	InputGoField string
	// Required reports whether the field is required (non-optional in ent)
	Required bool
	// UpsertKey reports whether the field participates in dedupe/upsert matching
	UpsertKey bool
	// LookupKey reports whether the field participates in stock ingest lookup matching
	LookupKey bool
	// FromIntegration reports whether the field value is injected from the integration record at ingest time
	FromIntegration bool
}

// integrationSchemaMeta carries the schema-level integration mapping metadata folded onto EntitySchema
type integrationSchemaMeta struct {
	// Mapped reports whether the schema has at least one integration mapping field
	Mapped bool
	// StockPersist reports whether the schema opts into the generated stock ingest persistence path
	StockPersist bool
	// RuntimeDefaults are the integration-injected field defaults applied during stock ingest preparation
	RuntimeDefaults []EntityRuntimeDefault
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
	}

	var runtimeDefaults []EntityRuntimeDefault

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
		fromIntegration := ant != nil && ant.FromIntegration

		meta[field.Name] = integrationFieldMeta{
			InputKey:        key,
			InputGoField:    goField,
			Required:        !field.Optional,
			UpsertKey:       ant != nil && ant.UpsertKey,
			LookupKey:       ant != nil && ant.LookupKey,
			FromIntegration: fromIntegration,
		}

		if stockPersist && fromIntegration {
			instField, err := integrationFieldForEntField(field.Name)
			if err != nil {
				return nil, schemaMeta, err
			}

			runtimeDefaults = append(runtimeDefaults, EntityRuntimeDefault{
				GoField:          goField,
				Required:         !field.Optional,
				IntegrationField: instField,
			})
		}
	}

	schemaMeta.Mapped = len(meta) > 0
	schemaMeta.StockPersist = stockPersist
	schemaMeta.RuntimeDefaults = runtimeDefaults

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

// integrationFieldForEntField maps an ent field name to the Go field name on *generated.Integration
func integrationFieldForEntField(entField string) (string, error) {
	switch entField {
	case "integration_id":
		return "ID", nil
	case "owner_id":
		return "OwnerID", nil
	case "platform_id":
		return "PlatformID", nil
	default:
		return "", ErrNoIntegrationFieldMapping
	}
}
