package entityops

import (
	"bytes"
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
	"github.com/stoewer/go-strcase"
	"golang.org/x/tools/imports"

	"github.com/theopenlane/entx"
)

//go:embed templates/*
var _templates embed.FS

const dirPermissions = 0o755

// EntityData holds all collected schema data for generation
type EntityData struct {
	// PackageName is the Go package name for generated files
	PackageName string
	// EntPackage is the ent generated package import path
	EntPackage string
	// GalaPackage is the gala package import path
	GalaPackage string
	// JsonxPackage is the jsonx package import path
	JsonxPackage string
	// LogxPackage is the logx package import path
	LogxPackage string
	// ContextxPackage is the contextx package import path
	ContextxPackage string
	// Schemas contains all schemas eligible for entity operations
	Schemas []EntitySchema
}

// EntitySchema represents one schema's metadata for entity operations generation
type EntitySchema struct {
	// Name is the PascalCase schema name (e.g., "ActionPlan")
	Name string
	// Snake is the snake_case form (e.g., "action_plan")
	Snake string
	// Camel is the camelCase form (e.g., "actionPlan")
	Camel string
	// Lower is the lowercase no-separator form (e.g., "actionplan")
	Lower string
	// Plural is the PascalCase plural form (e.g., "ActionPlans")
	Plural string
	// Table is the database table name (e.g., "action_plans")
	Table string
	// Label is the human-readable label (e.g., "Action Plan")
	Label string
	// HasCreate indicates a CreateInput type is generated
	HasCreate bool
	// HasUpdate indicates an UpdateInput type is generated
	HasUpdate bool
	// CreateInputType is the ent-generated CreateInput type name
	CreateInputType string
	// UpdateInputType is the ent-generated UpdateInput type name
	UpdateInputType string
	// PredicatePackage is the ent predicate package alias (lowercase schema name)
	PredicatePackage string
	// PredicateImport is the ent predicate package import path
	PredicateImport string
	// HasOwnerID indicates the schema has an owner_id field for org-scoped queries
	HasOwnerID bool
	// Fields contains mutable field name mappings (snake_case → PascalCase)
	Fields []EntityField
	// Edges contains linkable M2M and O2M edges for this schema
	Edges []EntityEdge
}

// EntityField represents a mutable field with its name variations
type EntityField struct {
	// Name is the PascalCase Go field name (e.g., "ReferenceID")
	Name string
	// Snake is the snake_case column name (e.g., "reference_id")
	Snake string
}

// EntityEdge represents one linkable edge on a schema
type EntityEdge struct {
	// Name is the edge name (e.g., "controls")
	Name string
	// TargetSchema is the target PascalCase name (e.g., "Control")
	TargetSchema string
	// Relationship is "M2M" or "O2M"
	Relationship string
}

