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
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"golang.org/x/tools/imports"

	"github.com/theopenlane/entx"
)

//go:embed templates/*
var _templates embed.FS

const dirPermissions = 0o755

// literalsTemplate holds the descriptor literal partials shared by every rendered template
const literalsTemplate = "templates/entity_literals.tpl"

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
	// EnumsPackage is the enums package import path for notification content types
	EnumsPackage string
	// EnumsPackageName is the Go package name for the generated WorkflowObjectType enum file
	EnumsPackageName string
	// SlateparserPackage is the slateparser package import path for mention scanning
	SlateparserPackage string
	// IntegrationTypesPackage is the integrations types package import path for notify spec dispatch
	IntegrationTypesPackage string
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

// ApprovalSpecEntry is one schema's annotation-declared approval flow configuration
type ApprovalSpecEntry struct {
	// StatusField is the enum field carrying the approval status
	StatusField string
	// ApproverField is the group-id field resolving the approvers
	ApproverField string
}

// EntitySchema represents one schema's metadata for entity operations generation
type EntitySchema struct {
	// Name is the PascalCase schema name (e.g., "ActionPlan")
	Name string
	// Snake is the snake_case form (e.g., "action_plan")
	Snake string
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
	// OwnerField is the foreign-key field of the owner edge on org-owned schemas, empty for schemas without an owner
	OwnerField string
	// SystemScoped indicates the owner field is optional and the schema carries the system_owned marker, so an empty owner scopes queries to system-owned rows
	SystemScoped bool
	// ObjectFields is the unified per-schema field catalog (every field with capability flags),
	// consumed by both the workflow builder and the integration cross-link config. It is the single
	// field list: update-input re-keying, key-match columns, link source context, and workflow-eligible
	// fields are all derived from it by filtering on the per-field flags
	ObjectFields []EntityField
	// StampedCreateFields are the Go names of stamped fields the GraphQL create input does not carry
	StampedCreateFields []string
	// StampedUpdateFields are the Go names of stamped fields the GraphQL update input does not carry
	StampedUpdateFields []string
	// Edges contains every edge to an entityops schema (any cardinality/direction, mutable or immutable)
	// plus workflow group-permission edges; the single edge list for linking, workflow, and runtime ops
	Edges []EntityEdge
	// WorkflowEligible indicates the schema participates in workflows via eligible fields or edges
	WorkflowEligible bool
	// IntegrationMapped indicates the schema participates in integration ingest mapping
	IntegrationMapped bool
	// LinkTarget indicates the schema is an edge target of an integration-mapped schema, making it
	// reachable by ingest link-rule target selection; with IntegrationMapped and WorkflowEligible it
	// gates emission of the runtime closures and projections only reachable through those capabilities
	LinkTarget bool
	// ConsoleRoute is present only when the schema explicitly declares a console route
	ConsoleRoute *ConsoleRouteEntry
	// OrgOwned indicates the schema carries the org-owned annotation, so owner_id is the owning
	// organization; it gates mention-source generation and is not emitted into runtime descriptors
	OrgOwned bool
	// MentionSpec is present only when the schema explicitly declares mention scanning
	MentionSpec *MentionSpecEntry
	// ApprovalSpec is present only when the schema explicitly declares approval fields
	ApprovalSpec *ApprovalSpecEntry
	// TaskRules are schema-level (unconditional) suggested-task rules declared via entx.SchemaTaskRule
	TaskRules []entx.TaskRuleSpec
	// IntegrationFKField is the FK column of the schema's mutable unique edge to Integration, if any
	IntegrationFKField string
	// IntegrationM2MEdge is the name of the schema's to-many mutable edge to Integration, if any
	IntegrationM2MEdge string
	// HasIntegrationID reports whether the schema carries an integration_id column
	HasIntegrationID bool
	// HasIntegrationRunID reports whether the schema carries an integration_run_id field
	HasIntegrationRunID bool
	// IntegrationRunM2MEdge is the name of the schema's to-many mutable edge to IntegrationRun, if any
	IntegrationRunM2MEdge string
	// LookupAlternatives are the schema's ordered composite ingest lookup keys as snake_case field name sets
	LookupAlternatives [][]string
	// InstanceScoped reports whether same-key records from different source instances are distinct rows
	InstanceScoped bool
	// RemovedAtField is the snake_case name of the schema's SnapshotRemoval-annotated field, if any
	RemovedAtField string
	// RemovedAtEpisodic reports whether removal is a recurring observation rather than a permanent tombstone
	RemovedAtEpisodic bool
	// CatalogPointer is the foreign-key field of the CatalogEdge-annotated self edge, empty when the schema has no catalogue
	CatalogPointer string
	// CatalogFields lists the snake_case fields copied from a catalogue row on adopt and refresh
	CatalogFields []string
	// CatalogVisibility is the snake_case bool field marking a catalogue row as visible to organizations, empty when absent
	CatalogVisibility string
	// CatalogKey is the snake_case field on adopted rows holding the catalogue row's lookup key, empty when absent
	CatalogKey string
	// CatalogLookupKey is the snake_case lookup-key field whose value identifies a catalogue row, empty when absent
	CatalogLookupKey string
	// HasCatalog reports whether the schema supports catalogue adoption: a pointer, copied fields, the system_owned marker, and the visibility and key fields
	HasCatalog bool
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
	// LookupKey reports whether the field is the ingest upsert lookup column for its schema
	LookupKey bool
	// Sanitizable reports whether ingest preparation pre-validates this mapped field and drops invalid values
	Sanitizable bool
	// SliceInput reports whether the field's create-input carrier is a slice rather than a pointer scalar
	SliceInput bool
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
	// SystemControlled excludes the field from provider mappings
	SystemControlled bool
	// SourceManaged reports whether the upstream source owns the value and overwrites it on refresh or reconcile
	SourceManaged bool
	// Volatile excludes the field from triggering an ingest change
	Volatile bool
	// Stamped reports the field is system-controlled by annotation and written by StampProvenance
	Stamped bool
	// CreateInputSkipped reports the field is absent from the GraphQL create input
	CreateInputSkipped bool
	// UpdateInputSkipped reports the field is absent from the GraphQL update input
	UpdateInputSkipped bool
	// CaseInsensitive compares the field case-insensitively in ingest change detection
	CaseInsensitive bool
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
	// ThroughSourceField is the join entity's field holding this schema's id
	ThroughSourceField string
	// ThroughTargetField is the join entity's field holding the target's id
	ThroughTargetField string
}

