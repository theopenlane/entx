package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/entx"
)

// WorkflowInstance is a minimal example schema
type WorkflowInstance struct {
	ent.Schema
}

// Fields of the WorkflowInstance
func (WorkflowInstance) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable(),
		field.String("status").
			Optional().
			Annotations(entx.FieldWorkflowEligible()),
	}
}

// Annotations of the WorkflowInstance
func (WorkflowInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
	}
}