// collectEntityData iterates the ent graph and collects schemas annotated with
// either OPENLANE_WORKFLOW_ELIGIBLE fields or OPENLANE_INTEGRATION_MAPPING_SCHEMA
func collectEntityData(g *gen.Graph, c *Config) (EntityData, error) {
	data := EntityData{
		PackageName:     c.PackageName,
		EntPackage:      c.EntPackage,
		GalaPackage:     c.GalaPackage,
		JsonxPackage:    c.JsonxPackage,
		LogxPackage:     c.LogxPackage,
		ContextxPackage: c.ContextxPackage,
		Schemas:         []EntitySchema{},
	}

	eligibleSchemas := make(map[string]struct{})

	for _, node := range g.Nodes {
		if skipNode(node) {
			continue
		}

		schema := findSchema(g, node.Name)
		if schema == nil {
			continue
		}

		source := classifySource(node, schema)
		if source.Workflow || source.Integration {
			eligibleSchemas[node.Name] = struct{}{}
		}
	}

	for _, node := range g.Nodes {
		if _, ok := eligibleSchemas[node.Name]; !ok {
			continue
		}

		hasCreate := !skipMutationCreateInput(node)
		hasUpdate := !skipMutationUpdateInput(node)

		schema := findSchema(g, node.Name)

		predAlias := strings.ToLower(node.Name)
		predImport := ""

		if c.EntPackage != "" {
			predImport = c.EntPackage + "/" + predAlias
		}

		entitySchema := EntitySchema{
			Name:             node.Name,
			Snake:            strcase.SnakeCase(node.Name),
			Camel:            lowerFirst(node.Name),
			Lower:            strings.ToLower(strings.ReplaceAll(strcase.SnakeCase(node.Name), "_", "")),
			Plural:           strcase.UpperCamelCase(node.Table()),
			Table:            node.Table(),
			Label:            humanize(node.Name),
			HasCreate:        hasCreate,
			HasUpdate:        hasUpdate,
			PredicatePackage: predAlias,
			PredicateImport:  predImport,
			HasOwnerID:       hasField(schema, "owner_id"),
		}

		for _, field := range node.Fields {
			if field.Immutable {
				continue
			}

			entitySchema.Fields = append(entitySchema.Fields, EntityField{
				Name:  field.Name,
				Snake: field.StorageKey(),
			})
		}

		slices.SortFunc(entitySchema.Fields, func(a, b EntityField) int {
			return cmp.Compare(a.Snake, b.Snake)
		})

		for _, edge := range node.Edges {
			if edge.IsInverse() {
				continue
			}

			if _, ok := eligibleSchemas[edge.Type.Name]; !ok {
				continue
			}

			switch edge.Rel.Type {
			case gen.M2M, gen.O2M:
				entitySchema.Edges = append(entitySchema.Edges, EntityEdge{
					Name:         edge.Name,
					TargetSchema: edge.Type.Name,
					Relationship: edge.Rel.Type.String(),
				})
			}
		}

		slices.SortFunc(entitySchema.Edges, func(a, b EntityEdge) int {
			return cmp.Compare(a.Name, b.Name)
		})

		if hasCreate {
			entitySchema.CreateInputType = "Create" + node.Name + "Input"
		}

		if hasUpdate {
			entitySchema.UpdateInputType = "Update" + node.Name + "Input"
		}

		data.Schemas = append(data.Schemas, entitySchema)
	}

	slices.SortFunc(data.Schemas, func(a, b EntitySchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return data, nil
}

// schemaSource tracks which annotation(s) caused a schema to be included
type schemaSource struct {
	Workflow    bool
	Integration bool
}

// classifySource determines why a schema should be included
func classifySource(node *gen.Type, schema *load.Schema) schemaSource {
	var source schemaSource

	for _, field := range node.Fields {
		if _, ok := field.Annotations[entx.WorkflowEligibleAnnotationName]; ok {
			source.Workflow = true
			break
		}
	}

	for _, edge := range node.Edges {
		if edge.IsInverse() {
			continue
		}

		if _, ok := edge.Annotations[entx.WorkflowEligibleAnnotationName]; ok {
			source.Workflow = true
			break
		}
	}

	if hasIntegrationMappingAnnotation(schema) {
		source.Integration = true
	}

	return source
}

// hasIntegrationMappingAnnotation checks if the schema has an OPENLANE_INTEGRATION_MAPPING_SCHEMA annotation
func hasIntegrationMappingAnnotation(schema *load.Schema) bool {
	if schema == nil {
		return false
	}

	for _, ant := range schema.Annotations {
		raw, ok := ant.(map[string]any)
		if !ok {
			continue
		}

		if _, found := raw[entx.IntegrationMappingSchemaAnnotationName]; found {
			return true
		}
	}

	for _, field := range schema.Fields {
		if field.Annotations == nil {
			continue
		}

		raw, ok := field.Annotations[entx.IntegrationMappingFieldAnnotationName]
		if !ok {
			continue
		}

		var ann entx.IntegrationMappingFieldAnnotation
		if err := ann.Decode(raw); err == nil && (ann.UpsertKey || ann.LookupKey) {
			return true
		}
	}

	return false
}

// generateEntityFiles renders all templates and writes them to the output directory
func generateEntityFiles(outputDir string, data EntityData) error {
	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		return fmt.Errorf("create output dir %s: %w", outputDir, err)
	}

	type templateSpec struct {
		name     string
		filename string
		tmplFile string
	}

	specs := []templateSpec{
		{name: "entity_schema", filename: "entity_schema.go", tmplFile: "templates/entity_schema.tpl"},
		{name: "entity_errors", filename: "entity_errors.go", tmplFile: "templates/entity_errors.tpl"},
		{name: "entity_registry", filename: "entity_registry.go", tmplFile: "templates/entity_registry.tpl"},
		{name: "entity_upsert", filename: "entity_upsert.go", tmplFile: "templates/entity_upsert.tpl"},
		{name: "entity_metadata", filename: "entity_metadata.go", tmplFile: "templates/entity_metadata.tpl"},
		{name: "entity_handlers", filename: "entity_handlers.go", tmplFile: "templates/entity_handlers.tpl"},
	}

	for _, spec := range specs {
		raw, err := _templates.ReadFile(spec.tmplFile)
		if err != nil {
			return fmt.Errorf("read template %s: %w", spec.tmplFile, err)
		}

		tmpl, err := template.New(spec.name).Funcs(gen.Funcs).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", spec.name, err)
		}

		if err := writeFile(outputDir, spec.filename, tmpl, data); err != nil {
			return err
		}
	}

	return nil
}

