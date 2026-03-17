package integrationmapping

import (
	"cmp"
	"embed"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/rs/zerolog/log"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/entx"
)

//go:embed templates/*
var _templates embed.FS

const dirPermissions = 0755

// integrationSystemFieldNames is the set of system-managed field names excluded from integration
// mapping unless they carry an explicit IntegrationMappingFieldAnnotation
var integrationSystemFieldNames = map[string]struct{}{
	"id":                       {},
	"owner_id":                 {},
	"organization_id":          {},
	"org_id":                   {},
	"created_at":               {},
	"updated_at":               {},
	"created_by":               {},
	"updated_by":               {},
	"deleted_at":               {},
	"deleted_by":               {},
	"workflow_eligible_marker": {},
}

// MappingData holds the data for generating integration mapping files
type MappingData struct {
	// PackageName is the Go package name for generated files
	PackageName string
	// EntPackage is the ent generated package import path used for typed ingest contracts
	EntPackage string
	// GalaPackage is the gala package import path used for typed ingest contracts
	GalaPackage string
	// GenerateIngestContracts controls whether typed ingest payload contracts are emitted
	GenerateIngestContracts bool
	// Schemas contains all schemas with integration mapping fields
	Schemas []MappingSchema
}

// MappingSchema represents a schema with integration mapping fields
type MappingSchema struct {
	// Name is the schema name (e.g., Vulnerability)
	Name string
	// ConstName is the generated Go constant name for the schema
	ConstName string
	// TableName is the SQL table name for the schema
	TableName string
	// IngestTopicConstName is the generated Go constant name for the schema ingest topic
	IngestTopicConstName string
	// IngestTopic is the schema-scoped ingest topic name
	IngestTopic string
	// IngestRequestTypeName is the generated Go type name for the typed ingest request payload
	IngestRequestTypeName string
	// IngestTopicVarName is the generated Go variable name for the typed ingest topic
	IngestTopicVarName string
	// InputTypeName is the ent-generated input type carried by the ingest request payload
	InputTypeName string
	// LookupFields are the ent/go field pairs used for stock ingest lookup matching
	LookupFields []IngestLookupField
	// RuntimeDefaults are the runtime-injected values applied during stock ingest preparation
	RuntimeDefaults []IngestRuntimeDefault
	// StockPersist indicates the schema can use the generated stock ingest persistence path
	StockPersist bool
	// Fields contains integration mapping field metadata
	Fields []MappingField
}

// IngestRuntimeDefault represents one runtime-injected field used by stock ingest persistence
type IngestRuntimeDefault struct {
	Field   string
	GoField string
	Source  string
}

// IngestLookupField represents one stock ingest lookup field
type IngestLookupField struct {
	Field   string
	GoField string
}

// MappingField represents a field that can be targeted by integration mappings
type MappingField struct {
	// InputKey is the GraphQL input field name (lowerCamel)
	InputKey string
	// GoField is the exported Go struct field name on ent input types
	GoField string
	// ConstName is the generated Go constant name for the input key
	ConstName string
	// EntField is the ent field name (snake_case)
	EntField string
	// Type is the ent field type (string, time, json, etc)
	Type string
	// Required indicates the field is required for input (non-optional in ent)
	Required bool
	// UpsertKey indicates the field participates in dedupe/upsert matching
	UpsertKey bool
	// LookupKey indicates the field participates in stock ingest lookup matching
	LookupKey bool
	// RuntimeDefault identifies the runtime source that injects this field during stock ingest
	RuntimeDefault string
}

