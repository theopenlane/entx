package integrationmapping

import (
	"cmp"
	"embed"
	"fmt"
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
	// IngestTopic is the schema-scoped ingest topic name
	IngestTopic string
	// IngestRequestTypeName is the generated Go type name for the typed ingest request payload
	IngestRequestTypeName string
	// IngestTopicVarName is the generated Go variable name for the typed ingest topic
	IngestTopicVarName string
	// InputTypeName is the ent-generated Create input type name
	InputTypeName string
	// UpdateInputTypeName is the ent-generated Update input type name
	UpdateInputTypeName string
	// EntSchemaPackageAlias is the Go identifier for the ent schema predicate package
	EntSchemaPackageAlias string
	// EntSchemaPackagePath is the import path of the ent schema predicate package
	EntSchemaPackagePath string
	// LookupFields are the ent/go field pairs used for stock ingest lookup matching
	LookupFields []IngestLookupField
	// RuntimeDefaults are the runtime-injected values applied during stock ingest preparation
	RuntimeDefaults []IngestRuntimeDefault
	// StockPersist indicates the schema can use the generated stock ingest persistence path
	StockPersist bool
	// ScopeByIntegrationID indicates the persist query should scope by integration_id rather than owner_id
	ScopeByIntegrationID bool
	// HasDirectorySyncRunID indicates the schema has a directory_sync_run_id field
	HasDirectorySyncRunID bool
	// DirectorySyncRunIDRequired indicates directory_sync_run_id is a required (non-pointer) field
	DirectorySyncRunIDRequired bool
	// NeedsManualOwnerID indicates owner_id must be injected from the integration at emit/persist time
	NeedsManualOwnerID bool
	// HasPrepareInput indicates the integrationgenerated package exports a Prepare*Input function
	HasPrepareInput bool
	// Fields contains integration mapping field metadata
	Fields []MappingField
}

// IngestRuntimeDefault represents one integration-injected field used by stock ingest persistence
type IngestRuntimeDefault struct {
	// Field is the ent field name for this runtime default
	Field string
	// GoField is the Go struct field name for this runtime default on ent input types
	GoField string
	// Required indicates the field is required (non-pointer) on the ent input type, which determines whether the zero value or nil is checked at ingest time before injection
	Required bool
	// IntegrationField is the Go struct field name on *ent.Integration that serves as the source value for this runtime default
	IntegrationField string
}

// IngestLookupField represents one stock ingest lookup field
type IngestLookupField struct {
	// Field is the ent field name for this lookup key
	Field string
	// GoField is the Go struct field name for this lookup key on ent input types
	GoField string
	// Required indicates the field is required (non-pointer) on the ent input type, which determines whether the zero value or nil is checked at ingest time before including the field in lookup matching
	Required bool
}

// IngestData holds the data for generating ingest operation files
type IngestData struct {
	*Config
	// Schemas contains all schemas participating in ingest
	Schemas []MappingSchema
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
	// FromIntegration indicates the field value is injected from the integration record at ingest time
	FromIntegration bool
}

