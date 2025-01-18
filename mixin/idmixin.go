package mixin

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/entx/customtypes"

	"github.com/theopenlane/entx"
)

// IDMixin holds the schema definition for the ID
type IDMixin struct {
	mixin.Schema
	// IncludeMappingID to include the mapping ID field to the schema that can be used without exposing the primary ID
	// by default, it is not included by default
	IncludeMappingID bool
	// IncludeHumanID to exclude the human ID field to
	HumanIdentifierPrefix string
}

// NewIDMixinWithPrefixedID creates a new IDMixin and includes an additional prefixed ID, e.g. TSK-000001
func NewIDMixinWithPrefixedID(prefix string) IDMixin {
	return IDMixin{HumanIdentifierPrefix: prefix}
}

// NewIDMixinWithMappingID creates a new IDMixin and includes an additional mapping ID
func NewIDMixinWithMappingID() IDMixin {
	return IDMixin{IncludeMappingID: true}
}

// Fields of the IDMixin.
func (i IDMixin) Fields() []ent.Field {
	fields := []ent.Field{
		field.String("id").
			Immutable().
			DefaultFunc(func() string { return ulids.New().String() }).
			Annotations(
				entx.FieldSearchable(),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
			),
	}

	if i.IncludeMappingID {
		fields = append(fields,
			field.String("mapping_id").
				Immutable().
				Annotations(
					entgql.Skip(),
				).
				Unique().
				DefaultFunc(func() string { return ulids.New().String() }),
		)
	}

	if i.HumanIdentifierPrefix != "" {
		fields = append(fields,
			field.String("identifier").
				Comment("a prefixed incremental field to use as a human readable identifier").
				SchemaType(map[string]string{
					dialect.Postgres: "SERIAL",
				}).
				ValueScanner(customtypes.NewPrefixedIdentifier(i.HumanIdentifierPrefix)).
				Immutable().
				Annotations(
					entx.FieldSearchable(),
					entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput),
					entsql.DefaultExpr("nextval('identifier_id_seq')"),
				).
				Unique(),
		)
	}

	return fields
}
