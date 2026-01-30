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
				Name:           "User",
				HasCreateInput: true,
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
	assert.Contains(t, contentStr, "LookupGroupByName")
	assert.Contains(t, contentStr, "CSVLookupRegistry")
}

func TestGenerateCSVHelperFileWithoutEntPackage(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "",
		Schemas: []CSVSchema{
			{
				Name:           "Simple",
				HasCreateInput: true,
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

func TestGenerateCSVHelperFileDuplicateDetection(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "github.com/example/ent/generated",
		Schemas: []CSVSchema{
			{
				Name:           "Evidence",
				HasCreateInput: true,
				Fields: []CSVReferenceField{
					{
						FieldName:    "controls",
						GoFieldName:  "ControlIDs",
						CSVColumn:    "ControlRefCodes",
						TargetEntity: "Control",
						MatchField:   "ref_code",
						IsSlice:      true,
					},
				},
			},
		},
		Lookups: []CSVLookup{
			{
				TargetEntity: "Control",
				MatchField:   "ref_code",
				OrgScoped:    true,
			},
		},
	}

	err := generateCSVHelperFile(tempDir, data)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "csv_generated.go"))
	require.NoError(t, err)

	contentStr := string(content)

	// Verify the lookup function is generated
	assert.Contains(t, contentStr, "LookupControlByRefCode")

	// Verify duplicate detection logic is present
	assert.Contains(t, contentStr, "existingID, exists := resolved[key]")
	assert.Contains(t, contentStr, "existingID != r.ID")

	// Verify error message guides user to use ID directly
	assert.Contains(t, contentStr, "matched multiple Control records")
	assert.Contains(t, contentStr, "use ControlID directly")
}

func TestGenerateCSVHelperFileWithSliceField(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "github.com/example/ent/generated",
		Schemas: []CSVSchema{
			{
				Name:           "Control",
				HasCreateInput: true,
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
		switch f.FieldName {
		case "platform_ids":
			platformField = f
		case "single_id":
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

func TestGenerateCSVFieldMappingsJSON(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		Schemas: []CSVSchema{
			{
				Name: "ActionPlan",
				Fields: []CSVReferenceField{
					{
						GoFieldName: "AssignedToUserID",
						CSVColumn:   "AssignedToUserEmail",
						IsSlice:     false,
					},
					{
						GoFieldName: "BlockedGroupIds",
						CSVColumn:   "BlockedGroupNames",
						IsSlice:     true,
					},
				},
			},
			{
				Name: "Control",
				Fields: []CSVReferenceField{
					{
						GoFieldName: "PlatformIds",
						CSVColumn:   "PlatformNames",
						IsSlice:     true,
					},
				},
			},
			{
				Name:   "Empty",
				Fields: []CSVReferenceField{},
			},
		},
	}

	err := generateCSVFieldMappingsJSON(tempDir, data)
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tempDir, "csv_field_mappings.json"))
	require.NoError(t, err)

	contentStr := string(content)

	assert.Contains(t, contentStr, `"ActionPlan"`)
	assert.Contains(t, contentStr, `"csvColumn": "AssignedToUserEmail"`)
	assert.Contains(t, contentStr, `"targetField": "AssignedToUserID"`)
	assert.Contains(t, contentStr, `"isSlice": false`)
	assert.Contains(t, contentStr, `"Control"`)
	assert.Contains(t, contentStr, `"csvColumn": "PlatformNames"`)
	assert.Contains(t, contentStr, `"isSlice": true`)
	assert.NotContains(t, contentStr, `"Empty"`)
}

func TestGenerateCSVHelperFileCreatesJSONFile(t *testing.T) {
	tempDir := t.TempDir()

	data := CSVSchemaData{
		PackageName: "testpkg",
		EntPackage:  "github.com/example/ent/generated",
		Schemas: []CSVSchema{
			{
				Name:           "User",
				HasCreateInput: true,
				Fields: []CSVReferenceField{
					{
						FieldName:    "group_id",
						GoFieldName:  "GroupID",
						CSVColumn:    "GroupName",
						TargetEntity: "Group",
						MatchField:   "name",
					},
				},
			},
		},
		Lookups: []CSVLookup{
			{
				TargetEntity: "Group",
				MatchField:   "name",
			},
		},
	}

	err := generateCSVHelperFile(tempDir, data)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tempDir, "csv_generated.go"))
	assert.FileExists(t, filepath.Join(tempDir, "csv_field_mappings.json"))

	jsonContent, err := os.ReadFile(filepath.Join(tempDir, "csv_field_mappings.json"))
	require.NoError(t, err)

	assert.Contains(t, string(jsonContent), `"User"`)
	assert.Contains(t, string(jsonContent), `"csvColumn": "GroupName"`)
}

