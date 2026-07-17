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
	entfield "entgo.io/ent/schema/field"
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
	// CelxPackage is the celx package import path for typed entity expression evaluation
	CelxPackage string
	// EnumsPackageName is the Go package name for the generated WorkflowObjectType enum file
	EnumsPackageName string
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
	// ObjectFields is the unified per-schema field catalog (every field with capability flags),
	// consumed by both the workflow builder and the integration cross-link config. It is the single
	// field list: update-input re-keying, key-match columns, link source context, and workflow-eligible
	// fields are all derived from it by filtering on the per-field flags
	ObjectFields []EntityField
	// Edges contains every edge to an entityops schema (any cardinality/direction, mutable or immutable)
	// plus workflow group-permission edges; the single edge list for linking, workflow, and runtime ops
	Edges []EntityEdge
	// WorkflowEligible indicates the schema participates in workflows via eligible fields or edges
	WorkflowEligible bool
	// IntegrationMapped indicates the schema participates in integration ingest mapping
	IntegrationMapped bool
	// StockPersist indicates the schema opts into the generated stock ingest persistence path
	StockPersist bool
	// RuntimeDefaults are integration-injected field defaults applied by the generated Prepare function
	RuntimeDefaults []EntityRuntimeDefault
	// IngestTopic is the gala ingest topic name (entityops.{snake}.ingest.requested) for integration schemas
	IngestTopic string
	// IngestRequestType is the generated Go type name for this schema's typed ingest event payload
	IngestRequestType string
	// IngestTopicVar is the generated Go variable name for this schema's typed ingest topic
	IngestTopicVar string
}

// EntityField represents a field with its name variations and capability flags
type EntityField struct {
	// Name is the PascalCase Go field name (e.g., "ReferenceID")
	Name string
	// Snake is the snake_case column name (e.g., "reference_id")
	Snake string
	// Type is the ent field type string (e.g. "string", "bool", "time.Time")
	Type string
	// WorkflowEligible reports whether the field may drive workflow conditions and triggers
	WorkflowEligible bool
	// MatchKey reports whether the field is a plain-string indexed column usable as a cross-link match key
	MatchKey bool
	// IntegrationMapped reports whether the field participates in integration ingest mapping
	IntegrationMapped bool
	// InputKey is the integration mapping create-input key (lowerCamel of the field name, or annotation override)
	InputKey string
	// InputGoField is the exported Go struct field name for the input key on ent create inputs
	InputGoField string
	// FromIntegration reports whether the field value is injected from the integration record at ingest time
	FromIntegration bool
	// LookupKey reports whether the field is the ingest upsert lookup column for its schema
	LookupKey bool
}

// EntityEdge represents one linkable edge on a schema, in either direction
type EntityEdge struct {
	// Name is the edge name (e.g., "controls")
	Name string
	// TargetSchema is the target PascalCase name (e.g., "Control")
	TargetSchema string
	// Unique reports whether this side references a single target, so linking sets one id
	// (Set<Edge>ID) rather than adding many (Add<Edge>IDs)
	Unique bool
	// Optional reports whether a unique edge may be cleared; gates ClearField emission since a
	// required unique edge has no Clear<Edge> on the update input
	Optional bool
	// Immutable reports whether the edge is set only at create time; gates the update-input
	// mutation key emission and the runtime link-existing guard
	Immutable bool
	// WorkflowEligible reports whether the edge may drive workflow conditions and triggers
	WorkflowEligible bool
	// Field is the foreign-key storage column on this schema's table for unique owning edges
	// (e.g. "control_id"); empty when the foreign key lives on the target table
	Field string
}

// workflowEligibleMarkerField is the name of the WorkflowApprovalMixin carrier field that flags a
// schema as workflow-eligible without being a real workflow-triggerable field
const workflowEligibleMarkerField = "workflow_eligible_marker"

// fieldWorkflowEligible reports whether a field carries a non-marker workflow-eligible annotation.
// marker is true when the field is the WorkflowApprovalMixin carrier field, which flags the schema
// as workflow-eligible without itself being a targetable field
func fieldWorkflowEligible(field *gen.Field) (eligible bool, marker bool, err error) {
	raw, ok := field.Annotations[entx.WorkflowEligibleAnnotationName]
	if !ok {
		return false, false, nil
	}

	ann := &entx.WorkflowEligibleAnnotation{}
	if err := ann.Decode(raw); err != nil {
		return false, false, err
	}

	if field.Name == workflowEligibleMarkerField {
		return false, true, nil
	}

	return ann.Eligible, false, nil
}

