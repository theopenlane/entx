package genhooks

import (
	"testing"

	"entgo.io/contrib/entgql"
	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
	"gotest.tools/assert"

	"github.com/theopenlane/entx"
)

func TestIsFieldTypeExcluded(t *testing.T) {
	testCases := []struct {
		name     string
		input    *load.Field
		expected bool
	}{
		{
			name: "bool, excluded",
			input: &load.Field{
				Info: &field.TypeInfo{
					Type: field.TypeBool,
				},
			},
			expected: true,
		},
		{
			name: "string, included",
			input: &load.Field{
				Info: &field.TypeInfo{
					Type: field.TypeString,
				},
			},
			expected: false,
		},
		{
			name: "time, excluded",
			input: &load.Field{
				Info: &field.TypeInfo{
					Type: field.TypeTime,
				},
			},
			expected: true,
		},
		{
			name: "enum, excluded",
			input: &load.Field{
				Info: &field.TypeInfo{
					Type: field.TypeEnum,
				},
			},
			expected: true,
		},
		{
			name: "json, excluded",
			input: &load.Field{
				Info: &field.TypeInfo{
					Type: field.TypeJSON,
				},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isFieldTypeExcluded(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestEntSkip(t *testing.T) {
	entAnt := &entgql.Annotation{}

	testCases := []struct {
		name     string
		input    *load.Field
		expected bool
	}{
		{
			name:     "no entql annotation, not skipped",
			input:    &load.Field{},
			expected: false,
		},
		{
			name: "skip update, not skipped",
			input: &load.Field{
				Annotations: map[string]interface{}{
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipMutationUpdateInput,
					},
				},
			},
			expected: false,
		},
		{
			name: "skip where input, skipped",
			input: &load.Field{
				Annotations: map[string]interface{}{
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipWhereInput,
					},
				},
			},
			expected: true,
		},
		{
			name: "skip all, skipped",
			input: &load.Field{
				Annotations: map[string]interface{}{
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipAll,
					},
				},
			},
			expected: true,
		},
		{
			name: "skip type, skipped",
			input: &load.Field{
				Annotations: map[string]interface{}{
					entAnt.Name(): &entgql.Annotation{
						Skip: entgql.SkipType,
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := entSkip(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}
func TestGetSearchFileName(t *testing.T) {
	testCases := []struct {
		name     string
		isAdmin  bool
		expected string
	}{
		{
			name:     "not admin",
			isAdmin:  false,
			expected: "search",
		},
		{
			name:     "admin",
			isAdmin:  true,
			expected: "adminsearch",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getSearchFileName(tc.isAdmin)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIncludeSchemaForSearch(t *testing.T) {
	schemaGenAnt := &entx.SchemaGenAnnotation{}

	testCases := []struct {
		name     string
		input    *gen.Type
		expected bool
	}{
		{
			name: "do not skip search, has annotation",
			input: &gen.Type{
				Annotations: map[string]interface{}{
					schemaGenAnt.Name(): &entx.SchemaGenAnnotation{
						SkipSearch: false,
					},
				},
			},
			expected: true,
		},
		{
			name: "skip search, has annotation",
			input: &gen.Type{
				Annotations: map[string]interface{}{
					schemaGenAnt.Name(): &entx.SchemaGenAnnotation{
						SkipSearch: true,
					},
				},
			},
			expected: false,
		},
		{
			name: "skip not set on annotation",
			input: &gen.Type{
				Annotations: map[string]interface{}{
					schemaGenAnt.Name(): &entx.SchemaGenAnnotation{},
				},
			},
			expected: true,
		},
		{
			name: "no annotation",
			input: &gen.Type{
				Annotations: map[string]interface{}{},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := includeSchemaForSearch(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsFieldSearchable(t *testing.T) {
	searchAnt := &entx.SearchFieldAnnotation{}

	testCases := []struct {
		name     string
		input    *load.Field
		expected bool
	}{
		{
			name: "searchable, has annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{
						Searchable: true,
					},
				},
			},
			expected: true,
		},
		{
			name: "searchable false, has annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{
						Searchable: false,
					},
				},
			},
			expected: false,
		},
		{
			name: "no searchable annotation, default false",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{},
				},
			},
			expected: false,
		},
		{
			name: "no annotation, default false",
			input: &load.Field{
				Annotations: map[string]interface{}{},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isFieldSearchable(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsAdminFieldSearchable(t *testing.T) {
	searchAnt := &entx.SearchFieldAnnotation{}

	testCases := []struct {
		name     string
		input    *load.Field
		expected bool
	}{
		{
			name: "not excluded, has annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{
						Searchable:   true,
						ExcludeAdmin: false,
					},
				},
			},
			expected: true,
		},
		{
			name: "excluded, has annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{
						ExcludeAdmin: true,
					},
				},
			},
			expected: false,
		},
		{
			name: "no excluded annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{
					searchAnt.Name(): &entx.SearchFieldAnnotation{},
				},
			},
			expected: true,
		},
		{
			name: "no annotation",
			input: &load.Field{
				Annotations: map[string]interface{}{},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isAdminFieldSearchable(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}
