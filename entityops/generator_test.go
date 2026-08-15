package entityops

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
	"text/template"

	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/require"
)

// TestStaticTemplatesRenderValidGo verifies the schema-independent templates render to
// syntactically valid Go for a fully configured generator
func TestStaticTemplatesRenderValidGo(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		MapxPackage:  "example.com/app/pkg/mapx",
	}

	for _, name := range []string{
		"entity_schema", "entity_errors", "entity_registry", "entity_workflow", "entity_tasks",
		"entity_links", "entity_integration", "entity_projection", "entity_metadata",
		"entity_changeset", "entity_mutation_events", "entity_listener",
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := _templates.ReadFile("templates/" + name + ".tpl")
			require.NoError(t, err)

			tmpl, err := template.New(name).Funcs(gen.Funcs).Parse(string(raw))
			require.NoError(t, err)

			var buf bytes.Buffer
			require.NoError(t, tmpl.Execute(&buf, data))

			_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
			require.NoError(t, err, buf.String())
		})
	}
}

// TestEntityMetadataTemplate verifies the metadata template renders console paths and
// mention specs into compilable map literals
func TestEntityMetadataTemplate(t *testing.T) {
	raw, err := _templates.ReadFile("templates/entity_metadata.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_metadata").Funcs(gen.Funcs).Parse(string(raw))
	require.NoError(t, err)

	data := EntityData{
		PackageName: "entityops",
	}

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, data))

	rendered := buf.String()
	require.Contains(t, rendered, "func ConsoleLanding(schemaType string) string")
	require.Contains(t, rendered, "func ConsoleObjectPath(schemaType, objectID string) string")
	require.Contains(t, rendered, "func MentionSpecFor(schemaType string) (MentionSpec, bool)")
}

// TestEntityMetadataTemplateEmpty verifies the metadata template renders with no annotated
// schemas, since most consuming repos start with empty metadata
func TestEntityMetadataTemplateEmpty(t *testing.T) {
	raw, err := _templates.ReadFile("templates/entity_metadata.tpl")
	require.NoError(t, err)

	tmpl, err := template.New("entity_metadata").Funcs(gen.Funcs).Parse(string(raw))
	require.NoError(t, err)

	var buf bytes.Buffer
	require.NoError(t, tmpl.Execute(&buf, EntityData{PackageName: "entityops"}))

	require.Contains(t, buf.String(), "schema, ok := LookupSchema(schemaType)")
}
