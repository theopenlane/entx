package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/mixin"
)

// Organization holds the schema definition for the Organization entity
type Organization struct {
	ent.Schema
}

// Fields of the Organization
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the organization").
			NotEmpty(),
		field.String("description").
			Comment("An optional description of the organization").
			Optional(),
		field.JSON("preferences", map[string]any{}).
			Comment("free-form organization preferences, including compliance setup answers").
			Optional().
			// field level task rule based on the values set on the field
			Annotations(entx.FieldTaskRule(organizationTaskRules...)),
	}
}

func (Organization) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").
			Unique(),
	}
}

// Annotations of the Organization
func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Mutations(entgql.MutationCreate(), entgql.MutationUpdate()),
		// schema level task rule that happens on create of the organization
		entx.SchemaTaskRule("setup-payment-method"),
	}
}

func (Organization) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.IDMixin{
			HumanIdentifierPrefix: "ORG",
			SingleFieldIndex:      true,
		},
		mixin.AuditMixin{},
	}
}

// organizationTaskRules are suggested-task rules driven by organization preferences
var organizationTaskRules = []entx.TaskRuleSpec{
	{
		RuleID:     "import-existing-policies",
		Expression: `value.policies.has_existing == true`,
		Trigger:    entx.TaskRuleOnCreateOnly,
	},
	{
		RuleID:      "framework",
		EachElement: `value.frameworks`,
	},
}
