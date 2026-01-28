package genhooks

import (
	"html/template"
	"log"
	"os"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/gertd/go-pluralize"
)

// bulkSchema holds data for the bulk mutations GraphQL schema template
type bulkSchema struct {
	Name       string
	PluralName string
}

// BulkSchemaConfig holds configuration options for bulk schema generation
type BulkSchemaConfig struct {
	// injectIntoExisting controls whether to inject bulk mutations into existing schema files.
	// When true (default), the hook will modify existing schemas that have at least one bulk
	// update/delete mutation to add any missing ones. When false, only new schemas created
	// by GenSchema will have bulk mutations (via the template). This allows running the
	// injection once to migrate existing schemas, then disabling it for subsequent runs.
	injectIntoExisting bool
}

// BulkSchemaOption is a functional option for configuring BulkSchemaConfig
type BulkSchemaOption func(*BulkSchemaConfig)

// WithBulkSchemaInjectExisting controls whether to inject bulk mutations into existing schemas.
// When enabled (default), existing schemas with at least one bulk update/delete mutation will
// have missing mutations injected. When disabled, only new schemas will get bulk mutations
// via the template. Use this to run injection once, then disable for subsequent code generation.
func WithBulkSchemaInjectExisting(enabled bool) BulkSchemaOption {
	return func(c *BulkSchemaConfig) {
		c.injectIntoExisting = enabled
	}
}

// GenBulkSchema generates GraphQL schema extensions for bulk update and delete mutations.
// This injects updateBulk, updateBulkCSV, and deleteBulk mutations into existing schemas
// that already have createBulkCSV mutations AND at least one existing bulk update/delete mutation.
// If a schema has createBulkCSV but no bulk update/delete mutations, it is assumed to be intentional.
// Use WithBulkSchemaInjectExisting(false) to disable injection into existing schemas after initial migration.
func GenBulkSchema(graphSchemaDir string, opts ...BulkSchemaOption) gen.Hook {
	cfg := &BulkSchemaConfig{
		injectIntoExisting: true, // default to enabled for backwards compatibility
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// Skip all processing if injection into existing schemas is disabled
			if !cfg.injectIntoExisting {
				return next.Generate(g)
			}

			pluralizer := pluralize.NewClient()

			for _, node := range g.Nodes {
				// Skip if schema generation is disabled for this node
				if checkSchemaGenSkip(node) {
					continue
				}

				filePath := getFileName(graphSchemaDir, node.Name)

				// Read existing schema file
				content, err := os.ReadFile(filePath)
				if err != nil {
					// File doesn't exist, skip (GenSchema will create it with all mutations)
					continue
				}

				existingContent := string(content)

				// Only process schemas that have createBulkCSV
				if !strings.Contains(existingContent, "createBulkCSV"+node.Name) {
					continue
				}

				// Only inject missing mutations if at least one bulk update/delete mutation already exists.
				// If none exist, assume it's intentional that this schema doesn't have bulk update/delete.
				hasUpdateBulk := strings.Contains(existingContent, "updateBulk"+node.Name+"(")
				hasUpdateBulkCSV := strings.Contains(existingContent, "updateBulkCSV"+node.Name+"(")
				hasDeleteBulk := strings.Contains(existingContent, "deleteBulk"+node.Name+"(")

				if !hasUpdateBulk && !hasUpdateBulkCSV && !hasDeleteBulk {
					continue
				}

				s := bulkSchema{
					Name:       node.Name,
					PluralName: pluralizer.Plural(node.Name),
				}

				updatedContent := existingContent

				// Inject bulk mutations if missing
				if !strings.Contains(existingContent, "updateBulk"+node.Name) ||
					!strings.Contains(existingContent, "updateBulkCSV"+node.Name) ||
					!strings.Contains(existingContent, "deleteBulk"+node.Name) {
					updatedContent = injectBulkMutations(updatedContent, s)
				}

				// Inject payload types if missing - check for type definitions, not just the name
				// (the name appears in mutation return types too)
				if !strings.Contains(updatedContent, "type "+node.Name+"BulkUpdatePayload") ||
					!strings.Contains(updatedContent, "type "+node.Name+"BulkDeletePayload") {
					updatedContent = injectBulkPayloadTypes(updatedContent, s)
				}

				if updatedContent == existingContent {
					continue
				}

				if err := os.WriteFile(filePath, []byte(updatedContent), 0600); err != nil { //nolint:mnd
					log.Fatalf("Unable to write file: %v", err)
				}
			}

			return next.Generate(g)
		})
	}
}