func TestGetCSVReferenceFieldsFromEdges(t *testing.T) {
	tests := []struct {
		name     string
		schema   *load.Schema
		expected []CSVReferenceField
	}{
		{
			name: "edge with CSV reference annotation",
			schema: &load.Schema{
				Name: "Evidence",
				Edges: []*load.Edge{
					{
						Name:  "controls",
						Type:  "Control",
						Field: "",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "ref_code",
								"CSVColumn":  "ControlRefCodes",
							},
						},
					},
				},
			},
			expected: []CSVReferenceField{
				{
					FieldName:       "controls",
					GoFieldName:     "ControlIDs", // "controls" → singularized "control" → "ControlIDs"
					CSVColumn:       "ControlRefCodes",
					TargetEntity:    "Control",
					MatchField:      "ref_code",
					IsSlice:         true,
					CreateIfMissing: false,
				},
			},
		},
		{
			name: "edge with backing field is skipped",
			schema: &load.Schema{
				Name: "Task",
				Edges: []*load.Edge{
					{
						Name:  "assignee",
						Type:  "User",
						Field: "assignee_id",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "email",
								"CSVColumn":  "AssigneeEmail",
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "edge without annotation is skipped",
			schema: &load.Schema{
				Name: "Control",
				Edges: []*load.Edge{
					{
						Name:        "procedures",
						Type:        "Procedure",
						Field:       "",
						Annotations: map[string]any{},
					},
				},
			},
			expected: nil,
		},
		{
			name: "edge with explicit target entity",
			schema: &load.Schema{
				Name: "Risk",
				Edges: []*load.Edge{
					{
						Name:  "mitigations",
						Type:  "Control",
						Field: "",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField":   "ref_code",
								"CSVColumn":    "MitigationRefCodes",
								"TargetEntity": "Control",
							},
						},
					},
				},
			},
			expected: []CSVReferenceField{
				{
					FieldName:       "mitigations",
					GoFieldName:     "MitigationIDs", // "mitigations" → singularized "mitigation" → "MitigationIDs"
					CSVColumn:       "MitigationRefCodes",
					TargetEntity:    "Control",
					MatchField:      "ref_code",
					IsSlice:         true,
					CreateIfMissing: false,
				},
			},
		},
		{
			name: "edge with create if missing",
			schema: &load.Schema{
				Name: "Control",
				Edges: []*load.Edge{
					{
						Name:  "tags",
						Type:  "Tag",
						Field: "",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField":      "name",
								"CSVColumn":       "TagNames",
								"CreateIfMissing": true,
							},
						},
					},
				},
			},
			expected: []CSVReferenceField{
				{
					FieldName:       "tags",
					GoFieldName:     "TagIDs", // "tags" → singularized "tag" → "TagIDs"
					CSVColumn:       "TagNames",
					TargetEntity:    "Tag",
					MatchField:      "name",
					IsSlice:         true,
					CreateIfMissing: true,
				},
			},
		},
		{
			name: "edge with missing match field is skipped",
			schema: &load.Schema{
				Name: "Test",
				Edges: []*load.Edge{
					{
						Name:  "items",
						Type:  "Item",
						Field: "",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"CSVColumn": "ItemNames",
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "edge with missing CSV column is skipped",
			schema: &load.Schema{
				Name: "Test",
				Edges: []*load.Edge{
					{
						Name:  "items",
						Type:  "Item",
						Field: "",
						Annotations: map[string]any{
							entx.CSVReferenceAnnotationName: map[string]any{
								"MatchField": "name",
							},
						},
					},
				},
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fields := getCSVReferenceFieldsFromEdges(tc.schema)
			assert.Equal(t, tc.expected, fields)
		})
	}
}

func TestGetCSVReferenceAnnotationFromEdge(t *testing.T) {
	tests := []struct {
		name     string
		edge     *load.Edge
		expected *entx.CSVReferenceAnnotation
	}{
		{
			name: "edge with annotation",
			edge: &load.Edge{
				Name: "controls",
				Type: "Control",
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField":      "ref_code",
						"CSVColumn":       "ControlRefCodes",
						"TargetEntity":    "Control",
						"CreateIfMissing": true,
					},
				},
			},
			expected: &entx.CSVReferenceAnnotation{
				MatchField:      "ref_code",
				CSVColumn:       "ControlRefCodes",
				TargetEntity:    "Control",
				CreateIfMissing: true,
			},
		},
		{
			name: "edge without annotation",
			edge: &load.Edge{
				Name:        "procedures",
				Type:        "Procedure",
				Annotations: map[string]any{},
			},
			expected: nil,
		},
		{
			name: "edge with nil annotations",
			edge: &load.Edge{
				Name: "items",
				Type: "Item",
			},
			expected: nil,
		},
		{
			name: "edge with invalid annotation value",
			edge: &load.Edge{
				Name: "bad_edge",
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: "invalid-not-a-map",
				},
			},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ann := getCSVReferenceAnnotationFromEdge(tc.edge)
			if tc.expected == nil {
				assert.Nil(t, ann)
			} else {
				assert.NotNil(t, ann)
				assert.Equal(t, tc.expected.MatchField, ann.MatchField)
				assert.Equal(t, tc.expected.CSVColumn, ann.CSVColumn)
				assert.Equal(t, tc.expected.TargetEntity, ann.TargetEntity)
				assert.Equal(t, tc.expected.CreateIfMissing, ann.CreateIfMissing)
			}
		})
	}
}