// workflowEligibleMarkerField is the name of the WorkflowApprovalMixin carrier field that flags a
// schema as workflow-eligible without being a real workflow-triggerable field
const workflowEligibleMarkerField = "workflow_eligible_marker"

// integrationTargetSchema is the PascalCase name of the Integration schema
const integrationTargetSchema = "Integration"

// integrationRunTargetSchema is the PascalCase name of the IntegrationRun schema
const integrationRunTargetSchema = "IntegrationRun"

// integrationIDFieldName is the provenance column recording the writing installation's id
const integrationIDFieldName = "integration_id"

// integrationRunIDFieldName is the provenance column recording the integration run that last wrote a record
const integrationRunIDFieldName = "integration_run_id"

// detectIntegrationEdges records the schema's mutable FK and to-many edges to Integration and IntegrationRun
func detectIntegrationEdges(schema *EntitySchema) {
	for _, edge := range schema.Edges {
		switch edge.TargetSchema {
		case integrationTargetSchema:
			switch {
			case edge.Unique && edge.Field != "" && !edge.Immutable:
				schema.IntegrationFKField = edge.Field
			case !edge.Unique && !edge.Immutable && edge.ThroughType == "":
				schema.IntegrationM2MEdge = edge.Name
			}
		case integrationRunTargetSchema:
			if !edge.Unique && !edge.Immutable && edge.ThroughType == "" {
				schema.IntegrationRunM2MEdge = edge.Name
			}
		}
	}
}

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

