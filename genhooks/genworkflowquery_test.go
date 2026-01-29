package genhooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func workflowSchemaNodes() []*gen.Type {
	return []*gen.Type{
		{
			Name: "WorkflowInstance",
			Fields: []*gen.Field{
				{Name: "workflow_definition_id"},
				{Name: "state"},
				{Name: "context"},
			},
		},
		{
			Name: "WorkflowEvent",
			Fields: []*gen.Field{
				{Name: "workflow_instance_id"},
				{Name: "event_type"},
				{Name: "payload"},
			},
		},
	}
}

func TestGenWorkflowQuery(t *testing.T) {
	tests := []struct {
		name             string
		setupSchema      func() *gen.Graph
		expectedContains []string
		expectedMissing  []string
		expectFile       bool
	}{
		{
			name: "generates workflow queries for workflow-eligible types",
			setupSchema: func() *gen.Graph {
				nodes := workflowSchemaNodes()
				nodes = append(nodes,
					&gen.Type{
						Name: "Campaign",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
					&gen.Type{
						Name: "Control",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
				)

				return &gen.Graph{Nodes: nodes}
			},
			expectedContains: []string{
				"query GetCampaignWorkflowStatus($campaignId: ID!)",
				"query GetCampaignWorkflowTimeline($campaignId: ID!",
				"query GetControlWorkflowStatus($controlId: ID!)",
				"query GetControlWorkflowTimeline($controlId: ID!",
				"hasPendingWorkflow",
				"hasWorkflowHistory",
				"activeWorkflowInstances",
				"workflowTimeline(",
			},
			expectFile: true,
		},
		{
			name: "skips non-workflow-eligible types",
			setupSchema: func() *gen.Graph {
				nodes := workflowSchemaNodes()
				nodes = append(nodes,
					&gen.Type{
						Name: "WorkflowEntity",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
					&gen.Type{
						Name: "RegularEntity",
						Fields: []*gen.Field{
							{Name: "name"},
						},
					},
				)

				return &gen.Graph{Nodes: nodes}
			},
			expectedContains: []string{
				"query GetWorkflowEntityWorkflowStatus",
			},
			expectedMissing: []string{
				"query GetRegularEntityWorkflowStatus",
			},
			expectFile: true,
		},
		{
			name: "skips history types",
			setupSchema: func() *gen.Graph {
				nodes := workflowSchemaNodes()
				nodes = append(nodes,
					&gen.Type{
						Name: "Campaign",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
					&gen.Type{
						Name: "CampaignHistory",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
						Annotations: gen.Annotations{
							"History": map[string]any{
								"isHistory": true,
							},
						},
					},
				)

				return &gen.Graph{Nodes: nodes}
			},
			expectedContains: []string{
				"query GetCampaignWorkflowStatus",
			},
			expectedMissing: []string{
				"query GetCampaignHistoryWorkflowStatus",
			},
			expectFile: true,
		},
		{
			name: "does not create file when no workflow types exist",
			setupSchema: func() *gen.Graph {
				nodes := workflowSchemaNodes()
				nodes = append(nodes,
					&gen.Type{
						Name: "RegularEntity",
						Fields: []*gen.Field{
							{Name: "name"},
						},
					},
				)

				return &gen.Graph{Nodes: nodes}
			},
			expectFile: false,
		},
		{
			name: "does not create file when WorkflowInstance is missing",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "WorkflowEvent",
							Fields: []*gen.Field{
								{Name: "event_type"},
							},
						},
						{
							Name: "Campaign",
							Fields: []*gen.Field{
								{Name: "workflow_eligible_marker"},
							},
						},
					},
				}
			},
			expectFile: false,
		},
		{
			name: "sorts types alphabetically",
			setupSchema: func() *gen.Graph {
				nodes := workflowSchemaNodes()
				nodes = append(nodes,
					&gen.Type{
						Name: "Zebra",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
					&gen.Type{
						Name: "Alpha",
						Fields: []*gen.Field{
							{Name: "workflow_eligible_marker"},
						},
					},
				)

				return &gen.Graph{Nodes: nodes}
			},
			expectedContains: []string{
				"query GetAlphaWorkflowStatus",
				"query GetZebraWorkflowStatus",
			},
			expectFile: true,
		},
		{
			name: "uses actual schema fields for WorkflowInstance and WorkflowEvent",
			setupSchema: func() *gen.Graph {
				return &gen.Graph{
					Nodes: []*gen.Type{
						{
							Name: "WorkflowInstance",
							Fields: []*gen.Field{
								{Name: "custom_field_one"},
								{Name: "custom_field_two"},
							},
						},
						{
							Name: "WorkflowEvent",
							Fields: []*gen.Field{
								{Name: "event_data"},
							},
						},
						{
							Name: "TestEntity",
							Fields: []*gen.Field{
								{Name: "workflow_eligible_marker"},
							},
						},
					},
				}
			},
			expectedContains: []string{
				"customFieldOne",
				"customFieldTwo",
				"eventData",
			},
			expectFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			queryDir := filepath.Join(tmpDir, "query")
			err := os.Mkdir(queryDir, 0755)
			require.NoError(t, err)

			graph := tt.setupSchema()

			hook := GenWorkflowQuery(queryDir)
			generator := hook(mockGenerator{})
			err = generator.Generate(graph)
			require.NoError(t, err)

			filePath := filepath.Join(queryDir, "workflow.graphql")
			_, err = os.Stat(filePath)

			if !tt.expectFile {
				assert.True(t, os.IsNotExist(err), "workflow.graphql should not be created")
				return
			}

			require.NoError(t, err, "workflow.graphql should exist")

			content, err := os.ReadFile(filePath)
			require.NoError(t, err)

			contentStr := string(content)

			for _, expected := range tt.expectedContains {
				assert.Contains(t, contentStr, expected, "expected content should be present: %s", expected)
			}

			for _, missing := range tt.expectedMissing {
				assert.NotContains(t, contentStr, missing, "unexpected content should not be present: %s", missing)
			}
		})
	}
}

