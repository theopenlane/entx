package entx

import (
	"encoding/json"
	"testing"

	"entgo.io/ent/entc/gen"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCascadeAnnotation(t *testing.T) {
	f := gofakeit.Name()
	ca := CascadeAnnotationField(f)

	assert.Equal(t, ca.Name(), CascadeAnnotationName)
	assert.Equal(t, ca.Field, f)
}

func TestCascadeThroughAnnotation(t *testing.T) {
	f := gofakeit.Name()
	s := gofakeit.Name()
	schemas := []ThroughCleanup{
		{
			Through: s,
			Field:   f,
		},
	}
	ca := CascadeThroughAnnotationField(schemas)

	assert.Equal(t, ca.Name(), CascadeThroughAnnotationName)
	assert.Equal(t, ca.Schemas[0].Field, f)
	assert.Equal(t, ca.Schemas[0].Through, s)
}

func TestSchemaGenAnnotation(t *testing.T) {
	s := gofakeit.Bool()
	sa := SchemaGenSkip(s)

	assert.Equal(t, sa.Name(), SchemaGenAnnotationName)
	assert.Equal(t, sa.Skip, s)
}

func TestExportableAnnotation(t *testing.T) {
	ea := &Exportable{}

	assert.Equal(t, ea.Name(), "Exportable")

	err := ea.Decode(map[string]any{})
	assert.NoError(t, err)
}

func TestWorkflowEligibleAnnotation(t *testing.T) {
	wea := FieldWorkflowEligible()

	assert.Equal(t, wea.Name(), WorkflowEligibleAnnotationName)
	assert.True(t, wea.Eligible)

	// Test Decode method
	decoded := &WorkflowEligibleAnnotation{}
	err := decoded.Decode(map[string]any{"Eligible": true})
	assert.NoError(t, err)
	assert.True(t, decoded.Eligible)
}

func TestCSVRefBuilder(t *testing.T) {
	tests := []struct {
		name               string
		builder            *CSVRefBuilder
		expectedMatchField string
		expectedColumn     string
		expectedTarget     string
		expectedCreate     bool
	}{
		{
			name: "basic user email lookup",
			builder: CSVRef().
				FromColumn("AssignedToUserEmail").
				MatchOn("email"),
			expectedMatchField: "email",
			expectedColumn:     "AssignedToUserEmail",
		},
		{
			name: "group name lookup",
			builder: CSVRef().
				FromColumn("BlockedGroupNames").
				MatchOn("name"),
			expectedMatchField: "name",
			expectedColumn:     "BlockedGroupNames",
		},
		{
			name: "platform with create if missing",
			builder: CSVRef().
				FromColumn("AccessPlatformNames").
				MatchOn("name").
				CreateIfMissing(),
			expectedMatchField: "name",
			expectedColumn:     "AccessPlatformNames",
			expectedCreate:     true,
		},
		{
			name: "entity name lookup",
			builder: CSVRef().
				FromColumn("EntityName").
				MatchOn("name"),
			expectedMatchField: "name",
			expectedColumn:     "EntityName",
		},
		{
			name: "control ref code with explicit target",
			builder: CSVRef().
				FromColumn("ControlRefCode").
				MatchOn("ref_code").
				TargetEntity("Control"),
			expectedMatchField: "ref_code",
			expectedColumn:     "ControlRefCode",
			expectedTarget:     "Control",
		},
		{
			name: "identity holder email lookup",
			builder: CSVRef().
				FromColumn("IdentityHolderEmail").
				MatchOn("email"),
			expectedMatchField: "email",
			expectedColumn:     "IdentityHolderEmail",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, CSVReferenceAnnotationName, tc.builder.Name())
			assert.Equal(t, tc.expectedMatchField, tc.builder.annotation.MatchField)
			assert.Equal(t, tc.expectedColumn, tc.builder.annotation.CSVColumn)
			assert.Equal(t, tc.expectedTarget, tc.builder.annotation.TargetEntity)
			assert.Equal(t, tc.expectedCreate, tc.builder.annotation.CreateIfMissing)
		})
	}
}

func TestCSVRefBuilderMarshalJSON(t *testing.T) {
	builder := CSVRef().
		FromColumn("BlockedGroupNames").
		MatchOn("name").
		TargetEntity("Group").
		CreateIfMissing()

	data, err := json.Marshal(builder)
	require.NoError(t, err)

	// Verify it marshals as CSVReferenceAnnotation, not as CSVRefBuilder
	var decoded CSVReferenceAnnotation

	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "name", decoded.MatchField)
	assert.Equal(t, "BlockedGroupNames", decoded.CSVColumn)
	assert.Equal(t, "Group", decoded.TargetEntity)
	assert.True(t, decoded.CreateIfMissing)
}

func TestCSVReferenceAnnotationDecode(t *testing.T) {
	decoded := &CSVReferenceAnnotation{}
	err := decoded.Decode(map[string]any{
		"MatchField":      "email",
		"CSVColumn":       "UserEmail",
		"TargetEntity":    "User",
		"CreateIfMissing": true,
	})

	assert.NoError(t, err)
	assert.Equal(t, "email", decoded.MatchField)
	assert.Equal(t, "UserEmail", decoded.CSVColumn)
	assert.Equal(t, "User", decoded.TargetEntity)
	assert.True(t, decoded.CreateIfMissing)
}

func TestGetAnnotation(t *testing.T) {
	ant := SchemaGenAnnotation{}

	tests := []struct {
		name        string
		node        *gen.Type
		expectedOk  bool
		expectedVal *SchemaGenAnnotation
	}{
		{
			name: "annotation exists as pointer",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): &SchemaGenAnnotation{Skip: true, SkipSearch: true},
				},
			},
			expectedOk:  true,
			expectedVal: &SchemaGenAnnotation{Skip: true, SkipSearch: true},
		},
		{
			name: "annotation missing",
			node: &gen.Type{
				Annotations: map[string]interface{}{},
			},
			expectedOk:  false,
			expectedVal: nil,
		},
		{
			name: "annotation stored as value, not pointer",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): SchemaGenAnnotation{Skip: true},
				},
			},
			expectedOk:  true,
			expectedVal: &SchemaGenAnnotation{Skip: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := GetAnnotation[*SchemaGenAnnotation](tt.node)
			assert.Equal(t, tt.expectedOk, ok)
			assert.Equal(t, tt.expectedVal, val)
		})
	}
}

func TestHasAnnotation(t *testing.T) {
	ant := OrgOwnedSchema{}

	tests := []struct {
		name     string
		node     *gen.Type
		expected bool
	}{
		{
			name: "has annotation",
			node: &gen.Type{
				Annotations: map[string]interface{}{
					ant.Name(): &OrgOwnedSchema{},
				},
			},
			expected: true,
		},
		{
			name: "missing annotation",
			node: &gen.Type{
				Annotations: map[string]interface{}{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, HasAnnotation[OrgOwnedSchema](tt.node))
		})
	}
}