// edgeCatalogPointer returns the foreign-key column of a CatalogEdge-annotated edge, which must be a
// unique self edge owning its foreign key; it returns empty when the edge carries no annotation
func edgeCatalogPointer(node *gen.Type, edge *gen.Edge) (string, error) {
	raw, ok := edge.Annotations[entx.CatalogEdgeAnnotationName]
	if !ok {
		return "", nil
	}

	ann := &entx.CatalogEdgeAnnotation{}
	if err := ann.Decode(raw); err != nil {
		return "", fmt.Errorf("decode catalog edge annotation on %s.%s: %w", node.Name, edge.Name, err)
	}

	if !edge.Unique || edge.Type.Name != node.Name || !edge.OwnFK() {
		return "", fmt.Errorf("%w: %s.%s", ErrCatalogEdgeInvalid, node.Name, edge.Name)
	}

	return edge.Rel.Column(), nil
}

// ownerEdgeName is the edge naming the organization that owns a row
const ownerEdgeName = "owner"

// schemaOwnerField returns the foreign-key column of the owner edge on an org-owned schema, empty when the schema is not org owned
func schemaOwnerField(node *gen.Type) (string, error) {
	if _, ok := node.Annotations[entx.OrgOwnedSchemaName]; !ok {
		return "", nil
	}

	for _, edge := range node.Edges {
		if edge.Name == ownerEdgeName && edge.Unique && edge.OwnFK() {
			return edge.Rel.Column(), nil
		}
	}

	return "", fmt.Errorf("%w: %s", ErrOwnerEdgeMissing, node.Name)
}

// systemOwnedFieldName is the marker column distinguishing catalogue rows from organization-owned rows
const systemOwnedFieldName = "system_owned"

// validateCatalog fails generation when a schema with a catalog edge lacks the inputs, markers, or lookup key adoption uses
func validateCatalog(schema EntitySchema) error {
	if schema.CatalogPointer == "" {
		return nil
	}

	switch {
	case !schema.HasCreate || !schema.HasUpdate:
		return fmt.Errorf("%w: %s", ErrCatalogInputsMissing, schema.Name)
	case schema.CatalogVisibility == "":
		return fmt.Errorf("%w: %s", ErrCatalogVisibilityMissing, schema.Name)
	case schema.CatalogKey == "":
		return fmt.Errorf("%w: %s", ErrCatalogKeyMissing, schema.Name)
	case schema.CatalogLookupKey == "":
		return fmt.Errorf("%w: %s", ErrCatalogLookupKeyMissing, schema.Name)
	case schema.OwnerField == "":
		return fmt.Errorf("%w: %s", ErrCatalogOwnerMissing, schema.Name)
	}

	return nil
}

// provenanceFieldNames are the record provenance columns every ingest-capable schema must carry
var provenanceFieldNames = []string{"source_definition_id", "source_definition_version", "source_instance_id", "managed_by"}

// validateProvenanceFields fails generation when an ingest-capable schema lacks the provenance fields
func validateProvenanceFields(schema EntitySchema) error {
	if !schema.IntegrationMapped || !schema.HasCreate {
		return nil
	}

	for _, name := range provenanceFieldNames {
		if !slices.ContainsFunc(schema.ObjectFields, func(f EntityField) bool { return f.Snake == name }) {
			return fmt.Errorf("%w: %s lacks %s", ErrProvenanceFieldsMissing, schema.Name, name)
		}
	}

	return nil
}

// synthesizeLookupAlternatives returns the declared alternatives or one synthesized from the LookupKey fields
func synthesizeLookupAlternatives(declared [][]string, lookupOrder []string) [][]string {
	if len(declared) > 0 {
		return declared
	}

	if len(lookupOrder) == 0 {
		return nil
	}

	return [][]string{lookupOrder}
}

