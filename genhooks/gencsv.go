package genhooks

import (
	"cmp"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"

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
	outputDir   string
	packageName string
	entPackage  string
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
// All lookups are automatically scoped to the organization context from the request.
type CSVLookup struct {
	// TargetEntity is the entity type to query (e.g., User, Group)
	TargetEntity string
	// MatchField is the field to match on (e.g., email, name)
	MatchField string
	// CreateIfMissing indicates if this lookup supports auto-creation
	CreateIfMissing bool
}

// CSVSchema represents a schema with CSV reference fields
type CSVSchema struct {
	// Name is the schema name (e.g., ActionPlan)
	Name string
	// Fields contains CSV reference field mappings
	Fields []CSVReferenceField
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

	for _, node := range g.Nodes {
		if checkSchemaGenSkip(node) || checkQueryGenSkip(node) {
			continue
		}

		schema := getEntSchema(g, node.Name)
		if schema == nil {
			continue
		}

		edgeTargets := buildEdgeTargetMap(node)
		fields := getCSVReferenceFieldsWithEdges(schema, edgeTargets)

		if len(fields) == 0 {
			continue
		}

		for _, f := range fields {
			key := f.TargetEntity + ":" + f.MatchField
			if _, exists := lookupSet[key]; !exists {
				lookupSet[key] = CSVLookup{
					TargetEntity:    f.TargetEntity,
					MatchField:      f.MatchField,
					CreateIfMissing: f.CreateIfMissing,
				}
			}
		}

		slices.SortFunc(fields, func(a, b CSVReferenceField) int {
			return cmp.Compare(a.CSVColumn, b.CSVColumn)
		})

		data.Schemas = append(data.Schemas, CSVSchema{
			Name:   node.Name,
			Fields: fields,
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
