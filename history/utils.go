package enthistory

import (
	"regexp"
	"strings"

	"entgo.io/ent/schema/field"

	"entgo.io/ent/entc/load"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// toSnakeCase converts a string to snake_case formatting
func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}

// copyRef makes a copy of a pointer reference
// nolint:unused
func copyRef[T any](ref *T) *T {
	if ref == nil {
		return nil
	}

	val := *ref

	return &val
}

// loadHistorySchema with provided id type of string or int
func loadHistorySchema(idType string) (*load.Schema, error) {
	schema := history{}

	switch idType {
	case "int":
		schema.ref = field.Int("ref").Immutable().Optional()
	case "string":
		schema.ref = field.String("ref").Immutable().Optional()
	default:
		return nil, ErrUnsupportedIDType
	}

	bytes, err := load.MarshalSchema(schema)
	if err != nil {
		return nil, err
	}

	historySchema, err := load.UnmarshalSchema(bytes)
	if err != nil {
		return nil, err
	}

	return historySchema, nil
}

// getUpdatedByField sets the updateBy field type to string or int, depending on the user id type
func getUpdatedByField(updatedByValueType string) (*load.Field, error) {
	if strings.ToLower(updatedByValueType) == "string" {
		return load.NewField(field.String("updated_by").Optional().Nillable().Immutable().Descriptor())
	}

	if strings.ToLower(updatedByValueType) == "int" {
		return load.NewField(field.Int("updated_by").Optional().Nillable().Immutable().Descriptor())
	}

	return nil, ErrUnsupportedType
}

// getHistoryAnnotations loads the annotations from the schema to reference if the schema
// is a history schema, or if it should be excluded entirely from history schemas, or neither
// meaning a history schema should be created for that schema
func getHistoryAnnotations(schema *load.Schema) Annotations {
	annotations := Annotations{}

	if historyAnnotations, ok := schema.Annotations["History"].(map[string]any); ok {
		if exclude, ok := historyAnnotations["exclude"].(bool); ok {
			annotations.Exclude = exclude
		}

		if isHistory, ok := historyAnnotations["isHistory"].(bool); ok {
			annotations.IsHistory = isHistory
		}
	}

	return annotations
}

// getSchemaTableName from the entSQL annotation
func getSchemaTableName(schema *load.Schema) string {
	if entSQLMap, ok := schema.Annotations["EntSQL"].(map[string]any); ok {
		if table, ok := entSQLMap["table"].(string); ok && table != "" {
			return table
		}
	}

	return toSnakeCase(schema.Name)
}

// getPkgFromSchemaPath returns the package from the schema path
func getPkgFromSchemaPath(schemaPath string) (string, error) {
	parts := strings.Split(schemaPath, "/")
	lastPart := parts[len(parts)-1]

	if len(lastPart) == 0 {
		return "", ErrInvalidSchemaPath
	}

	return lastPart, nil
}

// getIDType returns the id type, defaulting to a string
func getIDType(idType string) string {
	switch strings.ToLower(idType) {
	case "int":
		return "int"
	case "string":
		return "string"
	default:
		return "string"
	}
}
