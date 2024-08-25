package genhooks

import (
	"html/template"
	"log"
	"os"

	"entgo.io/ent/entc/gen"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/entx"
)

// schema data for template
type schema struct {
	Name string
}

// GenSchema generates graphql schemas when not specified to be skipped
func GenSchema(graphSchemaDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// create schema template
			tmpl := createTemplate()

			// loop through all nodes and generate schema if not specified to be skipped
			for _, node := range g.Nodes {
				// check skip annotation
				if checkSchemaGenSkip(node) {
					continue
				}

				filePath := getFileName(graphSchemaDir, node.Name)

				// check if schema already exists, skip generation so we don't overwrite manual changes
				if _, err := os.Stat(filePath); err == nil {
					continue
				}

				file, err := os.Create(filePath)
				if err != nil {
					log.Fatalf("Unable to create file: %v", err)
				}

				s := schema{
					Name: node.Name,
				}

				// execute template and write to file
				if err = tmpl.Execute(file, s); err != nil {
					log.Fatalf("Unable to execute template: %v", err)
				}
			}

			return next.Generate(g)
		})
	}
}

// checkSchemaGenSkip checks if the type has the Schema Skip annotation
func checkSchemaGenSkip(node *gen.Type) bool {
	schemaGenAnt := &entx.SchemaGenAnnotation{}

	if ant, ok := node.Annotations[schemaGenAnt.Name()]; ok {
		if err := schemaGenAnt.Decode(ant); err != nil {
			return false
		}

		return schemaGenAnt.Skip
	}

	return false
}

// createTemplate creates a new template for generating graphql schemas
func createTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"ToLowerCamel": strcase.LowerCamelCase,
		"ToPlural":     pluralize.NewClient().Plural,
	}

	// create schema template
	tmpl, err := template.New("graph.tpl").Funcs(fm).ParseFS(_templates, "templates/graph.tpl")
	if err != nil {
		log.Fatalf("Unable to parse template: %v", err)
	}

	return tmpl
}