// collectEntityData iterates the ent graph and collects schemas annotated with
// either OPENLANE_WORKFLOW_ELIGIBLE fields or OPENLANE_INTEGRATION_MAPPING_SCHEMA
func collectEntityData(g *gen.Graph, c *Config) (EntityData, error) {
	data := EntityData{
		PackageName:      c.PackageName,
		EntPackage:       c.EntPackage,
		GalaPackage:      c.GalaPackage,
		JsonxPackage:     c.JsonxPackage,
		LogxPackage:      c.LogxPackage,
		ContextxPackage:  c.ContextxPackage,
		CelxPackage:      c.CelxPackage,
		EnumsPackageName: c.EnumsPackageName,
		Schemas:          []EntitySchema{},
	}

	var eligibleSchemas []string

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
			eligibleSchemas = append(eligibleSchemas, node.Name)
		}
	}

	for _, node := range g.Nodes {
		if !slices.Contains(eligibleSchemas, node.Name) {
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

		// ObjectFields is the unified field catalog: every field with its type and capability flags,
		// consumed by both the workflow builder and the integration cross-link config
		var workflowMarker bool

		// integrationFields carries the per-field integration mapping metadata (keyed by ent field
		// name) and integrationMeta the schema-level mapping metadata, folded onto the unified catalog
		integrationFields, integrationMeta, err := collectIntegrationMapping(schema)
		if err != nil {
			return EntityData{}, fmt.Errorf("collect integration mapping for %s: %w", node.Name, err)
		}

		for _, field := range node.Fields {
			eligible, marker, err := fieldWorkflowEligible(field)
			if err != nil {
				return EntityData{}, fmt.Errorf("decode workflow eligible annotation on %s.%s: %w", node.Name, field.Name, err)
			}

			if marker {
				workflowMarker = true
			}

			fieldType := ""
			if field.Type != nil {
				fieldType = field.Type.String()
			}

			// MatchKey: plain-string indexed columns (e.g. external_id, ref_code) usable as cross-link
			// match keys; custom Go types and enums are excluded because their In predicates reject plain strings
			entityField := EntityField{
				Name:             field.StructField(),
				Snake:            field.StorageKey(),
				Type:             fieldType,
				WorkflowEligible: eligible,
				MatchKey:         field.Type != nil && field.Type.Type == entfield.TypeString && !field.HasGoType(),
			}

			if im, ok := integrationFields[field.Name]; ok {
				entityField.IntegrationMapped = true
				entityField.InputKey = im.InputKey
				entityField.InputGoField = im.InputGoField
				entityField.FromIntegration = im.FromIntegration
				entityField.LookupKey = im.LookupKey
			}

			entitySchema.ObjectFields = append(entitySchema.ObjectFields, entityField)
		}

		slices.SortFunc(entitySchema.ObjectFields, func(a, b EntityField) int {
			return cmp.Compare(a.Snake, b.Snake)
		})

		for _, edge := range node.Edges {
			// include every edge to an eligible target schema (cross-object linking, in either
			// direction and at any cardinality)
			if !slices.Contains(eligibleSchemas, edge.Type.Name) {
				continue
			}

			_, workflowEligible := edge.Annotations[entx.WorkflowEligibleAnnotationName]

			// immutable edges are included in the catalog (create-time injection can set them), but the
			// registry emits no Link/Unlink for them since the update builder has no setter; consumers
			// that mutate edges already nil-check Link
			fkColumn := ""
			if edge.Unique && edge.OwnFK() {
				fkColumn = edge.Rel.Column()
			}

			entitySchema.Edges = append(entitySchema.Edges, EntityEdge{
				Name:             edge.Name,
				TargetSchema:     edge.Type.Name,
				Unique:           edge.Unique,
				Optional:         edge.Optional,
				Immutable:        edge.Immutable,
				WorkflowEligible: workflowEligible,
				Field:            fkColumn,
			})
		}

		slices.SortFunc(entitySchema.Edges, func(a, b EntityEdge) int {
			return cmp.Compare(a.Name, b.Name)
		})

		// workflow eligibility is derived from the unified catalog: any workflow-eligible field or
		// edge, or the schema-level marker
		entitySchema.WorkflowEligible = workflowMarker ||
			slices.ContainsFunc(entitySchema.ObjectFields, func(f EntityField) bool { return f.WorkflowEligible }) ||
			slices.ContainsFunc(entitySchema.Edges, func(e EntityEdge) bool { return e.WorkflowEligible })

		if hasCreate {
			entitySchema.CreateInputType = "Create" + node.Name + "Input"
		}

		if hasUpdate {
			entitySchema.UpdateInputType = "Update" + node.Name + "Input"
		}

		entitySchema.IntegrationMapped = integrationMeta.Mapped
		entitySchema.StockPersist = integrationMeta.StockPersist
		entitySchema.RuntimeDefaults = integrationMeta.RuntimeDefaults

		if integrationMeta.Mapped && hasCreate {
			entitySchema.IngestTopic = "entityops." + entitySchema.Snake + ".ingest.requested"
			entitySchema.IngestRequestType = node.Name + "IngestRequested"
			entitySchema.IngestTopicVar = "Topic" + node.Name
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

		if _, ok := field.Annotations[entx.IntegrationMappingFieldAnnotationName]; ok {
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
		{name: "entity_handlers", filename: "entity_handlers.go", tmplFile: "templates/entity_handlers.tpl"},
		{name: "entity_workflow", filename: "entity_workflow.go", tmplFile: "templates/entity_workflow.tpl"},
		{name: "entity_links", filename: "entity_links.go", tmplFile: "templates/entity_links.tpl"},
		{name: "entity_integration", filename: "entity_integration.go", tmplFile: "templates/entity_integration.tpl"},
		{name: "entity_projection", filename: "entity_projection.go", tmplFile: "templates/entity_projection.tpl"},
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

// generateEnumFiles renders the WorkflowObjectType enum into the enums package, replacing the
// standalone workflowgen enum output with the same catalog-driven eligibility
func generateEnumFiles(outputDir string, data EntityData) error {
	raw, err := _templates.ReadFile("templates/entity_enums.tpl")
	if err != nil {
		return fmt.Errorf("read template templates/entity_enums.tpl: %w", err)
	}

	tmpl, err := template.New("entity_enums").Funcs(gen.Funcs).Parse(string(raw))
	if err != nil {
		return fmt.Errorf("parse template entity_enums: %w", err)
	}

	return writeFile(outputDir, "workflow_object_type.go", tmpl, data)
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
