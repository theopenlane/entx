package genhooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/assert"
)

func TestGenWorkflowSchema(t *testing.T) {
	tests := []struct {
		name                string
		setupSchema         func() *gen.Graph
		existingSchemaFile  string
		expectedContains    []string
		expectedNotContains []string
		shouldModify        bool
	}{
		{
			name: "adds workflow interface to entity with proposed_changes field",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "TestEntity",
							Fields: []*gen.Field{
								{Name: "proposed_changes"},
								{Name: "name"},
							},
						},
					},
				}
			},
			existingSchemaFile: `extend type Query {
    testEntity(id: ID!): TestEntity!
}
`,
			expectedContains: []string{
				"extend type TestEntity {",
				"hasPendingWorkflow: Boolean!",
				"activeWorkflowInstance: WorkflowInstance",
				"extend type Query",
			},
			shouldModify: true,
		},
		{
			name: "skips entity without proposed_changes field",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "SimpleEntity",
							Fields: []*gen.Field{
								{Name: "name"},
								{Name: "description"},
							},
						},
					},
				}
			},
			existingSchemaFile: `extend type Query {
    simpleEntity(id: ID!): SimpleEntity!
}
`,
			expectedNotContains: []string{
				"WorkflowEnabled",
				"hasPendingWorkflow",
				"activeWorkflowInstance",
			},
			shouldModify: false,
		},
		{
			name: "skips history types",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "EntityHistory",
							Fields: []*gen.Field{
								{Name: "proposed_changes"},
							},
							Annotations: gen.Annotations{
								"History": map[string]interface{}{
									"isHistory": true,
								},
							},
						},
					},
				}
			},
			existingSchemaFile: `extend type Query {
    entityHistory(id: ID!): EntityHistory!
}
`,
			expectedNotContains: []string{
				"WorkflowEnabled",
				"hasPendingWorkflow",
			},
			shouldModify: false,
		},
		{
			name: "skips if workflow fields already present",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "ExistingEntity",
							Fields: []*gen.Field{
								{Name: "proposed_changes"},
							},
						},
					},
				}
			},
			existingSchemaFile: `extend type ExistingEntity {
    hasPendingWorkflow: Boolean!
    activeWorkflowInstance: WorkflowInstance
}

extend type Query {
    existingEntity(id: ID!): ExistingEntity!
}
`,
			expectedContains: []string{
				"extend type ExistingEntity {",
			},
			expectedNotContains: []string{
				// Should not duplicate the interface
			},
			shouldModify: false,
		},
		{
			name: "handles multiple entities correctly",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "WorkflowEntity1",
							Fields: []*gen.Field{
								{Name: "proposed_changes"},
							},
						},
						{
							Name: "WorkflowEntity2",
							Fields: []*gen.Field{
								{Name: "proposed_changes"},
							},
						},
						{
							Name: "RegularEntity",
							Fields: []*gen.Field{
								{Name: "name"},
							},
						},
					},
				}
			},
			existingSchemaFile: "",
			shouldModify:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			schemaDir := filepath.Join(tmpDir, "schema")
			err := os.Mkdir(schemaDir, 0755)
			assert.NoError(t, err)

			graph := tt.setupSchema()

			// Create schema files for each node if there's an existing schema
			if tt.existingSchemaFile != "" {
				for _, node := range graph.Nodes {
					fileName := getFileName(schemaDir, node.Name)
					err := os.WriteFile(fileName, []byte(tt.existingSchemaFile), 0600)
					assert.NoError(t, err)
				}
			}

			// Run the hook
			hook := GenWorkflowSchema(schemaDir)
			gen := hook(mockGenerator{})
			err = gen.Generate(graph)
			assert.NoError(t, err)

			// Verify the output for each node
			for _, node := range graph.Nodes {
				fileName := getFileName(schemaDir, node.Name)

				// Check if file exists
				_, err := os.Stat(fileName)

				if tt.existingSchemaFile == "" {
					// File shouldn't be created if it didn't exist
					continue
				}

				assert.NoError(t, err)

				content, err := os.ReadFile(fileName)
				assert.NoError(t, err)

				contentStr := string(content)

				// Check expected content
				for _, expected := range tt.expectedContains {
					assert.Contains(t, contentStr, expected, "expected content should be present")
				}

				for _, notExpected := range tt.expectedNotContains {
					assert.NotContains(t, contentStr, notExpected, "unexpected content should not be present")
				}

				// Verify no duplication if workflow fields were already present
				if strings.Contains(tt.existingSchemaFile, "hasPendingWorkflow") {
					count := strings.Count(contentStr, "hasPendingWorkflow")
					assert.Equal(t, 1, count, "workflow fields should not be duplicated")
				}
			}
		})
	}
}

func TestContainsWorkflowFields(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "contains hasPendingWorkflow",
			content:  "hasPendingWorkflow: Boolean!",
			expected: true,
		},
		{
			name:     "contains activeWorkflowInstance",
			content:  "activeWorkflowInstance: WorkflowInstance",
			expected: true,
		},
		{
			name:     "does not contain any workflow fields",
			content:  "extend type Entity { name: String! }",
			expected: false,
		},
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "contains partial match",
			content:  "implements Something",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsWorkflowFields(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateWorkflowSchemaTemplate(t *testing.T) {
	tmpl := createWorkflowSchemaTemplate()
	assert.NotNil(t, tmpl)

	// Test template execution
	var buf strings.Builder

	data := workflowSchema{
		Name:        "TestEntity",
		HasWorkflow: true,
	}

	err := tmpl.Execute(&buf, data)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "extend type TestEntity {")
	assert.Contains(t, output, "hasPendingWorkflow: Boolean!")
	assert.Contains(t, output, "activeWorkflowInstance: WorkflowInstance")
	assert.Contains(t, output, "testEntity")
}

// mockGenerator is a mock implementation of gen.Generator for testing
type mockGenerator struct{}

func (m mockGenerator) Generate(*gen.Graph) error {
	return nil
}
