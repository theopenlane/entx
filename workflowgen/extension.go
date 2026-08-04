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

	"github.com/theopenlane/entx"
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

// Hook generates workflow domain helpers after ent codegen runs
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

			return e.generateHooks(ctx)
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

// generateHooks generates the workflow domain helpers; the object registry, edge extractor,
// create helpers, and object-type enum are owned by the entityops generator's unified catalog
func (e Extension) generateHooks(ctx templateContext) error {
	files := []templateFile{
		{
			name:      "workflow_domain",
			filename:  "workflow_domain.go",
			outputDir: e.config.HooksOutputDir,
			content:   workflowDomainTemplate,
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
	entgqlAnt, ok := entx.GetAnnotation[*entgql.Annotation](node)
	if !ok {
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
