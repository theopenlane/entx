package genhooks

import (
	"bytes"
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

	// check if schema already exists,update query to include manual changes to flat fields
	if _, err := os.Stat(filePath); err == nil {
		updateQuery(filePath, node, tmpl)
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

func updateQuery(filePath string, node *gen.Type, tmpl *template.Template) error {
	// Read file contents and parses for comparison and updating
	srcFile, err := os.ReadFile(filePath)
	if err != nil {
		return. fmt.Errorf("unable to read existing file: %w", err)
	}

	doc, err := parser.ParseQuery(&ast.Source{
	....
		Name:  filePath,
		Input: string(srcFile),
	})
	if err != nil {
		log.Fatalf("Unable to parse existing query file: %v", err)
	}

	// Load query selections into a map for easy access
	oldQuerySelections := make(map[string]*ast.OperationDefinition)

	for _, op := range doc.Operations {
		oldQuerySelections[op.Name] = op
	}

	//load new query  into memory for comparison
	var buf bytes.Buffer

	s := query{
		Name:             node.Name,
		Fields:           getFieldNames(node.Fields),
		IncludeMutations: checkEntqlMutation(node),
		IsHistory:        isHistorySchema(node),
	}

	if err = tmpl.Execute(&buf, s); err != nil {
		log.Fatalf("Unable to execute template: %v", err)
	}

	newDoc, err := parser.ParseQuery(&ast.Source{
		Input: buf.String(),
	})
	if err != nil {
		log.Fatalf("Unable to parse new query: %v", err)
	}

	// Load new query selections into a map for easy access
	newQuerySelections := make(map[string]*ast.OperationDefinition)

	for _, op := range newDoc.Operations {
		newQuerySelections[op.Name] = op
	}

	for queryName, oldQuery := range oldQuerySelections {

		newQuery, ok := newQuerySelections[queryName]

		//if query doesn't exist in nw queries, we add the old query to the newly created document
		if !ok {
			newDoc.Operations = append(newDoc.Operations, oldQuery)
			continue
		}

		// merge edges recursively
		writeMissingEdges(oldQuery.SelectionSet, &newQuery.SelectionSet)
	}

	//sort keys for consitency in output file, this prevents code from thinking file has changed.

	newQueryKeys := make([]string, 0, len(newQuerySelections))

	for key := range newQuerySelections {
		newQueryKeys = append(newQueryKeys, key)
	}

	sort.Strings(newQueryKeys)

	const filePerm = 0644
	
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, filePerm) //nolint: gosec
	if err != nil {
		log.Fatalf("Unable to open file for writing: %v", err)
	}
	defer f.Close()

	updatedDoc := &ast.QueryDocument{Operations: ast.OperationList{}}

	// add all new and updated queries to the document
	for _, keyName := range newQueryKeys {
		updatedDoc.Operations = append(updatedDoc.Operations, newQuerySelections[keyName])
	}

	formatter.NewFormatter(f).FormatQueryDocument(updatedDoc)
}

func writeMissingEdges(oldSel ast.SelectionSet, newSel *ast.SelectionSet) {

	for _, oldSelection := range oldSel {

		oldField, ok := oldSelection.(*ast.Field)
		if !ok {
			continue
		}

		if !isEdge(oldField) {
			continue
		}

		newField := findFieldInSelectionSet(*newSel, oldField.Name)

		fieldCopy := *oldField
		//Edge is missing if nil
		if newField == nil {
			*newSel = append(*newSel, &fieldCopy)
			continue
		}

		// recurse if and only if the edge existed in both old and new
		writeMissingEdges(oldField.SelectionSet, &newField.SelectionSet)
	}
}

// findFieldInSelectionSet goes through the fields and checks field by name, returns if found
func findFieldInSelectionSet(sel ast.SelectionSet, name string) *ast.Field {
	for _, s := range sel {

		if f, ok := s.(*ast.Field); ok {
			if f.Name == name {
				return f
			}
		}
	}
	return nil
}

// checks if field is an edge
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
