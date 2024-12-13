package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/theopenlane/iam/entfga"
)

type ActionPlan struct {
	ent.Schema
}

func (ActionPlan) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
	}
}

func (ActionPlan) Indexes() []ent.Index {
	return []ent.Index{}
}

func (ActionPlan) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.Annotations{
			IncludeHooks: false,
		},
	}
}
