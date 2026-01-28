package genhooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/assert"
)

func TestGenBulkSchema(t *testing.T) {
	tests := []struct {
		name                string
		setupSchema         func() *gen.Graph
		existingSchemaFile  string
		expectedContains    []string
		expectedNotContains []string
		shouldModify        bool
	}{
		{
			name: "adds missing bulk mutations when one already exists",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "Control",
						},
					},
				}
			},
			existingSchemaFile: `extend type Mutation {
    """
    Create multiple new controls via csv upload
    """
    createBulkCSVControl(
        """
        csv file containing values of the control
        """
        input: Upload!
    ): ControlBulkCreatePayload!
    """
    Delete multiple controls
    """
    deleteBulkControl(ids: [ID!]!): ControlBulkDeletePayload!
}
`,
			expectedContains: []string{
				"updateBulkControl(",
				"updateBulkCSVControl(",
				"deleteBulkControl(",
				"type ControlBulkUpdatePayload",
				"type ControlBulkDeletePayload",
				"updatedIDs: [ID!]",
				"deletedIDs: [ID!]!",
			},
			shouldModify: true,
		},
		{
			name: "skips schema with createBulkCSV but no existing bulk update/delete mutations",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "Policy",
						},
					},
				}
			},
			existingSchemaFile: `extend type Mutation {
    """
    Create multiple new policies via csv upload
    """
    createBulkCSVPolicy(
        """
        csv file containing values of the policy
        """
        input: Upload!
    ): PolicyBulkCreatePayload!
}
`,
			expectedNotContains: []string{
				"updateBulkPolicy(",
				"updateBulkCSVPolicy(",
				"deleteBulkPolicy(",
			},
			shouldModify: false,
		},
		{
			name: "skips schema without createBulkCSV",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "SimpleEntity",
						},
					},
				}
			},
			existingSchemaFile: `extend type Query {
    simpleEntity(id: ID!): SimpleEntity!
}

extend type Mutation {
    createSimpleEntity(input: CreateSimpleEntityInput!): SimpleEntityCreatePayload!
}
`,
			expectedNotContains: []string{
				"updateBulkSimpleEntity",
				"deleteBulkSimpleEntity",
				"BulkUpdatePayload",
				"BulkDeletePayload",
			},
			shouldModify: false,
		},
		{
			name: "does not duplicate existing bulk mutations",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "ActionPlan",
						},
					},
				}
			},
			existingSchemaFile: `extend type Mutation {
    createBulkCSVActionPlan(input: Upload!): ActionPlanBulkCreatePayload!
    updateBulkActionPlan(ids: [ID!]!, input: UpdateActionPlanInput!): ActionPlanBulkUpdatePayload!
    updateBulkCSVActionPlan(input: Upload!): ActionPlanBulkUpdatePayload!
    deleteBulkActionPlan(ids: [ID!]!): ActionPlanBulkDeletePayload!
}

type ActionPlanBulkUpdatePayload {
    actionPlans: [ActionPlan!]
    updatedIDs: [ID!]
}

type ActionPlanBulkDeletePayload {
    deletedIDs: [ID!]!
}
`,
			expectedContains: []string{
				"updateBulkActionPlan(",
				"updateBulkCSVActionPlan(",
				"deleteBulkActionPlan(",
			},
			expectedNotContains: []string{},
			shouldModify:        false,
		},
		{
			name: "adds missing mutations when some exist",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "Procedure",
						},
					},
				}
			},
			existingSchemaFile: `extend type Mutation {
    createBulkCSVProcedure(input: Upload!): ProcedureBulkCreatePayload!
    updateBulkProcedure(ids: [ID!]!, input: UpdateProcedureInput!): ProcedureBulkUpdatePayload!
}
`,
			expectedContains: []string{
				"updateBulkCSVProcedure(",
				"deleteBulkProcedure(",
				"type ProcedureBulkUpdatePayload",
				"type ProcedureBulkDeletePayload",
			},
			shouldModify: true,
		},
		{
			name: "skips types with SchemaGen skip annotation",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "InternalType",
							Annotations: gen.Annotations{
								"OPENLANE_SCHEMAGEN": map[string]any{
									"Skip": true,
								},
							},
						},
					},
				}
			},
			existingSchemaFile: `extend type Mutation {
    createBulkCSVInternalType(input: Upload!): InternalTypeBulkCreatePayload!
}
`,
			expectedNotContains: []string{
				"updateBulkInternalType",
				"deleteBulkInternalType",
			},
			shouldModify: false,
		},
		{
			name: "handles file not existing gracefully",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "NonExistent",
						},
					},
				}
			},
			existingSchemaFile: "", // No file will be created
			shouldModify:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			schemaDir := filepath.Join(tmpDir, "schema")
			err := os.Mkdir(schemaDir, 0755)
			assert.NoError(t, err)

			graph := tt.setupSchema()

			// Create schema files for each node if there's existing content
			if tt.existingSchemaFile != "" {
				for _, node := range graph.Nodes {
					fileName := getFileName(schemaDir, node.Name)
					err := os.WriteFile(fileName, []byte(tt.existingSchemaFile), 0600)
					assert.NoError(t, err)
				}
			}

			// Run the hook
			hook := GenBulkSchema(schemaDir)
			generator := hook(mockGenerator{})
			err = generator.Generate(graph)
			assert.NoError(t, err)

			// Verify the output for each node
			for _, node := range graph.Nodes {
				fileName := getFileName(schemaDir, node.Name)

				// Check if file exists
				_, err := os.Stat(fileName)

				if tt.existingSchemaFile == "" {
					// File shouldn't be created if it didn't exist
					assert.True(t, os.IsNotExist(err))
					continue
				}

				assert.NoError(t, err)

				content, err := os.ReadFile(fileName)
				assert.NoError(t, err)

				contentStr := string(content)

				// Check expected content
				for _, expected := range tt.expectedContains {
					assert.Contains(t, contentStr, expected, "expected content should be present: %s", expected)
				}

				for _, notExpected := range tt.expectedNotContains {
					assert.NotContains(t, contentStr, notExpected, "unexpected content should not be present: %s", notExpected)
				}

				// If should not modify, content should be the same
				if !tt.shouldModify {
					assert.Equal(t, tt.existingSchemaFile, contentStr, "content should not be modified")
				}
			}
		})
	}
}

func TestGenBulkSchemaWithInjectExistingDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "schema")
	err := os.Mkdir(schemaDir, 0755)
	assert.NoError(t, err)

	// Create a schema file that would normally be modified (has createBulkCSV and deleteBulk)
	existingContent := `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
    deleteBulkControl(ids: [ID!]!): ControlBulkDeletePayload!
}
`

	graph := &gen.Graph{
		Nodes: []*gen.Type{
			{Name: "Control"},
		},
	}

	fileName := getFileName(schemaDir, "Control")
	err = os.WriteFile(fileName, []byte(existingContent), 0600)
	assert.NoError(t, err)

	// Run the hook with injection disabled
	hook := GenBulkSchema(schemaDir, WithBulkSchemaInjectExisting(false))
	generator := hook(mockGenerator{})
	err = generator.Generate(graph)
	assert.NoError(t, err)

	// Verify the file was NOT modified
	content, err := os.ReadFile(fileName)
	assert.NoError(t, err)
	assert.Equal(t, existingContent, string(content), "content should not be modified when injection is disabled")

	// Verify that with injection enabled (default), it WOULD modify
	hook = GenBulkSchema(schemaDir)
	generator = hook(mockGenerator{})
	err = generator.Generate(graph)
	assert.NoError(t, err)

	content, err = os.ReadFile(fileName)
	assert.NoError(t, err)
	assert.NotEqual(t, existingContent, string(content), "content should be modified when injection is enabled")
	assert.Contains(t, string(content), "updateBulkControl(")
	assert.Contains(t, string(content), "updateBulkCSVControl(")
}

