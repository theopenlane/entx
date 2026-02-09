package workflowgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"golang.org/x/tools/imports"
)

// ExtensionOption is a function that modifies the Extension configuration
type ExtensionOption func(*Extension)

// Config is the configuration for workflow code generation
type Config struct {
	// HooksOutputDir is the output directory for workflow-generated hooks helpers
	HooksOutputDir string
	// HooksPackageName is the package name for workflow-generated hooks helpers
	HooksPackageName string
	// EnumsOutputDir is the output directory for workflow object type enums
	EnumsOutputDir string
	// EnumsPackageName is the package name for workflow object type enums
	EnumsPackageName string
	// EnumsImportPath is the import path for the enums package
	EnumsImportPath string
	// WorkflowsImportPath is the import path for the workflows package
	WorkflowsImportPath string
}

// Extension implements entc.Extension for workflow-related generated helpers
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// New creates a new workflowgen extension
func New(opts ...ExtensionOption) *Extension {
	ext := &Extension{
		config: &Config{
			HooksOutputDir:      "./internal/ent/workflowgenerated",
			HooksPackageName:    "workflowgenerated",
			EnumsOutputDir:      "./common/enums",
			EnumsPackageName:    "enums",
			EnumsImportPath:     "github.com/theopenlane/core/common/enums",
			WorkflowsImportPath: "github.com/theopenlane/core/internal/workflows",
		},
	}

	for _, opt := range opts {
		opt(ext)
	}

	return ext
}

// WithHooksOutputDir sets the output directory for workflow-generated hooks helpers
func WithHooksOutputDir(dir string) ExtensionOption {
	return func(e *Extension) {
		e.config.HooksOutputDir = dir
	}
}

// WithHooksPackageName sets the package name for workflow-generated hooks helpers
func WithHooksPackageName(name string) ExtensionOption {
	return func(e *Extension) {
		e.config.HooksPackageName = name
	}
}

// WithEnumsOutputDir sets the output directory for workflow object type enums
func WithEnumsOutputDir(dir string) ExtensionOption {
	return func(e *Extension) {
		e.config.EnumsOutputDir = dir
	}
}

// WithEnumsPackageName sets the package name for workflow object type enums
func WithEnumsPackageName(name string) ExtensionOption {
	return func(e *Extension) {
		e.config.EnumsPackageName = name
	}
}

// WithEnumsImportPath sets the import path for the enums package
func WithEnumsImportPath(path string) ExtensionOption {
	return func(e *Extension) {
		e.config.EnumsImportPath = path
	}
}

// WithWorkflowsImportPath sets the import path for the workflows package
func WithWorkflowsImportPath(path string) ExtensionOption {
	return func(e *Extension) {
		e.config.WorkflowsImportPath = path
	}
}

// Hooks satisfies the entc.Extension interface
func (e Extension) Hooks() []gen.Hook {
	return []gen.Hook{e.Hook()}
}

// Hook generates workflow registry helpers and enums after ent codegen runs
func (e Extension) Hook() gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			if err := next.Generate(g); err != nil {
				return err
			}

			ctx := templateContext{
				Graph:                   g,
				CreatableNodes:          buildCreatableNodes(g),
				HooksPackageName:        e.config.HooksPackageName,
				EnumsPackageName:        e.config.EnumsPackageName,
				EnumsImportPath:         e.config.EnumsImportPath,
				WorkflowsImportPath:     e.config.WorkflowsImportPath,
				GeneratedImportPath:     g.Package,
				WorkflowObjectRefImport: filepath.ToSlash(filepath.Join(g.Package, "workflowobjectref")),
			}

			if err := e.generateHooks(ctx); err != nil {
				return err
			}

			if err := e.generateEnums(ctx); err != nil {
				return err
			}

			return nil
		})
	}
}

// Annotations satisfies the entc.Extension interface.
func (Extension) Annotations() []entc.Annotation { return nil }

// Options satisfies the entc.Extension interface.
func (Extension) Options() []entc.Option { return nil }

// Templates satisfies the entc.Extension interface.
func (Extension) Templates() []*gen.Template { return nil }

