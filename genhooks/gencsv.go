package genhooks

import (
	"cmp"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/entx"
)

// dirPermissions is the permission mode for created directories
const dirPermissions = 0755

// CSVConfig holds configuration options for CSV reference generation
type CSVConfig struct {
	outputDir           string
	packageName         string
	entPackage          string
	generateAllWrappers bool
}

// CSVOption adds functional params for CSVConfig
type CSVOption func(*CSVConfig)

// WithCSVOutputDir sets the directory to output the generated CSV helper files
func WithCSVOutputDir(dir string) CSVOption {
	return func(c *CSVConfig) {
		c.outputDir = dir
	}
}

// WithCSVPackageName sets the Go package name for generated CSV files
func WithCSVPackageName(name string) CSVOption {
	return func(c *CSVConfig) {
		c.packageName = name
	}
}

// WithCSVEntPackage sets the import path for the ent generated package
func WithCSVEntPackage(pkg string) CSVOption {
	return func(c *CSVConfig) {
		c.entPackage = pkg
	}
}

// WithCSVGenerateAllWrappers enables wrapper type generation for all schemas,
// not just those with CSV reference annotations. This allows all CSV bulk
// resolvers to benefit from list preprocessing and header prefixing.
func WithCSVGenerateAllWrappers(enabled bool) CSVOption {
	return func(c *CSVConfig) {
		c.generateAllWrappers = enabled
	}
}

// CSVSchemaData holds the data for generating CSV helper files
type CSVSchemaData struct {
	// PackageName is the Go package name for generated files
	PackageName string
	// EntPackage is the import path for the ent generated package
	EntPackage string
	// Schemas contains all schemas with CSV reference fields
	Schemas []CSVSchema
	// Lookups contains unique target entity + match field combinations
	Lookups []CSVLookup
}

// CSVLookup represents a unique lookup type for resolving CSV values to IDs.
type CSVLookup struct {
	// TargetEntity is the entity type to query (e.g., User, Group)
	TargetEntity string
	// MatchField is the field to match on (e.g., email, name)
	MatchField string
	// CreateIfMissing indicates if this lookup supports auto-creation
	CreateIfMissing bool
	// OrgScoped indicates if the entity has an OwnerID field for org filtering
	OrgScoped bool
}

// CSVSchema represents a schema with CSV reference fields
type CSVSchema struct {
	// Name is the schema name (e.g., ActionPlan)
	Name string
	// Fields contains CSV reference field mappings
	Fields []CSVReferenceField
	// HasCreateInput indicates if the schema has a CreateInput type
	HasCreateInput bool
	// HasUpdateInput indicates if the schema has an UpdateInput type
	HasUpdateInput bool
}

// CSVReferenceField represents a field that can be resolved from CSV references.
// All lookups are automatically scoped to the organization context from the request.
type CSVReferenceField struct {
	// FieldName is the ent field name (e.g., assigned_to_user_id)
	FieldName string
	// GoFieldName is the Go field name (e.g., AssignedToUserID)
	GoFieldName string
	// CSVColumn is the friendly CSV header (e.g., AssignedToUserEmail)
	CSVColumn string
	// TargetEntity is the entity type to query (e.g., User, Group)
	TargetEntity string
	// MatchField is the field on target entity to match (e.g., email, name)
	MatchField string
	// IsSlice indicates if the field is a []string
	IsSlice bool
	// CreateIfMissing indicates if missing records should be created
	CreateIfMissing bool
}

// CSVFieldMappingJSON is a simplified field mapping structure for JSON export.
// This is consumed by bulkgen to generate sample CSV files with custom column headers.
type CSVFieldMappingJSON struct {
	// CSVColumn is the friendly CSV header name (e.g., AssignedToUserEmail)
	CSVColumn string `json:"csvColumn"`
	// TargetField is the Go field name this column maps to (e.g., AssignedToUserID)
	TargetField string `json:"targetField"`
	// IsSlice indicates if the field is a []string
	IsSlice bool `json:"isSlice"`
}

// CSVFieldMappingsJSON is the top-level structure for the JSON export file.
// Maps schema names to their CSV field mappings.
type CSVFieldMappingsJSON map[string][]CSVFieldMappingJSON

// GenCSVSchema generates CSV helper types and functions based on schema annotations
func GenCSVSchema(opts ...CSVOption) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			c := &CSVConfig{
				packageName: "csvgenerated",
			}

			for _, opt := range opts {
				opt(c)
			}

			if c.outputDir == "" {
				return next.Generate(g)
			}

			data := getCSVInputData(g, c)
			if len(data.Schemas) == 0 {
				return next.Generate(g)
			}

			if err := generateCSVHelperFile(c.outputDir, data); err != nil {
				return err
			}

			return next.Generate(g)
		})
	}
}

