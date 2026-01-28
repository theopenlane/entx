package genhooks

import (
	"html/template"
	"log"
	"os"
	"strings"

	"entgo.io/ent/entc/gen"
	"github.com/99designs/gqlgen/codegen/templates"
)

// workflowSchema holds data for the workflow GraphQL schema template
type workflowSchema struct {
	Name        string
	HasWorkflow bool
}

// GenWorkflowSchema generates GraphQL schema extensions for entities with ApprovalRequiredMixin
// This adds workflow fields for pending and active/history workflow instances
func GenWorkflowSchema(graphSchemaDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			tmpl := createWorkflowSchemaTemplate()

			for _, node := range g.Nodes {
				// Check if node has ApprovalRequiredMixin (indicated by proposed_changes field)
				hasWorkflowSupport := false

				for _, field := range node.Fields {
					if field.Name == "workflow_eligible_marker" {
						hasWorkflowSupport = true
						break
					}
				}

				if !hasWorkflowSupport {
					continue
				}

				// Skip history types
				if node.Annotations != nil {
					if historyAnt, ok := node.Annotations["History"]; ok {
						if historyMap, ok := historyAnt.(map[string]any); ok {
							if isHistory, ok := historyMap["isHistory"].(bool); ok && isHistory {
								continue
							}
						}
					}
				}

				filePath := getFileName(graphSchemaDir, node.Name)

				// Read existing schema file
				content, err := os.ReadFile(filePath)
				if err != nil {
					// File doesn't exist, skip
					continue
				}

				s := workflowSchema{
					Name:        node.Name,
					HasWorkflow: true,
				}

				newBlock, err := renderWorkflowSchemaTemplate(tmpl, s)
				if err != nil {
					log.Fatalf("Unable to execute template: %v", err)
				}

				existingContent := string(content)

				updatedContent, replaced := replaceWorkflowSchemaBlock(node.Name, existingContent, newBlock)
				if !replaced {
					updatedContent = joinWorkflowSchemaBlocks(newBlock, existingContent)
				}

				if updatedContent == existingContent {
					continue
				}

				if err := os.WriteFile(filePath, []byte(updatedContent), 0600); err != nil { // nolint:mnd
					log.Fatalf("Unable to write file: %v", err)
				}
			}

			return next.Generate(g)
		})
	}
}

// containsWorkflowFields checks if the schema already contains workflow fields
func containsWorkflowFields(content string) bool {
	return strings.Contains(content, "hasPendingWorkflow") ||
		strings.Contains(content, "hasWorkflowHistory") ||
		strings.Contains(content, "activeWorkflowInstances") ||
		strings.Contains(content, "workflowTimeline")
}

func renderWorkflowSchemaTemplate(tmpl *template.Template, data workflowSchema) (string, error) {
	var builder strings.Builder

	if err := tmpl.Execute(&builder, data); err != nil {
		return "", err
	}

	return builder.String(), nil
}

func replaceWorkflowSchemaBlock(typeName string, content string, newBlock string) (string, bool) {
	start, end := findWorkflowSchemaBlock(typeName, content)
	if start == -1 {
		return content, false
	}

	prefix := content[:start]
	suffix := content[end:]

	return prefix + joinWorkflowSchemaBlocks(newBlock, suffix), true
}

func findWorkflowSchemaBlock(typeName string, content string) (int, int) {
	needle := "extend type " + typeName
	searchFrom := 0

	for {
		index := strings.Index(content[searchFrom:], needle)
		if index == -1 {
			return -1, -1
		}

		index += searchFrom

		openBrace := strings.Index(content[index:], "{")
		if openBrace == -1 {
			return -1, -1
		}

		openBrace += index

		closeBrace := findMatchingBrace(content, openBrace)
		if closeBrace == -1 {
			return -1, -1
		}

		block := content[index : closeBrace+1]
		if containsWorkflowFields(block) {
			return index, closeBrace + 1
		}

		searchFrom = closeBrace + 1
	}
}

func findMatchingBrace(content string, openBrace int) int {
	depth := 0

	for i := openBrace; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}

	return -1
}

func joinWorkflowSchemaBlocks(block string, suffix string) string {
	block = strings.TrimRight(block, " \t\r\n")
	suffix = trimLeadingBlankLines(suffix)

	if suffix == "" {
		return block + "\n"
	}

	return block + "\n\n" + suffix
}

func trimLeadingBlankLines(content string) string {
	index := 0

	for index < len(content) {
		lineEnd := strings.IndexByte(content[index:], '\n')
		if lineEnd == -1 {
			if strings.TrimSpace(content[index:]) == "" {
				return ""
			}

			return content[index:]
		}

		line := content[index : index+lineEnd]
		if strings.TrimSpace(line) != "" {
			return content[index:]
		}

		index += lineEnd + 1
	}

	return ""
}

// createWorkflowSchemaTemplate creates the template for workflow schema extensions.
func createWorkflowSchemaTemplate() *template.Template {
	fm := template.FuncMap{
		"ToLowerCamel": templates.ToGoPrivate,
	}

	tmplStr := `extend type {{ .Name }} {
    """
    Indicates if this {{ .Name | ToLowerCamel }} has pending changes awaiting workflow approval
    """
    hasPendingWorkflow: Boolean!
    """
    Indicates if this {{ .Name | ToLowerCamel }} has any workflow history (completed or failed instances)
    """
    hasWorkflowHistory: Boolean!
    """
    Returns active workflow instances for this {{ .Name | ToLowerCamel }} (RUNNING or PAUSED)
    """
    activeWorkflowInstances: [WorkflowInstance!]!
    """
    Returns the workflow event timeline for this {{ .Name | ToLowerCamel }} across all workflow instances
    """
    workflowTimeline(
        after: Cursor
        first: Int
        before: Cursor
        last: Int
        orderBy: [WorkflowEventOrder!]
        where: WorkflowEventWhereInput
        includeEmitFailures: Boolean
    ): WorkflowEventConnection!
}
`

	tmpl, err := template.New("workflow_schema.tpl").Funcs(fm).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Unable to parse workflow schema template: %v", err)
	}

	return tmpl
}