// validateLookupAlternatives fails generation when a lookup alternative names an unknown or unmapped field
func validateLookupAlternatives(schema EntitySchema) error {
	if len(schema.LookupAlternatives) == 0 {
		return nil
	}

	if !schema.IntegrationMapped {
		return fmt.Errorf("%w: %s", ErrLookupAlternativeWithoutMapping, schema.Name)
	}

	for _, alternative := range schema.LookupAlternatives {
		for _, name := range alternative {
			field, ok := lo.Find(schema.ObjectFields, func(f EntityField) bool { return f.Snake == name })
			if !ok || !field.IntegrationMapped {
				return fmt.Errorf("%w: %s.%s", ErrLookupAlternativeFieldUnknown, schema.Name, name)
			}
		}
	}

	return nil
}

// fieldWebhookPayload reports whether a field is included in workflow webhook enrichment.
func fieldWebhookPayload(field *gen.Field) bool {
	ann, ok := entx.GetAnnotation[*entx.WebhookPayloadFieldAnnotation](field)

	return ok && ann.Include && !field.Sensitive()
}

// fieldCaseInsensitive reports whether ingest change detection compares the field case-insensitively
func fieldCaseInsensitive(field *gen.Field) bool {
	ann, ok := entx.GetAnnotation[*entx.CaseInsensitiveFieldAnnotation](field)

	return ok && ann.CaseInsensitive
}

// fieldSourceManaged reports whether the field carries entx.FieldSourceManaged
func fieldSourceManaged(field *gen.Field) bool {
	_, ok := field.Annotations[entx.FieldSourceManagedAnnotationName]

	return ok
}

// fieldCatalogVisibility reports whether the field carries entx.CatalogVisibilityField
func fieldCatalogVisibility(field *gen.Field) bool {
	_, ok := field.Annotations[entx.CatalogVisibilityFieldAnnotationName]

	return ok
}

// fieldCatalogKey reports whether the field carries entx.CatalogKeyField
func fieldCatalogKey(field *gen.Field) bool {
	_, ok := field.Annotations[entx.CatalogKeyFieldAnnotationName]

	return ok
}

// fieldMatchKey reports whether the field is a plain string column usable as a match key
func fieldMatchKey(field *gen.Field) bool {
	if field.Type == nil || field.Sensitive() {
		return false
	}

	return field.Type.Type == entfield.TypeString && !field.HasGoType()
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

	systemControlled := isIntegrationSystemField(field.StorageKey())
	volatile, stamped := false, false

	if ant, ok := entx.GetAnnotation[*entx.IntegrationMappingFieldAnnotation](field); ok {
		systemControlled = systemControlled || ant.SystemControlled
		volatile = ant.Volatile
		stamped = ant.SystemControlled
	}

	createInputSkipped, updateInputSkipped := false, false
	if gqlAnt, ok := entx.GetAnnotation[*entgql.Annotation](field); ok {
		createInputSkipped = gqlAnt.Skip.Is(entgql.SkipMutationCreateInput)
		updateInputSkipped = gqlAnt.Skip.Is(entgql.SkipMutationUpdateInput)
	}

	entityField = EntityField{
		Name:               field.StructField(),
		Snake:              field.StorageKey(),
		Type:               fieldType,
		WorkflowEligible:   eligible,
		MatchKey:           fieldMatchKey(field),
		Clearable:          field.Optional || field.Nillable,
		WebhookPayload:     fieldWebhookPayload(field),
		Projectable:        fieldProjectable(field),
		TaskRules:          taskRules,
		SystemControlled:   systemControlled,
		SourceManaged:      fieldSourceManaged(field),
		Volatile:           volatile,
		Stamped:            stamped,
		CreateInputSkipped: createInputSkipped,
		UpdateInputSkipped: updateInputSkipped,
		CaseInsensitive:    fieldCaseInsensitive(field),
	}

	if im, ok := integrationFields[field.Name]; ok {
		entityField.IntegrationMapped = true
		entityField.InputKey = im.InputKey
		entityField.InputGoField = im.InputGoField
		entityField.LookupKey = im.LookupKey
		entityField.SystemControlled = im.SystemControlled
		entityField.Sanitizable = ingestSanitizable(field, im)
		entityField.SliceInput = strings.HasPrefix(fieldType, "[]")
	}

	return entityField, marker, nil
}