// injectBulkMutations injects updateBulk, updateBulkCSV, and deleteBulk mutations into the schema
func injectBulkMutations(content string, s bulkSchema) string {
	// Find the Mutation block
	mutationStart := strings.Index(content, "extend type Mutation")
	if mutationStart == -1 {
		return content
	}

	openBrace := strings.Index(content[mutationStart:], "{")
	if openBrace == -1 {
		return content
	}

	openBrace += mutationStart

	closeBrace := findMatchingBrace(content, openBrace)
	if closeBrace == -1 {
		return content
	}

	mutationBlock := content[mutationStart : closeBrace+1]
	newMutations := ""

	// Add updateBulk if missing
	if !strings.Contains(mutationBlock, "updateBulk"+s.Name+"(") {
		newMutations += renderBulkUpdateMutation(s)
	}

	// Add updateBulkCSV if missing
	if !strings.Contains(mutationBlock, "updateBulkCSV"+s.Name+"(") {
		newMutations += renderBulkUpdateCSVMutation(s)
	}

	// Add deleteBulk if missing
	if !strings.Contains(mutationBlock, "deleteBulk"+s.Name+"(") {
		newMutations += renderBulkDeleteMutation(s)
	}

	if newMutations == "" {
		return content
	}

	// Insert new mutations before the closing brace of the Mutation block
	insertPoint := closeBrace
	prefix := content[:insertPoint]
	suffix := content[insertPoint:]

	// Ensure proper spacing
	prefix = strings.TrimRight(prefix, " \t\n")

	return prefix + "\n" + newMutations + suffix
}

// injectBulkPayloadTypes injects BulkUpdatePayload and BulkDeletePayload types into the schema
func injectBulkPayloadTypes(content string, s bulkSchema) string {
	newTypes := ""

	// Add BulkUpdatePayload if missing - check for type definition, not just the name
	// (the name appears in mutation return types too)
	if !strings.Contains(content, "type "+s.Name+"BulkUpdatePayload") {
		newTypes += renderBulkUpdatePayload(s)
	}

	// Add BulkDeletePayload if missing - check for type definition
	if !strings.Contains(content, "type "+s.Name+"BulkDeletePayload") {
		newTypes += renderBulkDeletePayload(s)
	}

	if newTypes == "" {
		return content
	}

	// Append to end of file
	content = strings.TrimRight(content, " \t\n")

	return content + "\n" + newTypes
}

func renderBulkUpdateMutation(s bulkSchema) string {
	tmpl := `    """
    Update multiple existing {{ .Name | ToLowerCamel }}s
    """
    updateBulk{{ .Name }}(
        """
        IDs of the {{ .Name | ToLowerCamel }}s to update
        """
        ids: [ID!]!
        """
        values to update the {{ .Name | ToLowerCamel }}s with
        """
        input: Update{{ .Name }}Input!
    ): {{ .Name }}BulkUpdatePayload!
`

	return renderBulkTemplate(tmpl, s)
}

func renderBulkUpdateCSVMutation(s bulkSchema) string {
	tmpl := `    """
    Update multiple existing {{ .Name | ToLowerCamel }}s via file upload
    """
    updateBulkCSV{{ .Name }}(
        """
        csv file containing values of the {{ .Name | ToLowerCamel }}, must include ID column
        """
        input: Upload!
    ): {{ .Name }}BulkUpdatePayload!
`

	return renderBulkTemplate(tmpl, s)
}

func renderBulkDeleteMutation(s bulkSchema) string {
	tmpl := `    """
    Delete multiple {{ .Name | ToLowerCamel }}s
    """
    deleteBulk{{ .Name }}(
        """
        IDs of the {{ .Name | ToLowerCamel }}s to delete
        """
        ids: [ID!]!
    ): {{ .Name }}BulkDeletePayload!
`

	return renderBulkTemplate(tmpl, s)
}

func renderBulkUpdatePayload(s bulkSchema) string {
	tmpl := `
"""
Return response for updateBulk{{ .Name }} mutation
"""
type {{ .Name }}BulkUpdatePayload {
    """
    Updated {{ .Name | ToLowerCamel }}s
    """
    {{ .PluralName | ToLowerCamel }}: [{{ .Name }}!]
    """
    IDs of the updated {{ .Name | ToLowerCamel }}s
    """
    updatedIDs: [ID!]
}
`

	return renderBulkTemplate(tmpl, s)
}

func renderBulkDeletePayload(s bulkSchema) string {
	tmpl := `
"""
Return response for deleteBulk{{ .Name }} mutation
"""
type {{ .Name }}BulkDeletePayload {
    """
    Deleted {{ .Name | ToLowerCamel }} IDs
    """
    deletedIDs: [ID!]!
}
`

	return renderBulkTemplate(tmpl, s)
}

func renderBulkTemplate(tmplStr string, s bulkSchema) string {
	fm := template.FuncMap{
		"ToLowerCamel": templates.ToGoPrivate,
	}

	tmpl, err := template.New("bulk").Funcs(fm).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Unable to parse bulk template: %v", err)
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, s); err != nil {
		log.Fatalf("Unable to execute bulk template: %v", err)
	}

	return builder.String()
}
