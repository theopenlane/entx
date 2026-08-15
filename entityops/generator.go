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
	"github.com/99designs/gqlgen/codegen/templates"
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
	// CelxPackage is the celx package import path for typed entity expression evaluation
	CelxPackage string
	// MapxPackage is the mapx package import path for map clone/merge helpers
	MapxPackage string
	// EnumsPackageName is the Go package name for the generated WorkflowObjectType enum file
	EnumsPackageName string
	// Schemas contains all schemas eligible for entity operations
	Schemas []EntitySchema
}

// ConsoleRouteEntry is one schema's annotation-declared console route.
type ConsoleRouteEntry struct {
	// Base is the console landing path (e.g. "automation/tasks")
	Base string
	// IDParam routes object links through a query parameter instead of a path segment
	IDParam string
	// Suffix is a path segment appended after the object ID
	Suffix string
}

// MentionSpecEntry is one schema's annotation-declared mention scan configuration.
type MentionSpecEntry struct {
	// NameField is the display-name field used in mention notification content
	NameField string
	// DetailsField is the plain-text rich-text field scanned for mentions
	DetailsField string
	// DetailsJSONField is the JSON rich-text field scanned for mentions
	DetailsJSONField string
	// OwnerField is the owning-organization field on the schema
	OwnerField string
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
	// RuntimeDefaults are integration-injected field defaults applied by the schema ingest capability
	RuntimeDefaults []EntityRuntimeDefault
	// ConsoleRoute is present only when the schema explicitly declares a console route
	ConsoleRoute *ConsoleRouteEntry
	// MentionSpec is present only when the schema explicitly declares mention scanning
	MentionSpec *MentionSpecEntry
	// TaskRules are schema-level (unconditional) suggested-task rules declared via entx.SchemaTaskRule
	TaskRules []entx.TaskRuleSpec
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
	// UpsertKey reports whether the field belongs to the schema's logical ingest identity
	UpsertKey bool
	// LookupKey reports whether the field is the ingest upsert lookup column for its schema
	LookupKey bool
	// DisplayKey reports whether the field is the schema's display-name source
	DisplayKey bool
	// Clearable reports whether update inputs support explicitly clearing this field
	Clearable bool
	// WebhookPayload reports whether workflow webhook payloads include this field
	WebhookPayload bool
	// Projectable reports whether the field may appear in CEL/jsonschema projections
	Projectable bool
	// TaskRules are suggested-task rules declared on this field via entx.FieldTaskRule
	TaskRules []entx.TaskRuleSpec
}

// EntityEdge represents one linkable edge on a schema, in either direction
type EntityEdge struct {
	// Name is the edge name (e.g., "controls")
	Name string
	// TargetSchema is the target PascalCase name (e.g., "Control")
	TargetSchema string
	// TargetInRegistry reports whether the target schema has its own registry entry; workflow-annotated
	// edges may target schemas outside the registry, which have no Schema var to reference
	TargetInRegistry bool
	// Unique reports whether this side references a single target, so linking sets one id
	// (Set<Edge>ID) rather than adding many (Add<Edge>IDs)
	Unique bool
	// Immutable reports whether the edge is set only at create time; it gates update-input keys
	Immutable bool
	// WorkflowEligible reports whether the edge may drive workflow conditions and triggers
	WorkflowEligible bool
	// Field is the foreign-key storage column on this schema's table for unique owning edges
	// (e.g. "control_id"); empty when the foreign key lives on the target table
	Field string
	// ThroughType is the join entity's Go type name when the edge goes through an edge schema
	// (e.g. "FindingControl"); empty for plain edges. Through edges are linked by creating join
	// entity rows, since batch edge adds cannot generate per-row entity ids
	ThroughType string
	// ThroughSourceSetter is the join create-builder setter binding this schema's id (e.g. "SetFindingID")
	ThroughSourceSetter string
	// ThroughTargetSetter is the join create-builder setter binding the target's id (e.g. "SetControlID")
	ThroughTargetSetter string
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

	if field.Sensitive() || field.Immutable {
		return false, false, nil
	}

	if ant, ok := entx.GetAnnotation[*entgql.Annotation](field); ok &&
		(ant.Skip.Is(entgql.SkipType) || ant.Skip.Is(entgql.SkipMutationUpdateInput)) {
		return false, false, nil
	}

	return ann.Eligible, false, nil
}

// edgeWorkflowEligible decodes the annotation value rather than treating its presence as opt-in.
func edgeWorkflowEligible(edge *gen.Edge) (bool, error) {
	raw, ok := edge.Annotations[entx.WorkflowEligibleAnnotationName]
	if !ok {
		return false, nil
	}

	ann := &entx.WorkflowEligibleAnnotation{}
	if err := ann.Decode(raw); err != nil {
		return false, err
	}

	return ann.Eligible, nil
}