// getCSVInputData collects CSV reference data from all schemas
func getCSVInputData(g *gen.Graph, c *CSVConfig) CSVSchemaData {
	data := CSVSchemaData{
		PackageName: c.packageName,
		EntPackage:  c.entPackage,
		Schemas:     []CSVSchema{},
	}

	lookupSet := make(map[string]CSVLookup)
	orgScopedEntities := buildOrgScopedEntityMap(g)

	for _, node := range g.Nodes {
		if checkSchemaGenSkip(node) || checkQueryGenSkip(node) {
			continue
		}

		// Check which mutation input types are available for this schema
		hasCreate := !checkSkipMutationCreateInput(node)
		hasUpdate := !checkSkipMutationUpdateInput(node)

		// Skip schemas that don't have any mutation input types
		if !hasCreate && !hasUpdate {
			continue
		}

		schema := getEntSchema(g, node.Name)
		if schema == nil {
			continue
		}

		edgeTargets := buildEdgeTargetMap(node)
		fields := getCSVReferenceFieldsWithEdges(schema, edgeTargets)

		// When generateAllWrappers is enabled, include schemas even without CSV reference fields.
		// This allows all CSV bulk resolvers to benefit from list preprocessing and header prefixing.
		if len(fields) == 0 && !c.generateAllWrappers {
			continue
		}

		for _, f := range fields {
			key := f.TargetEntity + ":" + f.MatchField
			if _, exists := lookupSet[key]; !exists {
				lookupSet[key] = CSVLookup{
					TargetEntity:    f.TargetEntity,
					MatchField:      f.MatchField,
					CreateIfMissing: f.CreateIfMissing,
					OrgScoped:       orgScopedEntities[f.TargetEntity],
				}
			}
		}

		slices.SortFunc(fields, func(a, b CSVReferenceField) int {
			return cmp.Compare(a.CSVColumn, b.CSVColumn)
		})

		data.Schemas = append(data.Schemas, CSVSchema{
			Name:           node.Name,
			Fields:         fields,
			HasCreateInput: hasCreate,
			HasUpdateInput: hasUpdate,
		})
	}

	slices.SortFunc(data.Schemas, func(a, b CSVSchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	for _, lookup := range lookupSet {
		data.Lookups = append(data.Lookups, lookup)
	}

	slices.SortFunc(data.Lookups, func(a, b CSVLookup) int {
		if a.TargetEntity != b.TargetEntity {
			return cmp.Compare(a.TargetEntity, b.TargetEntity)
		}

		return cmp.Compare(a.MatchField, b.MatchField)
	})

	return data
}

// checkSkipMutationCreateInput checks if the schema should be skipped for CSV create wrapper generation.
// Returns true if no CreateInput type is generated for this schema, which can happen when:
// 1. The schema has entgql.Skip(entgql.SkipMutationCreateInput) annotation
// 2. The schema has no MutationInputs configured (no entgql.Mutations() annotation)
// 3. The schema has MutationInputs but none are configured for create operations
func checkSkipMutationCreateInput(node *gen.Type) bool {
	entgqlAnt := &entgql.Annotation{}

	ant, ok := node.Annotations[entgqlAnt.Name()]
	if !ok {
		// No entgql annotation means no mutations configured
		return true
	}

	if err := entgqlAnt.Decode(ant); err != nil {
		return true
	}

	// Check if explicitly skipped
	if entgqlAnt.Skip.Is(entgql.SkipMutationCreateInput) {
		return true
	}

	// Check if mutations are configured - if nil, no mutations are generated
	if entgqlAnt.MutationInputs == nil {
		return true
	}

	// Check if any mutation input is configured for create operations
	for _, mi := range entgqlAnt.MutationInputs {
		if mi.IsCreate {
			return false
		}
	}

	// No create mutation found
	return true
}

// checkSkipMutationUpdateInput checks if the schema should be skipped for CSV update wrapper generation.
// Returns true if no UpdateInput type is generated for this schema, which can happen when:
// 1. The schema has entgql.Skip(entgql.SkipMutationUpdateInput) annotation
// 2. The schema has no MutationInputs configured (no entgql.Mutations() annotation)
// 3. The schema has MutationInputs but none are configured for update operations
func checkSkipMutationUpdateInput(node *gen.Type) bool {
	entgqlAnt := &entgql.Annotation{}

	ant, ok := node.Annotations[entgqlAnt.Name()]
	if !ok {
		// No entgql annotation means no mutations configured
		return true
	}

	if err := entgqlAnt.Decode(ant); err != nil {
		return true
	}

	// Check if explicitly skipped
	if entgqlAnt.Skip.Is(entgql.SkipMutationUpdateInput) {
		return true
	}

	// Check if mutations are configured - if nil, no mutations are generated
	if entgqlAnt.MutationInputs == nil {
		return true
	}

	// Check if any mutation input is configured for update operations
	for _, mi := range entgqlAnt.MutationInputs {
		if !mi.IsCreate {
			return false
		}
	}

	// No update mutation found
	return true
}

// buildOrgScopedEntityMap returns a map of entity names to whether they have an owner edge
func buildOrgScopedEntityMap(g *gen.Graph) map[string]bool {
	result := make(map[string]bool)

	for _, node := range g.Nodes {
		for _, edge := range node.Edges {
			if edge.Name == "owner" {
				result[node.Name] = true

				break
			}
		}
	}

	return result
}

// buildEdgeTargetMap creates a map from field names to their edge target entity types
func buildEdgeTargetMap(node *gen.Type) map[string]string {
	targets := make(map[string]string)

	for _, edge := range node.Edges {
		if f := edge.Field(); f != nil {
			targets[f.Name] = edge.Type.Name
		}
	}

	return targets
}

// getCSVReferenceFieldsWithEdges extracts CSV reference fields from a schema,
// inferring target entities from edge definitions when not explicitly specified
func getCSVReferenceFieldsWithEdges(schema *load.Schema, edgeTargets map[string]string) []CSVReferenceField {
	var fields []CSVReferenceField

	for _, field := range schema.Fields {
		ann := getCSVReferenceAnnotation(field)
		if ann == nil {
			continue
		}

		if ann.MatchField == "" || ann.CSVColumn == "" {
			continue
		}

		targetEntity := ann.TargetEntity
		if targetEntity == "" {
			targetEntity = edgeTargets[field.Name]
		}

		if targetEntity == "" {
			log.Warn().Str("field", field.Name).Str("schema", schema.Name).Msg("CSV reference field has no target entity and no edge found")
			continue
		}

		goFieldName := templates.ToGo(field.Name)
		isSlice := field.Info.String() == "[]string"

		fields = append(fields, CSVReferenceField{
			FieldName:       field.Name,
			GoFieldName:     goFieldName,
			CSVColumn:       ann.CSVColumn,
			TargetEntity:    targetEntity,
			MatchField:      ann.MatchField,
			IsSlice:         isSlice,
			CreateIfMissing: ann.CreateIfMissing,
		})
	}

	return fields
}

// getCSVReferenceAnnotation retrieves the CSV reference annotation from a field
func getCSVReferenceAnnotation(field *load.Field) *entx.CSVReferenceAnnotation {
	ann := &entx.CSVReferenceAnnotation{}
	if a, ok := field.Annotations[ann.Name()]; ok {
		if err := ann.Decode(a); err != nil {
			return nil
		}

		return ann
	}

	return nil
}

// generateCSVHelperFile creates the CSV helper Go file
func generateCSVHelperFile(outputDir string, data CSVSchemaData) error {
	tmpl := createCSVTemplate()

	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		log.Error().Err(err).Str("path", outputDir).Msg("failed to create CSV output directory")

		return err
	}

	filePath := filepath.Join(outputDir, "csv_generated.go")

	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("failed to create CSV helper file")

		return err
	}

	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Error().Err(err).Msg("failed to execute CSV template")

		return err
	}

	if err := generateCSVFieldMappingsJSON(outputDir, data); err != nil {
		return err
	}

	return nil
}

