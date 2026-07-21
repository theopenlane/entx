//go:build ignore

package main

import (
	"log"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/entityops"
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

	accessExt := accessmap.New(
		accessmap.WithGeneratedDir(entGeneratedDir),
		accessmap.WithPackageName("ent"),
	)

	entityOpsExt := entityops.New(
		entityops.WithOutputDir(entGeneratedDir+"entityops"),
		entityops.WithPackageName("entityops"),
		entityops.WithEntPackage("github.com/theopenlane/entx/vanilla/_example/ent"),
		entityops.WithGalaPackage("github.com/theopenlane/entx/vanilla/_example/pkg/gala"),
		entityops.WithJsonxPackage("github.com/theopenlane/entx/vanilla/_example/pkg/jsonx"),
		entityops.WithLogxPackage("github.com/theopenlane/entx/vanilla/_example/pkg/logx"),
		entityops.WithCelxPackage("github.com/theopenlane/entx/vanilla/_example/pkg/celx"),
	)

	if err := entc.Generate("./schema",
		&gen.Config{
			Features: []gen.Feature{gen.FeaturePrivacy},
			Hooks: []gen.Hook{
				genhooks.GenSchema(graphSchemaDir),
				genhooks.GenQuery(graphQueryDir, graphSchemaDir),
				accessExt.Hook(),
				entityOpsExt.Hook(),
			},
		},
		entc.Extensions(
			gqlExt,
		),
	); err != nil {
		log.Fatal("running ent codegen:", err)
	}
}
