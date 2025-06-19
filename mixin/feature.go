package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"

	"github.com/theopenlane/entx"
)

// Feature returns a mixin that annotates the schema with the provided feature identifiers
func Feature(features ...entx.FeatureModule) ent.Mixin {
	return featureMixin{features: features}
}

type featureMixin struct {
	ent.Mixin
	features []entx.FeatureModule
}

func (featureMixin) Hooks() []ent.Hook   { return nil }
func (featureMixin) Fields() []ent.Field { return nil }
func (featureMixin) Edges() []ent.Edge   { return nil }

func (f featureMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features(f.features...),
	}
}
