package genhooks

import (
	"cmp"
	"html/template"
	"log"
	"os"
	"slices"
	"strings"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/entx"
)

const (
	searchFilename = "search"
)

// schema data for template
type search struct {
	// Objects is a list of objects to generate bulk resolvers for
	Objects []Object
}

// Object is a struct to hold the object name for the bulk resolver
type Object struct {
	// Name of the object
	Name string
	// Fields that are searchable for object
	Fields []string
}

// GenSchema generates graphql schemas when specified to be searchable
func GenSearchSchema(graphSchemaDir, graphQueryDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// create schema template
			tmpl := createSearchTemplate()

			// create query template
			queryTmpl := createSearchQueryTemplate()

			// get input data
			inputData := getInputData(g)

			// sort objects by name so we have consistent output
			slices.SortFunc(inputData.Objects, func(a, b Object) int {
				return cmp.Compare(a.Name, b.Name)
			})

			// create search schema file
			genSearchSchemaTemplate(graphSchemaDir, tmpl, inputData)

			// create search query file
			genSearchQueryTemplate(graphQueryDir, queryTmpl, inputData)

			return next.Generate(g)
		})
	}
}

// getInputData returns the data to be used in the search templates
func getInputData(g *gen.Graph) search {
	inputData := search{
		Objects: []Object{},
	}

	for _, f := range g.Nodes {
		// check skip annotation and search annotation
		// to generate search schema the following conditions must be met:
		// skip must be false
		// skipSearch must be false
		// there must be at least one searchable field other than the ID field
		if checkSchemaGenSkip(f) || !includeSchemaForSearch(f) {
			continue
		}

		fields := GetSearchableFields(f.Name, g)

		// only add object if there are searchable fields other than the ID field (ID is always searchable)
		if len(fields) > 1 {
			inputData.Objects = append(inputData.Objects, Object{
				Name:   f.Name,
				Fields: fields,
			})
		}
	}

	return inputData
}

// genSearchSchemaTemplate generates the search schema file
func genSearchSchemaTemplate(graphSchemaDir string, tmpl *template.Template, inputData search) {
	// create search schema file
	filePath := getFileName(graphSchemaDir, searchFilename)

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}

	// execute template and write to file
	if err = tmpl.Execute(file, inputData); err != nil {
		log.Fatalf("Unable to execute template: %v", err)
	}
}

// genSearchQueryTemplate generates the search query file
func genSearchQueryTemplate(graphQueryDir string, tmpl *template.Template, inputData search) {
	// create search query file
	filePath := getFileName(graphQueryDir, searchFilename)

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}

	// execute query template and write to file
	if err = tmpl.Execute(file, inputData); err != nil {
		log.Fatalf("Unable to execute query template: %v", err)
	}
}

// createTemplate creates a new template for generating the search graphql schemas
func createSearchTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"toPlural":     pluralize.NewClient().Plural,
		"toLowerCamel": strcase.LowerCamelCase,
		"toUpperCamel": strcase.UpperCamelCase,
	}

	// create schema template
	tmpl, err := template.New("graph.tpl").Funcs(fm).ParseFS(_templates, "templates/search/graph.tpl")
	if err != nil {
		log.Fatalf("Unable to parse template: %v", err)
	}

	return tmpl
}

// createSearchQueryTemplate creates a new template for generating the search query
func createSearchQueryTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"toPlural":     pluralize.NewClient().Plural,
		"toLowerCamel": strcase.LowerCamelCase,
		"toUpperCamel": strcase.UpperCamelCase,
	}

	// create schema template
	tmpl, err := template.New("query.tpl").Funcs(fm).ParseFS(_templates, "templates/search/query.tpl")
	if err != nil {
		log.Fatalf("Unable to parse template: %v", err)
	}

	return tmpl
}

// includeSchemaForSearch checks if the type has the Schema Searchable annotation
func includeSchemaForSearch(node *gen.Type) bool {
	schemaGenAnt := &entx.SchemaGenAnnotation{}

	if ant, ok := node.Annotations[schemaGenAnt.Name()]; ok {
		if err := schemaGenAnt.Decode(ant); err != nil {
			return false
		}

		return !schemaGenAnt.SkipSearch
	}

	return true
}

// GetSearchableFields returns a list of searchable fields for a schema based on the search annotation
func GetSearchableFields(schemaName string, graph *gen.Graph) (fields []string) {
	// add the object name that is being searched
	schema := getEntSchema(graph, schemaName)

	for _, field := range schema.Fields {
		if isFieldSearchable(field) {
			fieldName := strcase.UpperCamelCase(field.Name)
			// capitalize ID field
			if strings.EqualFold(field.Name, "id") {
				fieldName = "ID"
			}

			fields = append(fields, fieldName)
		}
	}

	// sort fields so we have consistent output
	slices.Sort(fields)

	return
}

// isFieldSearchable checks if the field has the SearchField annotation
func isFieldSearchable(field *load.Field) bool {
	searchAnt := &entx.SearchFieldAnnotation{}

	if ant, ok := field.Annotations[searchAnt.Name()]; ok {
		if err := searchAnt.Decode(ant); err != nil {
			return false
		}

		return searchAnt.Searchable
	}

	return false
}

// getEntSchema returns the schema for a given name
func getEntSchema(graph *gen.Graph, name string) *load.Schema {
	for _, s := range graph.Schemas {
		if s.Name == name {
			return s
		}
	}

	return nil
}
