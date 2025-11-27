package genhooks

import (
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/load"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/entx"
)

func TestGetWorkflowEligibleAnnotation(t *testing.T) {
	workflowAnt := &entx.WorkflowEligibleAnnotation{}

	testCases := []struct {
		name     string
		input    *load.Field
		expected *entx.WorkflowEligibleAnnotation
	}{
		{
			name: "has workflow eligible annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
				},
			},
			expected: &entx.WorkflowEligibleAnnotation{
				Eligible: true,
			},
		},
		{
			name: "has workflow annotation set to false",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: false,
					},
				},
			},
			expected: &entx.WorkflowEligibleAnnotation{
				Eligible: false,
			},
		},
		{
			name: "no annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{},
			},
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getWorkflowEligibleAnnotation(tc.input)

			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tc.expected.Eligible, result.Eligible)
			}
		})
	}
}

func TestIsFieldWorkflowEligible(t *testing.T) {
	workflowAnt := &entx.WorkflowEligibleAnnotation{}
	entAnt := &entgql.Annotation{}

	testCases := []struct {
		name     string
		input    *load.Field
		expected bool
	}{
		{
			name: "eligible field",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "not eligible, annotation set to false",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: false,
					},
				},
			},
			expected: false,
		},
		{
			name: "no annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{},
			},
			expected: false,
		},
		{
			name: "sensitive field excluded",
			input: &load.Field{
				Sensitive: true,
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
				},
			},
			expected: false,
		},
		{
			name: "immutable field excluded",
			input: &load.Field{
				Immutable: true,
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
				},
			},
			expected: false,
		},
		{
			name: "entgql skip type excluded",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipType,
					},
				},
			},
			expected: false,
		},
		{
			name: "entgql skip where input excluded",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipWhereInput,
					},
				},
			},
			expected: false,
		},
		{
			name: "entgql skip mutation update allowed",
			input: &load.Field{
				Annotations: map[string]interface{}{
					workflowAnt.Name(): &entx.WorkflowEligibleAnnotation{
						Eligible: true,
					},
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipMutationUpdateInput,
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isFieldWorkflowEligible(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}
