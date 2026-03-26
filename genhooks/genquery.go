package genhooks

import (
	"html/template"
	"log"
	"os"
	"sort"
	"strings"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
)

// query data for template
type query struct {
	// Name of the type
	Name string
	// Fields to include in the query
	Fields []string
	// IncludeMutations to include mutation (create, update, delete) queries
	IncludeMutations bool
	// IsHistory to indicate if the type is a history type
	IsHistory bool
}

// GenQuery generates graphql queries when not specified to be skipped
func GenQuery(graphSchemaDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// create schema query
			tmpl := createQuery()

			// loop through all nodes and generate schema if not specified to be skipped
			for _, node := range g.Nodes {
				generateQuery(node, tmpl, graphSchemaDir)
			}

			return next.Generate(g)
		})
	}
}

// generateQuery generates the query file for the type
func generateQuery(node *gen.Type, tmpl *template.Template, graphSchemaDir string) {
	// check skip annotation
	if checkQueryGenSkip(node) {
		return
	}

	filePath := getFileName(graphSchemaDir, node.Name)

	// check if schema already exists, skip generation so we don't overwrite manual changes
	if _, err := os.Stat(filePath); err == nil {
		return
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}

	s := query{
		Name:             node.Name,
		Fields:           getFieldNames(node.Fields),
		IncludeMutations: checkEntqlMutation(node),
		IsHistory:        isHistorySchema(node),
	}

	// execute template and write to file
	if err = tmpl.Execute(file, s); err != nil {
		log.Fatalf("Unable to execute template: %v", err)
	}
}

// getFieldNames returns a list of field names from a list of fields in alphabetical order
func getFieldNames(fields []*gen.Field) []string {
	// field names should always include the id field
	fieldNames := []string{"id"}

	for _, f := range fields {
		// check for the Skip annotation for gqlgen
		if checkEntGqlSkip(f) {
			continue
		}

		// skip soft delete fields
		if isSoftDeleteField(f) {
			continue
		}

		// check if sensitive field
		if f.Sensitive() {
			continue
		}

		fieldNames = append(fieldNames, templates.ToGoPrivate(f.StructField()))
	}

	// sort field names
	sort.Strings(fieldNames)

	return fieldNames
}

// checkEntqlMutation checks if the type has the entgql.Mutation annotation
func checkEntqlMutation(node *gen.Type) bool {
	entqlAnt := &entgql.Annotation{}

	if ant, ok := node.Annotations[entqlAnt.Name()]; ok {
		if err := entqlAnt.Decode(ant); err != nil {
			return false
		}

		if entqlAnt.MutationInputs != nil {
			return true
		}
	}

	return false
}

// checkQueryGenSkip checks if the type has the QueryGen Skip annotation
func checkQueryGenSkip(node *gen.Type) bool {
	queryGenAnt := &entx.QueryGenAnnotation{}

	if ant, ok := node.Annotations[queryGenAnt.Name()]; ok {
		if err := queryGenAnt.Decode(ant); err != nil {
			return false
		}

		return queryGenAnt.Skip
	}

	return false
}

// checkEntGqlSkip checks if the field has the entgql.Skip annotation
// and returns true if it is set to SkipType
func checkEntGqlSkip(f *gen.Field) bool {
	entqlAnt := &entgql.Annotation{}

	if ant, ok := f.Annotations[entqlAnt.Name()]; ok {
		if err := entqlAnt.Decode(ant); err != nil {
			return false
		}

		if entqlAnt.Skip.Is(entgql.SkipType) {
			return true
		}

		return false
	}

	return false
}

// isHistorySchema checks if the type is a history type
func isHistorySchema(node *gen.Type) bool {
	return strings.Contains(node.Name, "History")
}

// createQuery creates a new template for generating graphql queries for the client
func createQuery() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"ToLowerCamel": templates.ToGoPrivate,
		"ToPlural":     pluralize.NewClient().Plural,
	}

	// create schema template
	tmpl, err := template.New("query.tpl").Funcs(fm).ParseFS(_templates, "templates/query.tpl")
	if err != nil {
		log.Fatalf("Unable to parse template: %v", err)
	}

	return tmpl
}