func TestInjectBulkMutations(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		schema           bulkSchema
		expectedContains []string
		shouldChange     bool
	}{
		{
			name: "injects all bulk mutations",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{
				"updateBulkControl(",
				"updateBulkCSVControl(",
				"deleteBulkControl(",
			},
			shouldChange: true,
		},
		{
			name: "skips if all mutations exist",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
    updateBulkControl(ids: [ID!]!, input: UpdateControlInput!): ControlBulkUpdatePayload!
    updateBulkCSVControl(input: Upload!): ControlBulkUpdatePayload!
    deleteBulkControl(ids: [ID!]!): ControlBulkDeletePayload!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{},
			shouldChange:     false,
		},
		{
			name:    "returns unchanged if no Mutation block",
			content: `type Query { control(id: ID!): Control! }`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{},
			shouldChange:     false,
		},
		{
			name: "injects only missing mutations",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
    updateBulkControl(ids: [ID!]!, input: UpdateControlInput!): ControlBulkUpdatePayload!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{
				"updateBulkCSVControl(",
				"deleteBulkControl(",
			},
			shouldChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectBulkMutations(tt.content, tt.schema)

			if tt.shouldChange {
				assert.NotEqual(t, tt.content, result, "content should be modified")
			} else {
				assert.Equal(t, tt.content, result, "content should not be modified")
			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "should contain: %s", expected)
			}
		})
	}
}

func TestInjectBulkPayloadTypes(t *testing.T) {
	tests := []struct {
		name             string
		content          string
		schema           bulkSchema
		expectedContains []string
		shouldChange     bool
	}{
		{
			name: "injects both payload types",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{
				"type ControlBulkUpdatePayload",
				"controls: [Control!]",
				"updatedIDs: [ID!]",
				"type ControlBulkDeletePayload",
				"deletedIDs: [ID!]!",
			},
			shouldChange: true,
		},
		{
			name: "skips if both types exist",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
}

type ControlBulkUpdatePayload {
    controls: [Control!]
    updatedIDs: [ID!]
}

type ControlBulkDeletePayload {
    deletedIDs: [ID!]!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{},
			shouldChange:     false,
		},
		{
			name: "injects only missing update payload",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
}

type ControlBulkDeletePayload {
    deletedIDs: [ID!]!
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{
				"type ControlBulkUpdatePayload",
			},
			shouldChange: true,
		},
		{
			name: "injects only missing delete payload",
			content: `extend type Mutation {
    createBulkCSVControl(input: Upload!): ControlBulkCreatePayload!
}

type ControlBulkUpdatePayload {
    controls: [Control!]
    updatedIDs: [ID!]
}`,
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expectedContains: []string{
				"type ControlBulkDeletePayload",
			},
			shouldChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := injectBulkPayloadTypes(tt.content, tt.schema)

			if tt.shouldChange {
				assert.NotEqual(t, tt.content, result, "content should be modified")
			} else {
				assert.Equal(t, tt.content, result, "content should not be modified")
			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected, "should contain: %s", expected)
			}
		})
	}
}

func TestRenderBulkUpdateMutation(t *testing.T) {
	s := bulkSchema{
		Name:       "Control",
		PluralName: "Controls",
	}

	result := renderBulkUpdateMutation(s)

	assert.Contains(t, result, "updateBulkControl(")
	assert.Contains(t, result, "ids: [ID!]!")
	assert.Contains(t, result, "input: UpdateControlInput!")
	assert.Contains(t, result, "ControlBulkUpdatePayload!")
	assert.Contains(t, result, "control")
}

func TestRenderBulkUpdateCSVMutation(t *testing.T) {
	s := bulkSchema{
		Name:       "Control",
		PluralName: "Controls",
	}

	result := renderBulkUpdateCSVMutation(s)

	assert.Contains(t, result, "updateBulkCSVControl(")
	assert.Contains(t, result, "input: Upload!")
	assert.Contains(t, result, "ControlBulkUpdatePayload!")
	assert.Contains(t, result, "csv file")
}

func TestRenderBulkDeleteMutation(t *testing.T) {
	s := bulkSchema{
		Name:       "Control",
		PluralName: "Controls",
	}

	result := renderBulkDeleteMutation(s)

	assert.Contains(t, result, "deleteBulkControl(")
	assert.Contains(t, result, "ids: [ID!]!")
	assert.Contains(t, result, "ControlBulkDeletePayload!")
}

func TestRenderBulkUpdatePayload(t *testing.T) {
	s := bulkSchema{
		Name:       "Control",
		PluralName: "Controls",
	}

	result := renderBulkUpdatePayload(s)

	assert.Contains(t, result, "type ControlBulkUpdatePayload")
	assert.Contains(t, result, "controls: [Control!]")
	assert.Contains(t, result, "updatedIDs: [ID!]")
	assert.Contains(t, result, "Updated control")
}