// collectMappingData collects mapping data from all schemas in the graph
func collectMappingData(g *gen.Graph, c *Config) (MappingData, error) {
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
			var err error

			runtimeDefaults, err = collectRuntimeDefaults(fields)
			if err != nil {
				return MappingData{}, fmt.Errorf("schema %s: %w", node.Name, err)
			}
		}

		scopeByIntID := integrationIDIsUpsertKey(fields)
		hasDirSyncRun, dirSyncRunRequired := schemaDirectorySyncRunInfo(schema)
		needsManualOwnerID := schemaHasFieldName(schema, "owner_id") && !hasRuntimeDefaultForEntField(runtimeDefaults, "owner_id")
		hasPrepareInput := stockPersist && len(runtimeDefaults) > 0

		entAlias := strings.ToLower(node.Name)
		entPkg := ""
		if c.EntPackage != "" {
			entPkg = c.EntPackage + "/" + entAlias
		}

		updateInputTypeName := ""
		if hasUpdate {
			updateInputTypeName = "Update" + node.Name + "Input"
		}

		data.Schemas = append(data.Schemas, MappingSchema{
			Name:                       node.Name,
			ConstName:                  schemaConstName(node.Name),
			IngestTopic:                schemaIngestTopicName(node.Name),
			IngestRequestTypeName:      schemaIngestRequestTypeName(node.Name),
			IngestTopicVarName:         schemaIngestTopicVarName(node.Name),
			InputTypeName:              inputTypeName,
			UpdateInputTypeName:        updateInputTypeName,
			EntSchemaPackageAlias:      entAlias,
			EntSchemaPackagePath:       entPkg,
			LookupFields:               lookupFields,
			RuntimeDefaults:            runtimeDefaults,
			StockPersist:               stockPersist,
			ScopeByIntegrationID:       scopeByIntID,
			HasDirectorySyncRunID:      hasDirSyncRun,
			DirectorySyncRunIDRequired: dirSyncRunRequired,
			NeedsManualOwnerID:         needsManualOwnerID,
			HasPrepareInput:            hasPrepareInput,
			Fields:                     fields,
		})
	}

	slices.SortFunc(data.Schemas, func(a, b MappingSchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return data, nil
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

// fieldIsIncluded reports whether a field should be collected given the schema's include/exclude/system-field rules.
// The include list takes full precedence: when present, only listed fields are included and the exclude list
// and system-field defaults do not apply. When no include list is set, excluded fields and system-managed
// fields are skipped unless the schema uses stock persistence and the field carries an explicit annotation.
func fieldIsIncluded(fieldName string, includeSet, excludeSet map[string]struct{}, hasInclude, stockPersist bool, ant *entx.IntegrationMappingFieldAnnotation) bool {
	if hasInclude {
		_, ok := includeSet[fieldName]
		return ok
	}

	if _, ok := excludeSet[fieldName]; ok {
		return false
	}

	if isSystemField(fieldName) {
		return stockPersist && ant != nil
	}

	return true
}

// collectMappingFields returns integration mapping fields for a schema.
func collectMappingFields(schema *load.Schema, schemaName string) []MappingField {
	if schema == nil {
		return nil
	}

	fields := make([]MappingField, 0)
	schemaAnt := getSchemaAnnotation(schema)
	stockPersist := schemaAnt != nil && schemaAnt.StockPersist

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
		if !isFieldMappingEligible(field) {
			continue
		}

		ant := getFieldAnnotation(field)

		if !fieldIsIncluded(field.Name, includeSet, excludeSet, hasInclude, stockPersist, ant) {
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
			goName := templates.ToGo(field.Name)
			key = templates.ToGoPrivate(goName)
		}

		upsert := ant != nil && ant.UpsertKey
		lookup := ant != nil && ant.LookupKey
		fromIntegration := ant != nil && ant.FromIntegration

		fields = append(fields, MappingField{
			InputKey:        key,
			GoField:         templates.ToGo(key),
			ConstName:       fieldConstName(schemaName, key),
			EntField:        field.Name,
			Type:            field.Info.Type.String(),
			Required:        !field.Optional,
			UpsertKey:       upsert,
			LookupKey:       lookup,
			FromIntegration: fromIntegration,
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
				Field:    field.EntField,
				GoField:  field.GoField,
				Required: field.Required,
			})
		}
	}

	return keys
}

// integrationIDIsUpsertKey reports whether the integration_id field is both an upsert key and sourced from the integration
func integrationIDIsUpsertKey(fields []MappingField) bool {
	for _, f := range fields {
		if f.EntField == "integration_id" && f.UpsertKey && f.FromIntegration {
			return true
		}
	}

	return false
}

// hasRuntimeDefaultForEntField reports whether runtimeDefaults contains an entry for the given ent field name
func hasRuntimeDefaultForEntField(defaults []IngestRuntimeDefault, entField string) bool {
	for _, d := range defaults {
		if d.Field == entField {
			return true
		}
	}

	return false
}

