package workflowgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"golang.org/x/tools/imports"
)

// ExtensionOption is a function that modifies the Extension configuration
type ExtensionOption func(*Extension)

// Config is the configuration for workflow code generation
type Config struct {
	HooksOutputDir      string
	HooksPackageName    string
	EnumsOutputDir      string
	EnumsPackageName    string
	EnumsImportPath     string
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
				HooksPackageName:        e.config.HooksPackageName,
				EnumsPackageName:        e.config.EnumsPackageName,
				EnumsImportPath:         e.config.EnumsImportPath,
				WorkflowsImportPath:     e.config.WorkflowsImportPath,
				GeneratedImportPath:     g.Config.Package,
				WorkflowObjectRefImport: filepath.ToSlash(filepath.Join(g.Config.Package, "workflowobjectref")),
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
	*gen.Graph
	HooksPackageName        string
	EnumsPackageName        string
	EnumsImportPath         string
	WorkflowsImportPath     string
	GeneratedImportPath     string
	WorkflowObjectRefImport string
}

// templateFile represents a template to be rendered and written to a file
type templateFile struct {
	name      string
	filename  string
	outputDir string
	content   string
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
			name:      "workflow_registry_test",
			filename:  "workflow_registry_test.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowRegistryTestTemplate,
		},
		{
			name:      "workflow_edge_extractor",
			filename:  "workflow_edge_extractor.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowEdgeExtractorTemplate,
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
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
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

	if err := os.WriteFile(outputPath, formatted, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outputPath, err)
	}

	return nil
}

const workflowRegistryTemplate = `{{/* Generate workflow object + CEL registry hooks for workflows */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_registry" }}
// Code generated by ent. DO NOT EDIT.
// This file is generated to keep workflow registries in sync with ent schemas.
package {{ .HooksPackageName }}

import (
	"context"

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
	{{- range $n := $.Nodes }}
		{{- $hasWorkflowFields := false }}
		{{- range $f := $n.Fields }}
			{{- if $f.Annotations.OPENLANE_WORKFLOW_ELIGIBLE }}{{ $hasWorkflowFields = true }}{{ end }}
		{{- end }}
		{{- if $hasWorkflowFields }}
	wf.RegisterCELContextBuilder(func(obj *wf.Object, changedFields []string, changedEdges []string, addedIDs, removedIDs map[string][]string, eventType, userID string) map[string]any {
		if obj == nil || obj.Node == nil {
			return nil
		}
		entObj, ok := obj.Node.(*generated.{{ $n.Name }})
		if !ok {
			return nil
		}
		return map[string]any{
			"object":         entObj,
			"changed_fields": changedFields,
			"changed_edges":  changedEdges,
			"added_ids":      addedIDs,
			"removed_ids":    removedIDs,
			"event_type":     eventType,
			"user_id":        userID,
		}
	})
		{{- end }}
	{{- end }}

	// Register assignment context builder for workflow runtime state in CEL expressions.
	wf.RegisterAssignmentContextBuilder(buildAssignmentContext)
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
		"assignments": summary,
		"instance":    instanceContext,
		"initiator":   initiator,
	}, nil
}

{{ end }}
`

const workflowRegistryTestTemplate = `{{/* minimal compile check for generated workflow registry */}}
{{/* gotype: entgo.io/ent/entc/gen.Graph */}}

{{ define "workflow_registry_test" }}
// Code generated by ent. DO NOT EDIT.
// Sanity compile check for workflow registry generation.
package {{ .HooksPackageName }}

import (
	"{{ .GeneratedImportPath }}"
)

func _workflowRegistryCompileCheck() {
	_ = generated.WorkflowObjectRef{}
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

var workflowEligibleEdges = map[string][]string{
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

// extractChangedEdges inspects the mutation to determine which edge relationships were modified.
// It returns: edge names, added IDs per edge, removed IDs per edge.
func extractChangedEdges(m ent.Mutation) ([]string, map[string][]string, map[string][]string) {
	var edgeNames []string
	addedIDs := make(map[string][]string)
	removedIDs := make(map[string][]string)

	eligibleEdges := workflowEligibleEdges[m.Type()]
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