// collectMappingData collects mapping data from all schemas in the graph
func collectMappingData(g *gen.Graph, c *Config) MappingData {
	data := MappingData{
		PackageName:             c.PackageName,
		EntPackage:              c.EntPackage,
		GalaPackage:             c.GalaPackage,
		GenerateIngestContracts: c.EntPackage != "" && c.GalaPackage != "",
		Schemas:                 []MappingSchema{},
	}

	for _, node := range g.Nodes {
		if checkSchemaGenSkip(node) || checkQueryGenSkip(node) {
			continue
		}

		if strings.HasSuffix(node.Name, "History") {
			continue
		}

		hasCreate := !checkSkipMutationCreateInput(node)
		hasUpdate := !checkSkipMutationUpdateInput(node)

		if !hasCreate && !hasUpdate {
			continue
		}

		schema := getEntSchema(g, node.Name)
		if schema == nil {
			continue
		}

		fields := collectMappingFields(schema, node.Name)
		if len(fields) == 0 {
			continue
		}

		schemaAnt := getSchemaAnnotation(schema)
		inputTypeName := ""

		switch {
		case hasCreate:
			inputTypeName = "Create" + node.Name + "Input"
		case hasUpdate:
			inputTypeName = "Update" + node.Name + "Input"
		}

		lookupFields := collectLookupFields(fields)
		stockPersist := schemaAnt != nil && schemaAnt.StockPersist
		runtimeDefaults := []IngestRuntimeDefault{}

		if stockPersist {
			runtimeDefaults = collectRuntimeDefaults(fields)
		}

		data.Schemas = append(data.Schemas, MappingSchema{
			Name:                  node.Name,
			ConstName:             schemaConstName(node.Name),
			TableName:             schemaTableName(schema),
			IngestTopicConstName:  schemaIngestTopicConstName(node.Name),
			IngestTopic:           schemaIngestTopicName(node.Name),
			IngestRequestTypeName: schemaIngestRequestTypeName(node.Name),
			IngestTopicVarName:    schemaIngestTopicVarName(node.Name),
			InputTypeName:         inputTypeName,
			LookupFields:          lookupFields,
			RuntimeDefaults:       runtimeDefaults,
			StockPersist:          stockPersist,
			Fields:                fields,
		})
	}

	slices.SortFunc(data.Schemas, func(a, b MappingSchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return data
}

// schemaConstName returns the generated Go constant name for a schema mapping identifier
func schemaConstName(schemaName string) string {
	if schemaName == "" {
		return ""
	}

	return "IntegrationMappingSchema" + templates.ToGo(schemaName)
}

// fieldConstName returns the generated Go constant name for a field mapping key within a schema
func fieldConstName(schemaName, inputKey string) string {
	if schemaName == "" || inputKey == "" {
		return ""
	}

	return "IntegrationMapping" + templates.ToGo(schemaName) + templates.ToGo(inputKey)
}

// schemaIngestTopicConstName returns the generated Go constant name for a schema's ingest topic
func schemaIngestTopicConstName(schemaName string) string {
	if schemaName == "" {
		return ""
	}

	return "IntegrationIngestTopic" + templates.ToGo(schemaName) + "Requested"
}

// schemaIngestTopicName returns the ingest topic string value for a schema
func schemaIngestTopicName(schemaName string) string {
	if schemaName == "" {
		return ""
	}

	return "integration.ingest." + strcase.SnakeCase(schemaName) + ".requested"
}

// schemaIngestRequestTypeName returns the generated Go type name for one schema's ingest request payload
func schemaIngestRequestTypeName(schemaName string) string {
	if schemaName == "" {
		return ""
	}

	return "IntegrationIngest" + templates.ToGo(schemaName) + "Requested"
}

// schemaIngestTopicVarName returns the generated Go variable name for one schema's typed ingest topic.
func schemaIngestTopicVarName(schemaName string) string {
	if schemaName == "" {
		return ""
	}

	return "IntegrationIngest" + templates.ToGo(schemaName) + "RequestedTopic"
}

// collectMappingFields returns integration mapping fields for a schema.
func collectMappingFields(schema *load.Schema, schemaName string) []MappingField {
	if schema == nil {
		return nil
	}

	fields := make([]MappingField, 0)
	schemaAnt := getSchemaAnnotation(schema)

	includeSet := make(map[string]struct{})
	excludeSet := make(map[string]struct{})
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

	for _, field := range schema.Fields {
		if hasInclude {
			if _, ok := includeSet[field.Name]; !ok {
				continue
			}
		}

		if _, ok := excludeSet[field.Name]; ok {
			continue
		}

		if !isFieldMappingEligible(field) {
			continue
		}

		ant := getFieldAnnotation(field)

		// system fields are excluded unless the schema opts into stock persistence
		// and the field carries an explicit annotation (typically with RuntimeDefault)
		if !hasInclude && isSystemField(field.Name) {
			if schemaAnt == nil || !schemaAnt.StockPersist || ant == nil {
				continue
			}
		}

		if schemaAnt == nil && ant == nil {
			continue
		}

		key := ""
		if ant != nil {
			key = ant.Key
		}

		if key == "" {
			goName := templates.ToGo(field.Name)
			key = templates.ToGoPrivate(goName)
		}

		upsert := ant != nil && ant.UpsertKey
		lookup := ant != nil && ant.LookupKey
		runtimeDefault := ""

		if ant != nil {
			runtimeDefault = ant.RuntimeDefault
		}

		fields = append(fields, MappingField{
			InputKey:       key,
			GoField:        templates.ToGo(key),
			ConstName:      fieldConstName(schemaName, key),
			EntField:       field.Name,
			Type:           field.Info.Type.String(),
			Required:       !field.Optional,
			UpsertKey:      upsert,
			LookupKey:      lookup,
			RuntimeDefault: runtimeDefault,
		})
	}

	slices.SortFunc(fields, func(a, b MappingField) int {
		return cmp.Compare(a.InputKey, b.InputKey)
	})

	return fields
}

// collectLookupFields returns the subset of fields marked as lookup keys
func collectLookupFields(fields []MappingField) []IngestLookupField {
	keys := make([]IngestLookupField, 0)

	for _, field := range fields {
		if field.LookupKey {
			keys = append(keys, IngestLookupField{
				Field:   field.EntField,
				GoField: field.GoField,
			})
		}
	}

	return keys
}

// collectRuntimeDefaults returns the subset of fields that carry a RuntimeDefault source
func collectRuntimeDefaults(fields []MappingField) []IngestRuntimeDefault {
	defaults := make([]IngestRuntimeDefault, 0)

	for _, field := range fields {
		if field.RuntimeDefault != "" {
			defaults = append(defaults, IngestRuntimeDefault{
				Field:   field.EntField,
				GoField: field.GoField,
				Source:  field.RuntimeDefault,
			})
		}
	}

	return defaults
}

// getFieldAnnotation retrieves the IntegrationMappingFieldAnnotation from a field
func getFieldAnnotation(field *load.Field) *entx.IntegrationMappingFieldAnnotation {
	ant := &entx.IntegrationMappingFieldAnnotation{}

	if raw, ok := field.Annotations[ant.Name()]; ok {
		if err := ant.Decode(raw); err != nil {
			return nil
		}

		return ant
	}

	return nil
}

// getSchemaAnnotation retrieves the IntegrationMappingSchemaAnnotation from a schema
func getSchemaAnnotation(schema *load.Schema) *entx.IntegrationMappingSchemaAnnotation {
	ant := &entx.IntegrationMappingSchemaAnnotation{}

	if raw, ok := schema.Annotations[ant.Name()]; ok {
		if err := ant.Decode(raw); err != nil {
			return nil
		}

		return ant
	}

	return nil
}

// isFieldMappingEligible checks if a field is eligible for integration mapping
func isFieldMappingEligible(field *load.Field) bool {
	if field.Sensitive {
		return false
	}

	if entSkipType(field) {
		return false
	}

	if entSkipMutationInputs(field) {
		return false
	}

	return true
}

// entSkipType checks if the field has entgql.SkipType or SkipAll
func entSkipType(field *load.Field) bool {
	entAnt := &entgql.Annotation{}

	if ant, ok := field.Annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			switch {
			case entAnt.Skip.Is(entgql.SkipType):
				return true
			case entAnt.Skip.Is(entgql.SkipAll):
				return true
			}
		}
	}

	return false
}

