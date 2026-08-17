package entityops

import (
	"bytes"
	"go/parser"
	"go/token"
	"strings"
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

// TestCapabilityGatedEmission verifies the registry and projection templates emit runtime
// closures and projections only for schemas whose capabilities can reach them: mapped schemas
// get create/update/query, workflow-eligible schemas get load-object, link targets get query
// and a projection, and capability-free schemas get only the universal load and catalogs
func TestCapabilityGatedEmission(t *testing.T) {
	data := EntityData{
		PackageName:  "entityops",
		EntPackage:   "example.com/app/ent/generated",
		GalaPackage:  "example.com/app/pkg/gala",
		JsonxPackage: "example.com/app/pkg/jsonx",
		LogxPackage:  "example.com/app/pkg/logx",
		CelxPackage:  "example.com/app/pkg/celx",
		MapxPackage:  "example.com/app/pkg/mapx",
		Schemas: []EntitySchema{
			{
				Name: "Asset", Snake: "asset", Lower: "asset",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateAssetInput", UpdateInputType: "UpdateAssetInput",
				PredicatePackage: "asset", PredicateImport: "example.com/app/ent/generated/asset",
				IntegrationMapped: true,
				ObjectFields:      []EntityField{{Name: "ExternalID", Snake: "external_id", Type: "string", MatchKey: true, IntegrationMapped: true, InputKey: "external_id", LookupKey: true}},
				Edges:             []EntityEdge{{Name: "controls", TargetSchema: "Control", TargetInRegistry: true}},
			},
			{
				Name: "Control", Snake: "control", Lower: "control",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateControlInput", UpdateInputType: "UpdateControlInput",
				PredicatePackage: "control", PredicateImport: "example.com/app/ent/generated/control",
				LinkTarget:       true,
				WorkflowEligible: true,
				ObjectFields:     []EntityField{{Name: "RefCode", Snake: "ref_code", Type: "string", MatchKey: true, Projectable: true, WorkflowEligible: true}},
			},
			{
				Name: "Note", Snake: "note", Lower: "note",
				HasCreate: true, HasUpdate: true, HasOwnerID: true,
				CreateInputType: "CreateNoteInput", UpdateInputType: "UpdateNoteInput",
				PredicatePackage: "note", PredicateImport: "example.com/app/ent/generated/note",
				ObjectFields: []EntityField{{Name: "Text", Snake: "text", Type: "string", Projectable: true}},
			},
		},
	}

	render := func(name string) string {
		raw, err := _templates.ReadFile("templates/" + name + ".tpl")
		require.NoError(t, err)

		tmpl, err := template.New(name).Funcs(gen.Funcs).Parse(string(raw))
		require.NoError(t, err)

		var buf bytes.Buffer
		require.NoError(t, tmpl.Execute(&buf, data))

		_, err = parser.ParseFile(token.NewFileSet(), name+".go", buf.Bytes(), parser.AllErrors)
		require.NoError(t, err, buf.String())

		return buf.String()
	}

	registry := render("entity_registry")

	asset := registry[strings.Index(registry, "SchemaAsset = &Schema{"):strings.Index(registry, "SchemaControl = &Schema{")]
	require.Contains(t, asset, "Create: func")
	require.Contains(t, asset, "Update: func")
	require.Contains(t, asset, "Query: func")
	require.Contains(t, asset, "Ingest: &IngestCapability")
	require.NotContains(t, asset, "LoadObject: func")

	control := registry[strings.Index(registry, "SchemaControl = &Schema{"):strings.Index(registry, "SchemaNote = &Schema{")]
	require.NotContains(t, control, "Create: func")
	require.Contains(t, control, "Update: func")
	require.Contains(t, control, "Query: func")
	require.Contains(t, control, "LoadObject: func")

	note := registry[strings.Index(registry, "SchemaNote = &Schema{"):strings.Index(registry, "// init wires")]
	require.Contains(t, note, "Load: func")
	require.NotContains(t, note, "Create: func")
	require.NotContains(t, note, "Update: func")
	require.NotContains(t, note, "Query: func")
	require.NotContains(t, note, "LoadObject: func")
	require.NotContains(t, note, "ProjectionType")

	require.Contains(t, registry, "SchemaAsset.QueryByKey = func")
	require.Contains(t, registry, "SchemaControl.QueryByKey = func")
	require.NotContains(t, registry, "SchemaNote.QueryByKey")
	require.NotContains(t, registry, "LoadMany")

	projections := render("entity_projection")
	require.Contains(t, projections, "type AssetProjection struct")
	require.Contains(t, projections, "type ControlProjection struct")
	require.NotContains(t, projections, "NoteProjection")
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
