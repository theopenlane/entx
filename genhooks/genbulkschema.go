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

// GenBulkSchema generates GraphQL schema extensions for bulk update and delete mutations.
// This injects updateBulk, updateBulkCSV, and deleteBulk mutations into existing schemas
// that already have createBulkCSV mutations.
func GenBulkSchema(graphSchemaDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
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

				// Only inject if schema has createBulkCSV but missing updateBulkCSV or deleteBulk
				if !strings.Contains(existingContent, "createBulkCSV"+node.Name) {
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