// schemaHasFieldName reports whether the schema declares a field with the given name (including mixin fields)
func schemaHasFieldName(schema *load.Schema, name string) bool {
	for _, f := range schema.Fields {
		if f.Name == name {
			return true
		}
	}

	return false
}

// schemaDirectorySyncRunInfo returns whether the schema has a directory_sync_run_id field and whether it is required
func schemaDirectorySyncRunInfo(schema *load.Schema) (has bool, required bool) {
	for _, f := range schema.Fields {
		if f.Name == "directory_sync_run_id" {
			return true, !f.Optional
		}
	}

	return false, false
}

// buildIngestData constructs IngestData for the ingest templates from config and collected schemas
func buildIngestData(c *Config, schemas []MappingSchema) IngestData {
	return IngestData{
		Config:  c,
		Schemas: schemas,
	}
}

// collectRuntimeDefaults returns fields marked FromIntegration for stock ingest preparation
func collectRuntimeDefaults(fields []MappingField) ([]IngestRuntimeDefault, error) {
	defaults := make([]IngestRuntimeDefault, 0)

	for _, field := range fields {
		if !field.FromIntegration {
			continue
		}

		instField, err := integrationFieldForEntField(field.EntField)
		if err != nil {
			return nil, err
		}

		defaults = append(defaults, IngestRuntimeDefault{
			Field:            field.EntField,
			GoField:          field.GoField,
			Required:         field.Required,
			IntegrationField: instField,
		})
	}

	return defaults, nil
}

// integrationFieldForEntField maps an ent field name to the Go field name on *ent.Integration
func integrationFieldForEntField(entField string) (string, error) {
	switch entField {
	case "integration_id":
		return "ID", nil
	case "owner_id":
		return "OwnerID", nil
	case "platform_id":
		return "PlatformID", nil
	default:
		return "", fmt.Errorf("field %q annotated FromIntegration has no known Integration field mapping", entField)
	}
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

	log.Info().Str("path", filePath).Msg("generated integration mapping file")

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

// writeIngestListenersFile creates the always-overwritten ingest_generated.go in the ingest output directory
func writeIngestListenersFile(outputDir string, data IngestData) error {
	tmpl, err := template.New("ingest_listeners.tpl").ParseFS(_templates, "templates/ingest_listeners.tpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse ingest listeners template")
	}

	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		log.Error().Err(err).Str("path", outputDir).Msg("failed to create ingest output directory")

		return err
	}

	filePath := filepath.Join(outputDir, "ingest_generated.go")

	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("failed to create ingest generated file")

		return err
	}

	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Error().Err(err).Msg("failed to execute ingest listeners template")

		return err
	}

	log.Info().Str("path", filePath).Msg("generated ingest listeners file")

	return nil
}

// writeIngestPersistFiles creates one per-schema persist file for each schema, skipping existing files
func writeIngestPersistFiles(outputDir string, data IngestData) error {
	tmpl, err := template.New("ingest_persist_schema.tpl").ParseFS(_templates, "templates/ingest_persist_schema.tpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse ingest persist schema template")
	}

	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		log.Error().Err(err).Str("path", outputDir).Msg("failed to create ingest output directory")

		return err
	}

	for _, schema := range data.Schemas {
		fileName := "ingest_" + strings.ToLower(schema.Name) + "_persist.go"
		filePath := filepath.Join(outputDir, fileName)

		if _, statErr := os.Stat(filePath); statErr == nil {
			log.Info().Str("path", filePath).Msg("skipping existing ingest persist file")

			continue
		}

		persistData := struct {
			IngestData
			Schema MappingSchema
		}{
			IngestData: data,
			Schema:     schema,
		}

		file, err := os.Create(filepath.Clean(filePath))
		if err != nil {
			log.Error().Err(err).Str("path", filePath).Msg("failed to create ingest persist file")

			return err
		}

		if err := tmpl.Execute(file, persistData); err != nil {
			file.Close()

			log.Error().Err(err).Str("schema", schema.Name).Msg("failed to execute ingest persist schema template")

			return err
		}

		file.Close()

		log.Info().Str("path", filePath).Msg("generated ingest persist file")
	}

	return nil
}
