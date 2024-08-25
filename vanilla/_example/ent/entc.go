//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/theopenlane/entx/genhooks"
	"go.uber.org/zap"

	"entgo.io/ent/entc"
)

const (
	graphSchemaDir = "./../schema/"
	graphQueryDir  = "./../query/"
)

func main() {
	gqlExt, err := entgql.NewExtension(
		// Tell Ent to generate a GraphQL schema for
		// the Ent schema in a file named ent.graphql.
		entgql.WithSchemaGenerator(),
		entgql.WithSchemaPath("../schema/ent.graphql"),
		entgql.WithConfigPath("../gqlgen.yml"),
		entgql.WithWhereInputs(true),
	)
	if err != nil {
		log.Fatalf("creating entgql extension: %v", err)
	}

	if err := entc.Generate("./schema",
		&gen.Config{
			Features: []gen.Feature{gen.FeaturePrivacy},
			Hooks: []gen.Hook{
				genhooks.GenSchema(graphSchemaDir),
				genhooks.GenQuery(graphQueryDir),
			},
		},
		entc.Dependency(
			entc.DependencyName("Logger"),
			entc.DependencyType(zap.SugaredLogger{}),
		),
		entc.Extensions(
			gqlExt,
		),
	); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
