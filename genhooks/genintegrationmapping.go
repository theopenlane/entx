package genhooks

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/entx"
)

var integrationSystemFieldNames = map[string]struct{}{
	"id":              {},
	"owner_id":        {},
	"organization_id": {},
	"org_id":          {},
	"created_at":      {},
	"updated_at":      {},
	"created_by":      {},
	"updated_by":      {},
	"deleted_at":      {},
	"deleted_by":      {},
	"workflow_eligible_marker": {},
}

// IntegrationMappingConfig holds configuration options for integration mapping generation.
type IntegrationMappingConfig struct {
	outputDir   string
	packageName string
}

// IntegrationMappingOption adds functional params for IntegrationMappingConfig.
type IntegrationMappingOption func(*IntegrationMappingConfig)

// WithIntegrationMappingOutputDir sets the directory to output the generated mapping files.
func WithIntegrationMappingOutputDir(dir string) IntegrationMappingOption {
	return func(c *IntegrationMappingConfig) {
		c.outputDir = dir
	}
}

// WithIntegrationMappingPackageName sets the Go package name for generated mapping files.
func WithIntegrationMappingPackageName(name string) IntegrationMappingOption {
	return func(c *IntegrationMappingConfig) {
		c.packageName = name
	}
}

// IntegrationMappingData holds the data for generating integration mapping files.
type IntegrationMappingData struct {
	// PackageName is the Go package name for generated files.
	PackageName string
	// Schemas contains all schemas with integration mapping fields.
	Schemas []IntegrationMappingSchema
}

// IntegrationMappingSchema represents a schema with integration mapping fields.
type IntegrationMappingSchema struct {
	// Name is the schema name (e.g., Vulnerability).
	Name string
	// Fields contains integration mapping field metadata.
	Fields []IntegrationMappingField
}

// IntegrationMappingField represents a field that can be targeted by integration mappings.
type IntegrationMappingField struct {
	// InputKey is the GraphQL input field name (lowerCamel).
	InputKey string
	// EntField is the ent field name (snake_case).
	EntField string
	// Type is the ent field type (string, time, json, etc).
	Type string
	// Required indicates the field is required for input (non-optional in ent).
	Required bool
	// UpsertKey indicates the field participates in dedupe/upsert matching.
	UpsertKey bool
}

// GenIntegrationMappingSchema generates mapping metadata for annotated fields.
func GenIntegrationMappingSchema(opts ...IntegrationMappingOption) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			c := &IntegrationMappingConfig{
				packageName: "integrationgenerated",
			}

			for _, opt := range opts {
				opt(c)
			}

			if c.outputDir == "" {
				return next.Generate(g)
			}

			data := getIntegrationMappingData(g, c)
			if len(data.Schemas) == 0 {
				return next.Generate(g)
			}

			if err := generateIntegrationMappingFile(c.outputDir, data); err != nil {
				return err
			}

			return next.Generate(g)
		})
	}
}

