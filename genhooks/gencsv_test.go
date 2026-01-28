package genhooks

import (
	"os"
	"path/filepath"
	"testing"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/entx"
)

func TestGetCSVReferenceFieldsWithEdges(t *testing.T) {
	tests := []struct {
		name        string
		schema      *load.Schema
		edgeTargets map[string]string
		expected    []CSVReferenceField
	}{
		{
			name: "schema with CSV reference fields and edge targets",
			schema: &load.Schema{
				Name: "ActionPlan",
				Fields: []*load.Field{
					{
						Name: "assigned_to_user_id",
						Info: &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "email",
								"CSVColumn":  "AssignedToUserEmail",
							},
						},
					},
					{
						Name: "blocked_group_ids",
						Info: &field.TypeInfo{Type: field.TypeJSON},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "name",
								"CSVColumn":  "BlockedGroupNames",
							},
						},
					},
					{
						Name:        "name",
						Info:        &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{},
					},
				},
			},
			edgeTargets: map[string]string{
				"assigned_to_user_id": "User",
				"blocked_group_ids":   "Group",
			},
			expected: []CSVReferenceField{
				{
					FieldName:    "assigned_to_user_id",
					GoFieldName:  "AssignedToUserID",
					CSVColumn:    "AssignedToUserEmail",
					TargetEntity: "User",
					MatchField:   "email",
					IsSlice:      false,
				},
				{
					FieldName:    "blocked_group_ids",
					GoFieldName:  "BlockedGroupIds",
					CSVColumn:    "BlockedGroupNames",
					TargetEntity: "Group",
					MatchField:   "name",
					IsSlice:      false,
				},
			},
		},
		{
			name: "schema with explicit target entity",
			schema: &load.Schema{
				Name: "Task",
				Fields: []*load.Field{
					{
						Name: "control_id",
						Info: &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField":   "ref_code",
								"CSVColumn":    "ControlRefCode",
								"TargetEntity": "Control",
							},
						},
					},
				},
			},
			edgeTargets: map[string]string{},
			expected: []CSVReferenceField{
				{
					FieldName:    "control_id",
					GoFieldName:  "ControlID",
					CSVColumn:    "ControlRefCode",
					TargetEntity: "Control",
					MatchField:   "ref_code",
					IsSlice:      false,
				},
			},
		},
		{
			name: "schema without CSV reference fields",
			schema: &load.Schema{
				Name: "User",
				Fields: []*load.Field{
					{
						Name:        "email",
						Info:        &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{},
					},
				},
			},
			edgeTargets: map[string]string{},
			expected:    nil,
		},
		{
			name: "schema with create if missing",
			schema: &load.Schema{
				Name: "Control",
				Fields: []*load.Field{
					{
						Name: "platform_ids",
						Info: &field.TypeInfo{Type: field.TypeJSON},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField":      "name",
								"CSVColumn":       "PlatformNames",
								"CreateIfMissing": true,
							},
						},
					},
				},
			},
			edgeTargets: map[string]string{
				"platform_ids": "Platform",
			},
			expected: []CSVReferenceField{
				{
					FieldName:       "platform_ids",
					GoFieldName:     "PlatformIds",
					CSVColumn:       "PlatformNames",
					TargetEntity:    "Platform",
					MatchField:      "name",
					IsSlice:         false,
					CreateIfMissing: true,
				},
			},
		},
		{
			name: "field with missing match field is skipped",
			schema: &load.Schema{
				Name: "Test",
				Fields: []*load.Field{
					{
						Name: "test_field",
						Info: &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"CSVColumn": "TestColumn",
							},
						},
					},
				},
			},
			edgeTargets: map[string]string{"test_field": "SomeEntity"},
			expected:    nil,
		},
		{
			name: "field with missing CSV column is skipped",
			schema: &load.Schema{
				Name: "Test",
				Fields: []*load.Field{
					{
						Name: "test_field",
						Info: &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "email",
							},
						},
					},
				},
			},
			edgeTargets: map[string]string{"test_field": "SomeEntity"},
			expected:    nil,
		},
		{
			name: "field with no target entity and no edge is skipped",
			schema: &load.Schema{
				Name: "Test",
				Fields: []*load.Field{
					{
						Name: "orphan_field",
						Info: &field.TypeInfo{Type: field.TypeString},
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "email",
								"CSVColumn":  "OrphanEmail",
							},
						},
					},
				},
			},
			edgeTargets: map[string]string{},
			expected:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fields := getCSVReferenceFieldsWithEdges(tc.schema, tc.edgeTargets)
			assert.Equal(t, tc.expected, fields)
		})
	}
}

func TestGetCSVReferenceAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		field    *load.Field
		expected *entx.CSVReferenceAnnotation
	}{
		{
			name: "field with annotation",
			field: &load.Field{
				Name: "user_id",
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField":      "email",
						"CSVColumn":       "UserEmail",
						"TargetEntity":    "User",
						"CreateIfMissing": false,
					},
				},
			},
			expected: &entx.CSVReferenceAnnotation{
				MatchField:   "email",
				CSVColumn:    "UserEmail",
				TargetEntity: "User",
			},
		},
		{
			name: "field without annotation",
			field: &load.Field{
				Name:        "name",
				Annotations: map[string]any{},
			},
			expected: nil,
		},
		{
			name: "field with nil annotations",
			field: &load.Field{
				Name: "name",
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ann := getCSVReferenceAnnotation(tc.field)
			if tc.expected == nil {
				assert.Nil(t, ann)
			} else {
				assert.NotNil(t, ann)
				assert.Equal(t, tc.expected.MatchField, ann.MatchField)
				assert.Equal(t, tc.expected.CSVColumn, ann.CSVColumn)
				assert.Equal(t, tc.expected.TargetEntity, ann.TargetEntity)
			}
		})
	}
}

func TestCSVConfigOptions(t *testing.T) {
	c := &CSVConfig{}

	WithCSVOutputDir("/tmp/output")(c)
	assert.Equal(t, "/tmp/output", c.outputDir)

	WithCSVPackageName("csvgenerated")(c)
	assert.Equal(t, "csvgenerated", c.packageName)

	WithCSVEntPackage("github.com/example/internal/ent/generated")(c)
	assert.Equal(t, "github.com/example/internal/ent/generated", c.entPackage)
}

func TestBuildEdgeTargetMap(t *testing.T) {
	t.Run("node with no edges returns empty map", func(t *testing.T) {
		node := &gen.Type{
			Name:  "Simple",
			Edges: []*gen.Edge{},
		}
		result := buildEdgeTargetMap(node)
		assert.Empty(t, result)
	})

	t.Run("node with nil edges returns empty map", func(t *testing.T) {
		node := &gen.Type{
			Name:  "Simple",
			Edges: nil,
		}
		result := buildEdgeTargetMap(node)
		assert.Empty(t, result)
	})
}

func TestCreateCSVTemplate(t *testing.T) {
	tmpl := createCSVTemplate()
	assert.NotNil(t, tmpl)
	assert.Equal(t, "csv.tpl", tmpl.Name())
}

func TestGenerateCSVHelperFile(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "github.com/example/ent/generated",
		Schemas: []CSVSchema{
			{
				Name: "User",
				Fields: []CSVReferenceField{
					{
						FieldName:       "group_id",
						GoFieldName:     "GroupID",
						CSVColumn:       "GroupName",
						TargetEntity:    "Group",
						MatchField:      "name",
						IsSlice:         false,
						CreateIfMissing: false,
					},
				},
			},
		},
		Lookups: []CSVLookup{
			{
				TargetEntity:    "Group",
				MatchField:      "name",
				CreateIfMissing: false,
			},
		},
	}

	err := generateCSVHelperFile(tempDir, data)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "csv_generated.go"))
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package testpkg")
	assert.Contains(t, contentStr, "github.com/example/ent/generated")
	assert.Contains(t, contentStr, `"User":`)
	assert.Contains(t, contentStr, "GroupName")
	assert.Contains(t, contentStr, "UserCSVInput")
	assert.Contains(t, contentStr, "CSVLookups()")
}

func TestGenerateCSVHelperFileWithoutEntPackage(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "",
		Schemas: []CSVSchema{
			{
				Name: "Simple",
				Fields: []CSVReferenceField{
					{
						FieldName:    "ref_id",
						GoFieldName:  "RefID",
						CSVColumn:    "RefName",
						TargetEntity: "Ref",
						MatchField:   "name",
					},
				},
			},
		},
		Lookups: []CSVLookup{
			{
				TargetEntity: "Ref",
				MatchField:   "name",
			},
		},
	}

	err := generateCSVHelperFile(tempDir, data)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "csv_generated.go"))
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "package testpkg")
	assert.NotContains(t, contentStr, "import (")
	assert.Contains(t, contentStr, "Input any")
}