func TestRenderBulkDeletePayload(t *testing.T) {
	s := bulkSchema{
		Name:       "Control",
		PluralName: "Controls",
	}

	result := renderBulkDeletePayload(s)

	assert.Contains(t, result, "type ControlBulkDeletePayload")
	assert.Contains(t, result, "deletedIDs: [ID!]!")
	assert.Contains(t, result, "Deleted control")
}

func TestRenderBulkTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		schema      bulkSchema
		expected    string
		shouldPanic bool
	}{
		{
			name:     "renders simple template",
			template: "{{ .Name }}",
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expected: "Control",
		},
		{
			name:     "renders template with ToLowerCamel",
			template: "{{ .Name | ToLowerCamel }}",
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expected: "control",
		},
		{
			name:     "renders plural name",
			template: "{{ .PluralName | ToLowerCamel }}",
			schema: bulkSchema{
				Name:       "Control",
				PluralName: "Controls",
			},
			expected: "controls",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBulkTemplate(tt.template, tt.schema)
			assert.Contains(t, result, tt.expected)
		})
	}
}

func TestBulkSchemaIntegration(t *testing.T) {
	// Integration test that verifies the full workflow
	tmpDir := t.TempDir()
	schemaDir := filepath.Join(tmpDir, "schema")
	err := os.Mkdir(schemaDir, 0755)
	assert.NoError(t, err)

	// Create a schema file with createBulkCSV and one existing bulk mutation (deleteBulk)
	// This should trigger injection of the missing bulk mutations (updateBulk, updateBulkCSV)
	existingContent := `"""
Control operations
"""
extend type Query {
    """
    Look up control by ID
    """
    control(id: ID!): Control!
}

extend type Mutation {
    """
    Create a single control
    """
    createControl(input: CreateControlInput!): ControlCreatePayload!
    """
    Create multiple new controls via csv upload
    """
    createBulkCSVControl(
        """
        csv file containing values of the control
        """
        input: Upload!
    ): ControlBulkCreatePayload!
    """
    Update an existing control
    """
    updateControl(id: ID!, input: UpdateControlInput!): ControlUpdatePayload!
    """
    Delete an existing control
    """
    deleteControl(id: ID!): ControlDeletePayload!
    """
    Delete multiple controls
    """
    deleteBulkControl(ids: [ID!]!): ControlBulkDeletePayload!
}

"""
Return response for createControl mutation
"""
type ControlCreatePayload {
    control: Control!
}

"""
Return response for createBulkCSVControl mutation
"""
type ControlBulkCreatePayload {
    controls: [Control!]
}
`

	graph := &gen.Graph{
		Nodes: []*gen.Type{
			{Name: "Control"},
		},
	}

	fileName := getFileName(schemaDir, "Control")
	err = os.WriteFile(fileName, []byte(existingContent), 0600)
	assert.NoError(t, err)

	// Run the hook
	hook := GenBulkSchema(schemaDir)
	generator := hook(mockGenerator{})
	err = generator.Generate(graph)
	assert.NoError(t, err)

	// Read the result
	content, err := os.ReadFile(fileName)
	assert.NoError(t, err)

	contentStr := string(content)

	// Verify all bulk mutations were added
	assert.Contains(t, contentStr, "updateBulkControl(")
	assert.Contains(t, contentStr, "updateBulkCSVControl(")
	assert.Contains(t, contentStr, "deleteBulkControl(")

	// Verify payload types were added
	assert.Contains(t, contentStr, "type ControlBulkUpdatePayload")
	assert.Contains(t, contentStr, "type ControlBulkDeletePayload")

	// Verify original content is preserved
	assert.Contains(t, contentStr, "createControl(")
	assert.Contains(t, contentStr, "createBulkCSVControl(")
	assert.Contains(t, contentStr, "updateControl(")
	assert.Contains(t, contentStr, "deleteControl(")

	// Verify no duplication
	assert.Equal(t, 1, strings.Count(contentStr, "updateBulkControl("))
	assert.Equal(t, 1, strings.Count(contentStr, "updateBulkCSVControl("))
	assert.Equal(t, 1, strings.Count(contentStr, "deleteBulkControl("))
	assert.Equal(t, 1, strings.Count(contentStr, "type ControlBulkUpdatePayload"))
	assert.Equal(t, 1, strings.Count(contentStr, "type ControlBulkDeletePayload"))
}
