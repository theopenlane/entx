package oscalgen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"golang.org/x/tools/imports"
)

const oscalGeneratedTemplatePath = "templates/oscal_generated.tmpl"

// oscalTemplates stores embedded generator templates
//
//go:embed templates/*.tmpl
var oscalTemplates embed.FS

// OSCALGenerator handles generation of OSCAL mapping registry helpers
// This generator scans ent schemas for OSCALModel, OSCALField, and OSCALRelationship
// annotations and produces a deterministic registry used by downstream import/export logic
type OSCALGenerator struct {
	// SchemaPath is the path to the schema directory
	SchemaPath string
	// OutputDir is the output directory for the generated file
	OutputDir string
	// Package is the package name for generated code
	Package string
}

// NewOSCALGenerator creates a new OSCALGenerator with default settings
func NewOSCALGenerator(schemaPath, outputDir string) *OSCALGenerator {
	return &OSCALGenerator{
		SchemaPath: schemaPath,
		OutputDir:  outputDir,
		Package:    "oscalgenerated",
	}
}

// WithPackage sets the package name for generated code
func (o *OSCALGenerator) WithPackage(pkg string) *OSCALGenerator {
	o.Package = pkg
	return o
}

// Generate scans schema annotations and generates OSCAL mapping helper code
func (o *OSCALGenerator) Generate(flags ...string) error {
	graph, err := entc.LoadGraph(o.SchemaPath, &gen.Config{
		BuildFlags: flags,
	})
	if err != nil {
		return fmt.Errorf("loading graph: %w", err)
	}

	schemas := o.collectOSCALSchemas(graph)

	if err := o.generateRegistryFile(schemas); err != nil {
		return err
	}

	return nil
}

// collectOSCALSchemas collects OSCAL mapping metadata from the ent graph
func (o *OSCALGenerator) collectOSCALSchemas(graph *gen.Graph) []oscalSchemaInfo {
	schemas := make([]oscalSchemaInfo, 0, len(graph.Schemas))

	for _, schema := range graph.Schemas {
		modelAnn := &OSCALModel{}
		hasSchemaAnn := false

		if raw, ok := schema.Annotations[modelAnn.Name()]; ok {
			if err := modelAnn.Decode(raw); err == nil {
				hasSchemaAnn = true
			}
		}

		fields := collectOSCALFieldMappings(schema.Fields)
		relationships := collectOSCALRelationshipMappings(schema.Edges)

		if !hasSchemaAnn && len(fields) == 0 && len(relationships) == 0 {
			continue
		}

		schemas = append(schemas, oscalSchemaInfo{
			name:          schema.Name,
			models:        modelTypesToStrings(modelAnn.Models),
			assembly:      modelAnn.Assembly,
			fields:        fields,
			relationships: relationships,
		})
	}

	sort.SliceStable(schemas, func(i, j int) bool {
		return schemas[i].name < schemas[j].name
	})

	return schemas
}

// collectOSCALFieldMappings collects field-level OSCAL mapping metadata
func collectOSCALFieldMappings(fields []*load.Field) []oscalFieldInfo {
	results := make([]oscalFieldInfo, 0, len(fields))

	for _, field := range fields {
		ann := &OSCALField{}
		raw, ok := field.Annotations[ann.Name()]
		if !ok {
			continue
		}

		if err := ann.Decode(raw); err != nil {
			continue
		}

		results = append(results, oscalFieldInfo{
			name:           field.Name,
			role:           string(ann.Role),
			models:         modelTypesToStrings(ann.Models),
			identityAnchor: ann.IdentityAnchor,
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].name < results[j].name
	})

	return results
}

// collectOSCALRelationshipMappings collects edge-level OSCAL mapping metadata
func collectOSCALRelationshipMappings(edges []*load.Edge) []oscalRelationshipInfo {
	results := make([]oscalRelationshipInfo, 0, len(edges))

	for _, edge := range edges {
		ann := &OSCALRelationship{}
		raw, ok := edge.Annotations[ann.Name()]
		if !ok {
			continue
		}

		if err := ann.Decode(raw); err != nil {
			continue
		}

		results = append(results, oscalRelationshipInfo{
			name:   edge.Name,
			role:   string(ann.Role),
			models: modelTypesToStrings(ann.Models),
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		return results[i].name < results[j].name
	})

	return results
}

// modelTypesToStrings normalizes OSCAL model values into sorted unique strings
func modelTypesToStrings(models []OSCALModelType) []string {
	seen := make(map[string]struct{}, len(models))
	out := make([]string, 0, len(models))

	for _, model := range models {
		name := string(model)
		if name == "" {
			continue
		}

		if _, ok := seen[name]; ok {
			continue
		}

		seen[name] = struct{}{}
		out = append(out, name)
	}

	sort.Strings(out)

	return out
}

// buildOSCALTemplateData converts internal mapping metadata into template-friendly structs
func buildOSCALTemplateData(pkg string, schemas []oscalSchemaInfo) oscalTemplateData {
	data := oscalTemplateData{
		Package: pkg,
		Schemas: make([]oscalTemplateSchema, 0, len(schemas)),
	}

	for _, schema := range schemas {
		templateSchema := oscalTemplateSchema{
			Name:          schema.name,
			Models:        schema.models,
			Assembly:      schema.assembly,
			Fields:        make([]oscalTemplateField, 0, len(schema.fields)),
			Relationships: make([]oscalTemplateRelationship, 0, len(schema.relationships)),
		}

		for _, field := range schema.fields {
			templateSchema.Fields = append(templateSchema.Fields, oscalTemplateField{
				Name:           field.name,
				Role:           field.role,
				Models:         field.models,
				IdentityAnchor: field.identityAnchor,
			})
		}

		for _, relationship := range schema.relationships {
			templateSchema.Relationships = append(templateSchema.Relationships, oscalTemplateRelationship{
				Name:   relationship.name,
				Role:   relationship.role,
				Models: relationship.models,
			})
		}

		data.Schemas = append(data.Schemas, templateSchema)
	}

	return data
}

// generateRegistryFile writes the generated OSCAL mapping helper file
func (o *OSCALGenerator) generateRegistryFile(schemas []oscalSchemaInfo) error {
	tmpl, err := template.New("oscal_generated").Funcs(template.FuncMap{"lower": strings.ToLower}).ParseFS(oscalTemplates, oscalGeneratedTemplatePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(o.OutputDir, 0o755); err != nil { //nolint:mnd
		return err
	}

	filePath := filepath.Join(o.OutputDir, "oscal_generated.go")

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	data := buildOSCALTemplateData(o.Package, schemas)

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "oscal_generated.tmpl", data); err != nil {
		return err
	}

	formatted, err := imports.Process(filePath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("%w: failed to format file", err)
	}

	if _, err := file.Write(formatted); err != nil {
		return err
	}

	return nil
}