func TestGetCSVReferenceFieldsWithEdgesIncludesEdgeAnnotations(t *testing.T) {
	schema := &load.Schema{
		Name: "Evidence",
		Fields: []*load.Field{
			{
				Name: "assignee_id",
				Info: &field.TypeInfo{Type: field.TypeString},
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField": "email",
						"CSVColumn":  "AssigneeEmail",
					},
				},
			},
		},
		Edges: []*load.Edge{
			{
				Name:  "controls",
				Type:  "Control",
				Field: "",
				Annotations: map[string]any{
					entx.CSVReferenceAnnotationName: map[string]any{
						"MatchField": "ref_code",
						"CSVColumn":  "ControlRefCodes",
					},
				},
			},
		},
	}

	edgeTargets := map[string]string{
		"assignee_id": "User",
	}

	fields := getCSVReferenceFieldsWithEdges(schema, edgeTargets)
	assert.Len(t, fields, 2)

	var assigneeField, controlsField CSVReferenceField

	for _, f := range fields {
		switch f.FieldName {
		case "assignee_id":
			assigneeField = f
		case "controls":
			controlsField = f
		}
	}

	assert.Equal(t, "AssigneeID", assigneeField.GoFieldName)
	assert.Equal(t, "AssigneeEmail", assigneeField.CSVColumn)
	assert.Equal(t, "User", assigneeField.TargetEntity)
	assert.False(t, assigneeField.IsSlice)

	assert.Equal(t, "ControlIDs", controlsField.GoFieldName)
	assert.Equal(t, "ControlRefCodes", controlsField.CSVColumn)
	assert.Equal(t, "Control", controlsField.TargetEntity)
	assert.True(t, controlsField.IsSlice)
}
