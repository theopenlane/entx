package genhooks

import (
	"cmp"
	"slices"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/99designs/gqlgen/codegen/templates"

	"github.com/theopenlane/entx"
)

// WorkflowField is a struct to hold the field name and type for workflow-eligible fields
type WorkflowField struct {
	// Name of the field
	Name string
	// Type of the field (string, json, etc)
	Type string
}

// GetWorkflowEligibleFields returns a list of workflow-eligible fields for a schema based on the WorkflowEligible annotation
func GetWorkflowEligibleFields(schemaName string, graph *gen.Graph) []WorkflowField {
	schema := getEntSchema(graph, schemaName)

	if schema == nil {
		return nil
	}

	var fields []WorkflowField

	for _, field := range schema.Fields {
		if !isFieldWorkflowEligible(field) {
			continue
		}

		fieldName := templates.ToGo(field.Name)

		f := WorkflowField{
			Name: fieldName,
			Type: field.Info.Type.String(),
		}

		fields = append(fields, f)
	}

	// sort fields so we have consistent output
	slices.SortFunc(fields, func(a, b WorkflowField) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return fields
}

// getWorkflowEligibleAnnotation retrieves the WorkflowEligible annotation from a field
func getWorkflowEligibleAnnotation(field *load.Field) *entx.WorkflowEligibleAnnotation {
	workflowAnt := &entx.WorkflowEligibleAnnotation{}
	if ant, ok := field.Annotations[workflowAnt.Name()]; ok {
		if err := workflowAnt.Decode(ant); err != nil {
			return nil
		}

		return workflowAnt
	}

	return nil
}

// isFieldWorkflowEligible checks if the field has the WorkflowEligible annotation set to true
func isFieldWorkflowEligible(field *load.Field) bool {
	// exclude sensitive fields
	if field.Sensitive {
		return false
	}

	// exclude immutable fields
	if field.Immutable {
		return false
	}

	// check for entgql skip annotations that would exclude the field
	if entSkip(field) {
		return false
	}

	workflowAnt := getWorkflowEligibleAnnotation(field)

	if workflowAnt == nil {
		return false
	}

	return workflowAnt.Eligible
}
