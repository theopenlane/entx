package enthistory

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"entgo.io/ent/entc/gen"
	"github.com/stoewer/go-strcase"
	"golang.org/x/tools/imports"
)

const (
	templateDir = "templates"
)

// extractUpdatedByKey gets the key that is used for the updated_by field
func extractUpdatedByKey(val any) string {
	updatedBy, ok := val.(*UpdatedBy)
	if !ok || updatedBy == nil {
		return ""
	}

	return updatedBy.key
}

// extractUpdatedByValueType gets the type (int or string) that the update_by
// field uses
func extractUpdatedByValueType(val any) string {
	updatedBy, ok := val.(*UpdatedBy)
	if !ok || updatedBy == nil {
		return ""
	}

	switch updatedBy.valueType {
	case ValueTypeInt:
		return "int"
	case ValueTypeString:
		return "string"
	default:
		return ""
	}
}

// fieldPropertiesNillable checks the config properties for the Nillable setting
func fieldPropertiesNillable(config Config) bool {
	return config.FieldProperties != nil && config.FieldProperties.Nillable
}

// isSlice checks if the string value of the type is prefixed with []
func isSlice(typeString string) bool {
	return strings.HasPrefix(typeString, "[]")
}

// in checks a string slice of a given string and returns true, if found
func in(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}

	return false
}

// parseTemplate parses the template and sets values in the template
func parseTemplate(name, path string) *gen.Template {
	t := gen.NewTemplate(name)
	t.Funcs(template.FuncMap{
		"extractUpdatedByKey":       extractUpdatedByKey,
		"extractUpdatedByValueType": extractUpdatedByValueType,
		"fieldPropertiesNillable":   fieldPropertiesNillable,
		"isSlice":                   isSlice,
		"in":                        in,
	})

	return gen.MustParse(t.ParseFS(_templates, path))
}

// parseSchemaTemplate parses the template and sets values in the template
func parseSchemaTemplate(info templateInfo, path string) error {
	name := "schema"
	templateName := fmt.Sprintf("%s.tmpl", name)

	t := template.New("schema")
	t.Funcs(template.FuncMap{
		"ToUpperCamel": strcase.UpperCamelCase,
		"ToLower":      strings.ToLower,
	})

	template.Must(t.ParseFS(_templates, fmt.Sprintf("%s/%s", templateDir, templateName)))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, templateName, info); err != nil {
		return fmt.Errorf("%w: failed to execute template: %v", ErrFailedToGenerateTemplate, err)
	}

	return writeAndFormatFile(buf, path)
}

// writeAndFormatFile formats the bytes using gofmt and goimports and writes them to the output file
func writeAndFormatFile(buf bytes.Buffer, outputPath string) error {
	// run gofmt and goimports on the file contents
	formatted, err := imports.Process(outputPath, buf.Bytes(), nil)
	if err != nil {
		return fmt.Errorf("%w: failed to format file: %v", ErrFailedToWriteTemplate, err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("%w: failed to create file: %v", ErrFailedToWriteTemplate, err)
	}

	// write the formatted source to the file
	if _, err := file.Write(formatted); err != nil {
		return fmt.Errorf("%w: failed to write to file: %v", ErrFailedToWriteTemplate, err)
	}

	return nil
}
