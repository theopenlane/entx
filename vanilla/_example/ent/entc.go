//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/theopenlane/entx/genhooks"

	"entgo.io/ent/entc"
)

const (
	graphSchemaDir  = "./../schema/"
	graphQueryDir   = "./../query/"
	entGeneratedDir = "./../ent/"
)

func main() {
	gqlExt, err := entgql.NewExtension(
		// Tell Ent to generate a GraphQL schema for
		// the Ent schema in a file named ent.graphql.
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath("../schema/ent.graphql"),
		entgql.WithWhereInputs(true),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}

	accessExt := genhooks.NewAccessMapExt(
		genhooks.WithGeneratedDir(entGeneratedDir),
	)

	if err := entc.Generate("./schema",
		&gen.Config{
			Features: []gen.Feature{gen.FeaturePrivacy},
			Hooks: []gen.Hook{
				genhooks.GenSchema(graphSchemaDir),
				genhooks.GenQuery(graphQueryDir),
				accessExt.Hook(),
			},
		},
		entc.Extensions(
			gqlExt,
		),
	); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
