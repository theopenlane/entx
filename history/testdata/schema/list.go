package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	enthistory "github.com/theopenlane/entx/history"
	"github.com/theopenlane/iam/entfga"
)

type List struct {
	ent.Schema
}

func (List) Fields() []ent.Field {
	return []ent.Field{
		field.String("item"),
		field.Time("due_date"),
	}
}

func (List) Indexes() []ent.Index {
	return []ent.Index{}
}

func (List) Annotations() []schema.Annotation {
	return []schema.Annotation{
		enthistory.Annotations{
			Exclude: false,
		},
		entfga.Annotations{
			ObjectType:   "organization",
			IDField:      "OwnerID",
			IncludeHooks: false,
		},
	}
}
