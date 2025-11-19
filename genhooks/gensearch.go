package genhooks

import (
	"cmp"
	"html/template"
	"os"
	"slices"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	entfield "entgo.io/ent/schema/field"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/gertd/go-pluralize"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/entx"
)

// schema data for template
type search struct {
	// Name of the search (e.g. Global, Admin)
	Name string
	// Objects is a list of objects to generate search resolvers for
	Objects []Object
}

// Object is a struct to hold the object name for the search resolver
type Object struct {
	// Name of the object
	Name string
	// Fields that are searchable for object
	Fields []Field
	// AdminFields are fields that are only searchable by admin
	AdminFields []Field
}

// Field is a struct to hold the field name and type
type Field struct {
	// Name of the field
	Name string
	// Type of the field (string, json, etc)
	Type string
	// Path for JSON fields
	Path string
	// DotPath for JSON fields
	DotPath string
}

// GenSchema generates graphql schemas when specified to be searchable
func GenSearchSchema(graphSchemaDir, graphQueryDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			// create schema template
			schemaTmpl := createSearchTemplate()

			// create query template
			queryTmpl := createSearchQueryTemplate()

			// get input data
			inputData := getInputData(g)

			// sort objects by name so we have consistent output
			slices.SortFunc(inputData.Objects, func(a, b Object) int {
				return cmp.Compare(a.Name, b.Name)
			})

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
		// skipQueryGen must be false
		// there must be at least one searchable field other than the ID field
		if checkSchemaGenSkip(f) || checkQueryGenSkip(f) || !includeSchemaForSearch(f) {
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
	fileName := getSearchFileName(isAdmin)

	inputData.Name = "Global"
	if isAdmin {
		inputData.Name = "Admin"
	}

	// create search schema file
	filePath := getFileName(graphSchemaDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal().Msgf("Unable to create file: %v", err)
	}

	// execute template and write to file
	if err = tmpl.Execute(file, inputData); err != nil {
		log.Fatal().Msgf("Unable to execute template: %v", err)
	}
}

// genSearchQueryTemplate generates the search query file
func genSearchQueryTemplate(graphQueryDir string, tmpl *template.Template, inputData search, isAdmin bool) {
	fileName := getSearchFileName(isAdmin)

	inputData.Name = "Global"
	if isAdmin {
		inputData.Name = "Admin"
	}

	// create search query file
	filePath := getFileName(graphQueryDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal().Msgf("Unable to create file: %v", err)
	}

	// execute query template and write to file
	if err = tmpl.Execute(file, inputData); err != nil {
		log.Fatal().Msgf("Unable to execute query template: %v", err)
	}
}

// createTemplate creates a new template for generating the search graphql schemas
func createSearchTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"toPlural":     pluralize.NewClient().Plural,
		"toLower":      templates.ToGoPrivate,
		"toUpperCamel": templates.ToGo,
	}

	// create schema template
	tmpl, err := template.New("graph.tpl").Funcs(fm).ParseFS(_templates, "templates/search/graph.tpl")
	if err != nil {
		log.Fatal().Msgf("Unable to parse template: %v", err)
	}

	return tmpl
}

// createSearchQueryTemplate creates a new template for generating the search query
func createSearchQueryTemplate() *template.Template {
	// function map for template
	fm := template.FuncMap{
		"toPlural":     pluralize.NewClient().Plural,
		"toLower":      templates.ToGoPrivate,
		"toUpperCamel": templates.ToGo,
	}

	// create schema template
	tmpl, err := template.New("query.tpl").Funcs(fm).ParseFS(_templates, "templates/search/query.tpl")
	if err != nil {
		log.Fatal().Msgf("Unable to parse template: %v", err)
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
func GetSearchableFields(schemaName string, graph *gen.Graph) (fields []Field, adminFields []Field) {
	// add the object name that is being searched
	schema := getEntSchema(graph, schemaName)

	if schema == nil {
		return
	}

	for _, field := range schema.Fields {
		if isFieldTypeExcluded(field) {
			continue
		}

		fieldName := templates.ToGo(field.Name)

		f := Field{
			Name:    fieldName,
			Type:    field.Info.Type.String(),
			Path:    getPathAnnotation(field),
			DotPath: getDotPathAnnotation(field),
		}

		if isFieldSearchable(field) {
			fields = append(fields, f)
			adminFields = append(adminFields, f)
		} else if isAdminFieldSearchable(field) {
			adminFields = append(adminFields, f)
		}
	}

	// sort fields so we have consistent output
	slices.SortFunc(fields, func(a, b Field) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return
}

func getSearchAnnotation(field *load.Field) *entx.SearchFieldAnnotation {
	searchAnt := &entx.SearchFieldAnnotation{}
	if ant, ok := field.Annotations[searchAnt.Name()]; ok {
		if err := searchAnt.Decode(ant); err != nil {
			return nil
		}

		return searchAnt
	}

	return nil
}

// isFieldSearchable checks if the field has the SearchField annotation
func isFieldSearchable(field *load.Field) bool {
	searchAnt := getSearchAnnotation(field)

	if searchAnt == nil {
		return false
	}

	return searchAnt.Searchable
}

// isAdminFieldSearchable checks if the field has the admin SearchField annotation
// it also checks for the SkipWhereInput annotation to exclude the field from the search schema
func isAdminFieldSearchable(field *load.Field) bool {
	if entSkip(field) {
		return false
	}

	searchAnt := getSearchAnnotation(field)

	if searchAnt == nil {
		return true
	}

	return !searchAnt.ExcludeAdmin
}

// getPathAnnotation checks if the field has the JSONPath field set on the SearchField annotation
func getPathAnnotation(field *load.Field) string {
	searchAnt := getSearchAnnotation(field)

	if searchAnt == nil {
		return ""
	}

	return searchAnt.JSONPath
}

// getDotPathAnnotation checks if the field has the JSONDotPath field set on the SearchField annotation
func getDotPathAnnotation(field *load.Field) string {
	searchAnt := getSearchAnnotation(field)

	if searchAnt == nil {
		return ""
	}

	return searchAnt.JSONDotPath
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

// getFileName returns the file path for the search schema file
func getSearchFileName(isAdmin bool) string {
	fileName := "search"
	if isAdmin {
		fileName = "admin" + fileName
	}

	return fileName
}

// entSkip checks if the field has the entgql.Skip annotation that would exclude the field from being searchable
func entSkip(field *load.Field) bool {
	// never include sensitive fields
	if field.Sensitive {
		return true
	}

	entAnt := &entgql.Annotation{}
	if ant, ok := field.Annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			switch {
			case entAnt.Skip.Is(entgql.SkipType):
				return true
			case entAnt.Skip.Is(entgql.SkipWhereInput):
				return true
			}
		}
	}

	return false
}

// isFieldTypeExcluded checks if the field type should be excluded from being searchable
func isFieldTypeExcluded(field *load.Field) bool {
	// include the following field types to be searchable
	includedTypes := []entfield.Type{
		entfield.TypeString,
		entfield.TypeJSON,
		entfield.TypeOther,
	}

	return !slices.Contains(includedTypes, field.Info.Type)
}