// templateContext holds data for rendering templates
type templateContext struct {
	// Graph is the ent code generation graph containing all schema information
	*gen.Graph
	// CreatableNodes is the list of nodes with generated CreateInput types
	CreatableNodes []*gen.Type
	// HooksPackageName is the package name for generated workflow hooks
	HooksPackageName string
	// EnumsPackageName is the package name for generated enum types
	EnumsPackageName string
	// EnumsImportPath is the import path for the enums package
	EnumsImportPath string
	// WorkflowsImportPath is the import path for the workflows runtime package
	WorkflowsImportPath string
	// GeneratedImportPath is the import path for the ent generated package
	GeneratedImportPath string
	// WorkflowObjectRefImport is the import path for the workflowobjectref subpackage
	WorkflowObjectRefImport string
}

// templateFile represents a template to be rendered and written to a file
type templateFile struct {
	// name is the template name used for parsing and execution
	name string
	// filename is the output filename for the generated file
	filename string
	// outputDir is the directory where the generated file will be written
	outputDir string
	// content is the raw template string content
	content string
}

// generateHooks generates the workflow registry hooks
func (e Extension) generateHooks(ctx templateContext) error {
	files := []templateFile{
		{
			name:      "workflow_registry",
			filename:  "workflow_registry.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowRegistryTemplate,
		},
		{
			name:      "workflow_edge_extractor",
			filename:  "workflow_edge_extractor.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowEdgeExtractorTemplate,
		},
		{
			name:      "workflow_domain",
			filename:  "workflow_domain.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowDomainTemplate,
		},
		{
			name:      "workflow_create_helpers",
			filename:  "workflow_create_helpers.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowCreateHelpersTemplate,
		},
	}

	return renderTemplates(files, ctx)
}

// generateEnums generates the workflow object type enums
func (e Extension) generateEnums(ctx templateContext) error {
	files := []templateFile{
		{
			name:      "workflow_object_type_enum",
			filename:  "workflow_object_type.go",
			outputDir: e.config.EnumsOutputDir,
			content:   workflowObjectTypeEnumTemplate,
		},
	}

	return renderTemplates(files, ctx)
}

// renderTemplates renders and writes multiple templates based on the provided context
func renderTemplates(files []templateFile, ctx templateContext) error {
	for _, file := range files {
		t, err := template.New(file.name).Funcs(gen.Funcs).Parse(file.content)
		if err != nil {
			return fmt.Errorf("parse %s template: %w", file.name, err)
		}

		if err := writeTemplate(file.outputDir, file.filename, file.name, t, ctx); err != nil {
			return err
		}
	}

	return nil
}

