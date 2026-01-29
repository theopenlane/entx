package genhooks

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"

	"entgo.io/ent/entc/gen"
	"github.com/99designs/gqlgen/codegen/templates"
)

// workflowQueryType holds data for a workflow-eligible type in the query template
type workflowQueryType struct {
	Name string
}

// workflowQueryData holds data for the workflow query template
type workflowQueryData struct {
	Types                  []workflowQueryType
	WorkflowInstanceFields []string
	WorkflowEventFields    []string
}

// GenWorkflowQuery generates GraphQL queries for entities with workflow support.
// This creates queries for checking workflow status and retrieving workflow timeline.
func GenWorkflowQuery(graphQueryDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			var workflowTypes []workflowQueryType

			var workflowInstanceFields []string

			var workflowEventFields []string

			for _, node := range g.Nodes {
				switch node.Name {
				case "WorkflowInstance":
					workflowInstanceFields = getFieldNames(node.Fields)
				case "WorkflowEvent":
					workflowEventFields = getFieldNames(node.Fields)
				}

				if !hasWorkflowSupport(node) {
					continue
				}

				if isHistoryType(node) {
					continue
				}

				workflowTypes = append(workflowTypes, workflowQueryType{
					Name: node.Name,
				})
			}

			if len(workflowTypes) == 0 {
				return next.Generate(g)
			}

			if len(workflowInstanceFields) == 0 || len(workflowEventFields) == 0 {
				return next.Generate(g)
			}

			sort.Slice(workflowTypes, func(i, j int) bool {
				return workflowTypes[i].Name < workflowTypes[j].Name
			})

			tmpl := createWorkflowQueryTemplate()
			filePath := filepath.Clean(filepath.Join(graphQueryDir, "workflow.graphql"))

			file, err := os.Create(filePath)
			if err != nil {
				log.Fatalf("Unable to create workflow query file: %v", err)
			}
			defer file.Close()

			data := workflowQueryData{
				Types:                  workflowTypes,
				WorkflowInstanceFields: workflowInstanceFields,
				WorkflowEventFields:    workflowEventFields,
			}

			if err := tmpl.Execute(file, data); err != nil {
				log.Fatalf("Unable to execute workflow query template: %v", err)
			}

			return next.Generate(g)
		})
	}
}

// hasWorkflowSupport checks if a node has workflow support via workflow_eligible_marker field
func hasWorkflowSupport(node *gen.Type) bool {
	for _, field := range node.Fields {
		if field.Name == "workflow_eligible_marker" {
			return true
		}
	}

	return false
}

// isHistoryType checks if a node is a history type
func isHistoryType(node *gen.Type) bool {
	if node.Annotations == nil {
		return false
	}

	historyAnt, ok := node.Annotations["History"]
	if !ok {
		return false
	}

	historyMap, ok := historyAnt.(map[string]any)
	if !ok {
		return false
	}

	isHistory, ok := historyMap["isHistory"].(bool)

	return ok && isHistory
}

// createWorkflowQueryTemplate creates the template for workflow GraphQL queries
func createWorkflowQueryTemplate() *template.Template {
	fm := template.FuncMap{
		"ToLowerCamel": templates.ToGoPrivate,
	}

	tmplStr := `{{- range .Types }}
query Get{{ .Name }}WorkflowStatus(${{ .Name | ToLowerCamel }}Id: ID!) {
  {{ .Name | ToLowerCamel }}(id: ${{ .Name | ToLowerCamel }}Id) {
    id
    hasPendingWorkflow
    hasWorkflowHistory
    activeWorkflowInstances {
      {{- range $.WorkflowInstanceFields }}
      {{ . }}
      {{- end }}
    }
  }
}

query Get{{ .Name }}WorkflowTimeline(${{ .Name | ToLowerCamel }}Id: ID!, $first: Int, $last: Int, $after: Cursor, $before: Cursor, $orderBy: [WorkflowEventOrder!], $where: WorkflowEventWhereInput, $includeEmitFailures: Boolean) {
  {{ .Name | ToLowerCamel }}(id: ${{ .Name | ToLowerCamel }}Id) {
    id
    workflowTimeline(
      first: $first
      last: $last
      after: $after
      before: $before
      orderBy: $orderBy
      where: $where
      includeEmitFailures: $includeEmitFailures
    ) {
      totalCount
      pageInfo {
        startCursor
        endCursor
        hasPreviousPage
        hasNextPage
      }
      edges {
        node {
          {{- range $.WorkflowEventFields }}
          {{ . }}
          {{- end }}
        }
      }
    }
  }
}
{{ end }}`

	tmpl, err := template.New("workflow_query.tpl").Funcs(fm).Parse(tmplStr)
	if err != nil {
		log.Fatalf("Unable to parse workflow query template: %v", err)
	}

	return tmpl
}