// ingestSanitizable reports whether ingest preparation pre-validates an optional non-lookup mapped
// field whose generated validator signature matches a plain string or string-slice carrier
func ingestSanitizable(field *gen.Field, im integrationFieldMeta) bool {
	if !field.Optional || field.Validators == 0 || im.LookupKey {
		return false
	}

	if field.Type == nil {
		return false
	}

	if field.Type.String() == "[]string" {
		return true
	}

	return field.Type.Type == entfield.TypeString && !field.HasGoType()
}

// collectEntityData iterates the ent graph and collects every primary schema. Optional
// workflow, integration, and task-rule annotations add capabilities to the canonical schema;
// they do not control whether the schema exists in the registry.
func collectEntityData(g *gen.Graph, c *Config) (EntityData, error) { //nolint:gocyclo
	data := EntityData{
		PackageName:             c.PackageName,
		EntPackage:              c.EntPackage,
		GalaPackage:             c.GalaPackage,
		JsonxPackage:            c.JsonxPackage,
		LogxPackage:             c.LogxPackage,
		CelxPackage:             c.CelxPackage,
		MapxPackage:             c.MapxPackage,
		EnumsPackage:            c.EnumsPackage,
		EnumsPackageName:        c.EnumsPackageName,
		SlateparserPackage:      c.SlateparserPackage,
		IntegrationTypesPackage: c.IntegrationTypesPackage,
		Schemas:                 []EntitySchema{},
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

		_, orgOwned := node.Annotations[entx.OrgOwnedSchemaName]

		ownerField, err := schemaOwnerField(node)
		if err != nil {
			return EntityData{}, err
		}

		entitySchema := EntitySchema{
			Name:                node.Name,
			Snake:               strcase.SnakeCase(node.Name),
			Lower:               strings.ToLower(strings.ReplaceAll(strcase.SnakeCase(node.Name), "_", "")),
			OrgOwned:            orgOwned,
			HasCreate:           hasCreate,
			HasUpdate:           hasUpdate,
			PredicatePackage:    predAlias,
			PredicateImport:     predImport,
			OwnerField:          ownerField,
			SystemScoped:        fieldOptional(schema, ownerField) && hasField(schema, systemOwnedFieldName),
			HasIntegrationID:    hasField(schema, integrationIDFieldName),
			HasIntegrationRunID: hasField(schema, integrationRunIDFieldName),
		}

		schemaRules, err := schemaTaskRules(schema)
		if err != nil {
			return EntityData{}, fmt.Errorf("decode schema task rule annotation on %s: %w", node.Name, err)
		}

		entitySchema.TaskRules = schemaRules

		// ObjectFields is the unified field catalog: every field with its type and capability flags,
		// consumed by both the workflow builder and the integration cross-link config
		var (
			workflowMarker bool
			lookupOrder    []string
		)

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

			if entityField.LookupKey {
				lookupOrder = append(lookupOrder, entityField.Snake)
			}

			if entityField.SourceManaged {
				entitySchema.CatalogFields = append(entitySchema.CatalogFields, entityField.Snake)
			}

			if fieldCatalogVisibility(field) {
				entitySchema.CatalogVisibility = entityField.Snake
			}

			if fieldCatalogKey(field) {
				entitySchema.CatalogKey = entityField.Snake
			}

			entitySchema.ObjectFields = append(entitySchema.ObjectFields, entityField)
		}

		entitySchema.CatalogLookupKey = lo.FirstOr(lookupOrder, "")

		slices.SortFunc(entitySchema.ObjectFields, func(a, b EntityField) int {
			return cmp.Compare(a.Snake, b.Snake)
		})

		for _, field := range entitySchema.ObjectFields {
			if !field.Stamped {
				continue
			}

			if field.CreateInputSkipped {
				entitySchema.StampedCreateFields = append(entitySchema.StampedCreateFields, field.Name)
			}

			if field.UpdateInputSkipped {
				entitySchema.StampedUpdateFields = append(entitySchema.StampedUpdateFields, field.Name)
			}
		}

		for _, edge := range node.Edges {
			// Include every edge to a registered target schema. Optional capability flags are
			// properties of the edge and never determine whether its target has a descriptor.
			workflowEligible, err := edgeWorkflowEligible(edge)
			if err != nil {
				return EntityData{}, fmt.Errorf("decode workflow eligible annotation on %s.%s: %w", node.Name, edge.Name, err)
			}

			catalogPointer, err := edgeCatalogPointer(node, edge)
			if err != nil {
				return EntityData{}, err
			}

			if catalogPointer != "" {
				if entitySchema.CatalogPointer != "" {
					return EntityData{}, fmt.Errorf("%w: %s.%s and %s.%s", ErrCatalogEdgeConflict, node.Name, entitySchema.CatalogPointer, node.Name, catalogPointer)
				}

				entitySchema.CatalogPointer = catalogPointer
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
				entityEdge.ThroughSourceField = templates.ToGo(sourceColumn)
				entityEdge.ThroughTargetField = templates.ToGo(targetColumn)
			}

			entitySchema.Edges = append(entitySchema.Edges, entityEdge)
		}

		slices.SortFunc(entitySchema.Edges, func(a, b EntityEdge) int {
			return cmp.Compare(a.Name, b.Name)
		})

		detectIntegrationEdges(&entitySchema)

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
		entitySchema.InstanceScoped = integrationMeta.InstanceScoped
		entitySchema.LookupAlternatives = synthesizeLookupAlternatives(integrationMeta.LookupAlternatives, lookupOrder)
		entitySchema.HasCatalog = entitySchema.CatalogPointer != "" && len(entitySchema.CatalogFields) > 0 && hasField(schema, systemOwnedFieldName) &&
			entitySchema.CatalogVisibility != "" && entitySchema.CatalogKey != ""

		if err := validateCatalog(entitySchema); err != nil {
			return EntityData{}, err
		}

		if err := validateLookupAlternatives(entitySchema); err != nil {
			return EntityData{}, err
		}

		if err := validateProvenanceFields(entitySchema); err != nil {
			return EntityData{}, err
		}

		data.Schemas = append(data.Schemas, entitySchema)
	}

	slices.SortFunc(data.Schemas, func(a, b EntitySchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	markLinkTargets(&data)

	if err := collectSchemaMetadata(g, &data); err != nil {
		return EntityData{}, err
	}

	return data, nil
}

// markLinkTargets flags every schema reachable as an edge target of an integration-mapped schema,
// since ingest link rules may select targets over any edge of a mapped source schema
func markLinkTargets(data *EntityData) {
	targets := map[string]struct{}{}

	for _, schema := range data.Schemas {
		if !schema.IntegrationMapped {
			continue
		}

		for _, edge := range schema.Edges {
			if edge.TargetInRegistry {
				targets[edge.TargetSchema] = struct{}{}
			}
		}
	}

	for i := range data.Schemas {
		_, data.Schemas[i].LinkTarget = targets[data.Schemas[i].Name]
	}
}

// collectSchemaMetadata gathers annotation-declared console-route, display, mention, and
// approval metadata. Catalog membership never implies that an entity is console-routable.
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

		if err := collectConsoleRoute(node, &data.Schemas[index]); err != nil {
			return err
		}

		markers, err := collectFieldMarkers(node)
		if err != nil {
			return err
		}

		if err := applyFieldMarkers(&data.Schemas[index], node.Name, markers); err != nil {
			return err
		}
	}

	return nil
}