// writeFile renders a template and writes the formatted output
func writeFile(outputDir, filename string, tmpl *template.Template, data any) error {
	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template for %s: %w", filename, err)
	}

	outputPath := filepath.Join(outputDir, filename)

	formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format %s: %w", outputPath, err)
	}

	return os.WriteFile(outputPath, formatted, 0o600) //nolint:mnd
}

// skipNode returns true if the node should be excluded from entity operations generation
func skipNode(node *gen.Type) bool {
	if strings.HasSuffix(node.Name, "History") {
		return true
	}

	schemaGenAnt := &entx.SchemaGenAnnotation{}
	if ant, ok := node.Annotations[schemaGenAnt.Name()]; ok {
		if err := schemaGenAnt.Decode(ant); err == nil && schemaGenAnt.Skip {
			return true
		}
	}

	queryGenAnt := &entx.QueryGenAnnotation{}
	if ant, ok := node.Annotations[queryGenAnt.Name()]; ok {
		if err := queryGenAnt.Decode(ant); err == nil && queryGenAnt.Skip {
			return true
		}
	}

	return false
}

// skipMutationCreateInput returns true if no CreateInput type is generated for this schema
func skipMutationCreateInput(node *gen.Type) bool {
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

// skipMutationUpdateInput returns true if no UpdateInput type is generated for this schema
func skipMutationUpdateInput(node *gen.Type) bool {
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

// findSchema returns the schema for a given name from the graph
func findSchema(g *gen.Graph, name string) *load.Schema {
	for _, s := range g.Schemas {
		if s.Name == name {
			return s
		}
	}

	return nil
}

// hasField checks if a schema declares a field with the given name
func hasField(schema *load.Schema, name string) bool {
	for _, f := range schema.Fields {
		if f.Name == name {
			return true
		}
	}

	return false
}

// lowerFirst returns the string with its first character lowered
func lowerFirst(s string) string {
	if s == "" {
		return s
	}

	return strings.ToLower(s[:1]) + s[1:]
}

// humanize converts PascalCase to a human-readable label with spaces
func humanize(s string) string {
	snake := strcase.SnakeCase(s)
	parts := strings.Split(snake, "_")

	for i, p := range parts {
		if i == 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}

	return strings.Join(parts, " ")
}
