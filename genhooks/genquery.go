package genhooks

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
	"github.com/vektah/gqlparser/v2/parser"
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

			schemaDir := "./internal/graphapi/schema/"

			schema, err := loadSchemasFromDir(schemaDir)
			if err != nil {
				panic(err)
			}

			// schemaToFields collects the extended flat fields for each schema
			schemaToFields := mapFieldsToSchema(schema)

			// loop through all nodes and generate schema if not specified to be skipped
			for _, node := range g.Nodes {
				generateQuery(schemaToFields[strings.ToLower(node.Name)], node, tmpl, graphSchemaDir)
			}

			return next.Generate(g)
		})
	}
}

// mapFieldsToSchema creates a mapping of schema to the fields that should be avoided deleting in the query update process
func mapFieldsToSchema(schema *ast.Schema) map[string]map[string]bool {
	schemaToFields := make(map[string]map[string]bool)

	for _, field := range schema.Types["Mutation"].Fields {
		path := field.Position.Src.Name

		schemaName := strings.ToLower(path[strings.LastIndex(path, "/")+1 : strings.LastIndex(path, ".graphql")])

		for _, flatFields := range schema.Types[field.Type.Name()].Fields {
			if schemaToFields[schemaName] == nil {
				schemaToFields[schemaName] = make(map[string]bool)
			}

			schemaToFields[schemaName][flatFields.Name] = true
		}
	}
	return schemaToFields
}