// entSkipMutationInputs checks if the field is skipped from both create and update inputs
func entSkipMutationInputs(field *load.Field) bool {
	entAnt := &entgql.Annotation{}

	if ant, ok := field.Annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			if entAnt.Skip.Is(entgql.SkipMutationCreateInput) && entAnt.Skip.Is(entgql.SkipMutationUpdateInput) {
				return true
			}
		}
	}

	return false
}

// isSystemField reports whether a field name is system-managed and excluded from mapping by default
func isSystemField(name string) bool {
	_, ok := integrationSystemFieldNames[name]

	return ok
}

// schemaTableName returns the SQL table name for a schema
func schemaTableName(schema *load.Schema) string {
	if schema == nil {
		return ""
	}

	if entSQLMap, ok := schema.Annotations["EntSQL"].(map[string]any); ok {
		if table, ok := entSQLMap["table"].(string); ok && table != "" {
			return table
		}
	}

	return strcase.SnakeCase(schema.Name)
}

// getEntSchema returns the schema for a given name from the graph's schema list
func getEntSchema(graph *gen.Graph, name string) *load.Schema {
	for _, s := range graph.Schemas {
		if s.Name == name {
			return s
		}
	}

	return nil
}

// checkSchemaGenSkip checks if the type has the Schema Skip annotation
func checkSchemaGenSkip(node *gen.Type) bool {
	schemaGenAnt := &entx.SchemaGenAnnotation{}

	if ant, ok := node.Annotations[schemaGenAnt.Name()]; ok {
		if err := schemaGenAnt.Decode(ant); err != nil {
			return false
		}

		return schemaGenAnt.Skip
	}

	return false
}

