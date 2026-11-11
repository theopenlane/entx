package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Note is a minimal example schema for comments that a CSV import creates from a column on the parent row
type Note struct {
	ent.Schema
}

// Fields of the Note
func (Note) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable(),
		field.Text("text").
			Comment("the text of the note").
			NotEmpty(),
	}
}

// Edges of the Note
func (Note) Edges() []ent.Edge {
	return []ent.Edge{
		// the unique inverse edge is what lets the CSV generator link each created note to its row
		edge.From("organization", Organization.Type).
			Ref("notes").
			Unique(),
	}
}
