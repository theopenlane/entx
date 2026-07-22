package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/entx"
)

// WorkflowObjectRef is a minimal example schema linking a WorkflowInstance to the object it
// targets
type WorkflowObjectRef struct {
	ent.Schema
}

// Fields of the WorkflowObjectRef
func (WorkflowObjectRef) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable(),
		field.String("workflow_instance_id").
			Immutable(),
		field.String("organization_id").
			Immutable().
			Optional(),
	}
}

// Edges of the WorkflowObjectRef
func (WorkflowObjectRef) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workflow_instance", WorkflowInstance.Type).
			Field("workflow_instance_id").
			Required().
			Unique().
			Immutable(),
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Unique().
			Immutable().
			Annotations(entx.FieldWorkflowEligible()),
	}
}

// Annotations of the WorkflowObjectRef
func (WorkflowObjectRef) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
	}
}