// writeTemplate renders and writes a template to the specified output directory and filename
func writeTemplate(outputDir, filename, templateName string, tmpl *template.Template, data any) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil { // nolint:mnd
		return fmt.Errorf("create output dir %s: %w", outputDir, err)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("execute %s template: %w", templateName, err)
	}

	outputPath := filepath.Join(outputDir, filename)

	formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("format %s: %w", outputPath, err)
	}

	if err := os.WriteFile(outputPath, formatted, 0o600); err != nil { // nolint:mnd
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

func buildCreatableNodes(g *gen.Graph) []*gen.Type {
	result := make([]*gen.Type, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		if checkSkipMutationCreateInput(node) {
			continue
		}

		result = append(result, node)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// checkSkipMutationCreateInput checks if the schema should be skipped for create helper generation.
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

const workflowRegistryTemplate = `{{/* Generate workflow object + CEL registry hooks for workflows */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_registry" }}
// Code generated by ent. DO NOT EDIT.
// This file is generated to keep workflow registries in sync with ent schemas.
package {{ .HooksPackageName }}

import (
	"context"
	"encoding/json"

	"{{ .GeneratedImportPath }}"
	"{{ .WorkflowObjectRefImport }}"
	wf "{{ .WorkflowsImportPath }}"
	"{{ .EnumsImportPath }}"
)

{{- $workflowTypes := dict }}
{{- range $n := $.Nodes }}
	{{- $hasWorkflowFields := false }}
	{{- $isHistory := false }}
	{{- if hasSuffix $n.Name "History" }}{{ $isHistory = true }}{{ end }}
	{{- range $f := $n.Fields }}
		{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
	{{- end }}
	{{- if and $hasWorkflowFields (not $isHistory) }}
		{{- $_ := set $workflowTypes $n.Name true }}
	{{- end }}
{{- end }}

func init() {
	// Register WorkflowObjectRef resolvers for workflow-addressable schemas.
	{{- range $n := $.Nodes }}
	{{- if eq $n.Name "WorkflowObjectRef" }}
	wf.RegisterObjectRefResolver(func(ref *generated.WorkflowObjectRef) (*wf.Object, bool) {
		{{- range $e := $n.Edges }}
			{{- if $e.Unique }}
				{{- $targetType := $e.Type.Name }}
				{{- if hasKey $workflowTypes $targetType }}
		if ref.{{ $e.StructField }}ID != "" {
			return &wf.Object{ID: ref.{{ $e.StructField }}ID, Type: enums.WorkflowObjectType{{ $targetType }}}, true
		}
				{{- end }}
			{{- end }}
		{{- end }}
		return nil, false
	})
	{{- end }}
	{{- end }}

	// Register WorkflowObjectRef query builders so object-based lookups avoid table scans.
	wf.RegisterObjectRefQueryBuilder(func(query *generated.WorkflowObjectRefQuery, obj *wf.Object) (*generated.WorkflowObjectRefQuery, bool) {
		if obj == nil {
			return nil, false
		}

		switch obj.Type {
		{{- range $n := $.Nodes }}
			{{- if hasKey $workflowTypes $n.Name }}
		case enums.WorkflowObjectType{{ $n.Name }}:
			return query.Where(workflowobjectref.{{ $n.Name }}IDEQ(obj.ID)), true
			{{- end }}
		{{- end }}
		default:
			return nil, false
		}
	})

	// Register CEL context builders so CEL expressions can work with typed objects.
	// Objects are converted to map[string]any via JSON to ensure:
	// - Field names match JSON tags (lowercase)
	// - Enum types are converted to strings
	{{- range $n := $.Nodes }}
		{{- $hasWorkflowFields := false }}
		{{- range $f := $n.Fields }}
			{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
		{{- end }}
		{{- if $hasWorkflowFields }}
	wf.RegisterCELContextBuilder(func(obj *wf.Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string, proposedChanges map[string]any) map[string]any {
		if obj == nil || obj.Node == nil {
			return nil
		}
		entObj, ok := obj.Node.(*generated.{{ $n.Name }})
		if !ok {
			return nil
		}
		objectMap, err := entObjectToMap(entObj)
		if err != nil {
			return nil
		}
		return map[string]any{
			"object":         objectMap,
			"changed_fields": changedFields,
			"changed_edges":  changedEdges,
			"added_ids":      addedIDs,
			"removed_ids":    removedIDs,
			"event_type":     eventType,
			"user_id":        userID,
			"proposed_changes": proposedChanges,
		}
	})
		{{- end }}
	{{- end }}

	// Register assignment context builder for workflow runtime state in CEL expressions.
	wf.RegisterAssignmentContextBuilder(buildAssignmentContext)

	// Register observability fields builder for structured logging.
	wf.RegisterObservabilityFieldsBuilder(buildObservabilityFields)

	// Register eligible fields from the generated workflow domain.
	wf.RegisterEligibleFields(WorkflowEligibleFields)
}

// buildObservabilityFields returns standard log fields for a workflow object.
func buildObservabilityFields(obj *wf.Object) map[string]any {
	if obj == nil {
		return nil
	}
	fields := map[string]any{
		"object_type": obj.Type.String(),
	}
	switch obj.Type {
	{{- range $n := $.Nodes }}
		{{- if hasKey $workflowTypes $n.Name }}
	case enums.WorkflowObjectType{{ $n.Name }}:
		fields[workflowobjectref.Field{{ $n.Name }}ID] = obj.ID
		{{- end }}
	{{- end }}
	}
	return fields
}

// entObjectToMap converts an ent entity to a map[string]any via JSON marshaling.
// This ensures field names match JSON tags (lowercase) and enums are converted to strings.
func entObjectToMap(obj any) (map[string]any, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// buildAssignmentContext builds the workflow runtime context (assignments, instance, initiator) for CEL evaluation.
// This is called when evaluating NOTIFY action When expressions that depend on assignment state.
func buildAssignmentContext(ctx context.Context, client *generated.Client, instanceID string) (map[string]any, error) {
	if client == nil || instanceID == "" {
		return nil, nil
	}

	// Build assignment summary
	summary, err := client.BuildAssignmentSummary(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Convert summary to map for CEL dynamic access
	// CEL cannot traverse Go structs directly - it needs map[string]any
	summaryMap, err := entObjectToMap(summary)
	if err != nil {
		return nil, err
	}

	// Get instance for context
	instance, err := client.WorkflowInstance.Get(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Build instance context map
	instanceContext := map[string]any{
		"id":                   instance.ID,
		"state":                instance.State.String(),
		"current_action_index": instance.CurrentActionIndex,
	}

	// Extract initiator from instance context
	initiator := ""
	if instance.Context.TriggerUserID != "" {
		initiator = instance.Context.TriggerUserID
	}

	return map[string]any{
		"assignments": summaryMap,
		"instance":    instanceContext,
		"initiator":   initiator,
	}, nil
}

{{ end }}
`

const workflowEdgeExtractorTemplate = `{{/* Generate edge mutation detection for workflow triggers */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_edge_extractor" }}

{{ $pkg := base $.Config.Package }}
// Code generated by ent. DO NOT EDIT.
// This should be above the package name declaration for go tooling to treat this as generated.
// This file is generated to detect edge mutations for workflow triggers.
package {{ .HooksPackageName }}

import (
	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
)

var WorkflowEligibleEdges = map[string][]string{
{{- range $n := $.Nodes }}
	{{- $isHistory := hasSuffix $n.Name "History" }}
	{{- if not $isHistory }}
		{{- $hasAnyWorkflowEdge := false }}
		{{- range $e := $n.Edges }}
			{{- if $e.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}
				{{- $hasAnyWorkflowEdge = true }}
			{{- end }}
		{{- end }}
		{{- if $hasAnyWorkflowEdge }}
	generated.Type{{ $n.Name }}: edgeList(
			{{- range $e := $n.Edges }}
				{{- if $e.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}
		"{{ $e.Name }}",
				{{- end }}
			{{- end }}
	),
		{{- end }}
	{{- end }}
{{- end }}
}

// ExtractChangedEdges inspects the mutation to determine which edge relationships were modified.
// It returns: edge names, added IDs per edge, removed IDs per edge.
func ExtractChangedEdges(m ent.Mutation) ([]string, map[string][]string, map[string][]string) {
	var edgeNames []string
	addedIDs := make(map[string][]string)
	removedIDs := make(map[string][]string)

	eligibleEdges := WorkflowEligibleEdges[m.Type()]
	if len(eligibleEdges) == 0 {
		return edgeNames, addedIDs, removedIDs
	}

	changedEdges := make(map[string]struct{})
	for _, edge := range m.AddedEdges() {
		changedEdges[edge] = struct{}{}
	}
	for _, edge := range m.RemovedEdges() {
		changedEdges[edge] = struct{}{}
	}
	for _, edge := range m.ClearedEdges() {
		changedEdges[edge] = struct{}{}
	}

	for _, edge := range eligibleEdges {
		if _, ok := changedEdges[edge]; !ok {
			continue
		}

		edgeNames = append(edgeNames, edge)

		if ids := toStringIDs(m.AddedIDs(edge)); len(ids) > 0 {
			addedIDs[edge] = ids
		}
		if ids := toStringIDs(m.RemovedIDs(edge)); len(ids) > 0 {
			removedIDs[edge] = ids
		}
		if m.EdgeCleared(edge) {
			removedIDs[edge] = []string{}
		}
	}

	return edgeNames, addedIDs, removedIDs
}

func edgeList(edges ...string) []string {
	return edges
}

func toStringIDs(values []ent.Value) []string {
	if len(values) == 0 {
		return nil
	}
	ids := make([]string, 0, len(values))
	for _, value := range values {
		id, ok := value.(string)
		if !ok {
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

{{ end }}
`

const workflowObjectTypeEnumTemplate = `{{/* Generate WorkflowObjectType enum values based on entities with WorkflowApprovalMixin */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_object_type_enum" }}
// Code generated by ent. DO NOT EDIT.
// This file is auto-generated to ensure WorkflowObjectType enum matches entities with workflow support.
package {{ .EnumsPackageName }}

import (
	"fmt"
	"io"
)

// WorkflowObjectType enumerates supported object types for workflows.
// These types are automatically generated based on entities that have the WorkflowApprovalMixin.
type WorkflowObjectType string

var (
{{- range $n := $.Nodes }}
	{{- $hasWorkflowFields := false }}
	{{- $isHistory := false }}
	{{- if hasSuffix $n.Name "History" }}{{ $isHistory = true }}{{ end }}
	{{- range $f := $n.Fields }}
		{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
	{{- end }}
	{{- if and $hasWorkflowFields (not $isHistory) }}
	WorkflowObjectType{{ $n.Name }} WorkflowObjectType = "{{ $n.Name }}"
	{{- end }}
{{- end }}
)

var WorkflowObjectTypes = []string{
{{- range $n := $.Nodes }}
	{{- $hasWorkflowFields := false }}
	{{- $isHistory := false }}
	{{- if hasSuffix $n.Name "History" }}{{ $isHistory = true }}{{ end }}
	{{- range $f := $n.Fields }}
		{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
	{{- end }}
	{{- if and $hasWorkflowFields (not $isHistory) }}
	string(WorkflowObjectType{{ $n.Name }}),
	{{- end }}
{{- end }}
}

// Values returns all workflow object type values.
func (WorkflowObjectType) Values() (vals []string) {
	return WorkflowObjectTypes
}

// String returns the string representation of the workflow object type.
func (r WorkflowObjectType) String() string { return string(r) }

// ToWorkflowObjectType converts a string to a WorkflowObjectType pointer.
func ToWorkflowObjectType(v string) *WorkflowObjectType {
	switch v {
	{{- range $n := $.Nodes }}
		{{- $hasWorkflowFields := false }}
		{{- $isHistory := false }}
		{{- if hasSuffix $n.Name "History" }}{{ $isHistory = true }}{{ end }}
		{{- range $f := $n.Fields }}
			{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
		{{- end }}
		{{- if and $hasWorkflowFields (not $isHistory) }}
	case WorkflowObjectType{{ $n.Name }}.String():
		return &WorkflowObjectType{{ $n.Name }}
		{{- end }}
	{{- end }}
	default:
		return nil
	}
}

// MarshalGQL marshals the workflow object type to GraphQL.
func (r WorkflowObjectType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte("\"" + r.String() + "\""))
}

// UnmarshalGQL unmarshals the workflow object type from GraphQL.
func (r *WorkflowObjectType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeWorkflowObjectType, v)
	}
	*r = WorkflowObjectType(str)
	return nil
}

{{ end }}
`

const workflowDomainTemplate = `{{/* Generate workflow domain types and validation for approval scopes */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_domain" }}
// Code generated by ent. DO NOT EDIT.
// This file is generated to provide type-safe workflow domains for approval scopes.
package {{ .HooksPackageName }}

import (
	"fmt"
	"sort"
	"strings"

	"{{ .EnumsImportPath }}"
)

// WorkflowDomain represents a canonical approval domain for workflow proposals.
// A domain is the combination of an object type and the fields requiring approval.
type WorkflowDomain struct {
	// ObjectType is the workflow object type this domain applies to
	ObjectType enums.WorkflowObjectType
	// Fields is the sorted list of field names in this domain
	Fields []string
}

// Key returns the canonical domain key including object type prefix.
// Format: "ObjectType:field1,field2,field3" (fields are sorted).
func (d WorkflowDomain) Key() string {
	if len(d.Fields) == 0 {
		return string(d.ObjectType)
	}
	fields := make([]string, len(d.Fields))
	copy(fields, d.Fields)
	sort.Strings(fields)
	return string(d.ObjectType) + ":" + strings.Join(fields, ",")
}

// WorkflowEligibleFields maps object types to their workflow-eligible field names.
// Use this to validate that fields in a domain are valid for the object type.
var WorkflowEligibleFields = map[enums.WorkflowObjectType]map[string]struct{}{
{{- range $n := $.Nodes }}
	{{- $isHistory := false }}
	{{- if hasSuffix $n.Name "History" }}{{ $isHistory = true }}{{ end }}
	{{- if not $isHistory }}
		{{- $eligibleFields := list }}
		{{- range $f := $n.Fields }}
			{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}
				{{- $eligibleFields = append $eligibleFields $f.Name }}
			{{- end }}
		{{- end }}
		{{- if $eligibleFields }}
	enums.WorkflowObjectType{{ $n.Name }}: {
			{{- range $field := $eligibleFields }}
		"{{ $field }}": {},
			{{- end }}
	},
		{{- end }}
	{{- end }}
{{- end }}
}

// ErrInvalidObjectType is returned when an object type is not workflow-eligible
var ErrInvalidObjectType = fmt.Errorf("object type is not workflow-eligible")

// ErrInvalidDomainField is returned when a field is not workflow-eligible for the object type
var ErrInvalidDomainField = fmt.Errorf("field is not workflow-eligible for object type")

// ErrEmptyDomainFields is returned when no fields are provided for a domain
var ErrEmptyDomainFields = fmt.Errorf("domain requires at least one field")

// NewWorkflowDomain creates a validated domain for an object type and fields.
// Fields are automatically sorted to ensure canonical ordering.
// Returns an error if the object type or any field is not workflow-eligible.
func NewWorkflowDomain(objectType enums.WorkflowObjectType, fields []string) (WorkflowDomain, error) {
	if len(fields) == 0 {
		return WorkflowDomain{}, ErrEmptyDomainFields
	}

	eligible, ok := WorkflowEligibleFields[objectType]
	if !ok {
		return WorkflowDomain{}, fmt.Errorf("%w: %s", ErrInvalidObjectType, objectType)
	}

	sorted := make([]string, len(fields))
	copy(sorted, fields)
	sort.Strings(sorted)

	for _, field := range sorted {
		if _, ok := eligible[field]; !ok {
			return WorkflowDomain{}, fmt.Errorf("%w: %s.%s", ErrInvalidDomainField, objectType, field)
		}
	}

	return WorkflowDomain{
		ObjectType: objectType,
		Fields:     sorted,
	}, nil
}

{{ end }}
`
const workflowCreateHelpersTemplate = `{{/* Generate workflow create helpers for CREATE_OBJECT actions */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_create_helpers" }}
// Code generated by ent. DO NOT EDIT.
// This file is generated to keep workflow create helpers in sync with ent schemas.
package {{ .HooksPackageName }}

import (
	"context"
	"errors"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"{{ .GeneratedImportPath }}"
)

var (
	ErrNilClient               = errors.New("ent client is required")
	ErrCreateObjectTypeInvalid = errors.New("create object type is invalid")
)

type createObjectEntry struct {
	create func(ctx context.Context, client *generated.Client, fields map[string]any) error
}

var creatableSchemaTypes = map[string]createObjectEntry{
{{- range $n := .CreatableNodes }}
	NormalizeSchemaType("{{ $n.Name }}"): makeCreateEntry[generated.Create{{ $n.Name }}Input](func(client *generated.Client) *generated.{{ $n.Name }}Create { return client.{{ $n.Name }}.Create() }),
{{- end }}
}

// NormalizeSchemaType returns a normalized schema identifier for matching.
func NormalizeSchemaType(value string) string {
	if value == "" {
		return ""
	}

	normalized := strings.ToLower(value)
	normalized = strings.ReplaceAll(normalized, "_", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	normalized = strings.ReplaceAll(normalized, " ", "")

	return normalized
}

// IsCreatableSchemaType reports whether the schema type maps to a creatable ent client.
func IsCreatableSchemaType(schemaType string) bool {
	if schemaType == "" {
		return false
	}

	_, ok := creatableSchemaTypes[NormalizeSchemaType(schemaType)]
	return ok
}

// CreateObject builds and executes a create mutation for the provided schema type.
func CreateObject(ctx context.Context, client *generated.Client, schemaType string, fields map[string]any) error {
	if client == nil {
		return ErrNilClient
	}

	entry, ok := creatableSchemaTypes[NormalizeSchemaType(schemaType)]
	if !ok {
		return ErrCreateObjectTypeInvalid
	}

	return entry.create(ctx, client, fields)
}

func makeCreateEntry[I any, B interface {
	SetInput(I) B
	Exec(context.Context) error
}](builder func(client *generated.Client) B) createObjectEntry {
	return createObjectEntry{
		create: func(ctx context.Context, client *generated.Client, fields map[string]any) error {
			input, err := decodeCreateInput[I](fields)
			if err != nil {
				return err
			}

			return builder(client).SetInput(input).Exec(ctx)
		},
	}
}

func decodeCreateInput[I any](fields map[string]any) (I, error) {
	var input I
	if len(fields) == 0 {
		return input, nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &input,
		WeaklyTypedInput: true,
		TagName:          "json",
		MatchName: func(mapKey, fieldName string) bool {
			return NormalizeSchemaType(mapKey) == NormalizeSchemaType(fieldName)
		},
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
		),
	})
	if err != nil {
		return input, err
	}

	if err := decoder.Decode(fields); err != nil {
		return input, err
	}

	return input, nil
}

{{ end }}
`