// generateCSVFieldMappingsJSON creates a JSON file with CSV field mappings for use by bulkgen.
// This allows bulkgen to include custom CSV columns in generated sample files.
func generateCSVFieldMappingsJSON(outputDir string, data CSVSchemaData) error {
	mappings := make(CSVFieldMappingsJSON)

	for _, schema := range data.Schemas {
		if len(schema.Fields) == 0 {
			continue
		}

		fields := make([]CSVFieldMappingJSON, 0, len(schema.Fields))
		for _, f := range schema.Fields {
			fields = append(fields, CSVFieldMappingJSON{
				CSVColumn:   f.CSVColumn,
				TargetField: f.GoFieldName,
				IsSlice:     f.IsSlice,
			})
		}

		mappings[schema.Name] = fields
	}

	filePath := filepath.Join(outputDir, "csv_field_mappings.json")

	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("failed to create CSV field mappings JSON file")

		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(mappings); err != nil {
		log.Error().Err(err).Msg("failed to encode CSV field mappings to JSON")

		return err
	}

	log.Debug().Str("path", filePath).Msg("generated CSV field mappings JSON file")

	return nil
}

// createCSVTemplate creates the template for CSV helper generation
func createCSVTemplate() *template.Template {
	fm := template.FuncMap{
		"toLower":      strings.ToLower,
		"toUpperCamel": templates.ToGo,
		"toLowerCamel": templates.ToGoPrivate,
	}

	tmpl, err := template.New("csv.tpl").Funcs(fm).ParseFS(_templates, "templates/csv/csv.tpl")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse CSV template")
	}

	return tmpl
}