// getIntegrationMappingData collects mapping data from all schemas.
func getIntegrationMappingData(g *gen.Graph, c *IntegrationMappingConfig) IntegrationMappingData {
	data := IntegrationMappingData{
		PackageName: c.packageName,
		Schemas:     []IntegrationMappingSchema{},
	}

	for _, node := range g.Nodes {
		if checkSchemaGenSkip(node) || checkQueryGenSkip(node) {
			continue
		}

		if strings.HasSuffix(node.Name, "History") {
			continue
		}

		hasCreate := !checkSkipMutationCreateInput(node)
		hasUpdate := !checkSkipMutationUpdateInput(node)
		if !hasCreate && !hasUpdate {
			continue
		}

		schema := getEntSchema(g, node.Name)
		if schema == nil {
			continue
		}

		fields := getIntegrationMappingFields(schema)
		if len(fields) == 0 {
			continue
		}

		data.Schemas = append(data.Schemas, IntegrationMappingSchema{
			Name:   node.Name,
			Fields: fields,
		})
	}

	// sort schemas for consistent output
	slices.SortFunc(data.Schemas, func(a, b IntegrationMappingSchema) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return data
}

// getIntegrationMappingFields returns integration mapping fields for a schema.
func getIntegrationMappingFields(schema *load.Schema) []IntegrationMappingField {
	if schema == nil {
		return nil
	}

	fields := make([]IntegrationMappingField, 0)
	schemaAnt := getIntegrationMappingSchemaAnnotation(schema)

	includeSet := make(map[string]struct{})
	excludeSet := make(map[string]struct{})
	upsertSet := make(map[string]struct{})
	hasInclude := false
	if schemaAnt != nil {
		for _, name := range schemaAnt.Include {
			includeSet[name] = struct{}{}
		}
		hasInclude = len(includeSet) > 0
		for _, name := range schemaAnt.Exclude {
			excludeSet[name] = struct{}{}
		}
		for _, name := range schemaAnt.UpsertKeys {
			upsertSet[name] = struct{}{}
		}
	}

	for _, field := range schema.Fields {
		if hasInclude {
			if _, ok := includeSet[field.Name]; !ok {
				continue
			}
		}
		if _, ok := excludeSet[field.Name]; ok {
			continue
		}

		if !isFieldIntegrationMappingEligible(field) {
			continue
		}

		// when using schema-level annotation without explicit includes, apply default exclusions
		if !hasInclude && isIntegrationSystemField(field.Name) {
			continue
		}

		ant := getIntegrationMappingAnnotation(field)
		if schemaAnt == nil && ant == nil {
			continue
		}

		key := ""
		if ant != nil {
			key = ant.Key
		}
		if key == "" {
			goName := templates.ToGo(field.Name)
			key = templates.ToGoPrivate(goName)
		}

		upsert := false
		if ant != nil && ant.UpsertKey {
			upsert = true
		}
		if _, ok := upsertSet[field.Name]; ok {
			upsert = true
		}

		fields = append(fields, IntegrationMappingField{
			InputKey:    key,
			EntField:    field.Name,
			Type:        field.Info.Type.String(),
			Required:    !field.Optional,
			UpsertKey:   upsert,
		})
	}

	slices.SortFunc(fields, func(a, b IntegrationMappingField) int {
		return cmp.Compare(a.InputKey, b.InputKey)
	})

	return fields
}

// getIntegrationMappingAnnotation retrieves the IntegrationMappingField annotation from a field.
func getIntegrationMappingAnnotation(field *load.Field) *entx.IntegrationMappingFieldAnnotation {
	ant := &entx.IntegrationMappingFieldAnnotation{}
	if raw, ok := field.Annotations[ant.Name()]; ok {
		if err := ant.Decode(raw); err != nil {
			return nil
		}

		return ant
	}

	return nil
}

// getIntegrationMappingSchemaAnnotation retrieves the IntegrationMappingSchema annotation from a schema.
func getIntegrationMappingSchemaAnnotation(schema *load.Schema) *entx.IntegrationMappingSchemaAnnotation {
	ant := &entx.IntegrationMappingSchemaAnnotation{}
	if raw, ok := schema.Annotations[ant.Name()]; ok {
		if err := ant.Decode(raw); err != nil {
			return nil
		}

		return ant
	}

	return nil
}

// isFieldIntegrationMappingEligible checks if a field is eligible for integration mapping.
func isFieldIntegrationMappingEligible(field *load.Field) bool {
	// exclude sensitive fields
	if field.Sensitive {
		return false
	}

	// exclude fields that are explicitly skipped from GraphQL types
	if entSkipType(field) {
		return false
	}

	// exclude fields skipped from both create and update inputs
	if entSkipMutationInputs(field) {
		return false
	}

	return true
}

// entSkipType checks if the field has entgql.SkipType or SkipAll.
func entSkipType(field *load.Field) bool {
	entAnt := &entgql.Annotation{}
	if ant, ok := field.Annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			switch {
			case entAnt.Skip.Is(entgql.SkipType):
				return true
			case entAnt.Skip.Is(entgql.SkipAll):
				return true
			}
		}
	}

	return false
}

// entSkipMutationInputs checks if the field is skipped from both create and update inputs.
func entSkipMutationInputs(field *load.Field) bool {
	entAnt := &entgql.Annotation{}
	if ant, ok := field.Annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			if entAnt.Skip.Is(entgql.SkipMutationCreateInput) && entAnt.Skip.Is(entgql.SkipMutationUpdateInput) {
				return true
			}
		}
	}

	return false
}

func isIntegrationSystemField(name string) bool {
	_, ok := integrationSystemFieldNames[name]
	return ok
}

// generateIntegrationMappingFile creates the integration mapping Go file.
func generateIntegrationMappingFile(outputDir string, data IntegrationMappingData) error {
	tmpl := createIntegrationMappingTemplate()

	if err := os.MkdirAll(outputDir, dirPermissions); err != nil {
		log.Error().Err(err).Str("path", outputDir).Msg("failed to create integration mapping output directory")
		return err
	}

	filePath := filepath.Join(outputDir, "integration_mapping_generated.go")
	file, err := os.Create(filepath.Clean(filePath))
	if err != nil {
		log.Error().Err(err).Str("path", filePath).Msg("failed to create integration mapping file")
		return err
	}

	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Error().Err(err).Msg("failed to execute integration mapping template")
		return err
	}

	log.Debug().Str("path", filePath).Msg("generated integration mapping file")

	return nil
}

// createIntegrationMappingTemplate creates the template for integration mapping generation.
func createIntegrationMappingTemplate() *template.Template {
	const mappingTemplate = `// Code generated by entx integration mapping generator. DO NOT EDIT.
package {{ .PackageName }}

// IntegrationMappingField describes an integration mapping target field.
type IntegrationMappingField struct {
	InputKey    string
	EntField    string
	Type        string
	Required    bool
	UpsertKey   bool
}

// IntegrationMappingSchema describes a schema with integration mapping fields.
type IntegrationMappingSchema struct {
	Name   string
	Fields []IntegrationMappingField
}

// IntegrationMappingSchemas maps schema names to their mapping metadata.
var IntegrationMappingSchemas = map[string]IntegrationMappingSchema{
{{- range .Schemas }}
	"{{ .Name }}": {
		Name: "{{ .Name }}",
		Fields: []IntegrationMappingField{
		{{- range .Fields }}
			{
				InputKey: {{ printf "%q" .InputKey }},
				EntField: {{ printf "%q" .EntField }},
				Type: {{ printf "%q" .Type }},
				Required: {{ .Required }},
				UpsertKey: {{ .UpsertKey }},
			},
		{{- end }}
		},
	},
{{- end }}
}
`

	tmpl, err := template.New("integration_mapping").Parse(mappingTemplate)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse integration mapping template")
	}

	return tmpl
}
