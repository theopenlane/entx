package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx/vanilla/_example/ent/enums"
)

// OrgMembership holds the schema definition for the OrgMembership entity
type OrgMembership struct {
	ent.Schema
}

// Fields of the OrgMembership
func (OrgMembership) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Immutable(),
		field.Enum("role").
			GoType(enums.Role("")).
			Default(string(enums.RoleMember)).
			Values(string(enums.RoleOwner)), // adds owner to possible values
		field.String("organization_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the OrgMembership
func (OrgMembership) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("organization", Organization.Type).
			Field("organization_id").
			Required().
			Unique().
			Immutable(),
	}
}

// Annotations of the OrgMembership
func (OrgMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
	}
}

func (OrgMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id").
			Unique(),
	}
}
