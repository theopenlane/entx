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
	entfield "entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/stoewer/go-strcase"

	"github.com/theopenlane/entx"
)

var (
	searchFilename = "search"
)

// schema data for template
type search struct {
	// Name of the search (e.g. Global, Admin)
	Name string
	// Objects is a list of objects to generate bulk resolvers for
	Objects []Object
}

// Object is a struct to hold the object name for the bulk resolver
type Object struct {
	// Name of the object
	Name string
	// Fields that are searchable for object
	Fields []string
	// AdminFields are fields that are only searchable by admin
	AdminFields []string
}

// GenSchema generates graphql schemas when specified to be searchable
func GenSearchSchema(graphSchemaDir, graphQueryDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// create schema template
			schemaTmpl := createSearchTemplate()

			// create schema template
			typeTmpl := createSearchTypesTemplate()

			// create query template
			queryTmpl := createSearchQueryTemplate()

			// get input data
			inputData := getInputData(g)

			// sort objects by name so we have consistent output
			slices.SortFunc(inputData.Objects, func(a, b Object) int {
				return cmp.Compare(a.Name, b.Name)
			})

			// create search type schema file
			genSearchTypeTemplate(graphSchemaDir, typeTmpl, inputData)

			// create search schema file for global and admin
			genSearchSchemaTemplate(graphSchemaDir, schemaTmpl, inputData, false)
			genSearchSchemaTemplate(graphSchemaDir, schemaTmpl, inputData, true)

			// create search query file for global and admin
			genSearchQueryTemplate(graphQueryDir, queryTmpl, inputData, false)
			genSearchQueryTemplate(graphQueryDir, queryTmpl, inputData, true)

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

		fields, adminFields := GetSearchableFields(f.Name, g)

		// only add object if there are searchable fields other than the ID field (ID is always searchable)
		if len(fields) > 1 {
			inputData.Objects = append(inputData.Objects, Object{
				Name:        f.Name,
				Fields:      fields,
				AdminFields: adminFields,
			})
		}
	}

	return inputData
}

// genSearchSchemaTemplate generates the search schema file
func genSearchSchemaTemplate(graphSchemaDir string, tmpl *template.Template, inputData search, isAdmin bool) {
	inputData.Name = "Global"
	if isAdmin {
		inputData.Name = "Admin"
		searchFilename = "admin" + searchFilename
	}

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
func genSearchQueryTemplate(graphQueryDir string, tmpl *template.Template, inputData search, isAdmin bool) {
	inputData.Name = "Global"
	if isAdmin {
		inputData.Name = "Admin"
		searchFilename = "admin" + searchFilename
	}

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

// genSearchTypeTemplate generates the search type schema file
func genSearchTypeTemplate(graphSchemaDir string, tmpl *template.Template, inputData search) {
	// create search query file
	filePath := getFileName(graphSchemaDir, "searchtypes")

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

// createSearchTypesTemplate creates a new template for generating the search types graphql schema
func createSearchTypesTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"toPlural":     pluralize.NewClient().Plural,
		"toLowerCamel": strcase.LowerCamelCase,
		"toUpperCamel": strcase.UpperCamelCase,
	}

	// create schema template
	tmpl, err := template.New("types.tpl").Funcs(fm).ParseFS(_templates, "templates/search/types.tpl")
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
// all fields will be included in the admin search
// whereas only fields with the search annotation will be included in the global search
func GetSearchableFields(schemaName string, graph *gen.Graph) (fields []string, adminFields []string) {
	// add the object name that is being searched
	schema := getEntSchema(graph, schemaName)

	if schema == nil {
		return
	}

	for _, field := range schema.Fields {
		if isFieldSearchable(field) {
			// exclude bool fields from search
			if field.Info.Type == entfield.TypeBool {
				continue
			}

			// skip enums for now
			if field.Info.Type == entfield.TypeEnum {
				continue
			}

			fieldName := strcase.UpperCamelCase(field.Name)
			// capitalize ID field
			fieldName = strings.Replace(fieldName, "Id", "ID", 1)

			fields = append(fields, fieldName)
			adminFields = append(adminFields, fieldName)
		} else if isAdminFieldSearchable(field) {
			fieldName := strcase.UpperCamelCase(field.Name)

			adminFields = append(adminFields, fieldName)
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

// isAdminFieldSearchable checks if the field has the admin SearchField annotation
func isAdminFieldSearchable(field *load.Field) bool {
	searchAnt := &entx.SearchFieldAnnotation{}

	if ant, ok := field.Annotations[searchAnt.Name()]; ok {
		if err := searchAnt.Decode(ant); err != nil {
			return false
		}

		return !searchAnt.ExcludeAdmin
	}

	return true
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