// fieldWebhookPayload reports whether a field is included in workflow webhook enrichment.
func fieldWebhookPayload(field *gen.Field) bool {
	ann, ok := entx.GetAnnotation[*entx.WebhookPayloadFieldAnnotation](field)

	return ok && ann.Include && !field.Sensitive()
}

// fieldProjectable excludes secrets and fields hidden from the GraphQL type surface.
func fieldProjectable(field *gen.Field) bool {
	if field.Sensitive() {
		return false
	}

	if ant, ok := entx.GetAnnotation[*entgql.Annotation](field); ok && ant.Skip.Is(entgql.SkipType) {
		return false
	}

	return true
}

// fieldTaskRules decodes the OPENLANE_TASK_RULE annotation on a field if it exists
func fieldTaskRules(field *gen.Field) ([]entx.TaskRuleSpec, error) {
	raw, ok := field.Annotations[entx.TaskRuleAnnotationName]
	if !ok {
		return nil, nil
	}

	ann := &entx.TaskRuleAnnotation{}
	if err := ann.Decode(raw); err != nil {
		return nil, err
	}

	return ann.Rules, nil
}

// schemaTaskRules decodes the OPENLANE_TASK_RULE annotation on the schema itself if it exists
func schemaTaskRules(schema *load.Schema) ([]entx.TaskRuleSpec, error) {
	if schema == nil {
		return nil, nil
	}

	raw, ok := schema.Annotations[entx.TaskRuleAnnotationName]
	if !ok {
		return nil, nil
	}

	ann := &entx.TaskRuleAnnotation{}
	if err := ann.Decode(raw); err != nil {
		return nil, err
	}

	return ann.Rules, nil
}

// hasTaskRuleAnnotation reports whether a schema carries a task rule via a field or the schema itself
func hasTaskRuleAnnotation(node *gen.Type, schema *load.Schema) bool {
	for _, field := range node.Fields {
		if _, ok := field.Annotations[entx.TaskRuleAnnotationName]; ok {
			return true
		}
	}

	if schema == nil {
		return false
	}

	_, ok := schema.Annotations[entx.TaskRuleAnnotationName]

	return ok
}

// buildEntityField constructs one field's catalog entry: capability flags decoded from its
// workflow/task-rule annotations, folded with its integration mapping metadata if any. marker
// reports whether this is the WorkflowApprovalMixin carrier field (see fieldWorkflowEligible)
func buildEntityField(node *gen.Type, field *gen.Field, integrationFields map[string]integrationFieldMeta) (entityField EntityField, marker bool, err error) {
	eligible, marker, err := fieldWorkflowEligible(field)
	if err != nil {
		return EntityField{}, false, fmt.Errorf("decode workflow eligible annotation on %s.%s: %w", node.Name, field.Name, err)
	}

	taskRules, err := fieldTaskRules(field)
	if err != nil {
		return EntityField{}, false, fmt.Errorf("decode task rule annotation on %s.%s: %w", node.Name, field.Name, err)
	}

	fieldType := ""
	if field.Type != nil {
		fieldType = field.Type.String()
	}

	// MatchKey: plain-string indexed columns (e.g. external_id, ref_code) usable as cross-link
	// match keys; custom Go types and enums are excluded because their In predicates reject plain strings
	entityField = EntityField{
		Name:             field.StructField(),
		Snake:            field.StorageKey(),
		Type:             fieldType,
		WorkflowEligible: eligible,
		MatchKey:         field.Type != nil && field.Type.Type == entfield.TypeString && !field.HasGoType() && !field.Sensitive(),
		Clearable:        field.Optional || field.Nillable,
		WebhookPayload:   fieldWebhookPayload(field),
		Projectable:      fieldProjectable(field),
		TaskRules:        taskRules,
	}

	if im, ok := integrationFields[field.Name]; ok {
		entityField.IntegrationMapped = true
		entityField.InputKey = im.InputKey
		entityField.InputGoField = im.InputGoField
		entityField.UpsertKey = im.UpsertKey
		entityField.LookupKey = im.LookupKey
	}

	return entityField, marker, nil
}

