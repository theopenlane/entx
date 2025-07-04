package mixin

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/utils/ulids"

	"github.com/theopenlane/entx"
)

// IDMixin holds the schema definition for the ID
type IDMixin struct {
	mixin.Schema
	// IncludeMappingID to include the mapping ID field to the schema that can be used without exposing the primary ID
	// by default, it is not included by default
	IncludeMappingID bool
	// HumanIdentifierPrefix is the prefix to use for the human identifier, if set a display_id field will be added
	// based on the original ID
	HumanIdentifierPrefix string
	// OverrideDefaultIndex to override the default index set on the display ID
	OverrideDefaultIndex string
	// SingleFieldIndex to set a single field index on the display ID
	SingleFieldIndex bool
	// OverrideDisplayID field name lets you customize the display ID field name
	OverrideDisplayID string
	// DisplayIDLength is the length of the display ID without the prefix, defaults to 6
	DisplayIDLength int
}

const humanIDFieldName = "display_id"

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
		displayField := field.String(humanIDFieldName).
			Comment("a shortened prefixed id field to use as a human readable identifier").
			NotEmpty(). // this is set by the hook
			Immutable().
			Annotations(
				entx.FieldSearchable(),
				entgql.Skip(entgql.SkipMutationCreateInput|entgql.SkipMutationUpdateInput), // do not allow users to set this field
			)

		if i.SingleFieldIndex {
			displayField = displayField.Unique()
		}

		fields = append(fields, displayField)
	}

	return fields
}

// Indexes of the IDMixin
func (i IDMixin) Indexes() []ent.Index {
	idx := []ent.Index{}

	if i.HumanIdentifierPrefix != "" && !i.SingleFieldIndex {
		idxField := "owner_id"
		if i.OverrideDefaultIndex != "" {
			idxField = i.OverrideDefaultIndex
		}

		idx = append(idx, index.Fields(humanIDFieldName, idxField).
			Unique())
	}

	return idx
}

// Hooks of the IDMixin
func (i IDMixin) Hooks() []ent.Hook {
	if i.HumanIdentifierPrefix == "" {
		// do not add hooks if the field is not used
		return []ent.Hook{}
	}

	return []ent.Hook{setIdentifierHook(i)}
}

type HookFunc func(i IDMixin) ent.Hook

var setIdentifierHook HookFunc = func(i IDMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			mut, ok := m.(mutationWithDisplayID)
			if ok {
				if id, exists := mut.ID(); exists {
					// default the length to 6 if not set
					length := 6
					if i.DisplayIDLength > 0 {
						length = i.DisplayIDLength
					}

					out := generateShortCharID(id, length)

					mut.SetDisplayID(fmt.Sprintf("%s-%s", i.HumanIdentifierPrefix, out))
				}
			}

			return next.Mutate(ctx, m)
		})
	}
}

// generateShortCharID generates a set-length alphanumeric string based on a ULID.
// Length 6: For up to 10,000 IDs, the collision probability is very low (~0.005%)
// Length 6: For up to 100,000 IDs, the collision probability is low (~0.5%)
func generateShortCharID(ulid string, length int) string {
	// Hash the ULID using SHA256
	hash := sha256.Sum256([]byte(ulid))

	// Encode the hash using Base32 to get an alphanumeric string
	encoded := base32.StdEncoding.EncodeToString(hash[:])

	// Remove padding and make it uppercase
	encoded = strings.ToUpper(strings.TrimRight(encoded, "="))

	// Return the first n characters
	return encoded[:length]
}

// mutationWithDisplayID is an interface that mutations can implement to get the identifier ID
type mutationWithDisplayID interface {
	SetDisplayID(string)
	ID() (id string, exists bool)
	Type() string
}