// checkQueryGenSkip checks if the type has the QueryGen Skip annotation
func checkQueryGenSkip(node *gen.Type) bool {
	queryGenAnt := &entx.QueryGenAnnotation{}

	if ant, ok := node.Annotations[queryGenAnt.Name()]; ok {
		if err := queryGenAnt.Decode(ant); err != nil {
			return false
		}

		return queryGenAnt.Skip
	}

	return false
}

// checkSkipMutationCreateInput returns true if no CreateInput type is generated for this schema
func checkSkipMutationCreateInput(node *gen.Type) bool {
	entgqlAnt := &entgql.Annotation{}

	ant, ok := node.Annotations[entgqlAnt.Name()]
	if !ok {
		return true
	}

	if err := entgqlAnt.Decode(ant); err != nil {
		return true
	}

	if entgqlAnt.Skip.Is(entgql.SkipMutationCreateInput) {
		return true
	}

	if entgqlAnt.MutationInputs == nil {
		return true
	}

	for _, mi := range entgqlAnt.MutationInputs {
		if mi.IsCreate {
			return false
		}
	}

	return true
}

// checkSkipMutationUpdateInput returns true if no UpdateInput type is generated for this schema
func checkSkipMutationUpdateInput(node *gen.Type) bool {
	entgqlAnt := &entgql.Annotation{}

	ant, ok := node.Annotations[entgqlAnt.Name()]
	if !ok {
		return true
	}

	if err := entgqlAnt.Decode(ant); err != nil {
		return true
	}

	if entgqlAnt.Skip.Is(entgql.SkipMutationUpdateInput) {
		return true
	}

	if entgqlAnt.MutationInputs == nil {
		return true
	}

	for _, mi := range entgqlAnt.MutationInputs {
		if !mi.IsCreate {
			return false
		}
	}

	return true
}

// writeMappingFile creates the integration mapping Go file
func writeMappingFile(outputDir string, data MappingData) error {
	tmpl := buildMappingTemplate()

	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		log.Error().Err(err).Str("path", outputDir).Msg("failed to create integration mapping output directory")

		return err
	}

	filePath := filepath.Join(outputDir, "integration_mapping_generated.go")

	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("failed to create integration mapping file")

		return err
	}

	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Error().Err(err).Msg("failed to execute integration mapping template")

		return err
	}

	log.Debug().Str("path", filePath).Msg("generated integration mapping file")

	return nil
}

// buildMappingTemplate parses and returns the integration mapping template
func buildMappingTemplate() *template.Template {
	tmpl, err := template.New("integration_mapping.tpl").ParseFS(_templates, "templates/integration_mapping.tpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse integration mapping template")
	}

	return tmpl
}