// collectConsoleRoute decodes and validates a node's console-route annotation onto its schema
func collectConsoleRoute(node *gen.Type, schema *EntitySchema) error {
	raw, ok := node.Annotations[entx.ConsoleRouteAnnotationName]
	if !ok {
		return nil
	}

	routeAnn := &entx.ConsoleRouteAnnotation{}
	if err := routeAnn.Decode(raw); err != nil {
		return fmt.Errorf("decode console route annotation on %s: %w", node.Name, err)
	}

	if routeAnn.IDParam != "" && routeAnn.Suffix != "" {
		return fmt.Errorf("%w: %s", ErrInvalidConsoleRoute, node.Name)
	}

	schema.ConsoleRoute = &ConsoleRouteEntry{
		Base:    cmp.Or(routeAnn.Base, node.Table()),
		IDParam: routeAnn.IDParam,
		Suffix:  routeAnn.Suffix,
	}

	return nil
}

// schemaFieldMarkers holds the storage names resolved from a node's field-marker annotations
type schemaFieldMarkers struct {
	// display is the field carrying the schema's display name
	display string
	// details is the plain-text mention source field
	details string
	// detailsJSON is the JSON mention source field
	detailsJSON string
	// status is the approval status enum field
	status string
	// approver is the approval approver group field
	approver string
	// removedAt is the snapshot-removal field
	removedAt string
	// removedAtEpisodic reports whether removal is a recurring observation rather than a permanent tombstone
	removedAtEpisodic bool
}

