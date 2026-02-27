package oscalgen

import (
	"os"
	"path/filepath"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectOSCALSchemas validates schema-level OSCAL graph collection
func TestCollectOSCALSchemas(t *testing.T) {
	graph := &gen.Graph{
		Schemas: []*load.Schema{
			{
				Name: "Platform",
				Annotations: map[string]any{
					OSCALModelAnnotationName: map[string]any{
						"Models":   []string{"ssp", "component-definition"},
						"Assembly": "system-characteristics",
					},
				},
				Fields: []*load.Field{
					{
						Name: "display_name",
						Annotations: map[string]any{
							OSCALFieldAnnotationName: map[string]any{
								"Role": "title",
							},
						},
					},
					{
						Name: "uuid",
						Annotations: map[string]any{
							OSCALFieldAnnotationName: map[string]any{
								"Role":           "uuid",
								"IdentityAnchor": true,
								"Models":         []string{"ssp"},
							},
						},
					},
				},
				Edges: []*load.Edge{
					{
						Name: "narratives",
						Annotations: map[string]any{
							OSCALRelationshipAnnotationName: map[string]any{
								"Role": "satisfies-control",
							},
						},
					},
				},
			},
			{
				Name: "User",
				Fields: []*load.Field{
					{Name: "email"},
				},
			},
		},
	}

	generator := NewOSCALGenerator("./schema", t.TempDir())
	schemas := generator.collectOSCALSchemas(graph)
	require.Len(t, schemas, 1)

	platform := schemas[0]
	assert.Equal(t, "Platform", platform.name)
	assert.Equal(t, []string{"component-definition", "ssp"}, platform.models)
	assert.Equal(t, "system-characteristics", platform.assembly)

	require.Len(t, platform.fields, 2)
	assert.Equal(t, "display_name", platform.fields[0].name)
	assert.Equal(t, "title", platform.fields[0].role)
	assert.Equal(t, "uuid", platform.fields[1].name)
	assert.Equal(t, "uuid", platform.fields[1].role)
	assert.Equal(t, []string{"ssp"}, platform.fields[1].models)
	assert.True(t, platform.fields[1].identityAnchor)

	require.Len(t, platform.relationships, 1)
	assert.Equal(t, "narratives", platform.relationships[0].name)
	assert.Equal(t, "satisfies-control", platform.relationships[0].role)
}

// TestGenerateOSCALRegistryFile validates OSCAL registry file generation
func TestGenerateOSCALRegistryFile(t *testing.T) {
	outputDir := t.TempDir()
	generator := NewOSCALGenerator("./schema", outputDir).WithPackage("testpkg")

	schemas := []oscalSchemaInfo{
		{
			name:     "Platform",
			models:   []string{"component-definition", "ssp"},
			assembly: "system-characteristics",
			fields: []oscalFieldInfo{
				{
					name:           "uuid",
					role:           "uuid",
					models:         []string{"ssp"},
					identityAnchor: true,
				},
			},
			relationships: []oscalRelationshipInfo{
				{
					name: "narratives",
					role: "satisfies-control",
				},
			},
		},
	}

	err := generator.generateRegistryFile(schemas)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(outputDir, "oscal_generated.go"))
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package testpkg")
	assert.Contains(t, contentStr, "var OSCALMappings = map[string]OSCALSchemaMapping{")
	assert.Contains(t, contentStr, "\"platform\":")
	assert.Contains(t, contentStr, "GetOSCALSchemaMapping")
	assert.Contains(t, contentStr, "SchemaSupportsOSCALModel")
	assert.Contains(t, contentStr, "GetOSCALFieldMapping")
	assert.Contains(t, contentStr, "GetOSCALRelationshipMapping")
	assert.Contains(t, contentStr, "IsOSCALIdentityAnchor")
}
