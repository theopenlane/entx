package genhooks

import (
	"fmt"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/stoewer/go-strcase"
	"github.com/vektah/gqlparser/v2/ast"
)

// WithStringSliceWhereOps is a schema hook that adds "Has" filter operations for fields of type []string in the generated GraphQL where input types. For example, if a node has a field "Tags" of type []string, this hook will add a "TagsHas" field to the corresponding where input type, allowing users to filter based on whether the "Tags" field contains a specific value.
func WithStringSliceWhereOps() entgql.SchemaHook {
	return func(g *gen.Graph, s *ast.Schema) error {
		for _, n := range g.Schemas {
			gqlTypeName := n.Name

			if entSkipWhere(n) {
				continue
			}

			whereName := gqlTypeName + "WhereInput"
			whereDef := s.Types[whereName]
			if whereDef == nil || whereDef.Kind != ast.InputObject {
				continue
			}

			existing := map[string]bool{}
			for _, f := range whereDef.Fields {
				existing[f.Name] = true
			}

			for _, f := range n.Fields {
				if f.Info == nil || f.Info.Ident != "[]string" {
					continue
				}

				if entSkipWhere(f) {
					continue
				}

				fieldName := strcase.LowerCamelCase(f.Name)

				addInputField(whereDef, existing, fieldName+"Has", ast.NamedType("String", nil))
			}
		}
		return nil
	}
}

// entSkipWhere checks if the field has entgql.SkipWhereInput, entgql.SkipType which indicate that the field should be skipped when generating where input filter types
func entSkipWhere[T any](t T) bool {
	entAnt := &entgql.Annotation{}
	annotations := map[string]any{}

	switch v := any(t).(type) {
	case *load.Schema:
		annotations = v.Annotations
	case *load.Field:
		annotations = v.Annotations
	case *gen.Type:
		annotations = v.Annotations
	case *gen.Field:
		annotations = v.Annotations
	}

	if ant, ok := annotations[entAnt.Name()]; ok {
		if err := entAnt.Decode(ant); err == nil {
			switch {
			case entAnt.Skip.Is(entgql.SkipWhereInput):
				return true
			case entAnt.Skip.Is(entgql.SkipType):
				return true
			}
		}
	}

	return false
}

// addInputField adds a new field to the input definition if it doesn't already exist
func addInputField(def *ast.Definition, existing map[string]bool, name string, t *ast.Type) {
	if existing[name] {
		return
	}

	def.Fields = append(def.Fields, &ast.FieldDefinition{
		Name:        name,
		Type:        t,
		Description: fmt.Sprintf("Filter for %s to contain a specific value", name),
	})

	existing[name] = true
}
