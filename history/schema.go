package enthistory

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// history holds the schema definition for the history entity
type history struct {
	ent.Schema
	ref ent.Field
}

// Fields of the history schema
func (h history) Fields() []ent.Field {
	return []ent.Field{
		field.Time("history_time").
			Default(time.Now).
			Immutable(),
		field.Enum("operation").
			GoType(OpType("")).
			Immutable(),
		h.ref,
	}
}