func TestGenerateCSVHelperFileWithSliceField(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "github.com/example/ent/generated",
		Schemas: []CSVSchema{
			{
				Name: "Control",
				Fields: []CSVReferenceField{
					{
						FieldName:       "platform_ids",
						GoFieldName:     "PlatformIds",
						CSVColumn:       "PlatformNames",
						TargetEntity:    "Platform",
						MatchField:      "name",
						IsSlice:         true,
						CreateIfMissing: true,
					},
				},
			},
		},
		Lookups: []CSVLookup{
			{
				TargetEntity:    "Platform",
				MatchField:      "name",
				CreateIfMissing: true,
			},
		},
	}

	err := generateCSVHelperFile(tempDir, data)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "csv_generated.go"))
	assert.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "PlatformNames []string")
	assert.Contains(t, contentStr, "IsSlice:         true")
	assert.Contains(t, contentStr, "CreateIfMissing: true")
}

func TestGetCSVReferenceFieldsWithEdgesSliceDetection(t *testing.T) {
	schema := &load.Schema{
		Name: "Control",
		Fields: []*load.Field{
			{
				Name: "platform_ids",
				Info: &field.TypeInfo{
					Type:  field.TypeJSON,
					Ident: "[]string",
				},
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField": "name",
						"CSVColumn":  "PlatformNames",
					},
				},
			},
			{
				Name: "single_id",
				Info: &field.TypeInfo{
					Type:  field.TypeString,
					Ident: "",
				},
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField": "email",
						"CSVColumn":  "SingleEmail",
					},
				},
			},
		},
	}

	edgeTargets := map[string]string{
		"platform_ids": "Platform",
		"single_id":    "User",
	}

	fields := getCSVReferenceFieldsWithEdges(schema, edgeTargets)
	assert.Len(t, fields, 2)

	var platformField, singleField CSVReferenceField

	for _, f := range fields {
		if f.FieldName == "platform_ids" {
			platformField = f
		} else if f.FieldName == "single_id" {
			singleField = f
		}
	}

	assert.True(t, platformField.IsSlice, "platform_ids should be detected as slice")
	assert.False(t, singleField.IsSlice, "single_id should not be detected as slice")
}

func TestGetCSVReferenceAnnotationDecodeError(t *testing.T) {
	f := &load.Field{
		Name: "bad_field",
		Annotations: map[string]any{
			entx.CSVReferenceAnnotationName: "invalid-not-a-map",
		},
	}

	ann := getCSVReferenceAnnotation(f)
	assert.Nil(t, ann, "should return nil when annotation decode fails")
}

func TestCSVSchemaDataLookupDeduplication(t *testing.T) {
	data := CSVSchemaData{
		PackageName: "test",
		Schemas: []CSVSchema{
			{
				Name: "Schema1",
				Fields: []CSVReferenceField{
					{TargetEntity: "User", MatchField: "email"},
					{TargetEntity: "Group", MatchField: "name"},
				},
			},
			{
				Name: "Schema2",
				Fields: []CSVReferenceField{
					{TargetEntity: "User", MatchField: "email"},
					{TargetEntity: "Platform", MatchField: "name"},
				},
			},
		},
	}

	lookupSet := make(map[string]CSVLookup)

	for _, schema := range data.Schemas {
		for _, f := range schema.Fields {
			key := f.TargetEntity + ":" + f.MatchField
			if _, exists := lookupSet[key]; !exists {
				lookupSet[key] = CSVLookup{
					TargetEntity: f.TargetEntity,
					MatchField:   f.MatchField,
				}
			}
		}
	}

	assert.Len(t, lookupSet, 3)
	assert.Contains(t, lookupSet, "User:email")
	assert.Contains(t, lookupSet, "Group:name")
	assert.Contains(t, lookupSet, "Platform:name")
}

func TestGenerateCSVHelperFileInvalidPath(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file where we expect a directory - MkdirAll will fail
	blockerFile := filepath.Join(tempDir, "blocker")
	err := os.WriteFile(blockerFile, []byte("x"), 0600)
	require.NoError(t, err)

	data := CSVSchemaData{
		PackageName: "test",
		Schemas:     []CSVSchema{},
	}

	// Attempt to create output inside the file (impossible)
	err = generateCSVHelperFile(filepath.Join(blockerFile, "subdir"), data)
	assert.Error(t, err)
}