func TestHasWorkflowSupport(t *testing.T) {
	tests := []struct {
		name     string
		node     *gen.Type
		expected bool
	}{
		{
			name: "has workflow_eligible_marker",
			node: &gen.Type{
				Name: "TestEntity",
				Fields: []*gen.Field{
					{Name: "workflow_eligible_marker"},
					{Name: "name"},
				},
			},
			expected: true,
		},
		{
			name: "does not have workflow_eligible_marker",
			node: &gen.Type{
				Name: "TestEntity",
				Fields: []*gen.Field{
					{Name: "name"},
					{Name: "description"},
				},
			},
			expected: false,
		},
		{
			name: "empty fields",
			node: &gen.Type{
				Name:   "TestEntity",
				Fields: []*gen.Field{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasWorkflowSupport(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHistoryType(t *testing.T) {
	tests := []struct {
		name     string
		node     *gen.Type
		expected bool
	}{
		{
			name: "is history type",
			node: &gen.Type{
				Name: "EntityHistory",
				Annotations: gen.Annotations{
					"History": map[string]any{
						"isHistory": true,
					},
				},
			},
			expected: true,
		},
		{
			name: "is not history type",
			node: &gen.Type{
				Name: "Entity",
				Annotations: gen.Annotations{
					"History": map[string]any{
						"isHistory": false,
					},
				},
			},
			expected: false,
		},
		{
			name: "no history annotation",
			node: &gen.Type{
				Name:        "Entity",
				Annotations: gen.Annotations{},
			},
			expected: false,
		},
		{
			name: "nil annotations",
			node: &gen.Type{
				Name:        "Entity",
				Annotations: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHistoryType(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateWorkflowQueryTemplate(t *testing.T) {
	tmpl := createWorkflowQueryTemplate()
	assert.NotNil(t, tmpl)

	var buf strings.Builder

	data := workflowQueryData{
		Types: []workflowQueryType{
			{Name: "TestEntity"},
		},
		WorkflowInstanceFields: []string{"id", "state", "context"},
		WorkflowEventFields:    []string{"id", "eventType", "payload"},
	}

	err := tmpl.Execute(&buf, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "query GetTestEntityWorkflowStatus($testEntityId: ID!)")
	assert.Contains(t, output, "query GetTestEntityWorkflowTimeline($testEntityId: ID!")
	assert.Contains(t, output, "testEntity(id: $testEntityId)")
	assert.Contains(t, output, "hasPendingWorkflow")
	assert.Contains(t, output, "hasWorkflowHistory")
	assert.Contains(t, output, "activeWorkflowInstances")
	assert.Contains(t, output, "workflowTimeline(")
	assert.Contains(t, output, "state")
	assert.Contains(t, output, "context")
	assert.Contains(t, output, "eventType")
	assert.Contains(t, output, "payload")
}
