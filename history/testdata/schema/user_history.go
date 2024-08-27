package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	enthistory "github.com/theopenlane/entx/history"
)

type UserHistory struct {
	ent.Schema
}

func (UserHistory) Fields() []ent.Field {
	return []ent.Field{
		field.String("ref"),
		field.String("operation"),
		field.Int("age"),
		field.String("name"),
		field.String("nickname").
			Unique(),
	}
}

func (UserHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("age", "name").
			Unique(),
	}
}

func (UserHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		enthistory.Annotations{
			IsHistory: true,
		},
	}
}