// collectFieldMarkers scans a node's fields for display, mention, and approval markers,
// rejecting duplicate markers and wrongly typed fields
func collectFieldMarkers(node *gen.Type) (schemaFieldMarkers, error) {
	markers := schemaFieldMarkers{}

	for _, f := range node.Fields {
		storage := f.StorageKey()

		if _, ok := f.Annotations[entx.DisplayNameAnnotationName]; ok {
			if markers.display != "" {
				return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrDisplayNameConflict, node.Name, markers.display, node.Name, storage)
			}

			markers.display = storage
		}

		if _, ok := f.Annotations[entx.MentionSourceAnnotationName]; ok {
			switch {
			case f.Type != nil && f.Type.Type == entfield.TypeJSON:
				if markers.detailsJSON != "" {
					return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrMentionSourceConflict, node.Name, markers.detailsJSON, node.Name, storage)
				}

				markers.detailsJSON = storage
			case f.Type != nil && f.Type.Type == entfield.TypeString:
				if markers.details != "" {
					return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrMentionSourceConflict, node.Name, markers.details, node.Name, storage)
				}

				markers.details = storage
			default:
				return markers, fmt.Errorf("%w: %s.%s", ErrMentionSourceType, node.Name, storage)
			}
		}

		if _, ok := f.Annotations[entx.ApprovalStatusAnnotationName]; ok {
			if markers.status != "" {
				return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrApprovalFieldConflict, node.Name, markers.status, node.Name, storage)
			}

			if f.Type == nil || f.Type.Type != entfield.TypeEnum {
				return markers, fmt.Errorf("%w: %s.%s", ErrApprovalFieldType, node.Name, storage)
			}

			markers.status = storage
		}

		if _, ok := f.Annotations[entx.ApprovalApproverAnnotationName]; ok {
			if markers.approver != "" {
				return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrApprovalFieldConflict, node.Name, markers.approver, node.Name, storage)
			}

			if f.Type == nil || f.Type.Type != entfield.TypeString {
				return markers, fmt.Errorf("%w: %s.%s", ErrApprovalFieldType, node.Name, storage)
			}

			markers.approver = storage
		}

		if raw, ok := f.Annotations[entx.SnapshotRemovalAnnotationName]; ok {
			if markers.removedAt != "" {
				return markers, fmt.Errorf("%w: %s.%s and %s.%s", ErrSnapshotRemovalConflict, node.Name, markers.removedAt, node.Name, storage)
			}

			ann := &entx.SnapshotRemovalAnnotation{}
			if err := ann.Decode(raw); err != nil {
				return markers, fmt.Errorf("decode snapshot removal annotation on %s.%s: %w", node.Name, storage, err)
			}

			markers.removedAt = storage
			markers.removedAtEpisodic = ann.Episodic
		}
	}

	return markers, nil
}