// collectEntityData iterates the ent graph and collects every primary schema. Optional
// workflow, integration, and task-rule annotations add capabilities to the canonical schema;
// they do not control whether the schema exists in the registry.
func collectEntityData(g *gen.Graph, c *Config) (EntityData, error) {
	data := EntityData{
		PackageName:      c.PackageName,
		EntPackage:       c.EntPackage,
		GalaPackage:      c.GalaPackage,
		JsonxPackage:     c.JsonxPackage,
		LogxPackage:      c.LogxPackage,
		CelxPackage:      c.CelxPackage,
		MapxPackage:      c.MapxPackage,
		EnumsPackageName: c.EnumsPackageName,
		Schemas:          []EntitySchema{},
	}

	var registeredSchemas []string

	for _, node := range g.Nodes {
		if skipNode(node) {
			continue
		}

		registeredSchemas = append(registeredSchemas, node.Name)
	}

	for _, node := range g.Nodes {
		if !slices.Contains(registeredSchemas, node.Name) {
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
			HasCreate:        hasCreate,
			HasUpdate:        hasUpdate,
			PredicatePackage: predAlias,
			PredicateImport:  predImport,
			HasOwnerID:       hasField(schema, "owner_id"),
		}

		schemaRules, err := schemaTaskRules(schema)
		if err != nil {
			return EntityData{}, fmt.Errorf("decode schema task rule annotation on %s: %w", node.Name, err)
		}

		entitySchema.TaskRules = schemaRules

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
			entityField, marker, err := buildEntityField(node, field, integrationFields)
			if err != nil {
				return EntityData{}, err
			}

			if marker {
				workflowMarker = true
			}

			entitySchema.ObjectFields = append(entitySchema.ObjectFields, entityField)
		}

		slices.SortFunc(entitySchema.ObjectFields, func(a, b EntityField) int {
			return cmp.Compare(a.Snake, b.Snake)
		})

		if displayField := displayFieldName(node); displayField != "" {
			for i := range entitySchema.ObjectFields {
				if entitySchema.ObjectFields[i].Snake == displayField {
					entitySchema.ObjectFields[i].DisplayKey = true
					break
				}
			}
		}

		for _, edge := range node.Edges {
			// Include every edge to a registered target schema. Optional capability flags are
			// properties of the edge and never determine whether its target has a descriptor.
			workflowEligible, err := edgeWorkflowEligible(edge)
			if err != nil {
				return EntityData{}, fmt.Errorf("decode workflow eligible annotation on %s.%s: %w", node.Name, edge.Name, err)
			}

			targetInRegistry := slices.Contains(registeredSchemas, edge.Type.Name)
			if !targetInRegistry && !workflowEligible {
				continue
			}

			// immutable edges are included in the catalog (create-time injection can set them), but the
			// registry emits no Link/Unlink for them since the update builder has no setter; consumers
			// that mutate edges already nil-check Link
			fkColumn := ""
			if edge.Unique && edge.OwnFK() {
				fkColumn = edge.Rel.Column()
			}

			entityEdge := EntityEdge{
				Name:             edge.Name,
				TargetSchema:     edge.Type.Name,
				TargetInRegistry: targetInRegistry,
				Unique:           edge.Unique,
				Immutable:        edge.Immutable,
				WorkflowEligible: workflowEligible,
				Field:            fkColumn,
			}

			// through edges are linked by creating rows of the join entity, so capture the join
			// type and the create-builder setters for each side; the relation columns are ordered
			// owner-first, so the inverse side's own column is the second
			if edge.Through != nil && len(edge.Rel.Columns) == 2 {
				sourceColumn, targetColumn := edge.Rel.Columns[0], edge.Rel.Columns[1]
				if edge.IsInverse() {
					sourceColumn, targetColumn = targetColumn, sourceColumn
				}

				entityEdge.ThroughType = edge.Through.Name
				entityEdge.ThroughSourceSetter = "Set" + templates.ToGo(sourceColumn)
				entityEdge.ThroughTargetSetter = "Set" + templates.ToGo(targetColumn)
			}

			entitySchema.Edges = append(entitySchema.Edges, entityEdge)
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
		entitySchema.RuntimeDefaults = integrationMeta.RuntimeDefaults

		data.Schemas = append(data.Schemas, entitySchema)
	}

	slices.SortFunc(data.Schemas, func(a, b EntitySchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	if err := collectSchemaMetadata(g, &data); err != nil {
		return EntityData{}, err
	}

	return data, nil
}

// displayFieldConvention is the ordered field-name chain used to resolve a schema's
// display-name field
var displayFieldConvention = []string{"name", "title", "display_name"}

// displayFieldName resolves a node's display-name field via the name/title/display_name
// convention chain; empty when no candidate field exists
func displayFieldName(node *gen.Type) string {
	for _, candidate := range displayFieldConvention {
		if nodeHasField(node, candidate) {
			return candidate
		}
	}

	return ""
}

// collectSchemaMetadata gathers annotation-declared console-route and mention metadata.
// Catalog membership never implies that an entity is console-routable.
func collectSchemaMetadata(g *gen.Graph, data *EntityData) error {
	schemaIndexes := make(map[string]int, len(data.Schemas))
	for i := range data.Schemas {
		schemaIndexes[data.Schemas[i].Name] = i
	}

	for _, node := range g.Nodes {
		// only history shadows are excluded here: schemas skipped for schema/query generation
		// (e.g. types used through extended types) still carry console and mention metadata
		if strings.HasSuffix(node.Name, "History") {
			continue
		}

		index, cataloged := schemaIndexes[node.Name]
		if !cataloged {
			continue
		}

		routeAnn, hasRouteAnn := (*entx.ConsoleRouteAnnotation)(nil), false

		if raw, ok := node.Annotations[entx.ConsoleRouteAnnotationName]; ok {
			routeAnn = &entx.ConsoleRouteAnnotation{}
			if err := routeAnn.Decode(raw); err != nil {
				return fmt.Errorf("decode console route annotation on %s: %w", node.Name, err)
			}

			if routeAnn.IDParam != "" && routeAnn.Suffix != "" {
				return fmt.Errorf("%w: %s", ErrInvalidConsoleRoute, node.Name)
			}

			hasRouteAnn = true
		}

		if hasRouteAnn {
			entry := ConsoleRouteEntry{Base: node.Table()}
			entry.Base = cmp.Or(routeAnn.Base, entry.Base)
			entry.IDParam = routeAnn.IDParam
			entry.Suffix = routeAnn.Suffix
			data.Schemas[index].ConsoleRoute = &entry
		}

		if raw, ok := node.Annotations[entx.MentionableAnnotationName]; ok {
			ann := &entx.MentionableAnnotation{}
			if err := ann.Decode(raw); err != nil {
				return fmt.Errorf("decode mentionable annotation on %s: %w", node.Name, err)
			}

			entry := MentionSpecEntry{
				NameField:        cmp.Or(ann.NameField, displayFieldName(node)),
				DetailsField:     cmp.Or(ann.DetailsField, "details"),
				DetailsJSONField: cmp.Or(ann.DetailsJSONField, "details_json"),
				OwnerField:       cmp.Or(ann.OwnerField, "owner_id"),
			}

			for _, fieldName := range []string{entry.NameField, entry.DetailsField, entry.DetailsJSONField, entry.OwnerField} {
				if fieldName != "" && !nodeHasField(node, fieldName) {
					return fmt.Errorf("%w: %s.%s", ErrMentionFieldMissing, node.Name, fieldName)
				}
			}

			data.Schemas[index].MentionSpec = &entry
		}
	}

	return nil
}

// nodeHasField reports whether the graph node declares a field with the given storage name
func nodeHasField(node *gen.Type, name string) bool {
	for _, f := range node.Fields {
		if f.Name == name || f.StorageKey() == name {
			return true
		}
	}

	return false
}

// schemaSource tracks which annotation(s) caused a schema to be included
type schemaSource struct {
	Workflow    bool
	Integration bool
	TaskRules   bool
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

	if hasTaskRuleAnnotation(node, schema) {
		source.TaskRules = true
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
		{name: "entity_workflow", filename: "entity_workflow.go", tmplFile: "templates/entity_workflow.tpl"},
		{name: "entity_tasks", filename: "entity_tasks.go", tmplFile: "templates/entity_tasks.tpl"},
		{name: "entity_links", filename: "entity_links.go", tmplFile: "templates/entity_links.tpl"},
		{name: "entity_integration", filename: "entity_integration.go", tmplFile: "templates/entity_integration.tpl"},
		{name: "entity_projection", filename: "entity_projection.go", tmplFile: "templates/entity_projection.tpl"},
		{name: "entity_metadata", filename: "entity_metadata.go", tmplFile: "templates/entity_metadata.tpl"},
		{name: "entity_changeset", filename: "entity_changeset.go", tmplFile: "templates/entity_changeset.tpl"},
		{name: "entity_mutation_events", filename: "entity_mutation_events.go", tmplFile: "templates/entity_mutation_events.tpl"},
		{name: "entity_listener", filename: "entity_listener.go", tmplFile: "templates/entity_listener.tpl"},
	}

	for _, filename := range []string{"entity_handlers.go"} {
		if err := os.Remove(filepath.Join(outputDir, filename)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove obsolete generated file %s: %w", filename, err)
		}
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

// skipNode returns true if the node should be excluded from the primary entity catalog.
// GraphQL schema/query generation flags are intentionally irrelevant here: entityops is the
// runtime catalog for Ent entities, including types exposed only through extended GraphQL APIs.
func skipNode(node *gen.Type) bool {
	return strings.HasSuffix(node.Name, "History")
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