// generateQuery generates the query file for the type
func generateQuery(fieldsToAvoidDeleting map[string]bool, node *gen.Type, tmpl *template.Template, graphSchemaDir string) {
	// check skip annotation
	if checkQueryGenSkip(node) {
		return
	}

	filePath := getFileName(graphSchemaDir, node.Name)

	// check if schema already exists,update query to include manual changes to flat fields
	if _, err := os.Stat(filePath); err == nil {
		if err := updateQuery(fieldsToAvoidDeleting, filePath, node, tmpl); err != nil {
			log.Fatalf("unable to update file: %v", err)
		}

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

// loadSchemasFromDir loads all the schemas from a root directory
func loadSchemasFromDir(dir string) (*ast.Schema, error) {
	var sources []*ast.Source

	fsys := os.DirFS(dir)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".graphql" {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}

		sources = append(sources, &ast.Source{
			Name:  path,
			Input: string(content),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return gqlparser.LoadSchema(sources...)
}

// updateQuery updates the query by keeping old edges and updating flat fields
func updateQuery(fieldsToAvoidDeleting map[string]bool, filePath string, node *gen.Type, tmpl *template.Template) error {
	// Read file contents and parses for comparison and updating
	srcFile, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("unable to read existing file: %w", err)
	}

	oldDoc, err := parser.ParseQuery(&ast.Source{
		Name:  filePath,
		Input: string(srcFile),
	})
	if err != nil {
		return fmt.Errorf("unable to parse existing query file: %w", err)
	}

	// Load query selections into a map for easy access
	oldQuerySelections := make(map[string]*ast.OperationDefinition)

	for _, op := range oldDoc.Operations {
		oldQuerySelections[op.Name] = op
	}

	// Load new query into memory for comparison
	var buf bytes.Buffer

	s := query{
		Name:             node.Name,
		Fields:           getFieldNames(node.Fields),
		IncludeMutations: checkEntqlMutation(node),
		IsHistory:        isHistorySchema(node),
	}

	if err = tmpl.Execute(&buf, s); err != nil {
		return fmt.Errorf("unable to execute query template: %w", err)
	}

	newDoc, err := parser.ParseQuery(&ast.Source{
		Input: buf.String(),
	})
	if err != nil {
		return fmt.Errorf("unable to parse new query file: %w", err)
	}

	// Load new query selections into a map for easy access
	newQuerySelections := make(map[string]*ast.OperationDefinition)

	for _, op := range newDoc.Operations {
		newQuerySelections[op.Name] = op
	}

	// track query names we care about.
	newQueryKeys := make([]string, 0, len(newQuerySelections))

	for queryName, oldQuery := range oldQuerySelections {
		newQuery, ok := newQuerySelections[queryName]

		newQueryKeys = append(newQueryKeys, queryName)

		// if query doesn't exist in new queries, we add the old query to the newly created document
		if !ok {
			newQuerySelections[queryName] = oldQuery
			continue
		}

		// if the new query signature parameters differ from the old query's, keep the old query.
		if !compareSignatureParams(oldQuery.VariableDefinitions, newQuery.VariableDefinitions) {
			newQuerySelections[queryName] = oldQuery
			continue
		}

		// merge edges recursively
		writeMissingFields(oldQuery.SelectionSet, &newQuery.SelectionSet, fieldsToAvoidDeleting)
	}

	// sort keys for consistency in output file, this prevents code from thinking file has changed.
	sort.Strings(newQueryKeys)

	const filePerm = 0644

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, filePerm) //nolint: gosec
	if err != nil {
		return fmt.Errorf("unable to open file for writing: %w", err)
	}
	defer f.Close()

	updatedDoc := &ast.QueryDocument{Operations: ast.OperationList{}}

	// add all new and updated queries to the document
	for _, keyName := range newQueryKeys {
		updatedDoc.Operations = append(updatedDoc.Operations, newQuerySelections[keyName])
	}

	formatter.NewFormatter(f).FormatQueryDocument(updatedDoc)

	return nil
}

// compareSelectionSets compares the parameters between two querys, returns true if equal, false otherwise
func compareSignatureParams(list1 ast.VariableDefinitionList, list2 ast.VariableDefinitionList) bool {
	if len(list1) != len(list2) {
		return false
	}

	m := make(map[string]string)

	for _, v := range list1 {
		m[v.Variable] = v.Type.String()
	}

	for _, v := range list2 {
		if typ, ok := m[v.Variable]; !ok || typ != v.Type.String() {
			return false
		}
	}

	return true
}

// writeMissingFields adds edges and flat fields from oldSel into newSel
func writeMissingFields(oldSel ast.SelectionSet, newSel *ast.SelectionSet, fieldsToAvoidDeleting map[string]bool) {
	for _, oldSelection := range oldSel {
		oldField, ok := oldSelection.(*ast.Field)
		if !ok {
			continue
		}

		// if this is false, the field was effectively deleted in the new query
		if !isEdge(oldField) && !fieldsToAvoidDeleting[oldField.Name] {
			continue
		}

		newField := findFieldInSelectionSet(*newSel, oldField.Name)

		fieldCopy := *oldField

		// if missing entirely, add it
		if newField == nil {
			*newSel = append(*newSel, &fieldCopy)
			continue
		}

		newField.Name = oldField.Name
		newField.Alias = oldField.Alias

		// only recurse if it's actually an edge
		if isEdge(oldField) {
			writeMissingFields(oldField.SelectionSet, &newField.SelectionSet, fieldsToAvoidDeleting)
		}
	}
}

// findFieldInSelectionSet goes through the fields and checks field by name, returns if found
func findFieldInSelectionSet(sel ast.SelectionSet, name string) *ast.Field {
	for _, s := range sel {
		if f, ok := s.(*ast.Field); ok {
			if strings.EqualFold(f.Name, name) {
				return f
			}
		}
	}

	return nil
}

// isEdge checks if field is an edge
func isEdge(f *ast.Field) bool {
	return len(f.SelectionSet) > 0
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

		firstWord := GetFirstWord(f.Name)

		filteredName := f.StructField()

		fieldNames = append(fieldNames, lowerSubstring(filteredName, len(firstWord)))
	}

	// sort field names
	sort.Strings(fieldNames)

	return fieldNames
}

// lowerSubstring lowers the a substring from s, where the portion lowored starts from the first character
// and lowers wordLen characters in s
func lowerSubstring(s string, wordLen int) string {
	return strings.ToLower(s[:wordLen]) + s[wordLen:]
}

// isSeparator checks if character is a underscore, hyphen or a space
func isSeparator(r rune) bool {
	return r == '_' || r == '-' || unicode.IsSpace(r)
}

// getFirstWord returns the first work of the string before the separator
func GetFirstWord(name string) string {
	words := strings.FieldsFunc(name, isSeparator)
	return words[0]
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
