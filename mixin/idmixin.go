package mixin

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql/sqlgraph"
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
	// DisplayIDIndexWhere, when set, makes the default display ID unique index partial with the given predicate
	DisplayIDIndexWhere string
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

		displayIdx := index.Fields(humanIDFieldName, idxField).
			Unique()

		if i.DisplayIDIndexWhere != "" {
			displayIdx = displayIdx.Annotations(entsql.IndexWhere(i.DisplayIDIndexWhere))
		}

		idx = append(idx, displayIdx)
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

const (
	// defaultDisplayIDLength is the display ID length used when the mixin does not set one
	defaultDisplayIDLength = 6
	// maxDisplayIDAttempts bounds display ID regeneration retries on unique-constraint collisions
	maxDisplayIDAttempts = 3
)

var setIdentifierHook HookFunc = func(i IDMixin) ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if !m.Op().Is(ent.OpCreate) {
				return next.Mutate(ctx, m)
			}

			mut, ok := m.(mutationWithDisplayID)
			if !ok {
				return next.Mutate(ctx, m)
			}

			id, exists := mut.ID()
			if !exists {
				return next.Mutate(ctx, m)
			}

			length := defaultDisplayIDLength
			if i.DisplayIDLength > 0 {
				length = i.DisplayIDLength
			}

			var (
				v   ent.Value
				err error
			)

			for attempt := range maxDisplayIDAttempts {
				mut.SetDisplayID(fmt.Sprintf("%s-%s", i.HumanIdentifierPrefix, generateShortCharID(saltForAttempt(id, attempt), length)))

				v, err = next.Mutate(ctx, m)
				if err == nil || !isDisplayIDConflict(err) {
					return v, err
				}
			}

			return v, err
		})
	}
}

// saltForAttempt returns the hash input for a display ID generation attempt
func saltForAttempt(id string, attempt int) string {
	if attempt == 0 {
		return id
	}

	return fmt.Sprintf("%s#%d", id, attempt)
}

// isDisplayIDConflict reports whether err is a unique-constraint violation on the display ID column
func isDisplayIDConflict(err error) bool {
	return sqlgraph.IsUniqueConstraintError(err) && strings.Contains(err.Error(), humanIDFieldName)
}

// generateShortCharID generates a set-length alphanumeric string based on a ULID.
// Length 6: For up to 10,000 IDs, the collision probability is ~4.6%, reaching ~50% near 40,000 IDs
// Length 8: For up to 100,000 IDs, the collision probability is ~0.5%
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