// applyFieldMarkers folds resolved field markers onto a schema: the display key flag and the
// mention and approval specs, both gated on org ownership
func applyFieldMarkers(schema *EntitySchema, name string, markers schemaFieldMarkers) error {
	if markers.display != "" {
		for i := range schema.ObjectFields {
			if schema.ObjectFields[i].Snake == markers.display {
				schema.ObjectFields[i].DisplayKey = true
				break
			}
		}
	}

	if markers.details != "" || markers.detailsJSON != "" {
		if !schema.OrgOwned {
			return fmt.Errorf("%w: %s", ErrMentionOrgOwnedRequired, name)
		}

		schema.MentionSpec = &MentionSpecEntry{
			NameField:        markers.display,
			DetailsField:     markers.details,
			DetailsJSONField: markers.detailsJSON,
			OwnerField:       schema.OwnerField,
		}
	}

	if (markers.status != "") != (markers.approver != "") {
		return fmt.Errorf("%w: %s", ErrApprovalSpecIncomplete, name)
	}

	if markers.status != "" && markers.approver != "" {
		if !schema.OrgOwned {
			return fmt.Errorf("%w: %s", ErrApprovalOrgOwnedRequired, name)
		}

		schema.ApprovalSpec = &ApprovalSpecEntry{
			StatusField:   markers.status,
			ApproverField: markers.approver,
		}
	}

	if markers.removedAt != "" {
		schema.RemovedAtField = markers.removedAt
		schema.RemovedAtEpisodic = markers.removedAtEpisodic
	}

	return nil
}

// generateEntityFiles renders all templates and writes them to the output directory
func generateEntityFiles(outputDir string, data EntityData) error {
	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		return fmt.Errorf("create output dir %s: %w", outputDir, err)
	}

	type templateSpec struct {
		name     string
		filename string
	}

	specs := []templateSpec{
		{name: "entity_schema", filename: "entity_schema.go"},
		{name: "entity_errors", filename: "entity_errors.go"},
		{name: "entity_registry", filename: "entity_registry.go"},
		{name: "entity_workflow", filename: "entity_workflow.go"},
		{name: "entity_tasks", filename: "entity_tasks.go"},
		{name: "entity_links", filename: "entity_links.go"},
		{name: "entity_integration", filename: "entity_integration.go"},
		{name: "entity_projection", filename: "entity_projection.go"},
		{name: "entity_metadata", filename: "entity_metadata.go"},
		{name: "entity_changeset", filename: "entity_changeset.go"},
		{name: "entity_mutation_events", filename: "entity_mutation_events.go"},
		{name: "entity_listener", filename: "entity_listener.go"},
		{name: "entity_notification", filename: "entity_notification.go"},
	}

	for _, filename := range []string{"entity_handlers.go"} {
		if err := os.Remove(filepath.Join(outputDir, filename)); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove obsolete generated file %s: %w", filename, err)
		}
	}

	for _, spec := range specs {
		tmpl, err := parseTemplate(spec.name)
		if err != nil {
			return err
		}

		if err := writeFile(outputDir, spec.filename, tmpl, data); err != nil {
			return err
		}
	}

	return nil
}

// parseTemplate parses the named template together with the shared literal partial
func parseTemplate(name string) (*template.Template, error) {
	tmpl := template.New(name).Funcs(gen.Funcs)

	for _, file := range []string{literalsTemplate, "templates/" + name + ".tpl"} {
		raw, err := _templates.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read template %s: %w", file, err)
		}

		if _, err := tmpl.Parse(string(raw)); err != nil {
			return nil, fmt.Errorf("parse template %s: %w", file, err)
		}
	}

	return tmpl, nil
}

// generateEnumFiles renders the WorkflowObjectType enum into the enums package, replacing the
// standalone workflowgen enum output with the same catalog-driven eligibility
func generateEnumFiles(outputDir string, data EntityData) error {
	tmpl, err := parseTemplate("entity_enums")
	if err != nil {
		return err
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

// fieldOptional reports whether a schema declares the named field as optional
func fieldOptional(schema *load.Schema, name string) bool {
	for _, f := range schema.Fields {
		if f.Name == name {
			return f.Optional
		}
	}

	return false
}
